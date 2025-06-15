// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package cache

//go:generate mockgen -destination mock_cache.go -package cache github.com/ashwinyue/one-auth/internal/apiserver/cache ICache,UserCache,SessionCache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet 是一个 Wire 的 Provider 集合，用于声明依赖注入的规则.
var ProviderSet = wire.NewSet(
	NewCache,
	NewSessionManager,
	NewLoginSecurityManager,
	wire.Bind(new(ICache), new(*dataCache)),
)

var (
	once sync.Once
	// C 全局变量，方便其它包直接调用已初始化好的 cache 实例.
	C *dataCache
)

// ICache 定义了缓存层需要实现的方法.
type ICache interface {
	// 基础缓存操作
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// 业务缓存接口
	User() UserCache
	Session() SessionCache
}

// dataCache 是 ICache 的具体实现.
type dataCache struct {
	client *redis.Client
}

// 确保 dataCache 实现了 ICache 接口.
var _ ICache = (*dataCache)(nil)

// NewCache 创建一个 ICache 类型的实例.
func NewCache(client *redis.Client) *dataCache {
	// 确保 C 只被初始化一次
	once.Do(func() {
		C = &dataCache{client: client}
	})

	return C
}

// Set 设置缓存值.
func (c *dataCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var data string
	switch v := value.(type) {
	case string:
		data = v
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data = string(bytes)
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存值.
func (c *dataCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del 删除缓存.
func (c *dataCache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查key是否存在.
func (c *dataCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Expire 设置key的过期时间.
func (c *dataCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// User 返回一个实现了 UserCache 接口的实例.
func (c *dataCache) User() UserCache {
	return newUserCache(c)
}

// Session 返回一个实现了 SessionCache 接口的实例.
func (c *dataCache) Session() SessionCache {
	return newSessionCache(c)
}
