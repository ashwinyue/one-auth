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

// RoleStore 定义了 role 模块在 store 层所实现的方法.
type RoleStore interface {
	Create(ctx context.Context, obj *model.RoleM) error
	Update(ctx context.Context, obj *model.RoleM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.RoleM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.RoleM, error)

	RoleExpansion
}

// RoleExpansion 定义了角色操作的附加方法.
type RoleExpansion interface {
	// 按租户获取角色列表
	GetRolesByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.RoleM, error)

	// 检查角色名称是否已存在（在指定租户内）
	CheckNameExists(ctx context.Context, name string, tenantID int64) (bool, error)

	// 获取活跃角色列表（状态为启用的角色）
	GetActiveRoles(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.RoleM, error)

	// 批量获取角色
	GetRolesByIDs(ctx context.Context, roleIDs []int64) ([]*model.RoleM, error)
}

// roleStore 是 RoleStore 接口的实现.
type roleStore struct {
	*genericstore.Store[model.RoleM]
}

// 确保 roleStore 实现了 RoleStore 接口.
var _ RoleStore = (*roleStore)(nil)

// newRoleStore 创建 roleStore 的实例.
func newRoleStore(store *datastore) *roleStore {
	return &roleStore{
		Store: genericstore.NewStore[model.RoleM](store, NewLogger()),
	}
}

// GetRolesByTenant 按租户获取角色列表
func (s *roleStore) GetRolesByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.RoleM, error) {
	// 使用通用的List方法，添加租户条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID)
	return s.List(ctx, opts)
}

// CheckNameExists 检查角色名称是否已存在
func (s *roleStore) CheckNameExists(ctx context.Context, name string, tenantID int64) (bool, error) {
	// 使用通用的Get方法检查是否存在
	_, err := s.Get(ctx, where.F("name", name, "tenant_id", tenantID))
	if err != nil {
		// 如果是记录不存在错误，返回false
		return false, nil
	}
	return true, nil
}

// GetActiveRoles 获取活跃角色列表
func (s *roleStore) GetActiveRoles(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.RoleM, error) {
	// 使用通用的List方法，添加租户和状态条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID, "status", true)
	return s.List(ctx, opts)
}

// GetRolesByIDs 批量获取角色
func (s *roleStore) GetRolesByIDs(ctx context.Context, roleIDs []int64) ([]*model.RoleM, error) {
	// 使用通用的List方法，通过IN查询
	if len(roleIDs) == 0 {
		return []*model.RoleM{}, nil
	}

	// 转换为interface{}切片
	ids := make([]interface{}, len(roleIDs))
	for i, id := range roleIDs {
		ids[i] = id
	}

	_, roles, err := s.List(ctx, where.NewWhere().Q("id IN (?)", ids))
	return roles, err
}
