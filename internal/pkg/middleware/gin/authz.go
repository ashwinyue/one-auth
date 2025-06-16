// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package gin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/one-auth/pkg/core"

	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// Authorizer 用于定义授权接口的实现.
type Authorizer interface {
	// 使用domain进行授权检查
	AuthorizeWithDomain(subject, domain, object, action string) (bool, error)
}

// APIAuthorizer 是一个额外的接口，用于基于API路径的权限检查
type APIAuthorizer interface {
	// 使用API路径进行授权检查
	CheckAPIAccess(subject, domain, object, action string) (bool, error)
}

// AuthzMiddleware 是一个 Gin 中间件，用于进行请求授权.
func AuthzMiddleware(authorizer Authorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject := strconv.FormatInt(contextx.UserID(c.Request.Context()), 10)
		domain := contextx.TenantID(c.Request.Context()) // 获取租户ID作为domain
		object := c.Request.URL.Path
		action := c.Request.Method

		// 权限检查不再跳过任何路径，所有通过认证的接口都需要权限验证

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

		// 首先检查是否是基于API路径的权限检查
		if apiAuthorizer, ok := authorizer.(APIAuthorizer); ok {
			// 使用API权限检查
			allowed, err := apiAuthorizer.CheckAPIAccess(subject, domain, object, action)
			if err != nil || !allowed {
				core.WriteResponse(c, nil, errno.ErrPermissionDenied.WithMessage(
					"access denied: subject=%s, domain=%s, object=%s, action=%s, reason=%v",
					subject,
					domain,
					object,
					action,
					err,
				))
				c.Abort()
				return
			}
		} else {
			// 使用传统的domain授权检查
			allowed, err := authorizer.AuthorizeWithDomain(subject, domain, object, action)
			if err != nil || !allowed {
				core.WriteResponse(c, nil, errno.ErrPermissionDenied.WithMessage(
					"access denied: subject=%s, domain=%s, object=%s, action=%s, reason=%v",
					subject,
					domain,
					object,
					action,
					err,
				))
				c.Abort()
				return
			}
		}

		c.Next() // 继续处理请求
	}
}
