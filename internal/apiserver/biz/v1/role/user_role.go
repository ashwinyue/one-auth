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

	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"gorm.io/gorm"
)

// GetUserRoles 获取用户角色
func (b *roleBiz) GetUserRoles(ctx context.Context, rq *apiv1.GetUserRolesRequest) (*apiv1.GetUserRolesResponse, error) {
	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 将字符串用户ID转换为数字ID
	userIDInt, err := strconv.ParseInt(rq.UserId, 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid user_id format")
	}

	// 构建用户标识符和租户标识符
	userIdentifier := fmt.Sprintf("u%d", userIDInt)
	tenantIdentifier := fmt.Sprintf("t%s", tenantID)

	// 从Casbin获取用户角色
	roleIdentifiers, err := b.authz.GetRolesForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user roles from Casbin", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to get user roles")
	}

	var roles []*apiv1.Role

	// 解析角色标识符并查询角色详情
	for _, roleIdentifier := range roleIdentifiers {
		if len(roleIdentifier) > 1 && roleIdentifier[0] == 'r' {
			// 解析角色ID
			roleIDStr := roleIdentifier[1:]
			if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
				// 从数据库查询角色详情
				roleM, err := b.store.Role().Get(ctx, where.F("id", roleID))
				if err == nil {
					description := ""
					if roleM.Description != nil {
						description = *roleM.Description
					}

					status := int32(0)
					if roleM.Status {
						status = 1
					}

					roles = append(roles, &apiv1.Role{
						Id:          roleM.ID,
						TenantId:    roleM.TenantID,
						RoleCode:    roleM.RoleCode,
						Name:        roleM.Name,
						Description: description,
						Status:      status,
					})
				}
			}
		}
	}

	return &apiv1.GetUserRolesResponse{
		Roles: roles,
	}, nil
}

// AssignUserRoles 分配用户角色
func (b *roleBiz) AssignUserRoles(ctx context.Context, rq *apiv1.AssignUserRolesRequest) (*apiv1.AssignUserRolesResponse, error) {
	// 参数验证
	if rq.UserId == "" {
		return nil, errno.ErrInvalidArgument.WithMessage("user_id cannot be empty")
	}
	if len(rq.RoleIds) == 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("role_ids cannot be empty")
	}

	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 将字符串用户ID转换为数字ID
	userIDInt, err := strconv.ParseInt(rq.UserId, 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid user_id format")
	}

	// 验证用户是否存在
	_, err = b.store.User().Get(ctx, where.F("id", userIDInt))
	if err != nil {
		log.W(ctx).Errorw("User not found", "user_id", userIDInt, "err", err)
		return nil, errno.ErrUserNotFound.WithMessage("User not found")
	}

	// 验证角色是否存在且属于正确的租户
	for _, roleID := range rq.RoleIds {
		roleM, err := b.store.Role().Get(ctx, where.F("id", roleID))
		if err != nil {
			log.W(ctx).Errorw("Role not found", "role_id", roleID, "err", err)
			return nil, errno.ErrRoleNotFound.WithMessage(fmt.Sprintf("Role %d not found", roleID))
		}

		// 检查租户权限（将字符串租户ID转为int64进行比较）
		tenantIDInt, _ := strconv.ParseInt(tenantID, 10, 64)
		if roleM.TenantID != tenantIDInt {
			log.W(ctx).Errorw("Role not belongs to current tenant", "role_id", roleID, "role_tenant", roleM.TenantID, "current_tenant", tenantIDInt)
			return nil, errno.ErrPermissionDenied.WithMessage(fmt.Sprintf("Role %d not accessible in current tenant", roleID))
		}
	}

	// 构建用户标识符和租户标识符
	userIdentifier := fmt.Sprintf("u%d", userIDInt)
	tenantIdentifier := fmt.Sprintf("t%s", tenantID)

	// 使用数据库事务确保操作的原子性
	err = b.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除用户现有的所有角色
		success := b.authz.DeleteAllRolesForUser(userIdentifier, tenantIdentifier)
		if !success {
			log.W(ctx).Errorw("Failed to delete existing user roles", "user_id", userIdentifier, "tenant", tenantIdentifier)
			return errno.ErrDBWrite.WithMessage("Failed to clear existing roles")
		}

		// 为用户分配新的角色
		for _, roleID := range rq.RoleIds {
			// 构建角色标识符
			roleIdentifier := fmt.Sprintf("r%d", roleID)

			// 添加用户角色到Casbin
			_, err := b.authz.AddRoleForUser(userIdentifier, roleIdentifier, tenantIdentifier)
			if err != nil {
				log.W(ctx).Errorw("Failed to add role for user",
					"user_id", userIdentifier,
					"role", roleIdentifier,
					"tenant", tenantIdentifier,
					"err", err)
				// 如果某个角色分配失败，回滚整个事务
				return errno.ErrDBWrite.WithMessage(fmt.Sprintf("Failed to assign role %d to user", roleID))
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.W(ctx).Infow("User roles assigned successfully", "user_id", userIdentifier, "role_ids", rq.RoleIds, "tenant", tenantIdentifier)

	return &apiv1.AssignUserRolesResponse{
		Success: true,
	}, nil
}

// GetRolesByUser 获取当前用户的角色
func (b *roleBiz) GetRolesByUser(ctx context.Context, rq *apiv1.GetRolesByUserRequest) (*apiv1.GetRolesByUserResponse, error) {
	// 从上下文获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not authenticated")
	}

	// 获取租户信息
	tenantID := rq.TenantId
	if tenantID == 0 {
		// 如果没有指定租户，使用上下文中的租户或默认租户
		contextTenantID := contextx.TenantID(ctx)
		if contextTenantID != "" {
			if tid, err := strconv.ParseInt(contextTenantID, 10, 64); err == nil {
				tenantID = tid
			}
		}
		if tenantID == 0 {
			tenantID = 1 // 默认租户
		}
	}

	// 构建用户标识符和租户标识符
	userIdentifier := fmt.Sprintf("u%d", userID)
	tenantIdentifier := fmt.Sprintf("t%d", tenantID)

	// 从Casbin获取用户角色
	roleIdentifiers, err := b.authz.GetRolesForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user roles from Casbin", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to get user roles")
	}

	var roles []*apiv1.Role

	// 解析角色标识符并查询角色详情
	for _, roleIdentifier := range roleIdentifiers {
		if len(roleIdentifier) > 1 && roleIdentifier[0] == 'r' {
			// 解析角色ID
			roleIDStr := roleIdentifier[1:]
			if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
				// 从数据库查询角色详情
				roleM, err := b.store.Role().Get(ctx, where.F("id", roleID))
				if err == nil {
					description := ""
					if roleM.Description != nil {
						description = *roleM.Description
					}

					status := int32(0)
					if roleM.Status {
						status = 1
					}

					roles = append(roles, &apiv1.Role{
						Id:          roleM.ID,
						TenantId:    roleM.TenantID,
						RoleCode:    roleM.RoleCode,
						Name:        roleM.Name,
						Description: description,
						Status:      status,
					})
				}
			}
		}
	}

	return &apiv1.GetRolesByUserResponse{
		Roles: roles,
	}, nil
}

// RefreshPrivilegeData 刷新权限数据
func (b *roleBiz) RefreshPrivilegeData(ctx context.Context, rq *apiv1.RefreshPrivilegeDataRequest) (*apiv1.RefreshPrivilegeDataResponse, error) {
	// 重新加载Casbin策略
	err := b.authz.LoadPolicy()
	if err != nil {
		log.W(ctx).Errorw("Failed to reload Casbin policy", "err", err)
		return &apiv1.RefreshPrivilegeDataResponse{
			Success: false,
		}, errno.ErrDBWrite.WithMessage("Failed to reload policy")
	}

	log.W(ctx).Infow("Privilege data refreshed successfully")

	return &apiv1.RefreshPrivilegeDataResponse{
		Success: true,
	}, nil
}
