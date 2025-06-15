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

// GetUserTenants 获取用户所属的租户列表
func (h *Handler) GetUserTenants(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.TenantV1().GetUserTenants, h.val.ValidateGetUserTenantsRequest)
}

// SwitchTenant 切换用户当前工作租户
func (h *Handler) SwitchTenant(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.TenantV1().SwitchTenant, h.val.ValidateSwitchTenantRequest)
}

// GetUserProfile 获取用户完整信息（包含当前租户、角色、权限）
func (h *Handler) GetUserProfile(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.TenantV1().GetUserProfile, h.val.ValidateGetUserProfileRequest)
}

// ListTenants 获取租户列表（管理员用）
func (h *Handler) ListTenants(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.TenantV1().ListTenants, h.val.ValidateListTenantsRequest)
}
