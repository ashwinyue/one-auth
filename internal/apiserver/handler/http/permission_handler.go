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

// GetUserPermissions 获取用户权限
func (h *Handler) GetUserPermissions(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.PermissionV1().GetUserPermissions)
}

// CheckPermissions 批量检查权限
func (h *Handler) CheckPermissions(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.PermissionV1().CheckPermissions)
}

// CheckAPIAccess 检查API访问权限
func (h *Handler) CheckAPIAccess(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.PermissionV1().CheckAPIAccess)
}
