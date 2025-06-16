// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package http

import (
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/one-auth/pkg/core"
)

// CreateMenu 创建菜单
func (h *Handler) CreateMenu(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.MenuV1().CreateMenu)
}

// UpdateMenu 更新菜单
func (h *Handler) UpdateMenu(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.MenuV1().UpdateMenu)
}

// DeleteMenu 删除菜单
func (h *Handler) DeleteMenu(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.MenuV1().DeleteMenu)
}

// GetMenu 获取菜单详情
func (h *Handler) GetMenu(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.MenuV1().GetMenu)
}

// ListMenus 获取菜单列表
func (h *Handler) ListMenus(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.MenuV1().ListMenus)
}

// GetUserMenus 获取用户菜单（权限过滤后的菜单树）
func (h *Handler) GetUserMenus(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.MenuV1().GetUserMenus)
}

// GetMenuTree 获取菜单树
func (h *Handler) GetMenuTree(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.MenuV1().GetMenuTree)
}

// UpdateMenuSort 批量更新菜单排序
func (h *Handler) UpdateMenuSort(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.MenuV1().UpdateMenuSort)
}

// CopyMenu 复制菜单
func (h *Handler) CopyMenu(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.MenuV1().CopyMenu)
}

// MoveMenu 移动菜单
func (h *Handler) MoveMenu(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.MenuV1().MoveMenu)
}
