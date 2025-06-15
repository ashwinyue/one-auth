// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// Authorizer 用于定义授权接口的实现.
type Authorizer interface {
	// 使用domain进行授权检查
	AuthorizeWithDomain(subject, domain, object, action string) (bool, error)
}

// AuthzInterceptor 是一个 gRPC 拦截器，用于进行请求授权.
func AuthzInterceptor(authorizer Authorizer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID := contextx.UserID(ctx)   // 获取用户ID
		domain := contextx.TenantID(ctx) // 获取租户ID作为domain
		object := info.FullMethod        // 获取请求资源
		action := "CALL"                 // 默认操作

		// 构建用户标识符
		subject := fmt.Sprintf("u%d", userID)

		// 如果没有租户ID，使用默认租户
		if domain == "" {
			domain = "default"
		}

		// 记录授权上下文信息
		log.Debugw("Build authorize context",
			"subject", subject,
			"domain", domain,
			"object", object,
			"action", action)

		// 使用domain进行授权检查
		allowed, err := authorizer.AuthorizeWithDomain(subject, domain, object, action)

		if err != nil || !allowed {
			return nil, errno.ErrPermissionDenied.WithMessage(
				"access denied: subject=%s, domain=%s, object=%s, action=%s, reason=%v",
				subject,
				domain,
				object,
				action,
				err,
			)
		}

		// 继续处理请求
		return handler(ctx, req)
	}
}
