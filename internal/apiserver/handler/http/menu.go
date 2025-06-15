// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package http

import (
	"github.com/ashwinyue/one-auth/pkg/core"
	"github.com/gin-gonic/gin"
)

// GetUserMenus 获取用户可访问的菜单
func (h *Handler) GetUserMenus(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.MenuV1().GetUserMenus, h.val.ValidateGetUserMenusRequest)
}

// ListMenus 获取菜单列表
func (h *Handler) ListMenus(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.MenuV1().ListMenus, h.val.ValidateListMenusRequest)
}
