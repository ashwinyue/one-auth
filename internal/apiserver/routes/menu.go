// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/one-auth/internal/apiserver/handler/http"
)

// RegisterMenuRoutes 注册菜单管理相关路由
func RegisterMenuRoutes(v1 *gin.RouterGroup, h *http.Handler, authMiddlewares ...gin.HandlerFunc) {
	menuRoutes := v1.Group("/menus")
	menuRoutes.Use(authMiddlewares...)

	// 菜单CRUD操作
	menuRoutes.POST("", h.CreateMenu)       // 创建菜单
	menuRoutes.GET("/:id", h.GetMenu)       // 获取菜单详情
	menuRoutes.PUT("/:id", h.UpdateMenu)    // 更新菜单
	menuRoutes.DELETE("/:id", h.DeleteMenu) // 删除菜单
	menuRoutes.GET("", h.ListMenus)         // 获取菜单列表

	// 菜单树和用户菜单
	menuRoutes.GET("/tree", h.GetMenuTree)  // 获取菜单树
	menuRoutes.GET("/user", h.GetUserMenus) // 获取用户菜单

	// 菜单管理操作
	menuRoutes.PUT("/sort", h.UpdateMenuSort) // 批量更新排序
	menuRoutes.POST("/copy", h.CopyMenu)      // 复制菜单
	menuRoutes.PUT("/move", h.MoveMenu)       // 移动菜单
}
