// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
)

// UserCache 定义用户缓存操作接口.
type UserCache interface {
	// SetUser 缓存用户信息
	SetUser(ctx context.Context, user *model.UserM, expiration time.Duration) error
	// GetUser 从缓存获取用户信息
	GetUser(ctx context.Context, userID string) (*model.UserM, error)
	// DelUser 删除用户缓存
	DelUser(ctx context.Context, userID string) error
	// SetUserSession 缓存用户会话
	SetUserSession(ctx context.Context, userID string, token string, expiration time.Duration) error
	// GetUserByToken 根据token获取用户ID
	GetUserByToken(ctx context.Context, token string) (string, error)
	// DelUserSession 删除用户会话
	DelUserSession(ctx context.Context, token string) error
}

// userCache 用户缓存实现.
type userCache struct {
	cache *dataCache
}

// 确保 userCache 实现了 UserCache 接口.
var _ UserCache = (*userCache)(nil)

// newUserCache 创建用户缓存实例.
func newUserCache(cache *dataCache) *userCache {
	return &userCache{cache: cache}
}

// userKey 生成用户缓存key.
func (u *userCache) userKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

// tokenKey 生成token缓存key.
func (u *userCache) tokenKey(token string) string {
	return fmt.Sprintf("token:%s", token)
}

// SetUser 缓存用户信息.
func (u *userCache) SetUser(ctx context.Context, user *model.UserM, expiration time.Duration) error {
	key := u.userKey(fmt.Sprintf("%d", user.ID))
	return u.cache.Set(ctx, key, user, expiration)
}

// GetUser 从缓存获取用户信息.
func (u *userCache) GetUser(ctx context.Context, userID string) (*model.UserM, error) {
	key := u.userKey(userID)
	data, err := u.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var user model.UserM
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// DelUser 删除用户缓存.
func (u *userCache) DelUser(ctx context.Context, userID string) error {
	key := u.userKey(userID)
	return u.cache.Del(ctx, key)
}

// SetUserSession 缓存用户会话.
func (u *userCache) SetUserSession(ctx context.Context, userID string, token string, expiration time.Duration) error {
	key := u.tokenKey(token)
	return u.cache.Set(ctx, key, userID, expiration)
}

// GetUserByToken 根据token获取用户ID.
func (u *userCache) GetUserByToken(ctx context.Context, token string) (string, error) {
	key := u.tokenKey(token)
	return u.cache.Get(ctx, key)
}

// DelUserSession 删除用户会话.
func (u *userCache) DelUserSession(ctx context.Context, token string) error {
	key := u.tokenKey(token)
	return u.cache.Del(ctx, key)
}
