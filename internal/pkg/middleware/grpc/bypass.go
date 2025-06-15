// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package grpc

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/known"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// AuthnBypassInterceptor 是一个 gRPC 拦截器，模拟所有请求都通过认证。
func AuthnBypassInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 从请求头中获取用户 ID
		userIDStr := "1" // 默认用户 ID（数字）
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			// 获取 header 中指定的用户 ID，假设 Header 名为 "x-user-id"
			if values := md.Get(known.XUserID); len(values) > 0 {
				userIDStr = values[0]
			}
		}

		// 将字符串转换为数字ID
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			userID = 1 // 如果解析失败，使用默认用户ID
		}

		log.Debugw("Simulated authentication successful", "userID", userID)

		// 将默认的用户信息存入上下文
		//nolint: staticcheck
		ctx = context.WithValue(ctx, known.XUserID, userIDStr)

		// 为 log 和 contextx 提供用户上下文支持
		ctx = contextx.WithUserID(ctx, userID)

		// 继续处理请求
		return handler(ctx, req)
	}
}
