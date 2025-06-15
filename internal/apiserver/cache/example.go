// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
)

// UserServiceWithCache 展示如何在业务服务中集成缓存.
// 这是一个示例服务，展示了缓存的常见使用模式.
type UserServiceWithCache struct {
	store store.IStore
	cache ICache
}

// NewUserServiceWithCache 创建带缓存的用户服务实例.
func NewUserServiceWithCache(store store.IStore, cache ICache) *UserServiceWithCache {
	return &UserServiceWithCache{
		store: store,
		cache: cache,
	}
}

// GetUserWithCache 获取用户信息，优先从缓存读取.
func (s *UserServiceWithCache) GetUserWithCache(ctx context.Context, userID string) (*model.UserM, error) {
	// 1. 尝试从缓存获取
	user, err := s.cache.User().GetUser(ctx, userID)
	if err == nil {
		// 缓存命中，直接返回
		return user, nil
	}

	// 2. 缓存未命中（或出错），从数据库获取
	if !errors.Is(err, redis.Nil) {
		// 如果不是缓存未命中错误，记录日志
		fmt.Printf("Cache get user error: %v\n", err)
	}

	// 从数据库查询
	user, err = s.store.User().Get(ctx, nil) // 这里需要根据实际的store接口调整
	if err != nil {
		return nil, fmt.Errorf("failed to get user from database: %w", err)
	}

	// 3. 将结果写入缓存（异步或同步都可以）
	if cacheErr := s.cache.User().SetUser(ctx, user, 30*time.Minute); cacheErr != nil {
		// 缓存写入失败不影响主要业务逻辑，只记录日志
		fmt.Printf("Failed to cache user: %v\n", cacheErr)
	}

	return user, nil
}

// LoginWithCache 用户登录示例，展示会话管理.
func (s *UserServiceWithCache) LoginWithCache(ctx context.Context, username, password string) (string, error) {
	// 1. 检查登录尝试次数（防暴力破解）
	attemptKey := fmt.Sprintf("login_attempt:%s", username)
	attempts, err := s.cache.Session().GetLoginAttempt(ctx, attemptKey)
	if err == nil {
		// 如果尝试次数过多，拒绝登录
		var attemptCount int
		if json.Unmarshal([]byte(attempts), &attemptCount) == nil && attemptCount >= 5 {
			return "", errors.New("too many login attempts, please try again later")
		}
	}

	// 2. 验证用户凭据（这里简化处理）
	// user, err := s.validateCredentials(ctx, username, password)
	// if err != nil {
	//     // 登录失败，增加尝试次数
	//     s.incrementLoginAttempts(ctx, username)
	//     return "", fmt.Errorf("invalid credentials: %w", err)
	// }

	// 3. 登录成功，生成会话token
	token := fmt.Sprintf("token_%s_%d", username, time.Now().Unix())

	// 4. 将token存储到缓存
	if err := s.cache.User().SetUserSession(ctx, username, token, 2*time.Hour); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// 5. 清除登录尝试记录
	_ = s.cache.Session().DelLoginAttempt(ctx, attemptKey)

	return token, nil
}

// LogoutWithCache 用户退出登录.
func (s *UserServiceWithCache) LogoutWithCache(ctx context.Context, token string) error {
	// 删除会话token
	return s.cache.User().DelUserSession(ctx, token)
}

// GetUserByTokenWithCache 根据token获取用户信息.
func (s *UserServiceWithCache) GetUserByTokenWithCache(ctx context.Context, token string) (*model.UserM, error) {
	// 1. 从缓存获取用户ID
	userID, err := s.cache.User().GetUserByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// 2. 根据用户ID获取用户信息（优先缓存）
	return s.GetUserWithCache(ctx, userID)
}

// RefreshTokenWithCache 刷新token有效期.
func (s *UserServiceWithCache) RefreshTokenWithCache(ctx context.Context, token string) error {
	// 1. 验证token是否存在
	userID, err := s.cache.User().GetUserByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// 2. 延长token有效期
	return s.cache.User().SetUserSession(ctx, userID, token, 2*time.Hour)
}

// ClearUserCache 清除用户相关的所有缓存.
func (s *UserServiceWithCache) ClearUserCache(ctx context.Context, userID string) error {
	// 可以清除用户信息缓存
	return s.cache.User().DelUser(ctx, userID)
}

// BatchCacheUsers 批量缓存用户信息.
func (s *UserServiceWithCache) BatchCacheUsers(ctx context.Context, users []*model.UserM) error {
	for _, user := range users {
		if err := s.cache.User().SetUser(ctx, user, 30*time.Minute); err != nil {
			// 记录错误，但继续处理其他用户
			fmt.Printf("Failed to cache user %d: %v\n", user.ID, err)
		}
	}
	return nil
}
