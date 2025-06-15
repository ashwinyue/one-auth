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
	"strings"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"gorm.io/gorm"
)

// GetRolePermissions 获取角色权限
func (b *roleBiz) GetRolePermissions(ctx context.Context, rq *apiv1.GetRolePermissionsRequest) (*apiv1.GetRolePermissionsResponse, error) {
	// 参数验证
	if rq.RoleId <= 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("role_id must be greater than 0")
	}

	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 验证角色是否存在且属于正确的租户
	roleM, err := b.store.Role().Get(ctx, where.F("id", rq.RoleId))
	if err != nil {
		log.W(ctx).Errorw("Role not found", "role_id", rq.RoleId, "err", err)
		return nil, errno.ErrRoleNotFound.WithMessage("Role not found")
	}

	// 检查租户权限
	tenantIDInt, _ := strconv.ParseInt(tenantID, 10, 64)
	if roleM.TenantID != tenantIDInt {
		log.W(ctx).Errorw("Role not belongs to current tenant", "role_id", rq.RoleId, "role_tenant", roleM.TenantID, "current_tenant", tenantIDInt)
		return nil, errno.ErrPermissionDenied.WithMessage("Role not accessible in current tenant")
	}

	// 构建角色标识符
	roleIdentifier := fmt.Sprintf("r%d", rq.RoleId)

	// 从Casbin获取角色的权限
	permissions := b.authz.GetPermissionsForUserInDomain(roleIdentifier, fmt.Sprintf("t%s", tenantID))

	var permissionList []*apiv1.Permission

	// 解析权限并查询详细信息
	for _, perm := range permissions {
		if len(perm) >= 2 {
			// perm[1] 是权限标识符，格式为 a{id}
			permissionCode := perm[1]
			if len(permissionCode) > 1 && permissionCode[0] == 'a' {
				// 解析权限ID
				permissionIDStr := permissionCode[1:]
				if permissionID, err := strconv.ParseInt(permissionIDStr, 10, 64); err == nil {
					// 从数据库查询权限详情
					permissionM, err := b.store.Permission().Get(ctx, where.F("id", permissionID))
					if err == nil {
						description := ""
						if permissionM.Description != nil {
							description = *permissionM.Description
						}

						status := int32(0)
						if permissionM.Status {
							status = 1
						}

						permissionList = append(permissionList, &apiv1.Permission{
							Id:             permissionM.ID,
							TenantId:       permissionM.TenantID,
							MenuId:         permissionM.MenuID,
							PermissionCode: permissionM.PermissionCode,
							Name:           permissionM.Name,
							Description:    description,
							Status:         status,
						})
					}
				}
			}
		}
	}

	return &apiv1.GetRolePermissionsResponse{
		Permissions: permissionList,
	}, nil
}

// AssignRolePermissions 分配角色权限
func (b *roleBiz) AssignRolePermissions(ctx context.Context, rq *apiv1.AssignRolePermissionsRequest) (*apiv1.AssignRolePermissionsResponse, error) {
	// 参数验证
	if rq.RoleId <= 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("role_id must be greater than 0")
	}
	if len(rq.PermissionIds) == 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("permission_ids cannot be empty")
	}

	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 验证角色是否存在且属于正确的租户
	roleM, err := b.store.Role().Get(ctx, where.F("id", rq.RoleId))
	if err != nil {
		log.W(ctx).Errorw("Role not found", "role_id", rq.RoleId, "err", err)
		return nil, errno.ErrRoleNotFound.WithMessage("Role not found")
	}

	// 检查租户权限
	tenantIDInt, _ := strconv.ParseInt(tenantID, 10, 64)
	if roleM.TenantID != tenantIDInt {
		log.W(ctx).Errorw("Role not belongs to current tenant", "role_id", rq.RoleId, "role_tenant", roleM.TenantID, "current_tenant", tenantIDInt)
		return nil, errno.ErrPermissionDenied.WithMessage("Role not accessible in current tenant")
	}

	// 验证权限是否存在且属于正确的租户
	for _, permissionID := range rq.PermissionIds {
		permissionM, err := b.store.Permission().Get(ctx, where.F("id", permissionID))
		if err != nil {
			log.W(ctx).Errorw("Permission not found", "permission_id", permissionID, "err", err)
			return nil, errno.ErrInvalidArgument.WithMessage(fmt.Sprintf("Permission %d not found", permissionID))
		}

		// 检查租户权限
		if permissionM.TenantID != tenantIDInt {
			log.W(ctx).Errorw("Permission not belongs to current tenant", "permission_id", permissionID, "permission_tenant", permissionM.TenantID, "current_tenant", tenantIDInt)
			return nil, errno.ErrPermissionDenied.WithMessage(fmt.Sprintf("Permission %d not accessible in current tenant", permissionID))
		}
	}

	// 构建角色标识符
	roleIdentifier := fmt.Sprintf("r%d", rq.RoleId)
	tenantIdentifier := fmt.Sprintf("t%s", tenantID)

	// 使用数据库事务确保操作的原子性
	err = b.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除角色现有的所有权限
		_, err := b.authz.DeletePermissionsForUser(roleIdentifier)
		if err != nil {
			log.W(ctx).Errorw("Failed to delete existing role permissions", "role", roleIdentifier, "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to clear existing permissions")
		}

		// 为角色分配新的权限
		for _, permissionID := range rq.PermissionIds {
			// 构建权限标识符
			permissionIdentifier := fmt.Sprintf("a%d", permissionID)

			// 添加权限到Casbin
			_, err = b.authz.AddPermissionForUser(roleIdentifier, permissionIdentifier, tenantIdentifier)
			if err != nil {
				log.W(ctx).Errorw("Failed to add permission for role",
					"role", roleIdentifier,
					"permission", permissionIdentifier,
					"err", err)
				// 如果某个权限分配失败，回滚整个事务
				return errno.ErrDBWrite.WithMessage(fmt.Sprintf("Failed to assign permission %d to role", permissionID))
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.W(ctx).Infow("Role permissions assigned successfully", "role_id", rq.RoleId, "permission_ids", rq.PermissionIds, "tenant", tenantIdentifier)

	return &apiv1.AssignRolePermissionsResponse{
		Success: true,
	}, nil
}

// GetRoleMenus 获取角色菜单
func (b *roleBiz) GetRoleMenus(ctx context.Context, rq *apiv1.GetRoleMenusRequest) (*apiv1.GetRoleMenusResponse, error) {
	// 参数验证
	if rq.RoleId <= 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("role_id must be greater than 0")
	}

	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 验证角色是否存在且属于正确的租户
	roleM, err := b.store.Role().Get(ctx, where.F("id", rq.RoleId))
	if err != nil {
		log.W(ctx).Errorw("Role not found", "role_id", rq.RoleId, "err", err)
		return nil, errno.ErrRoleNotFound.WithMessage("Role not found")
	}

	// 检查租户权限
	tenantIDInt, _ := strconv.ParseInt(tenantID, 10, 64)
	if roleM.TenantID != tenantIDInt {
		log.W(ctx).Errorw("Role not belongs to current tenant", "role_id", rq.RoleId, "role_tenant", roleM.TenantID, "current_tenant", tenantIDInt)
		return nil, errno.ErrPermissionDenied.WithMessage("Role not accessible in current tenant")
	}

	// 构建角色标识符
	roleIdentifier := fmt.Sprintf("r%d", rq.RoleId)

	// 从Casbin获取角色的权限
	permissions := b.authz.GetPermissionsForUserInDomain(roleIdentifier, fmt.Sprintf("t%s", tenantID))

	// 收集菜单ID
	menuIDSet := make(map[int64]bool)
	for _, perm := range permissions {
		if len(perm) >= 2 {
			permissionCode := perm[1]
			if len(permissionCode) > 1 && permissionCode[0] == 'a' {
				// 解析权限ID
				permissionIDStr := permissionCode[1:]
				if permissionID, err := strconv.ParseInt(permissionIDStr, 10, 64); err == nil {
					// 查询权限对应的菜单ID
					permissionM, err := b.store.Permission().Get(ctx, where.F("id", permissionID))
					if err == nil && permissionM.MenuID > 0 {
						menuIDSet[permissionM.MenuID] = true
					}
				}
			}
		}
	}

	// 查询菜单详情
	var menus []*apiv1.Menu
	for menuID := range menuIDSet {
		menuM, err := b.store.Menu().Get(ctx, where.F("id", menuID))
		if err == nil {
			status := int32(0)
			if menuM.Status {
				status = 1
			}

			parentID := int64(0)
			if menuM.ParentID != nil {
				parentID = *menuM.ParentID
			}

			routePath := ""
			if menuM.RoutePath != nil {
				routePath = *menuM.RoutePath
			}

			apiPath := ""
			if menuM.APIPath != nil {
				apiPath = *menuM.APIPath
			}

			httpMethods := ""
			if menuM.HTTPMethods != nil {
				httpMethods = *menuM.HTTPMethods
			}

			component := ""
			if menuM.Component != nil {
				component = *menuM.Component
			}

			icon := ""
			if menuM.Icon != nil {
				icon = *menuM.Icon
			}

			menus = append(menus, &apiv1.Menu{
				Id:          menuM.ID,
				TenantId:    menuM.TenantID,
				ParentId:    parentID,
				MenuCode:    menuM.MenuCode,
				Title:       menuM.Title,
				RoutePath:   routePath,
				ApiPath:     apiPath,
				HttpMethods: httpMethods,
				RequireAuth: menuM.RequireAuth,
				Component:   component,
				Icon:        icon,
				SortOrder:   int32(menuM.SortOrder),
				MenuType: func() int32 {
					if menuM.MenuType {
						return 1
					} else {
						return 0
					}
				}(),
				Visible: menuM.Visible,
				Status:  status,
			})
		}
	}

	return &apiv1.GetRoleMenusResponse{
		Menus: menus,
	}, nil
}

// UpdateRoleMenus 更新角色菜单（改进版）
func (b *roleBiz) UpdateRoleMenus(ctx context.Context, rq *apiv1.UpdateRoleMenusRequest) (*apiv1.UpdateRoleMenusResponse, error) {
	// 参数验证
	if rq.RoleId <= 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("role_id must be greater than 0")
	}
	if len(rq.MenuIds) == 0 {
		return nil, errno.ErrInvalidArgument.WithMessage("menu_ids cannot be empty")
	}

	// 获取租户信息
	tenantID := contextx.TenantID(ctx)
	if tenantID == "" {
		tenantID = "1" // 默认租户
	}

	// 验证角色是否存在且属于正确的租户
	roleM, err := b.store.Role().Get(ctx, where.F("id", rq.RoleId))
	if err != nil {
		log.W(ctx).Errorw("Role not found", "role_id", rq.RoleId, "err", err)
		return nil, errno.ErrRoleNotFound.WithMessage("Role not found")
	}

	// 检查租户权限
	tenantIDInt, _ := strconv.ParseInt(tenantID, 10, 64)
	if roleM.TenantID != tenantIDInt {
		log.W(ctx).Errorw("Role not belongs to current tenant", "role_id", rq.RoleId, "role_tenant", roleM.TenantID, "current_tenant", tenantIDInt)
		return nil, errno.ErrPermissionDenied.WithMessage("Role not accessible in current tenant")
	}

	// 批量验证菜单是否存在且属于正确的租户
	menuMap := make(map[int64]*model.MenuM)
	for _, menuID := range rq.MenuIds {
		menuM, err := b.store.Menu().Get(ctx, where.F("id", menuID))
		if err != nil {
			log.W(ctx).Errorw("Menu not found", "menu_id", menuID, "err", err)
			return nil, errno.ErrInvalidArgument.WithMessage(fmt.Sprintf("Menu %d not found", menuID))
		}

		// 检查租户权限
		if menuM.TenantID != tenantIDInt {
			log.W(ctx).Errorw("Menu not belongs to current tenant", "menu_id", menuID, "menu_tenant", menuM.TenantID, "current_tenant", tenantIDInt)
			return nil, errno.ErrPermissionDenied.WithMessage(fmt.Sprintf("Menu %d not accessible in current tenant", menuID))
		}

		menuMap[menuID] = menuM
	}

	// 构建角色和租户标识符
	roleIdentifier := fmt.Sprintf("r%d", rq.RoleId)
	tenantIdentifier := fmt.Sprintf("t%s", tenantID)

	// 使用数据库事务确保操作的原子性
	var assignedPermissions []string
	err = b.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除角色现有的所有权限
		deletedCount, err := b.authz.DeletePermissionsForUser(roleIdentifier)
		if err != nil {
			log.W(ctx).Errorw("Failed to delete existing role permissions", "role", roleIdentifier, "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to clear existing permissions")
		}
		log.W(ctx).Infow("Deleted existing permissions", "role", roleIdentifier, "count", deletedCount)

		// 收集所有需要分配的权限
		var allPermissions []*model.PermissionM
		for _, menuID := range rq.MenuIds {
			// 查询菜单对应的权限
			_, permissions, err := b.store.Permission().List(ctx, where.F("menu_id", menuID).F("status", true))
			if err != nil {
				log.W(ctx).Errorw("Failed to get menu permissions", "menu_id", menuID, "err", err)
				return errno.ErrDBRead.WithMessage(fmt.Sprintf("Failed to get permissions for menu %d", menuID))
			}

			if len(permissions) == 0 {
				menuM := menuMap[menuID]
				log.W(ctx).Warnw("Menu has no permissions, consider creating default permissions",
					"menu_id", menuID,
					"menu_code", menuM.MenuCode,
					"menu_title", menuM.Title)
				// 这里可以选择自动生成权限或者跳过
				continue
			}

			allPermissions = append(allPermissions, permissions...)
		}

		// 验证权限命名约定并记录统计信息
		standardPermissions := 0
		legacyPermissions := 0

		// 为角色分配所有权限
		for _, permission := range allPermissions {
			// 验证权限编码格式
			if isStandardPermissionCode(permission.PermissionCode) {
				standardPermissions++
			} else {
				legacyPermissions++
				log.W(ctx).Warnw("Legacy permission code format detected",
					"permission_code", permission.PermissionCode,
					"suggestion", fmt.Sprintf("Consider updating to format like '%s:action'", getMenuCodeFromPermission(permission)))
			}

			permissionIdentifier := fmt.Sprintf("a%d", permission.ID)
			_, err = b.authz.AddPermissionForUser(roleIdentifier, permissionIdentifier, tenantIdentifier)
			if err != nil {
				log.W(ctx).Errorw("Failed to add permission for role",
					"role", roleIdentifier,
					"permission", permissionIdentifier,
					"permission_code", permission.PermissionCode,
					"err", err)
				// 如果某个权限分配失败，回滚整个事务
				return errno.ErrDBWrite.WithMessage(fmt.Sprintf("Failed to assign permission %s to role", permission.PermissionCode))
			}

			assignedPermissions = append(assignedPermissions, permission.PermissionCode)
		}

		// 记录权限分配统计
		log.W(ctx).Infow("Role permissions assignment completed",
			"role_id", rq.RoleId,
			"role_code", roleM.RoleCode,
			"menu_count", len(rq.MenuIds),
			"total_permissions", len(allPermissions),
			"standard_format_permissions", standardPermissions,
			"legacy_format_permissions", legacyPermissions,
			"tenant", tenantIdentifier)

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.W(ctx).Infow("Role menus updated successfully",
		"role_id", rq.RoleId,
		"menu_ids", rq.MenuIds,
		"assigned_permissions", assignedPermissions,
		"tenant", tenantIdentifier)

	return &apiv1.UpdateRoleMenusResponse{
		Success: true,
	}, nil
}

// isStandardPermissionCode 检查权限编码是否符合标准格式 {module}:{action}
func isStandardPermissionCode(permissionCode string) bool {
	parts := strings.Split(permissionCode, ":")
	return len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != ""
}

// getMenuCodeFromPermission 从权限中推断菜单编码（用于建议）
func getMenuCodeFromPermission(permission *model.PermissionM) string {
	// 简单实现：从权限编码中提取模块部分
	if strings.Contains(permission.PermissionCode, ":") {
		parts := strings.Split(permission.PermissionCode, ":")
		return parts[0]
	}

	// 对于legacy格式，尝试从权限编码中推断
	code := permission.PermissionCode
	if strings.Contains(code, "_") {
		parts := strings.Split(code, "_")
		if len(parts) >= 2 {
			return parts[0]
		}
	}

	return "unknown"
}
