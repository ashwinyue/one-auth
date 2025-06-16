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

// CreateRole 创建角色
func (h *Handler) CreateRole(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().Create)
}

// UpdateRole 更新角色
func (h *Handler) UpdateRole(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().Update)
}

// DeleteRole 删除角色
func (h *Handler) DeleteRole(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.RoleV1().Delete)
}

// ListRoles 获取角色列表
func (h *Handler) ListRoles(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.RoleV1().List)
}
