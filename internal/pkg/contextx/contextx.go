// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package contextx

import (
	"context"
)

// 定义用于上下文的键.
type (
	// usernameKey 定义用户名的上下文键.
	usernameKey struct{}
	// userIDKey 定义用户 ID 的上下文键.
	userIDKey struct{}
	// accessTokenKey 定义访问令牌的上下文键.
	accessTokenKey struct{}
	// requestIDKey 定义请求 ID 的上下文键.
	requestIDKey struct{}
	// tenantIDKey 定义租户 ID 的上下文键.
	tenantIDKey struct{}
)

// WithUserID 将用户 ID 存放到上下文中.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID 从上下文中提取用户 ID.
func UserID(ctx context.Context) int64 {
	userID, _ := ctx.Value(userIDKey{}).(int64)
	return userID
}

// WithUserIDString 将字符串用户 ID 存放到上下文中（向后兼容）.
func WithUserIDString(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserIDString 从上下文中提取字符串用户 ID（向后兼容）.
func UserIDString(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey{}).(string)
	return userID
}

// WithUsername 将用户名存放到上下文中.
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey{}, username)
}

// Username User 从上下文中提取用户名.
func Username(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey{}).(string)
	return username
}

// WithAccessToken 将访问令牌存放到上下文中.
func WithAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, accessToken)
}

// AccessToken 从上下文中提取访问令牌.
func AccessToken(ctx context.Context) string {
	accessToken, _ := ctx.Value(accessTokenKey{}).(string)
	return accessToken
}

// WithRequestID 将请求 ID 存放到上下文中.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestID 从上下文中提取请求 ID.
func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

// WithTenantID 将租户 ID 存放到上下文中.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

// TenantID 从上下文中提取租户 ID.
func TenantID(ctx context.Context) string {
	tenantID, _ := ctx.Value(tenantIDKey{}).(string)
	return tenantID
}
