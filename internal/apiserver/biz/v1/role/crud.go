// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package role

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"gorm.io/gorm"
)

// ListRoles 获取角色列表
func (b *roleBiz) ListRoles(ctx context.Context, rq *apiv1.ListRolesRequest) (*apiv1.ListRolesResponse, error) {
	opts := where.NewWhere().T(ctx) // 使用where.T自动添加租户过滤

	// 如果请求中指定了租户ID，则覆盖上下文中的租户ID
	if rq.TenantId > 0 {
		opts = opts.F("tenant_id", rq.TenantId)
	}

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
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 转换为响应格式
	var roleList []*apiv1.Role
	for _, role := range roles {
		description := ""
		if role.Description != nil {
			description = *role.Description
		}

		status := int32(0)
		if role.Status {
			status = 1
		}

		roleList = append(roleList, &apiv1.Role{
			Id:          role.ID,
			TenantId:    role.TenantID,
			RoleCode:    role.RoleCode,
			Name:        role.Name,
			Description: description,
			Status:      status,
		})
	}

	return &apiv1.ListRolesResponse{
		Roles:      roleList,
		TotalCount: count,
	}, nil
}

// CreateRole 创建角色
func (b *roleBiz) CreateRole(ctx context.Context, rq *apiv1.CreateRoleRequest) (*apiv1.CreateRoleResponse, error) {
	// 参数验证
	if rq.TenantId <= 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("tenant_id must be greater than 0")
	}
	if rq.RoleCode == "" {
		return nil, errno.ErrInvalidArgument.WithMessage("role_code cannot be empty")
	}
	if rq.Name == "" {
		return nil, errno.ErrInvalidArgument.WithMessage("name cannot be empty")
	}

	// 角色编码格式校验：只能包含字母、数字、下划线、连字符，长度3-50
	if !isValidRoleCode(rq.RoleCode) {
		return nil, errno.ErrInvalidRoleCode.WithMessage("role_code format is invalid. Must be 3-50 characters, containing only letters, numbers, underscores, and hyphens")
	}

	// 角色名称格式校验：长度2-100
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

	// 验证请求的租户ID与当前上下文租户ID是否匹配
	if rq.TenantId != tenantIDInt {
		return nil, errno.ErrPermissionDenied.WithMessage("cannot create role for different tenant")
	}

	// 使用数据库事务确保操作的原子性
	var roleM *model.RoleM
	err = b.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查角色编码是否已存在（在当前租户内）
		existingRole, err := b.store.Role().Get(ctx, where.F("role_code", rq.RoleCode).F("tenant_id", rq.TenantId))
		if err == nil && existingRole != nil {
			return errno.ErrRoleAlreadyExists.WithMessage("role code already exists in current tenant")
		}

		// 检查角色名称是否已存在（在当前租户内）
		existingRole, err = b.store.Role().Get(ctx, where.F("name", rq.Name).F("tenant_id", rq.TenantId))
		if err == nil && existingRole != nil {
			return errno.ErrRoleAlreadyExists.WithMessage("role name already exists in current tenant")
		}

		// 创建角色
		roleM = &model.RoleM{
			TenantID:    rq.TenantId,
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
		Role: &apiv1.Role{
			Id:          roleM.ID,
			TenantId:    roleM.TenantID,
			RoleCode:    roleM.RoleCode,
			Name:        roleM.Name,
			Description: *roleM.Description,
			Status:      1,
		},
	}, nil
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

// UpdateRole 更新角色
func (b *roleBiz) UpdateRole(ctx context.Context, rq *apiv1.UpdateRoleRequest) (*apiv1.UpdateRoleResponse, error) {
	// 获取现有角色
	roleM, err := b.store.Role().Get(ctx, where.F("id", rq.RoleId))
	if err != nil {
		return nil, errno.ErrRoleNotFound.WithMessage("role not found")
	}

	// 更新字段
	roleM.Name = rq.Name
	roleM.Description = &rq.Description

	if err := b.store.Role().Update(ctx, roleM); err != nil {
		log.W(ctx).Errorw("Failed to update role", "err", err)
		return nil, errno.ErrDBWrite.WithMessage(err.Error())
	}

	// 转换为响应格式
	return &apiv1.UpdateRoleResponse{
		Role: &apiv1.Role{
			Id:          roleM.ID,
			TenantId:    roleM.TenantID,
			RoleCode:    roleM.RoleCode,
			Name:        roleM.Name,
			Description: *roleM.Description,
			Status:      1,
		},
	}, nil
}

// DeleteRole 删除角色
func (b *roleBiz) DeleteRole(ctx context.Context, rq *apiv1.DeleteRoleRequest) (*apiv1.DeleteRoleResponse, error) {
	// 检查角色是否可以删除
	canDelete, err := b.CheckDeleteRole(ctx, &apiv1.CheckDeleteRoleRequest{RoleId: rq.RoleId})
	if err != nil {
		return nil, err
	}
	if !canDelete.CanDelete {
		return nil, errno.ErrRoleInUse.WithMessage(canDelete.Reason)
	}

	// 删除角色
	err = b.store.Role().Delete(ctx, where.F("id", rq.RoleId))
	if err != nil {
		log.W(ctx).Errorw("Failed to delete role", "role_id", rq.RoleId, "err", err)
		return nil, errno.ErrDBWrite.WithMessage("Failed to delete role")
	}

	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
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

	return &apiv1.DeleteRoleResponse{
		Success: true,
	}, nil
}

// CheckDeleteRole 检查角色是否可以删除
func (b *roleBiz) CheckDeleteRole(ctx context.Context, rq *apiv1.CheckDeleteRoleRequest) (*apiv1.CheckDeleteRoleResponse, error) {
	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 构建角色标识符
	roleIdentifier := fmt.Sprintf("r%d", rq.RoleId)
	tenantIdentifier := fmt.Sprintf("t%s", tenantID)

	// 检查是否有用户使用该角色
	users, _ := b.authz.GetUsersForRole(roleIdentifier, tenantIdentifier)
	if len(users) > 0 {
		return &apiv1.CheckDeleteRoleResponse{
			CanDelete: false,
			Reason:    fmt.Sprintf("角色正在被 %d 个用户使用，无法删除", len(users)),
		}, nil
	}

	// 检查是否是系统内置角色
	roleM, err := b.store.Role().Get(ctx, where.F("id", rq.RoleId))
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("Failed to get role")
	}

	// 检查是否是系统角色（假设role_code以system_开头的是系统角色）
	if len(roleM.RoleCode) > 7 && roleM.RoleCode[:7] == "system_" {
		return &apiv1.CheckDeleteRoleResponse{
			CanDelete: false,
			Reason:    "系统内置角色不能删除",
		}, nil
	}

	return &apiv1.CheckDeleteRoleResponse{
		CanDelete: true,
		Reason:    "",
	}, nil
}
