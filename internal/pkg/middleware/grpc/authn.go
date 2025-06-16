// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package grpc

import (
	"context"
	"strconv"

	"github.com/ashwinyue/one-auth/pkg/store/where"
	"github.com/ashwinyue/one-auth/pkg/token"
	"google.golang.org/grpc"

	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/known"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// AuthnInterceptor 是一个 gRPC 拦截器，用于进行认证.
func AuthnInterceptor(userStore store.UserStore) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 解析 JWT Token
		userID, err := token.ParseRequest(ctx)
		if err != nil {
			log.Errorw("Failed to parse request", "err", err)
			return nil, errno.ErrTokenInvalid.WithMessage(err.Error())
		}

		log.Debugw("Token parsing successful", "userID", userID)

		// 获取用户信息
		user, err := userStore.Get(ctx, where.F("id", userID))
		if err != nil {
			return nil, errno.ErrUnauthenticated.WithMessage(err.Error())
		}

		// 获取用户的租户ID
		tenantID, err := userStore.GetUserTenantID(ctx, userID)
		if err != nil {
			log.Errorw("Failed to get user tenant ID", "userID", userID, "err", err)
			// 租户ID获取失败不阻止认证，但需要记录日志
			tenantID = 0
		}

		// 将用户信息存入上下文
		//nolint: staticcheck
		ctx = context.WithValue(ctx, known.XUsername, user.Username)
		//nolint: staticcheck
		ctx = context.WithValue(ctx, known.XUserID, userID)

		// 供 log 和 contextx 使用
		ctx = contextx.WithUserID(ctx, user.ID)
		ctx = contextx.WithUsername(ctx, user.Username)
		if tenantID > 0 {
			ctx = contextx.WithTenantID(ctx, strconv.FormatInt(tenantID, 10))
		}

		// 继续处理请求
		return handler(ctx, req)
	}
}
