// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package role

//go:generate mockgen -destination mock_role.go -package role github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/role RoleBiz

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// RoleBiz 定义处理角色相关请求所需的方法.
type RoleBiz interface {
	// 角色CRUD
	Create(ctx context.Context, rq *apiv1.CreateRoleRequest) (*apiv1.CreateRoleResponse, error)
	Update(ctx context.Context, rq *apiv1.UpdateRoleRequest) (*apiv1.UpdateRoleResponse, error)
	Delete(ctx context.Context, rq *apiv1.DeleteRoleRequest) (*apiv1.DeleteRoleResponse, error)
	List(ctx context.Context, rq *apiv1.ListRolesRequest) (*apiv1.ListRolesResponse, error)
}

// roleBiz 是 RoleBiz 接口的实现.
type roleBiz struct {
	store       store.IStore
	authz       *authz.Authz
	idConverter *authz.IDConverter
}

// 确保 roleBiz 实现了 RoleBiz 接口.
var _ RoleBiz = (*roleBiz)(nil)

// New 创建一个新的 RoleBiz 实例.
func New(store store.IStore, authorizer *authz.Authz) *roleBiz {
	return &roleBiz{
		store:       store,
		authz:       authorizer,
		idConverter: authz.NewIDConverter(),
	}
}

// Create 创建角色
func (b *roleBiz) Create(ctx context.Context, rq *apiv1.CreateRoleRequest) (*apiv1.CreateRoleResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 参数验证
	if rq.RoleCode == "" {
		return nil, errno.ErrInvalidArgument.WithMessage("role_code cannot be empty")
	}
	if rq.Name == "" {
		return nil, errno.ErrInvalidArgument.WithMessage("name cannot be empty")
	}

	// 角色编码格式校验
	if !isValidRoleCode(rq.RoleCode) {
		return nil, errno.ErrInvalidArgument.WithMessage("role_code format is invalid. Must be 3-50 characters, containing only letters, numbers, underscores, and hyphens")
	}

	// 角色名称格式校验
	if !isValidRoleName(rq.Name) {
		return nil, errno.ErrInvalidArgument.WithMessage("role_name format is invalid. Must be 2-100 characters")
	}

	// 获取当前租户ID
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	tenantIDInt, err := strconv.ParseInt(tenantID, 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid tenant_id format")
	}

	// 使用数据库事务确保操作的原子性
	var roleM *model.RoleM
	err = b.store.TX(ctx, func(ctx context.Context) error {
		// 检查角色编码是否已存在（在当前租户内）
		exists, err := b.store.Role().IsRoleCodeExists(ctx, rq.RoleCode, tenantIDInt)
		if err != nil {
			return errno.ErrDBRead.WithMessage("Failed to check role code existence")
		}
		if exists {
			return errno.ErrRoleAlreadyExists.WithMessage("role code already exists in current tenant")
		}

		// 创建角色
		roleM = &model.RoleM{
			TenantID:    tenantIDInt,
			RoleCode:    rq.RoleCode,
			Name:        rq.Name,
			Description: &rq.Description,
			Status:      true, // 默认启用
		}

		if err := b.store.Role().Create(ctx, roleM); err != nil {
			log.W(ctx).Errorw("Failed to create role", "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to create role")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.W(ctx).Infow("Role created successfully", "role_id", roleM.ID, "role_code", roleM.RoleCode, "name", roleM.Name, "tenant_id", roleM.TenantID)

	// 转换为响应格式
	return &apiv1.CreateRoleResponse{
		Role: convertRoleToAPI(roleM),
	}, nil
}

// Update 更新角色
func (b *roleBiz) Update(ctx context.Context, rq *apiv1.UpdateRoleRequest) (*apiv1.UpdateRoleResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 参数验证
	if rq.RoleId <= 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("role_id must be greater than 0")
	}

	// 角色名称格式校验
	if rq.Name != "" && !isValidRoleName(rq.Name) {
		return nil, errno.ErrInvalidArgument.WithMessage("role_name format is invalid. Must be 2-100 characters")
	}

	// 获取现有角色
	roleM, err := b.store.Role().Get(ctx, where.F("id", rq.RoleId))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errno.ErrRoleNotFound.WithMessage("role not found")
		}
		log.W(ctx).Errorw("Failed to get role", "role_id", rq.RoleId, "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to get role")
	}

	// 更新字段
	if rq.Name != "" {
		roleM.Name = rq.Name
	}
	if rq.Description != "" {
		roleM.Description = &rq.Description
	}

	if err := b.store.Role().Update(ctx, roleM); err != nil {
		log.W(ctx).Errorw("Failed to update role", "role_id", rq.RoleId, "err", err)
		return nil, errno.ErrDBWrite.WithMessage("Failed to update role")
	}

	log.W(ctx).Infow("Role updated successfully", "role_id", roleM.ID, "role_code", roleM.RoleCode, "name", roleM.Name)

	// 转换为响应格式
	return &apiv1.UpdateRoleResponse{
		Role: convertRoleToAPI(roleM),
	}, nil
}

// Delete 删除角色
func (b *roleBiz) Delete(ctx context.Context, rq *apiv1.DeleteRoleRequest) (*apiv1.DeleteRoleResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 参数验证
	if rq.RoleId <= 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("role_id must be greater than 0")
	}

	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 检查角色是否可以删除
	canDelete, reason, err := b.checkDeleteRole(ctx, rq.RoleId)
	if err != nil {
		return nil, err
	}

	if !canDelete {
		return nil, errno.ErrOperationFailed.WithMessage(reason)
	}

	// 使用事务处理删除
	err = b.store.TX(ctx, func(ctx context.Context) error {
		// 删除角色
		err := b.store.Role().Delete(ctx, where.F("id", rq.RoleId))
		if err != nil {
			log.W(ctx).Errorw("Failed to delete role", "role_id", rq.RoleId, "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to delete role")
		}

		// 构建角色标识符
		roleIdentifier := fmt.Sprintf("r%d", rq.RoleId)
		tenantIdentifier := fmt.Sprintf("t%s", tenantID)

		// 删除Casbin中的相关策略
		// 删除角色的所有权限
		_, err = b.authz.DeletePermissionsForUser(roleIdentifier)
		if err != nil {
			log.W(ctx).Errorw("Failed to delete role permissions from Casbin", "role", roleIdentifier, "err", err)
		}

		// 删除所有用户的该角色
		users, _ := b.authz.GetUsersForRole(roleIdentifier, tenantIdentifier)
		for _, user := range users {
			_, err = b.authz.DeleteRoleForUser(user, roleIdentifier, tenantIdentifier)
			if err != nil {
				log.W(ctx).Errorw("Failed to delete role for user", "user", user, "role", roleIdentifier, "err", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.W(ctx).Infow("Role deleted successfully", "role_id", rq.RoleId)

	return &apiv1.DeleteRoleResponse{
		Success: true,
	}, nil
}

// List 获取角色列表
func (b *roleBiz) List(ctx context.Context, rq *apiv1.ListRolesRequest) (*apiv1.ListRolesResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 获取当前租户ID
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	tenantIDInt, err := strconv.ParseInt(tenantID, 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid tenant_id format")
	}

	opts := where.NewWhere().F("tenant_id", tenantIDInt)

	// 添加分页
	if rq.Offset > 0 {
		opts = opts.O(int(rq.Offset))
	}
	if rq.Limit > 0 {
		opts = opts.L(int(rq.Limit))
	}

	count, roles, err := b.store.Role().List(ctx, opts)
	if err != nil {
		log.W(ctx).Errorw("Failed to list roles", "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to list roles")
	}

	// 转换为响应格式
	var roleList []*apiv1.Role
	for _, role := range roles {
		roleList = append(roleList, convertRoleToAPI(role))
	}

	return &apiv1.ListRolesResponse{
		Roles:      roleList,
		TotalCount: count,
	}, nil
}

// checkDeleteRole 检查角色是否可以删除
func (b *roleBiz) checkDeleteRole(ctx context.Context, roleID int64) (bool, string, error) {
	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 构建角色标识符
	roleIdentifier := fmt.Sprintf("r%d", roleID)
	tenantIdentifier := fmt.Sprintf("t%s", tenantID)

	// 检查是否有用户使用该角色
	users, _ := b.authz.GetUsersForRole(roleIdentifier, tenantIdentifier)
	if len(users) > 0 {
		return false, fmt.Sprintf("角色正在被 %d 个用户使用，无法删除", len(users)), nil
	}

	// 检查是否是系统内置角色
	roleM, err := b.store.Role().Get(ctx, where.F("id", roleID))
	if err != nil {
		return false, "", errno.ErrDBRead.WithMessage("Failed to get role")
	}

	// 检查是否是系统角色（假设role_code以system_开头的是系统角色）
	if len(roleM.RoleCode) > 7 && roleM.RoleCode[:7] == "system_" {
		return false, "系统内置角色不能删除", nil
	}

	return true, "", nil
}

// convertRoleToAPI 将数据库模型转换为API响应模型
func convertRoleToAPI(roleM *model.RoleM) *apiv1.Role {
	description := ""
	if roleM.Description != nil {
		description = *roleM.Description
	}

	status := int32(0)
	if roleM.Status {
		status = 1
	}

	return &apiv1.Role{
		Id:          roleM.ID,
		TenantId:    roleM.TenantID,
		RoleCode:    roleM.RoleCode,
		Name:        roleM.Name,
		Description: description,
		Status:      status,
		CreatedAt:   timestamppb.New(roleM.CreatedAt),
		UpdatedAt:   timestamppb.New(roleM.UpdatedAt),
	}
}

// isValidRoleCode 验证角色编码格式
func isValidRoleCode(code string) bool {
	if len(code) < 3 || len(code) > 50 {
		return false
	}
	// 只能包含字母、数字、下划线、连字符
	for _, char := range code {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}
	return true
}

// isValidRoleName 验证角色名称格式
func isValidRoleName(name string) bool {
	return len(name) >= 2 && len(name) <= 100
}
