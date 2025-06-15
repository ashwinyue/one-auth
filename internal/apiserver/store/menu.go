// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package store

import (
	"context"

	genericstore "github.com/ashwinyue/one-auth/pkg/store"
	"github.com/ashwinyue/one-auth/pkg/store/where"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
)

// MenuStore 定义了 menu 模块在 store 层所实现的方法.
type MenuStore interface {
	Create(ctx context.Context, obj *model.MenuM) error
	Update(ctx context.Context, obj *model.MenuM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.MenuM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.MenuM, error)

	MenuExpansion
}

// MenuExpansion 定义了菜单操作的附加方法.
type MenuExpansion interface {
	// 按租户获取菜单列表
	GetMenusByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.MenuM, error)

	// 检查菜单代码是否已存在（在指定租户内）
	IsMenuCodeExists(ctx context.Context, menuCode string, tenantID int64) (bool, error)

	// 根据菜单代码获取菜单（在指定租户内）
	GetByMenuCode(ctx context.Context, menuCode string, tenantID int64) (*model.MenuM, error)

	// 获取根菜单列表（无父菜单的菜单）
	GetRootMenus(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.MenuM, error)

	// 获取指定父菜单的子菜单
	GetChildMenus(ctx context.Context, parentID int64, opts *where.Options) ([]*model.MenuM, error)

	// 获取活跃菜单列表（状态为启用且可见的菜单）
	GetActiveMenus(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.MenuM, error)

	// 批量获取菜单
	GetMenusByIDs(ctx context.Context, menuIDs []int64) ([]*model.MenuM, error)

	// 按API路径获取菜单
	GetMenusByAPIPath(ctx context.Context, apiPath string) ([]*model.MenuM, error)

	// 获取指定菜单的所有子菜单（递归查询）
	GetAllChildMenus(ctx context.Context, parentID int64) ([]*model.MenuM, error)

	// 按菜单类型获取菜单（如：目录、菜单、按钮等）
	GetMenusByType(ctx context.Context, tenantID int64, menuType bool, opts *where.Options) ([]*model.MenuM, error)
}

// menuStore 是 MenuStore 接口的实现.
type menuStore struct {
	*genericstore.Store[model.MenuM]
}

// 确保 menuStore 实现了 MenuStore 接口.
var _ MenuStore = (*menuStore)(nil)

// newMenuStore 创建 menuStore 的实例.
func newMenuStore(store *datastore) *menuStore {
	return &menuStore{
		Store: genericstore.NewStore[model.MenuM](store, NewLogger()),
	}
}

// GetMenusByTenant 按租户获取菜单列表
func (s *menuStore) GetMenusByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.MenuM, error) {
	// 使用通用的List方法，添加租户条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID)
	return s.List(ctx, opts)
}

// IsMenuCodeExists 检查菜单代码是否已存在
func (s *menuStore) IsMenuCodeExists(ctx context.Context, menuCode string, tenantID int64) (bool, error) {
	// 使用通用的Get方法检查是否存在
	_, err := s.Get(ctx, where.F("menu_code", menuCode, "tenant_id", tenantID))
	if err != nil {
		// 如果是记录不存在错误，返回false
		return false, nil
	}
	return true, nil
}

// GetByMenuCode 根据菜单代码获取菜单
func (s *menuStore) GetByMenuCode(ctx context.Context, menuCode string, tenantID int64) (*model.MenuM, error) {
	// 使用通用的Get方法
	return s.Get(ctx, where.F("menu_code", menuCode, "tenant_id", tenantID))
}

// GetRootMenus 获取根菜单列表
func (s *menuStore) GetRootMenus(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.MenuM, error) {
	// 使用通用的List方法，查询parent_id为NULL或0的记录
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID).Q("(parent_id IS NULL OR parent_id = 0)")
	return s.List(ctx, opts)
}

// GetChildMenus 获取指定父菜单的子菜单
func (s *menuStore) GetChildMenus(ctx context.Context, parentID int64, opts *where.Options) ([]*model.MenuM, error) {
	// 使用通用的List方法
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("parent_id", parentID)
	_, menus, err := s.List(ctx, opts)
	return menus, err
}

// GetActiveMenus 获取活跃菜单列表
func (s *menuStore) GetActiveMenus(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.MenuM, error) {
	// 使用通用的List方法，添加租户、状态和可见性条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID, "status", true, "visible", true)
	return s.List(ctx, opts)
}

// GetMenusByIDs 批量获取菜单
func (s *menuStore) GetMenusByIDs(ctx context.Context, menuIDs []int64) ([]*model.MenuM, error) {
	// 使用通用的List方法，通过IN查询
	if len(menuIDs) == 0 {
		return []*model.MenuM{}, nil
	}

	// 转换为interface{}切片
	ids := make([]interface{}, len(menuIDs))
	for i, id := range menuIDs {
		ids[i] = id
	}

	_, menus, err := s.List(ctx, where.NewWhere().Q("id IN (?)", ids))
	return menus, err
}

// GetMenusByAPIPath 按API路径获取菜单
func (s *menuStore) GetMenusByAPIPath(ctx context.Context, apiPath string) ([]*model.MenuM, error) {
	// 使用通用的List方法
	_, menus, err := s.List(ctx, where.F("api_path", apiPath))
	return menus, err
}

// GetAllChildMenus 获取指定菜单的所有子菜单（递归）
func (s *menuStore) GetAllChildMenus(ctx context.Context, parentID int64) ([]*model.MenuM, error) {
	// 这里使用简单的实现，实际可能需要递归查询
	// 使用通用的List方法查询直接子菜单
	_, menus, err := s.List(ctx, where.F("parent_id", parentID))
	if err != nil {
		return nil, err
	}

	// 递归查询每个子菜单的子菜单
	var allMenus []*model.MenuM
	allMenus = append(allMenus, menus...)

	for _, menu := range menus {
		childMenus, err := s.GetAllChildMenus(ctx, menu.ID)
		if err == nil {
			allMenus = append(allMenus, childMenus...)
		}
	}

	return allMenus, nil
}

// GetMenusByType 按菜单类型获取菜单
func (s *menuStore) GetMenusByType(ctx context.Context, tenantID int64, menuType bool, opts *where.Options) ([]*model.MenuM, error) {
	// 使用通用的List方法，添加租户和类型条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID, "menu_type", menuType)
	_, menus, err := s.List(ctx, opts)
	return menus, err
}
