// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package store

import (
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"

	genericstore "github.com/ashwinyue/one-auth/pkg/store"
	"github.com/ashwinyue/one-auth/pkg/store/where"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// TenantStore 定义了 tenant 模块在 store 层所实现的方法.
type TenantStore interface {
	Create(ctx context.Context, obj *model.TenantM) error
	Update(ctx context.Context, obj *model.TenantM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.TenantM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.TenantM, error)

	TenantExpansion
}

// TenantExpansion 定义了租户操作的附加方法.
type TenantExpansion interface {
	// 获取用户所属的租户列表
	GetUserTenants(ctx context.Context, userID string) ([]*model.TenantM, error)

	// 检查用户是否属于指定租户
	CheckUserTenant(ctx context.Context, userID string, tenantID int64) (bool, error)

	// 为用户添加租户关联
	AddUserTenant(ctx context.Context, userID string, tenantID int64) error

	// 移除用户租户关联
	RemoveUserTenant(ctx context.Context, userID string, tenantID int64) error
}

// tenantStore 是 TenantStore 接口的实现.
type tenantStore struct {
	*genericstore.Store[model.TenantM]
	store           *datastore
	userTenantStore *genericstore.Store[model.UserTenantM]
}

// 确保 tenantStore 实现了 TenantStore 接口.
var _ TenantStore = (*tenantStore)(nil)

// newTenantStore 创建 tenantStore 的实例.
func newTenantStore(store *datastore) *tenantStore {
	return &tenantStore{
		Store:           genericstore.NewStore[model.TenantM](store, NewLogger()),
		store:           store,
		userTenantStore: genericstore.NewStore[model.UserTenantM](store, NewLogger()),
	}
}

// GetUserTenants 获取用户所属的租户列表
func (s *tenantStore) GetUserTenants(ctx context.Context, userID string) ([]*model.TenantM, error) {
	// 解析用户ID为数字
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid user ID")
	}

	// 首先获取用户的租户关联
	_, userTenants, err := s.userTenantStore.List(ctx, where.F("user_id", uid, "status", true))
	if err != nil {
		log.Errorw("Failed to get user tenant relations", "user_id", userID, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	if len(userTenants) == 0 {
		return []*model.TenantM{}, nil
	}

	// 提取租户ID列表
	tenantIDs := make([]interface{}, len(userTenants))
	for i, ut := range userTenants {
		tenantIDs[i] = ut.TenantID
	}

	// 获取租户详情
	_, tenants, err := s.List(ctx, where.NewWhere().Q("id IN (?)", tenantIDs))
	if err != nil {
		log.Errorw("Failed to get tenants", "tenant_ids", tenantIDs, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	return tenants, nil
}

// CheckUserTenant 检查用户是否属于指定租户
func (s *tenantStore) CheckUserTenant(ctx context.Context, userID string, tenantID int64) (bool, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return false, errno.ErrInvalidArgument.WithMessage("invalid user ID")
	}

	// 使用通用store的Get方法检查关联是否存在
	_, err = s.userTenantStore.Get(ctx, where.F("user_id", uid, "tenant_id", tenantID, "status", true))

	if err != nil {
		// 如果是记录不存在，返回false
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		log.Errorw("Failed to check user tenant", "user_id", userID, "tenant_id", tenantID, "err", err)
		return false, errno.ErrDBRead.WithMessage(err.Error())
	}

	return true, nil
}

// AddUserTenant 为用户添加租户关联
func (s *tenantStore) AddUserTenant(ctx context.Context, userID string, tenantID int64) error {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return errno.ErrInvalidArgument.WithMessage("invalid user ID")
	}

	userTenant := &model.UserTenantM{
		UserID:   uid,
		TenantID: tenantID,
		Status:   true,
	}

	if err := s.userTenantStore.Create(ctx, userTenant); err != nil {
		log.Errorw("Failed to add user tenant", "user_id", userID, "tenant_id", tenantID, "err", err)
		return errno.ErrDBWrite.WithMessage(err.Error())
	}

	return nil
}

// RemoveUserTenant 移除用户租户关联
func (s *tenantStore) RemoveUserTenant(ctx context.Context, userID string, tenantID int64) error {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return errno.ErrInvalidArgument.WithMessage("invalid user ID")
	}

	err = s.userTenantStore.Delete(ctx, where.F("user_id", uid, "tenant_id", tenantID))

	if err != nil {
		log.Errorw("Failed to remove user tenant", "user_id", userID, "tenant_id", tenantID, "err", err)
		return errno.ErrDBWrite.WithMessage(err.Error())
	}

	return nil
}
