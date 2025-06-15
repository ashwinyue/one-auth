// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package cache

import (
	"context"
	"fmt"
	"time"
)

// SessionCache 定义会话缓存操作接口.
type SessionCache interface {
	// SetSession 设置会话数据
	SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error
	// GetSession 获取会话数据
	GetSession(ctx context.Context, sessionID string) (string, error)
	// DelSession 删除会话数据
	DelSession(ctx context.Context, sessionID string) error
	// RefreshSession 刷新会话过期时间
	RefreshSession(ctx context.Context, sessionID string, expiration time.Duration) error
	// SetLoginAttempt 设置登录尝试次数（用于防暴力破解）
	SetLoginAttempt(ctx context.Context, userID string, attempts int, expiration time.Duration) error
	// GetLoginAttempt 获取登录尝试次数
	GetLoginAttempt(ctx context.Context, userID string) (string, error)
	// DelLoginAttempt 清除登录尝试记录
	DelLoginAttempt(ctx context.Context, userID string) error
}

// sessionCache 会话缓存实现.
type sessionCache struct {
	cache *dataCache
}

// 确保 sessionCache 实现了 SessionCache 接口.
var _ SessionCache = (*sessionCache)(nil)

// newSessionCache 创建会话缓存实例.
func newSessionCache(cache *dataCache) *sessionCache {
	return &sessionCache{cache: cache}
}

// sessionKey 生成会话缓存key.
func (s *sessionCache) sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// loginAttemptKey 生成登录尝试缓存key.
func (s *sessionCache) loginAttemptKey(userID string) string {
	return fmt.Sprintf("login_attempt:%s", userID)
}

// SetSession 设置会话数据.
func (s *sessionCache) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	key := s.sessionKey(sessionID)
	return s.cache.Set(ctx, key, data, expiration)
}

// GetSession 获取会话数据.
func (s *sessionCache) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := s.sessionKey(sessionID)
	return s.cache.Get(ctx, key)
}

// DelSession 删除会话数据.
func (s *sessionCache) DelSession(ctx context.Context, sessionID string) error {
	key := s.sessionKey(sessionID)
	return s.cache.Del(ctx, key)
}

// RefreshSession 刷新会话过期时间.
func (s *sessionCache) RefreshSession(ctx context.Context, sessionID string, expiration time.Duration) error {
	key := s.sessionKey(sessionID)
	return s.cache.Expire(ctx, key, expiration)
}

// SetLoginAttempt 设置登录尝试次数.
func (s *sessionCache) SetLoginAttempt(ctx context.Context, userID string, attempts int, expiration time.Duration) error {
	key := s.loginAttemptKey(userID)
	return s.cache.Set(ctx, key, attempts, expiration)
}

// GetLoginAttempt 获取登录尝试次数.
func (s *sessionCache) GetLoginAttempt(ctx context.Context, userID string) (string, error) {
	key := s.loginAttemptKey(userID)
	return s.cache.Get(ctx, key)
}

// DelLoginAttempt 清除登录尝试记录.
func (s *sessionCache) DelLoginAttempt(ctx context.Context, userID string) error {
	key := s.loginAttemptKey(userID)
	return s.cache.Del(ctx, key)
}
