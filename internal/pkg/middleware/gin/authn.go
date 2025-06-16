// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package gin

import (
	"strconv"

	"github.com/ashwinyue/one-auth/pkg/core"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"github.com/ashwinyue/one-auth/pkg/token"
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// AuthnMiddleware 是一个认证中间件，用于从 gin.Context 中提取 token 并验证 token 是否合法.
func AuthnMiddleware(userStore store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 JWT Token
		userID, err := token.ParseRequest(c)
		if err != nil {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()
			return
		}

		log.Debugw("Token parsing successful", "userID", userID)

		// 获取用户信息
		user, err := userStore.Get(c.Request.Context(), where.F("id", userID))
		if err != nil {
			core.WriteResponse(c, errno.ErrUnauthenticated, nil)
			c.Abort()
			return
		}

		// 获取用户的租户ID
		tenantID, err := userStore.GetUserTenantID(c.Request.Context(), userID)
		if err != nil {
			log.Errorw("Failed to get user tenant ID", "userID", userID, "err", err)
			// 租户ID获取失败不阻止认证，但需要记录日志
			tenantID = 0
		}

		// 将用户信息存入上下文
		c.Set("userID", userID)
		c.Set("username", user.Username)
		if tenantID > 0 {
			c.Set("tenantID", strconv.FormatInt(tenantID, 10))
		}

		// 供 log 和 contextx 使用
		ctx := contextx.WithUserID(c.Request.Context(), user.ID)
		ctx = contextx.WithUsername(ctx, user.Username)
		if tenantID > 0 {
			ctx = contextx.WithTenantID(ctx, strconv.FormatInt(tenantID, 10))
		}
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
