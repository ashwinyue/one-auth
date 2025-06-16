// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package routes

import (
	"github.com/gin-gonic/gin"

	handler "github.com/ashwinyue/one-auth/internal/apiserver/handler/http"
)

// InstallTenantRoutes 设置租户相关的路由.
func InstallTenantRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddlewares ...gin.HandlerFunc) {
	// 用户租户相关 - 需要认证
	userGroup := v1.Group("/user")
	userGroup.Use(authMiddlewares...)
	{
		// 获取用户所属的租户列表
		userGroup.GET("/tenants", h.GetUserTenants)

		// 获取用户完整信息（包含当前租户、角色、权限）
		userGroup.GET("/profile", h.GetUserProfile)
	}

	// 租户切换 - 需要认证
	tenantGroup := v1.Group("/tenant")
	tenantGroup.Use(authMiddlewares...)
	{
		// 切换用户当前工作租户
		tenantGroup.POST("/switch", h.SwitchTenant)
	}

	// 租户管理 - 需要认证和权限验证
	tenantsGroup := v1.Group("/tenants")
	tenantsGroup.Use(authMiddlewares...)
	{
		// 获取租户列表
		tenantsGroup.GET("", h.ListTenants)
	}
}
