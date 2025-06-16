// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package menu

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	v1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MenuBiz 定义了菜单相关的业务逻辑接口
type MenuBiz interface {
	// 创建菜单
	CreateMenu(ctx context.Context, r *v1.CreateMenuRequest) (*v1.CreateMenuResponse, error)

	// 更新菜单
	UpdateMenu(ctx context.Context, r *v1.UpdateMenuRequest) (*v1.UpdateMenuResponse, error)

	// 删除菜单
	DeleteMenu(ctx context.Context, r *v1.DeleteMenuRequest) (*v1.DeleteMenuResponse, error)

	// 获取菜单详情
	GetMenu(ctx context.Context, r *v1.GetMenuRequest) (*v1.GetMenuResponse, error)

	// 获取菜单列表
	ListMenus(ctx context.Context, r *v1.ListMenusRequest) (*v1.ListMenusResponse, error)

	// 获取用户菜单（权限过滤后的菜单树）
	GetUserMenus(ctx context.Context, r *v1.GetUserMenusRequest) (*v1.GetUserMenusResponse, error)

	// 获取菜单树
	GetMenuTree(ctx context.Context, r *v1.GetMenuTreeRequest) (*v1.GetMenuTreeResponse, error)

	// 批量更新菜单排序
	UpdateMenuSort(ctx context.Context, r *v1.UpdateMenuSortRequest) (*v1.UpdateMenuSortResponse, error)

	// 复制菜单
	CopyMenu(ctx context.Context, r *v1.CopyMenuRequest) (*v1.CopyMenuResponse, error)

	// 移动菜单
	MoveMenu(ctx context.Context, r *v1.MoveMenuRequest) (*v1.MoveMenuResponse, error)
}

// menuBiz 是 MenuBiz 接口的实现
type menuBiz struct {
	ds    store.IStore
	authz *authz.Authz
}

// 确保 menuBiz 实现了 MenuBiz 接口
var _ MenuBiz = (*menuBiz)(nil)

// NewMenuBiz 创建一个新的菜单业务逻辑实例
func NewMenuBiz(ds store.IStore, authz *authz.Authz) *menuBiz {
	return &menuBiz{ds: ds, authz: authz}
}

// CreateMenu 创建菜单
func (b *menuBiz) CreateMenu(ctx context.Context, r *v1.CreateMenuRequest) (*v1.CreateMenuResponse, error) {
	// 验证菜单编码是否已存在
	exists, err := b.ds.Menu().IsMenuCodeExists(ctx, r.MenuCode, r.TenantId)
	if err != nil {
		log.W(ctx).Errorw("Failed to check menu code existence", "menu_code", r.MenuCode, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}
	if exists {
		return nil, errno.ErrMenuCodeExists.WithMessage(fmt.Sprintf("菜单编码 %s 已存在", r.MenuCode))
	}

	// 如果指定了父菜单，验证父菜单是否存在
	if r.ParentId > 0 {
		_, err := b.ds.Menu().Get(ctx, where.F("id", r.ParentId, "tenant_id", r.TenantId))
		if err != nil {
			log.W(ctx).Errorw("Parent menu not found", "parent_id", r.ParentId, "err", err)
			return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("父菜单ID %d 不存在", r.ParentId))
		}
	}

	// 创建菜单模型
	menu := &model.MenuM{
		TenantID:  r.TenantId,
		MenuCode:  r.MenuCode,
		Title:     r.Title,
		MenuType:  r.MenuType,
		SortOrder: r.SortOrder,
		Visible:   r.Visible,
		Status:    r.Status,
	}

	// 处理可选字段
	if r.ParentId > 0 {
		menu.ParentID = &r.ParentId
	}
	if r.RoutePath != "" {
		menu.RoutePath = &r.RoutePath
	}
	if r.Component != "" {
		menu.Component = &r.Component
	}
	if r.Icon != "" {
		menu.Icon = &r.Icon
	}
	if r.Remark != "" {
		menu.Remark = &r.Remark
	}

	// 保存菜单
	err = b.ds.Menu().Create(ctx, menu)
	if err != nil {
		log.W(ctx).Errorw("Failed to create menu", "menu_code", r.MenuCode, "err", err)
		return nil, errno.ErrDBWrite.WithMessage(err.Error())
	}

	log.W(ctx).Infow("Menu created successfully", "menu_id", menu.ID, "menu_code", r.MenuCode)

	return &v1.CreateMenuResponse{
		MenuId:  menu.ID,
		Message: "菜单创建成功",
	}, nil
}

// UpdateMenu 更新菜单
func (b *menuBiz) UpdateMenu(ctx context.Context, r *v1.UpdateMenuRequest) (*v1.UpdateMenuResponse, error) {
	// 获取现有菜单
	menu, err := b.ds.Menu().Get(ctx, where.F("id", r.MenuId))
	if err != nil {
		log.W(ctx).Errorw("Menu not found", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("菜单ID %d 不存在", r.MenuId))
	}

	// 如果更新菜单编码，检查是否与其他菜单冲突
	if r.MenuCode != "" && r.MenuCode != menu.MenuCode {
		exists, err := b.ds.Menu().IsMenuCodeExists(ctx, r.MenuCode, menu.TenantID)
		if err != nil {
			log.W(ctx).Errorw("Failed to check menu code existence", "menu_code", r.MenuCode, "err", err)
			return nil, errno.ErrDBRead.WithMessage(err.Error())
		}
		if exists {
			return nil, errno.ErrMenuCodeExists.WithMessage(fmt.Sprintf("菜单编码 %s 已存在", r.MenuCode))
		}
		menu.MenuCode = r.MenuCode
	}

	// 如果更新父菜单，验证父菜单是否存在且不会形成循环引用
	if r.ParentId != nil {
		if *r.ParentId > 0 {
			// 检查父菜单是否存在
			_, err := b.ds.Menu().Get(ctx, where.F("id", *r.ParentId, "tenant_id", menu.TenantID))
			if err != nil {
				log.W(ctx).Errorw("Parent menu not found", "parent_id", *r.ParentId, "err", err)
				return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("父菜单ID %d 不存在", *r.ParentId))
			}

			// 检查是否会形成循环引用
			if err := b.checkMenuHierarchyLoop(ctx, menu.ID, *r.ParentId); err != nil {
				return nil, err
			}
		}
		menu.ParentID = r.ParentId
	}

	// 更新其他字段
	if r.Title != "" {
		menu.Title = r.Title
	}
	if r.MenuType != nil {
		menu.MenuType = *r.MenuType
	}
	if r.RoutePath != nil {
		menu.RoutePath = r.RoutePath
	}
	if r.Component != nil {
		menu.Component = r.Component
	}
	if r.Icon != nil {
		menu.Icon = r.Icon
	}
	if r.SortOrder != nil {
		menu.SortOrder = *r.SortOrder
	}
	if r.Visible != nil {
		menu.Visible = *r.Visible
	}
	if r.Status != nil {
		menu.Status = *r.Status
	}
	if r.Remark != nil {
		menu.Remark = r.Remark
	}

	// 保存更新
	err = b.ds.Menu().Update(ctx, menu)
	if err != nil {
		log.W(ctx).Errorw("Failed to update menu", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrDBWrite.WithMessage(err.Error())
	}

	log.W(ctx).Infow("Menu updated successfully", "menu_id", r.MenuId)

	return &v1.UpdateMenuResponse{
		Success: true,
		Message: "菜单更新成功",
	}, nil
}

// DeleteMenu 删除菜单
func (b *menuBiz) DeleteMenu(ctx context.Context, r *v1.DeleteMenuRequest) (*v1.DeleteMenuResponse, error) {
	// 检查菜单是否存在
	menu, err := b.ds.Menu().Get(ctx, where.F("id", r.MenuId))
	if err != nil {
		log.W(ctx).Errorw("Menu not found", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("菜单ID %d 不存在", r.MenuId))
	}

	// 检查是否有子菜单
	childMenus, err := b.ds.Menu().GetChildMenus(ctx, r.MenuId, nil)
	if err != nil {
		log.W(ctx).Errorw("Failed to check child menus", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	if len(childMenus) > 0 && !r.Force {
		return nil, errno.ErrMenuHasChildren.WithMessage("菜单下存在子菜单，请先删除子菜单或使用强制删除")
	}

	// 如果强制删除，递归删除所有子菜单
	if r.Force && len(childMenus) > 0 {
		for _, child := range childMenus {
			_, err := b.DeleteMenu(ctx, &v1.DeleteMenuRequest{
				MenuId: child.ID,
				Force:  true,
			})
			if err != nil {
				log.W(ctx).Errorw("Failed to delete child menu", "child_menu_id", child.ID, "err", err)
				return nil, err
			}
		}
	}

	// 删除菜单权限关联
	err = b.ds.MenuPermission().ClearMenuPermissions(ctx, r.MenuId)
	if err != nil {
		log.W(ctx).Errorw("Failed to clear menu permissions", "menu_id", r.MenuId, "err", err)
		// 不阻止删除，只记录警告
	}

	// 删除菜单
	err = b.ds.Menu().Delete(ctx, where.F("id", r.MenuId))
	if err != nil {
		log.W(ctx).Errorw("Failed to delete menu", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrDBWrite.WithMessage(err.Error())
	}

	log.W(ctx).Infow("Menu deleted successfully", "menu_id", r.MenuId, "menu_code", menu.MenuCode)

	return &v1.DeleteMenuResponse{
		Success: true,
		Message: "菜单删除成功",
	}, nil
}

// GetMenu 获取菜单详情
func (b *menuBiz) GetMenu(ctx context.Context, r *v1.GetMenuRequest) (*v1.GetMenuResponse, error) {
	menu, err := b.ds.Menu().Get(ctx, where.F("id", r.MenuId))
	if err != nil {
		log.W(ctx).Errorw("Menu not found", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("菜单ID %d 不存在", r.MenuId))
	}

	// 转换为API格式
	apiMenu := convertMenuToAPI(menu)

	return &v1.GetMenuResponse{
		Menu: apiMenu,
	}, nil
}

// ListMenus 获取菜单列表
func (b *menuBiz) ListMenus(ctx context.Context, r *v1.ListMenusRequest) (*v1.ListMenusResponse, error) {
	opts := where.NewWhere()

	// 租户过滤
	if r.TenantId > 0 {
		opts = opts.F("tenant_id", r.TenantId)
	}

	// 父菜单过滤
	if r.ParentId != nil {
		if *r.ParentId == 0 {
			opts = opts.Q("(parent_id IS NULL OR parent_id = 0)")
		} else {
			opts = opts.F("parent_id", *r.ParentId)
		}
	}

	// 菜单类型过滤
	if r.MenuType != nil {
		opts = opts.F("menu_type", *r.MenuType)
	}

	// 状态过滤
	if r.Status != nil {
		opts = opts.F("status", *r.Status)
	}

	// 可见性过滤
	if r.Visible != nil {
		opts = opts.F("visible", *r.Visible)
	}

	// 分页
	if r.Offset > 0 {
		opts = opts.O(int(r.Offset))
	}
	if r.Limit > 0 {
		opts = opts.L(int(r.Limit))
	} else {
		opts = opts.L(20) // 默认限制
	}

	count, menus, err := b.ds.Menu().List(ctx, opts)
	if err != nil {
		log.W(ctx).Errorw("Failed to list menus", "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 转换为API格式
	var apiMenus []*v1.Menu
	for _, menu := range menus {
		apiMenus = append(apiMenus, convertMenuToAPI(menu))
	}

	return &v1.ListMenusResponse{
		TotalCount: count,
		Menus:      apiMenus,
	}, nil
}

// GetUserMenus 获取用户菜单（权限过滤后的菜单树）
func (b *menuBiz) GetUserMenus(ctx context.Context, r *v1.GetUserMenusRequest) (*v1.GetUserMenusResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated
	}

	// 获取租户信息
	tenantID := r.TenantId
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

	// 构建用户标识符
	userIdentifier := fmt.Sprintf("u%d", userID)

	// 直接使用存储层的GetUserAccessibleMenus方法
	menuWithPermissions, err := b.ds.MenuPermission().GetUserAccessibleMenus(ctx, userIdentifier, tenantID)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user accessible menus", "user_id", userIdentifier, "tenant_id", tenantID, "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to get user accessible menus")
	}

	// 转换为API格式
	var menus []*v1.Menu
	for _, menuWithPerm := range menuWithPermissions {
		if menuWithPerm.Status {
			menus = append(menus, convertMenuToAPI(&menuWithPerm.MenuM))
		}
	}

	// 构建菜单树结构
	menuTree := buildMenuTree(menus)

	return &v1.GetUserMenusResponse{
		Menus: menuTree,
	}, nil
}

// GetMenuTree 获取菜单树
func (b *menuBiz) GetMenuTree(ctx context.Context, r *v1.GetMenuTreeRequest) (*v1.GetMenuTreeResponse, error) {
	opts := where.NewWhere()

	// 租户过滤
	if r.TenantId > 0 {
		opts = opts.F("tenant_id", r.TenantId)
	}

	// 状态过滤
	if r.OnlyActive {
		opts = opts.F("status", true, "visible", true)
	}

	// 菜单类型过滤
	if len(r.MenuTypes) > 0 {
		opts = opts.Q("menu_type IN (?)", r.MenuTypes)
	}

	_, menus, err := b.ds.Menu().List(ctx, opts)
	if err != nil {
		log.W(ctx).Errorw("Failed to get menu tree", "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 转换为API格式
	var apiMenus []*v1.Menu
	for _, menu := range menus {
		apiMenus = append(apiMenus, convertMenuToAPI(menu))
	}

	// 构建菜单树结构
	menuTree := buildMenuTree(apiMenus)

	return &v1.GetMenuTreeResponse{
		Menus: menuTree,
	}, nil
}

// UpdateMenuSort 批量更新菜单排序
func (b *menuBiz) UpdateMenuSort(ctx context.Context, r *v1.UpdateMenuSortRequest) (*v1.UpdateMenuSortResponse, error) {
	var updatedCount int32

	for _, sortItem := range r.SortItems {
		menu, err := b.ds.Menu().Get(ctx, where.F("id", sortItem.MenuId))
		if err != nil {
			log.W(ctx).Errorw("Menu not found for sort update", "menu_id", sortItem.MenuId, "err", err)
			continue
		}

		menu.SortOrder = sortItem.SortOrder
		err = b.ds.Menu().Update(ctx, menu)
		if err != nil {
			log.W(ctx).Errorw("Failed to update menu sort", "menu_id", sortItem.MenuId, "err", err)
			continue
		}

		updatedCount++
	}

	return &v1.UpdateMenuSortResponse{
		Success:      true,
		Message:      fmt.Sprintf("成功更新 %d 个菜单的排序", updatedCount),
		UpdatedCount: updatedCount,
	}, nil
}

// CopyMenu 复制菜单
func (b *menuBiz) CopyMenu(ctx context.Context, r *v1.CopyMenuRequest) (*v1.CopyMenuResponse, error) {
	// 获取源菜单
	sourceMenu, err := b.ds.Menu().Get(ctx, where.F("id", r.SourceMenuId))
	if err != nil {
		log.W(ctx).Errorw("Source menu not found", "source_menu_id", r.SourceMenuId, "err", err)
		return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("源菜单ID %d 不存在", r.SourceMenuId))
	}

	// 生成新的菜单编码
	newMenuCode := r.NewMenuCode
	if newMenuCode == "" {
		newMenuCode = fmt.Sprintf("%s_copy_%d", sourceMenu.MenuCode, time.Now().Unix())
	}

	// 检查新菜单编码是否已存在
	exists, err := b.ds.Menu().IsMenuCodeExists(ctx, newMenuCode, r.TargetTenantId)
	if err != nil {
		log.W(ctx).Errorw("Failed to check new menu code existence", "new_menu_code", newMenuCode, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}
	if exists {
		return nil, errno.ErrMenuCodeExists.WithMessage(fmt.Sprintf("菜单编码 %s 已存在", newMenuCode))
	}

	// 创建新菜单
	newMenu := &model.MenuM{
		TenantID:  r.TargetTenantId,
		MenuCode:  newMenuCode,
		Title:     sourceMenu.Title,
		MenuType:  sourceMenu.MenuType,
		RoutePath: sourceMenu.RoutePath,
		Component: sourceMenu.Component,
		Icon:      sourceMenu.Icon,
		SortOrder: sourceMenu.SortOrder,
		Visible:   sourceMenu.Visible,
		Status:    sourceMenu.Status,
		Remark:    sourceMenu.Remark,
	}

	// 处理父菜单
	if r.TargetParentId != nil {
		newMenu.ParentID = r.TargetParentId
	} else {
		newMenu.ParentID = sourceMenu.ParentID
	}

	// 保存新菜单
	err = b.ds.Menu().Create(ctx, newMenu)
	if err != nil {
		log.W(ctx).Errorw("Failed to copy menu", "source_menu_id", r.SourceMenuId, "err", err)
		return nil, errno.ErrDBWrite.WithMessage(err.Error())
	}

	// 如果需要复制权限
	if r.CopyPermissions {
		sourcePermissions, err := b.ds.MenuPermission().GetMenuPermissions(ctx, r.SourceMenuId)
		if err == nil && len(sourcePermissions) > 0 {
			var permConfigs []model.MenuPermissionConfig
			for _, perm := range sourcePermissions {
				permConfigs = append(permConfigs, model.MenuPermissionConfig{
					PermissionCode: perm.PermissionCode,
					IsRequired:     true, // 简化处理，都设为必需权限
					AutoCreate:     true,
				})
			}
			_ = b.ds.MenuPermission().ConfigureMenuPermissions(ctx, newMenu.ID, permConfigs)
		}
	}

	log.W(ctx).Infow("Menu copied successfully", "source_menu_id", r.SourceMenuId, "new_menu_id", newMenu.ID)

	return &v1.CopyMenuResponse{
		NewMenuId: newMenu.ID,
		Message:   "菜单复制成功",
	}, nil
}

// MoveMenu 移动菜单
func (b *menuBiz) MoveMenu(ctx context.Context, r *v1.MoveMenuRequest) (*v1.MoveMenuResponse, error) {
	// 获取要移动的菜单
	menu, err := b.ds.Menu().Get(ctx, where.F("id", r.MenuId))
	if err != nil {
		log.W(ctx).Errorw("Menu not found", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("菜单ID %d 不存在", r.MenuId))
	}

	// 如果指定了新的父菜单，验证父菜单是否存在且不会形成循环引用
	if r.NewParentId != nil && *r.NewParentId > 0 {
		// 检查父菜单是否存在
		_, err := b.ds.Menu().Get(ctx, where.F("id", *r.NewParentId, "tenant_id", menu.TenantID))
		if err != nil {
			log.W(ctx).Errorw("New parent menu not found", "new_parent_id", *r.NewParentId, "err", err)
			return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("新父菜单ID %d 不存在", *r.NewParentId))
		}

		// 检查是否会形成循环引用
		if err := b.checkMenuHierarchyLoop(ctx, menu.ID, *r.NewParentId); err != nil {
			return nil, err
		}
	}

	// 更新父菜单
	menu.ParentID = r.NewParentId

	// 如果指定了新的排序位置
	if r.NewSortOrder != nil {
		menu.SortOrder = *r.NewSortOrder
	}

	// 保存更新
	err = b.ds.Menu().Update(ctx, menu)
	if err != nil {
		log.W(ctx).Errorw("Failed to move menu", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrDBWrite.WithMessage(err.Error())
	}

	log.W(ctx).Infow("Menu moved successfully", "menu_id", r.MenuId)

	return &v1.MoveMenuResponse{
		Success: true,
		Message: "菜单移动成功",
	}, nil
}

// checkMenuHierarchyLoop 检查菜单层级是否会形成循环引用
func (b *menuBiz) checkMenuHierarchyLoop(ctx context.Context, menuID, newParentID int64) error {
	if menuID == newParentID {
		return errno.ErrMenuHierarchyLoop.WithMessage("菜单不能设置自己为父菜单")
	}

	// 递归检查新父菜单的所有祖先菜单
	currentParentID := newParentID
	for currentParentID > 0 {
		if currentParentID == menuID {
			return errno.ErrMenuHierarchyLoop.WithMessage("菜单层级设置会形成循环引用")
		}

		parentMenu, err := b.ds.Menu().Get(ctx, where.F("id", currentParentID))
		if err != nil {
			break // 父菜单不存在，跳出循环
		}

		if parentMenu.ParentID == nil {
			break // 到达根菜单
		}

		currentParentID = *parentMenu.ParentID
	}

	return nil
}

// convertMenuToAPI 转换菜单模型为API格式
func convertMenuToAPI(menu *model.MenuM) *v1.Menu {
	apiMenu := &v1.Menu{
		Id:        menu.ID,
		TenantId:  menu.TenantID,
		MenuCode:  menu.MenuCode,
		Title:     menu.Title,
		MenuType:  menu.MenuType,
		SortOrder: menu.SortOrder,
		Visible:   menu.Visible,
		Status:    boolToInt32(menu.Status),
		CreatedAt: timestamppb.New(menu.CreatedAt),
		UpdatedAt: timestamppb.New(menu.UpdatedAt),
	}

	// 处理可选字段
	if menu.ParentID != nil {
		apiMenu.ParentId = *menu.ParentID
	}
	if menu.RoutePath != nil {
		apiMenu.RoutePath = *menu.RoutePath
	}
	if menu.Component != nil {
		apiMenu.Component = *menu.Component
	}
	if menu.Icon != nil {
		apiMenu.Icon = *menu.Icon
	}

	return apiMenu
}

// buildMenuTree 构建菜单树结构
func buildMenuTree(menus []*v1.Menu) []*v1.Menu {
	// 创建菜单映射
	menuMap := make(map[int64]*v1.Menu)
	for _, menu := range menus {
		menuMap[menu.Id] = menu
	}

	// 构建树结构
	var rootMenus []*v1.Menu
	for _, menu := range menus {
		if menu.ParentId == 0 {
			// 根菜单
			rootMenus = append(rootMenus, menu)
		} else {
			// 子菜单
			if parent, exists := menuMap[menu.ParentId]; exists {
				if parent.Children == nil {
					parent.Children = make([]*v1.Menu, 0)
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
func sortMenus(menus []*v1.Menu) {
	for i := 0; i < len(menus)-1; i++ {
		for j := i + 1; j < len(menus); j++ {
			if menus[i].SortOrder > menus[j].SortOrder {
				menus[i], menus[j] = menus[j], menus[i]
			}
		}
	}
}

// sortMenuChildren 递归排序菜单的子菜单
func sortMenuChildren(menu *v1.Menu) {
	if menu.Children != nil && len(menu.Children) > 0 {
		sortMenus(menu.Children)
		for _, child := range menu.Children {
			sortMenuChildren(child)
		}
	}
}

// boolToInt32 将布尔值转换为int32
func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
