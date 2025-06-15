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

// PermissionStore 定义了 permission 模块在 store 层所实现的方法.
type PermissionStore interface {
	Create(ctx context.Context, obj *model.PermissionM) error
	Update(ctx context.Context, obj *model.PermissionM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.PermissionM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.PermissionM, error)

	PermissionExpansion
}

// PermissionExpansion 定义了权限操作的附加方法.
type PermissionExpansion interface {
	// 按租户获取权限列表
	GetPermissionsByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.PermissionM, error)

	// 检查权限代码是否已存在（在指定租户内）
	IsPermissionCodeExists(ctx context.Context, permissionCode string, tenantID int64) (bool, error)

	// 根据权限代码获取权限（在指定租户内）
	GetByPermissionCode(ctx context.Context, permissionCode string, tenantID int64) (*model.PermissionM, error)

	// 按菜单ID获取权限列表
	GetPermissionsByMenuID(ctx context.Context, menuID int64) ([]*model.PermissionM, error)

	// 按API路径和HTTP方法获取权限
	GetPermissionsByAPI(ctx context.Context, apiPath, httpMethod string) ([]*model.PermissionM, error)

	// 获取活跃权限列表（状态为启用的权限）
	GetActivePermissions(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.PermissionM, error)

	// 批量获取权限
	GetPermissionsByIDs(ctx context.Context, permissionIDs []int64) ([]*model.PermissionM, error)

	// 获取指定菜单的所有权限（包括子菜单权限）
	GetAllPermissionsByMenuIDs(ctx context.Context, menuIDs []int64) ([]*model.PermissionM, error)
}

// permissionStore 是 PermissionStore 接口的实现.
type permissionStore struct {
	*genericstore.Store[model.PermissionM]
}

// 确保 permissionStore 实现了 PermissionStore 接口.
var _ PermissionStore = (*permissionStore)(nil)

// newPermissionStore 创建 permissionStore 的实例.
func newPermissionStore(store *datastore) *permissionStore {
	return &permissionStore{
		Store: genericstore.NewStore[model.PermissionM](store, NewLogger()),
	}
}

// GetPermissionsByTenant 按租户获取权限列表
func (s *permissionStore) GetPermissionsByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.PermissionM, error) {
	// 使用通用的List方法，添加租户条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID)
	return s.List(ctx, opts)
}

// IsPermissionCodeExists 检查权限代码是否已存在
func (s *permissionStore) IsPermissionCodeExists(ctx context.Context, permissionCode string, tenantID int64) (bool, error) {
	// 使用通用的Get方法检查是否存在
	_, err := s.Get(ctx, where.F("permission_code", permissionCode, "tenant_id", tenantID))
	if err != nil {
		// 如果是记录不存在错误，返回false
		return false, nil
	}
	return true, nil
}

// GetByPermissionCode 根据权限代码获取权限
func (s *permissionStore) GetByPermissionCode(ctx context.Context, permissionCode string, tenantID int64) (*model.PermissionM, error) {
	// 使用通用的Get方法
	return s.Get(ctx, where.F("permission_code", permissionCode, "tenant_id", tenantID))
}

// GetPermissionsByMenuID 按菜单ID获取权限列表
func (s *permissionStore) GetPermissionsByMenuID(ctx context.Context, menuID int64) ([]*model.PermissionM, error) {
	// 使用通用的List方法
	_, permissions, err := s.List(ctx, where.F("menu_id", menuID))
	return permissions, err
}

// GetPermissionsByAPI 按API路径和HTTP方法获取权限
func (s *permissionStore) GetPermissionsByAPI(ctx context.Context, apiPath, httpMethod string) ([]*model.PermissionM, error) {
	// 使用通用的List方法，查询api_path和http_methods字段
	_, permissions, err := s.List(ctx, where.F("api_path", apiPath, "http_methods", httpMethod))
	return permissions, err
}

// GetActivePermissions 获取活跃权限列表
func (s *permissionStore) GetActivePermissions(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.PermissionM, error) {
	// 使用通用的List方法，添加租户和状态条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID, "status", true)
	return s.List(ctx, opts)
}

// GetPermissionsByIDs 批量获取权限
func (s *permissionStore) GetPermissionsByIDs(ctx context.Context, permissionIDs []int64) ([]*model.PermissionM, error) {
	// 使用通用的List方法，通过IN查询
	if len(permissionIDs) == 0 {
		return []*model.PermissionM{}, nil
	}

	// 转换为interface{}切片
	ids := make([]interface{}, len(permissionIDs))
	for i, id := range permissionIDs {
		ids[i] = id
	}

	_, permissions, err := s.List(ctx, where.NewWhere().Q("id IN (?)", ids))
	return permissions, err
}

// GetAllPermissionsByMenuIDs 获取指定菜单的所有权限
func (s *permissionStore) GetAllPermissionsByMenuIDs(ctx context.Context, menuIDs []int64) ([]*model.PermissionM, error) {
	// 使用通用的List方法，通过IN查询
	if len(menuIDs) == 0 {
		return []*model.PermissionM{}, nil
	}

	// 转换为interface{}切片
	ids := make([]interface{}, len(menuIDs))
	for i, id := range menuIDs {
		ids[i] = id
	}

	_, permissions, err := s.List(ctx, where.NewWhere().Q("menu_id IN (?)", ids))
	return permissions, err
}
