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

// UserStatusStore 定义了用户状态存储层方法
type UserStatusStore interface {
	Create(ctx context.Context, obj *model.UserStatusM) error
	Update(ctx context.Context, obj *model.UserStatusM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.UserStatusM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.UserStatusM, error)
}

// userStatusStore 是 UserStatusStore 接口的实现
type userStatusStore struct {
	*genericstore.Store[model.UserStatusM]
}

// 确保 userStatusStore 实现了 UserStatusStore 接口
var _ UserStatusStore = (*userStatusStore)(nil)

// newUserStatusStore 创建 userStatusStore 的实例
func newUserStatusStore(store *datastore) *userStatusStore {
	return &userStatusStore{
		Store: genericstore.NewStore[model.UserStatusM](store, NewLogger()),
	}
}
