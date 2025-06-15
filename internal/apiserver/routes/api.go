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

// InstallUserRoutes 安装用户相关的路由
func InstallUserRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddlewares ...gin.HandlerFunc) {
	// 用户管理路由
	userGroup := v1.Group("/users")
	{
		// 创建用户。这里要注意：创建用户是不用进行认证和授权的
		userGroup.POST("", h.CreateUser)
		userGroup.Use(authMiddlewares...)                          // 应用中间件。之后的接口需要认证和授权
		userGroup.PUT(":userID/change-password", h.ChangePassword) // 修改用户密码
		userGroup.PUT(":userID", h.UpdateUser)                     // 更新用户信息
		userGroup.DELETE(":userID", h.DeleteUser)                  // 删除用户
		userGroup.GET(":userID", h.GetUser)                        // 查询用户详情
		userGroup.GET("", h.ListUser)                              // 查询用户列表

		// 用户角色管理
		userGroup.GET("/:userID/roles", h.GetUserRoles) // 获取用户角色
		userGroup.POST("/roles", h.AssignUserRoles)     // 分配用户角色
	}

	// 当前用户相关路由
	currentUserGroup := v1.Group("/user", authMiddlewares...)
	{
		// 租户相关
		currentUserGroup.GET("/tenants", h.GetUserTenants)      // 获取用户租户列表
		currentUserGroup.POST("/switch-tenant", h.SwitchTenant) // 切换租户
		currentUserGroup.GET("/profile", h.GetUserProfile)      // 获取用户完整信息

		// 权限相关
		currentUserGroup.GET("/permissions", h.GetUserPermissions) // 获取用户权限

		// 菜单相关
		currentUserGroup.GET("/menus", h.GetUserMenus) // 获取用户菜单
	}
}

// InstallTenantRoutes 安装租户相关的路由
func InstallTenantRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddlewares ...gin.HandlerFunc) {
	// 租户管理路由
	tenantGroup := v1.Group("/tenants", authMiddlewares...)
	{
		tenantGroup.GET("", h.ListTenants) // 获取租户列表
	}
}

// InstallRoleRoutes 安装角色相关的路由
func InstallRoleRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddlewares ...gin.HandlerFunc) {
	// 角色管理路由
	roleGroup := v1.Group("/roles", authMiddlewares...)
	{
		roleGroup.GET("", h.ListRoles)                             // 获取角色列表
		roleGroup.POST("", h.CreateRole)                           // 创建角色
		roleGroup.PUT("/:role_id", h.UpdateRole)                   // 更新角色
		roleGroup.DELETE("/:role_id", h.DeleteRole)                // 删除角色
		roleGroup.GET("/:role_id/check-delete", h.CheckDeleteRole) // 检查角色是否可删除

		// 角色权限管理
		roleGroup.GET("/:role_id/permissions", h.GetRolePermissions) // 获取角色权限
		roleGroup.POST("/permissions", h.AssignRolePermissions)      // 分配角色权限

		// 角色菜单管理
		roleGroup.GET("/:role_id/menus", h.GetRoleMenus) // 获取角色菜单
		roleGroup.PUT("/menus", h.UpdateRoleMenus)       // 更新角色菜单
	}

	// 当前用户角色相关路由
	currentUserRoleGroup := v1.Group("/user/roles", authMiddlewares...)
	{
		currentUserRoleGroup.GET("", h.GetRolesByUser) // 获取当前用户角色
	}

	// 系统管理路由
	systemGroup := v1.Group("/system", authMiddlewares...)
	{
		systemGroup.POST("/refresh-privilege", h.RefreshPrivilegeData) // 刷新权限数据
	}
}

// InstallPermissionRoutes 安装权限相关的路由
func InstallPermissionRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddlewares ...gin.HandlerFunc) {
	// 权限检查路由
	permissionGroup := v1.Group("/permissions", authMiddlewares...)
	{
		permissionGroup.POST("/check", h.CheckPermissions) // 批量检查权限
	}

	// API访问检查路由
	apiGroup := v1.Group("/api", authMiddlewares...)
	{
		apiGroup.GET("/check-access", h.CheckAPIAccess) // 检查API访问权限
	}
}

// InstallMenuRoutes 安装菜单相关的路由
func InstallMenuRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddlewares ...gin.HandlerFunc) {
	// 菜单管理路由
	menuGroup := v1.Group("/menus", authMiddlewares...)
	{
		menuGroup.GET("", h.ListMenus) // 获取菜单列表
	}
}

// InstallPostRoutes 安装博客相关的路由
func InstallPostRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddlewares ...gin.HandlerFunc) {
	postGroup := v1.Group("/posts", authMiddlewares...) // 所有博客相关接口都需要认证和授权
	{
		postGroup.POST("", h.CreatePost)       // 创建博客
		postGroup.PUT(":postID", h.UpdatePost) // 更新博客
		postGroup.DELETE("", h.DeletePost)     // 删除博客
		postGroup.GET(":postID", h.GetPost)    // 查询博客详情
		postGroup.GET("", h.ListPost)          // 查询博客列表
	}
}
