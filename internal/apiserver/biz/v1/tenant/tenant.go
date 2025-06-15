// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package tenant

//go:generate mockgen -destination mock_tenant.go -package tenant github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/tenant TenantBiz

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/ashwinyue/one-auth/pkg/store/where"
)

// TenantBiz 定义处理租户相关请求所需的方法.
type TenantBiz interface {
	// 用户租户相关
	GetUserTenants(ctx context.Context, rq *apiv1.GetUserTenantsRequest) (*apiv1.GetUserTenantsResponse, error)
	SwitchTenant(ctx context.Context, rq *apiv1.SwitchTenantRequest) (*apiv1.SwitchTenantResponse, error)
	GetUserProfile(ctx context.Context, rq *apiv1.GetUserProfileRequest) (*apiv1.GetUserProfileResponse, error)

	// 租户管理
	ListTenants(ctx context.Context, rq *apiv1.ListTenantsRequest) (*apiv1.ListTenantsResponse, error)
}

// tenantBiz 是 TenantBiz 接口的实现.
type tenantBiz struct {
	store store.IStore
	authz *authz.Authz
}

// 确保 tenantBiz 实现了 TenantBiz 接口.
var _ TenantBiz = (*tenantBiz)(nil)

func New(store store.IStore, authz *authz.Authz) *tenantBiz {
	return &tenantBiz{store: store, authz: authz}
}

// GetUserTenants 获取用户所属的租户列表
func (b *tenantBiz) GetUserTenants(ctx context.Context, rq *apiv1.GetUserTenantsRequest) (*apiv1.GetUserTenantsResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated
	}

	// 使用tenant store的扩展方法获取用户租户
	tenants, err := b.store.Tenant().GetUserTenants(ctx, fmt.Sprintf("%d", userID))
	if err != nil {
		log.W(ctx).Errorw("Failed to get user tenants", "user_id", userID, "err", err)
		return nil, err
	}

	// 转换为响应格式
	var tenantList []*apiv1.Tenant
	for _, tenant := range tenants {
		description := ""
		if tenant.Description != nil {
			description = *tenant.Description
		}

		status := int32(0)
		if tenant.Status {
			status = 1
		}

		tenantList = append(tenantList, &apiv1.Tenant{
			Id:          tenant.ID,
			TenantCode:  tenant.TenantCode,
			Name:        tenant.Name,
			Description: description,
			Status:      status,
		})
	}

	return &apiv1.GetUserTenantsResponse{Tenants: tenantList}, nil
}

// SwitchTenant 切换用户当前工作租户
func (b *tenantBiz) SwitchTenant(ctx context.Context, rq *apiv1.SwitchTenantRequest) (*apiv1.SwitchTenantResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated
	}

	// 验证用户是否属于该租户
	hasAccess, err := b.store.Tenant().CheckUserTenant(ctx, fmt.Sprintf("%d", userID), rq.TenantId)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errno.ErrPermissionDenied.WithMessage("user does not belong to this tenant")
	}

	// 这里可以将租户ID存储到会话或缓存中
	// 暂时返回成功，实际实现中可能需要更新JWT token或会话信息
	return &apiv1.SwitchTenantResponse{Success: true}, nil
}

// GetUserProfile 获取用户完整信息（包含当前租户、角色、权限）
func (b *tenantBiz) GetUserProfile(ctx context.Context, rq *apiv1.GetUserProfileRequest) (*apiv1.GetUserProfileResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated
	}

	// 获取当前租户ID（从请求参数或上下文中获取）
	tenantID := rq.TenantId
	if tenantID == 0 {
		// 如果没有指定租户，获取用户的默认租户
		tenantID = 1 // 默认租户
	}

	// 获取用户基本信息
	userM, err := b.store.User().Get(ctx, where.F("id", userID))
	if err != nil {
		return nil, err
	}

	// 获取用户在当前租户下的角色
	userIdentifier := fmt.Sprintf("u%d", userID)
	tenantIdentifier := fmt.Sprintf("t%d", tenantID)
	roles, err := b.authz.GetRolesForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user roles", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		roles = []string{} // 如果获取失败，返回空角色列表
	}

	// 实现获取用户权限和菜单的逻辑
	permissions, menus := b.getUserPermissionsAndMenus(ctx, userIdentifier, tenantIdentifier)

	// 获取当前租户信息
	tenantM, err := b.store.Tenant().Get(ctx, where.F("id", tenantID))
	if err != nil {
		log.W(ctx).Errorw("Failed to get tenant info", "tenant_id", tenantID, "err", err)
		// 使用默认租户信息
		tenantM = nil
	}

	var currentTenant *apiv1.Tenant
	if tenantM != nil {
		description := ""
		if tenantM.Description != nil {
			description = *tenantM.Description
		}
		status := int32(0)
		if tenantM.Status {
			status = 1
		}
		currentTenant = &apiv1.Tenant{
			Id:          tenantM.ID,
			TenantCode:  tenantM.TenantCode,
			Name:        tenantM.Name,
			Description: description,
			Status:      status,
		}
	} else {
		currentTenant = &apiv1.Tenant{
			Id:         tenantID,
			TenantCode: fmt.Sprintf("t%d", tenantID),
			Name:       "Default Tenant",
		}
	}

	return &apiv1.GetUserProfileResponse{
		User: &apiv1.UserProfile{
			Id:            userM.ID,
			UserId:        fmt.Sprintf("%d", userM.ID), // 使用数字ID转字符串
			Username:      userM.Username,
			Nickname:      userM.Nickname,
			Email:         userM.Email,
			Phone:         userM.Phone,
			CurrentTenant: currentTenant,
		},
		Roles:       roles,
		Permissions: permissions,
		Menus:       menus,
	}, nil
}

// getUserPermissionsAndMenus 获取用户权限和菜单
func (b *tenantBiz) getUserPermissionsAndMenus(ctx context.Context, userIdentifier, tenantIdentifier string) ([]*apiv1.Permission, []*apiv1.Menu) {
	// 从Casbin获取用户的所有权限（包括通过角色继承的权限）
	permissions, err := b.authz.GetImplicitPermissionsForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user permissions", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		return []*apiv1.Permission{}, []*apiv1.Menu{}
	}

	var permissionList []*apiv1.Permission
	var menuIDSet = make(map[int64]bool)
	permissionSet := make(map[int64]bool) // 用于去重

	// 解析权限并查询详细信息
	for _, perm := range permissions {
		if len(perm) >= 2 {
			// perm[1] 是权限标识符，格式为 a{id}
			permissionCode := perm[1]
			if len(permissionCode) > 1 && permissionCode[0] == 'a' {
				// 解析权限ID
				permissionIDStr := permissionCode[1:]
				if permissionID, err := strconv.ParseInt(permissionIDStr, 10, 64); err == nil {
					// 避免重复
					if permissionSet[permissionID] {
						continue
					}
					permissionSet[permissionID] = true

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

						// 收集菜单ID
						if permissionM.MenuID > 0 {
							menuIDSet[permissionM.MenuID] = true
						}
					}
				}
			}
		}
	}

	// 查询菜单详情
	var menus []*apiv1.Menu
	for menuID := range menuIDSet {
		menuM, err := b.store.Menu().Get(ctx, where.F("id", menuID))
		if err == nil && menuM.Status {
			// 处理可选字段
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

			status := int32(0)
			if menuM.Status {
				status = 1
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

	return permissionList, menus
}

// ListTenants 获取租户列表
func (b *tenantBiz) ListTenants(ctx context.Context, rq *apiv1.ListTenantsRequest) (*apiv1.ListTenantsResponse, error) {
	opts := where.NewWhere()

	// 添加分页
	if rq.Offset > 0 {
		opts = opts.O(int(rq.Offset))
	}
	if rq.Limit > 0 {
		opts = opts.L(int(rq.Limit))
	}

	count, tenants, err := b.store.Tenant().List(ctx, opts)
	if err != nil {
		log.W(ctx).Errorw("Failed to list tenants", "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 转换为响应格式
	var tenantList []*apiv1.Tenant
	for _, tenant := range tenants {
		description := ""
		if tenant.Description != nil {
			description = *tenant.Description
		}

		status := int32(0)
		if tenant.Status {
			status = 1
		}

		tenantList = append(tenantList, &apiv1.Tenant{
			Id:          tenant.ID,
			TenantCode:  tenant.TenantCode,
			Name:        tenant.Name,
			Description: description,
			Status:      status,
		})
	}

	return &apiv1.ListTenantsResponse{
		Tenants:    tenantList,
		TotalCount: count,
	}, nil
}
