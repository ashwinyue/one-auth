// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package menu

//go:generate mockgen -destination mock_menu.go -package menu github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/menu MenuBiz

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

// MenuBiz 定义处理菜单相关请求所需的方法.
type MenuBiz interface {
	// 菜单相关
	GetUserMenus(ctx context.Context, rq *apiv1.GetUserMenusRequest) (*apiv1.GetUserMenusResponse, error)
	ListMenus(ctx context.Context, rq *apiv1.ListMenusRequest) (*apiv1.ListMenusResponse, error)
}

// menuBiz 是 MenuBiz 接口的实现.
type menuBiz struct {
	store store.IStore
	authz *authz.Authz
}

// 确保 menuBiz 实现了 MenuBiz 接口.
var _ MenuBiz = (*menuBiz)(nil)

func New(store store.IStore, authz *authz.Authz) *menuBiz {
	return &menuBiz{store: store, authz: authz}
}

// GetUserMenus 获取用户菜单
func (b *menuBiz) GetUserMenus(ctx context.Context, rq *apiv1.GetUserMenusRequest) (*apiv1.GetUserMenusResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated
	}

	// 获取租户信息
	tenantID := rq.TenantId
	if tenantID == 0 {
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

	// 从Casbin获取用户的所有权限（包括通过角色继承的权限）
	permissions, err := b.authz.GetImplicitPermissionsForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user permissions", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to get user permissions")
	}

	// 收集菜单ID
	menuIDSet := make(map[int64]bool)
	for _, perm := range permissions {
		if len(perm) >= 2 {
			// perm[1] 是权限标识符，格式为 a{id}
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

	// 查询菜单详情并构建菜单树
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

	// 构建菜单树结构
	menuTree := buildMenuTree(menus)

	return &apiv1.GetUserMenusResponse{
		Menus: menuTree,
	}, nil
}

// buildMenuTree 构建菜单树结构
func buildMenuTree(menus []*apiv1.Menu) []*apiv1.Menu {
	// 创建菜单映射
	menuMap := make(map[int64]*apiv1.Menu)
	for _, menu := range menus {
		menuMap[menu.Id] = menu
	}

	// 构建树结构
	var rootMenus []*apiv1.Menu
	for _, menu := range menus {
		if menu.ParentId == 0 {
			// 根菜单
			rootMenus = append(rootMenus, menu)
		} else {
			// 子菜单
			if parent, exists := menuMap[menu.ParentId]; exists {
				if parent.Children == nil {
					parent.Children = make([]*apiv1.Menu, 0)
				}
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	// 按排序字段排序
	sortMenus(rootMenus)
	for _, menu := range rootMenus {
		sortMenuChildren(menu)
	}

	return rootMenus
}

// sortMenus 对菜单列表进行排序
func sortMenus(menus []*apiv1.Menu) {
	for i := 0; i < len(menus)-1; i++ {
		for j := i + 1; j < len(menus); j++ {
			if menus[i].SortOrder > menus[j].SortOrder {
				menus[i], menus[j] = menus[j], menus[i]
			}
		}
	}
}

// sortMenuChildren 递归排序菜单的子菜单
func sortMenuChildren(menu *apiv1.Menu) {
	if menu.Children != nil && len(menu.Children) > 0 {
		sortMenus(menu.Children)
		for _, child := range menu.Children {
			sortMenuChildren(child)
		}
	}
}

// ListMenus 获取菜单列表
func (b *menuBiz) ListMenus(ctx context.Context, rq *apiv1.ListMenusRequest) (*apiv1.ListMenusResponse, error) {
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

	count, menus, err := b.store.Menu().List(ctx, opts)
	if err != nil {
		log.W(ctx).Errorw("Failed to list menus", "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 转换为响应格式
	var menuList []*apiv1.Menu
	for _, menu := range menus {
		icon := ""
		if menu.Icon != nil {
			icon = *menu.Icon
		}

		routePath := ""
		if menu.RoutePath != nil {
			routePath = *menu.RoutePath
		}

		apiPath := ""
		if menu.APIPath != nil {
			apiPath = *menu.APIPath
		}

		httpMethods := ""
		if menu.HTTPMethods != nil {
			httpMethods = *menu.HTTPMethods
		}

		component := ""
		if menu.Component != nil {
			component = *menu.Component
		}

		status := int32(0)
		if menu.Status {
			status = 1
		}

		parentID := int64(0)
		if menu.ParentID != nil {
			parentID = *menu.ParentID
		}

		menuType := int32(1)
		if menu.MenuType {
			menuType = 1
		}

		menuList = append(menuList, &apiv1.Menu{
			Id:          menu.ID,
			TenantId:    menu.TenantID,
			ParentId:    parentID,
			MenuCode:    menu.MenuCode,
			Title:       menu.Title,
			RoutePath:   routePath,
			ApiPath:     apiPath,
			HttpMethods: httpMethods,
			RequireAuth: menu.RequireAuth,
			Component:   component,
			Icon:        icon,
			SortOrder:   menu.SortOrder,
			MenuType:    menuType,
			Visible:     menu.Visible,
			Status:      status,
		})
	}

	return &apiv1.ListMenusResponse{
		Menus:      menuList,
		TotalCount: count,
	}, nil
}
