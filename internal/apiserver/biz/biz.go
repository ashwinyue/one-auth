// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package biz

//go:generate mockgen -destination mock_biz.go -package biz github.com/ashwinyue/one-auth/internal/apiserver/biz IBiz

import (
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/google/wire"

	menuv1 "github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/menu"
	permissionv1 "github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/permission"
	postv1 "github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/post"
	rolev1 "github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/role"
	tenantv1 "github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/tenant"
	userv1 "github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/user"
	"github.com/ashwinyue/one-auth/internal/apiserver/cache"

	// Post V2 版本（未实现，仅展示用）
	// postv2 "github.com/ashwinyue/one-auth/internal/apiserver/biz/v2/post".
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
)

// ProviderSet 是一个 Wire 的 Provider 集合，用于声明依赖注入的规则.
// 包含 NewBiz 构造函数，用于生成 biz 实例.
// wire.Bind 用于将接口 IBiz 与具体实现 *biz 绑定，
// 这样依赖 IBiz 的地方会自动注入 *biz 实例.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz 定义了业务层需要实现的方法.
type IBiz interface {
	// UserV1 获取用户业务接口.
	UserV1() userv1.UserBiz
	// PostV1 获取帖子业务接口.
	PostV1() postv1.PostBiz

	// RBAC相关业务接口
	TenantV1() tenantv1.TenantBiz
	RoleV1() rolev1.RoleBiz
	PermissionV1() permissionv1.PermissionBiz
	MenuV1() menuv1.MenuBiz

	// PostV2 获取帖子业务接口（V2 版本）.
	// PostV2() post.PostBiz
}

// biz 是 IBiz 的一个具体实现.
type biz struct {
	store store.IStore
	authz *authz.Authz
	cache cache.ICache
}

// 确保 biz 实现了 IBiz 接口.
var _ IBiz = (*biz)(nil)

// NewBiz 创建一个 IBiz 类型的实例.
func NewBiz(store store.IStore, authz *authz.Authz, cache cache.ICache) *biz {
	return &biz{store: store, authz: authz, cache: cache}
}

// UserV1 返回一个实现了 UserBiz 接口的实例.
func (b *biz) UserV1() userv1.UserBiz {
	sessionManager := cache.NewSessionManager(b.cache)
	loginSecurity := cache.NewLoginSecurityManager(b.cache)
	return userv1.New(b.store, b.authz, sessionManager, loginSecurity)
}

// PostV1 返回一个实现了 PostBiz 接口的实例.
func (b *biz) PostV1() postv1.PostBiz {
	return postv1.New(b.store)
}

// TenantV1 返回一个实现了 TenantBiz 接口的实例.
func (b *biz) TenantV1() tenantv1.TenantBiz {
	return tenantv1.New(b.store, b.authz)
}

// RoleV1 返回一个实现了 RoleBiz 接口的实例.
func (b *biz) RoleV1() rolev1.RoleBiz {
	return rolev1.New(b.store, b.authz)
}

// PermissionV1 返回一个实现了 PermissionBiz 接口的实例.
func (b *biz) PermissionV1() permissionv1.PermissionBiz {
	return permissionv1.New(b.store, b.authz)
}

// MenuV1 返回一个实现了 MenuBiz 接口的实例.
func (b *biz) MenuV1() menuv1.MenuBiz {
	return menuv1.New(b.store, b.authz)
}
