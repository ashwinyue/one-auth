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

// ListRoles 获取角色列表
func (h *Handler) ListRoles(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.RoleV1().ListRoles, h.val.ValidateListRolesRequest)
}

// GetRolePermissions 获取角色的权限列表
func (h *Handler) GetRolePermissions(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.RoleV1().GetRolePermissions, h.val.ValidateGetRolePermissionsRequest)
}

// AssignRolePermissions 为角色分配权限
func (h *Handler) AssignRolePermissions(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().AssignRolePermissions, h.val.ValidateAssignRolePermissionsRequest)
}

// GetUserRoles 获取用户的角色列表
func (h *Handler) GetUserRoles(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.RoleV1().GetUserRoles, h.val.ValidateGetUserRolesRequest)
}

// AssignUserRoles 为用户分配角色
func (h *Handler) AssignUserRoles(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().AssignUserRoles, h.val.ValidateAssignUserRolesRequest)
}

// CreateRole 创建角色
func (h *Handler) CreateRole(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().CreateRole, h.val.ValidateCreateRoleRequest)
}

// UpdateRole 更新角色
func (h *Handler) UpdateRole(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().UpdateRole, h.val.ValidateUpdateRoleRequest)
}

// DeleteRole 删除角色
func (h *Handler) DeleteRole(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.RoleV1().DeleteRole, h.val.ValidateDeleteRoleRequest)
}

// CheckDeleteRole 检查角色是否可以删除
func (h *Handler) CheckDeleteRole(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.RoleV1().CheckDeleteRole, h.val.ValidateCheckDeleteRoleRequest)
}

// GetRoleMenus 获取角色菜单
func (h *Handler) GetRoleMenus(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.RoleV1().GetRoleMenus, h.val.ValidateGetRoleMenusRequest)
}

// UpdateRoleMenus 更新角色菜单
func (h *Handler) UpdateRoleMenus(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().UpdateRoleMenus, h.val.ValidateUpdateRoleMenusRequest)
}

// GetRolesByUser 获取当前用户的角色
func (h *Handler) GetRolesByUser(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.RoleV1().GetRolesByUser, h.val.ValidateGetRolesByUserRequest)
}

// RefreshPrivilegeData 刷新权限数据
func (h *Handler) RefreshPrivilegeData(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.RoleV1().RefreshPrivilegeData, h.val.ValidateRefreshPrivilegeDataRequest)
}
