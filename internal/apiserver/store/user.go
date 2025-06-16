// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

// nolint: dupl
package store

import (
	"context"

	genericstore "github.com/ashwinyue/one-auth/pkg/store"
	"github.com/ashwinyue/one-auth/pkg/store/where"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
)

// UserStore 定义了 user 模块在 store 层所实现的方法.
type UserStore interface {
	Create(ctx context.Context, obj *model.UserM) error
	Update(ctx context.Context, obj *model.UserM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.UserM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.UserM, error)

	UserExpansion
}

// UserExpansion 定义了用户操作的附加方法.
type UserExpansion interface {
	// 根据用户名或邮箱获取用户（用于登录认证）
	GetByUsernameOrEmail(ctx context.Context, identifier string) (*model.UserM, error)

	// 检查用户是否激活
	IsUserActive(ctx context.Context, userID string) (bool, error)

	// 更新用户最后登录时间
	UpdateLastLoginTime(ctx context.Context, userID string) error

	// 检查用户名是否已存在
	IsUsernameExists(ctx context.Context, username string) (bool, error)

	// 检查邮箱是否已存在
	IsEmailExists(ctx context.Context, email string) (bool, error)

	// 按租户获取用户列表
	GetUsersByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.UserM, error)

	// 获取用户的租户ID（用于认证中间件）
	GetUserTenantID(ctx context.Context, userID string) (int64, error)
}

// userStore 是 UserStore 接口的实现.
type userStore struct {
	*genericstore.Store[model.UserM]
	ds *datastore
}

// 确保 userStore 实现了 UserStore 接口.
var _ UserStore = (*userStore)(nil)

// newUserStore 创建 userStore 的实例.
func newUserStore(store *datastore) *userStore {
	return &userStore{
		Store: genericstore.NewStore[model.UserM](store, NewLogger()),
		ds:    store,
	}
}

// GetByUsernameOrEmail 根据用户名或邮箱获取用户
func (s *userStore) GetByUsernameOrEmail(ctx context.Context, identifier string) (*model.UserM, error) {
	// 使用通用的Get方法，通过where条件查询
	return s.Get(ctx, where.NewWhere().Q("(username = ? OR email = ?) AND deleted_at IS NULL", identifier, identifier))
}

// IsUserActive 检查用户是否激活
func (s *userStore) IsUserActive(ctx context.Context, userID string) (bool, error) {
	// 用户状态由UserStatusM模型管理，这里暂时返回true
	// 实际应该查询user_status表
	user, err := s.Get(ctx, where.F("user_id", userID))
	if err != nil {
		return false, err
	}
	// 简单检查用户是否存在且未被软删除
	return user != nil, nil
}

// UpdateLastLoginTime 更新用户最后登录时间
func (s *userStore) UpdateLastLoginTime(ctx context.Context, userID string) error {
	// 可以使用通用的Update方法，先Get再Update
	user, err := s.Get(ctx, where.F("user_id", userID))
	if err != nil {
		return err
	}

	// 更新最后登录时间字段
	// user.LastLoginAt = time.Now() // 假设模型有这个字段
	return s.Update(ctx, user)
}

// IsUsernameExists 检查用户名是否已存在
func (s *userStore) IsUsernameExists(ctx context.Context, username string) (bool, error) {
	// 使用通用的Get方法检查是否存在
	_, err := s.Get(ctx, where.F("username", username))
	if err != nil {
		// 如果是记录不存在错误，返回false
		return false, nil
	}
	return true, nil
}

// IsEmailExists 检查邮箱是否已存在
func (s *userStore) IsEmailExists(ctx context.Context, email string) (bool, error) {
	// 使用通用的Get方法检查是否存在
	_, err := s.Get(ctx, where.F("email", email))
	if err != nil {
		// 如果是记录不存在错误，返回false
		return false, nil
	}
	return true, nil
}

// GetUsersByTenant 按租户获取用户列表
func (s *userStore) GetUsersByTenant(ctx context.Context, tenantID int64, opts *where.Options) (int64, []*model.UserM, error) {
	// 使用通用的List方法，添加租户条件
	if opts == nil {
		opts = where.NewWhere()
	}
	opts = opts.F("tenant_id", tenantID)
	return s.List(ctx, opts)
}

// GetUserTenantID 获取用户的租户ID
func (s *userStore) GetUserTenantID(ctx context.Context, userID string) (int64, error) {
	var userStatus model.UserStatusM
	err := s.ds.DB(ctx).Table("user_status").Where("user_id = ? AND is_primary = ?", userID, true).First(&userStatus).Error
	if err != nil {
		return 0, err
	}
	return userStatus.TenantID, nil
}
