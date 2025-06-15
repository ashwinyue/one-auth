// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package authz

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// 基础前缀（简化版，参考旧项目auth-center但去掉复杂的组织架构）
	PrefixUserID     = "u" // 用户ID：u123
	PrefixRoleID     = "r" // 角色ID：r1
	PrefixResourceID = "a" // 资源/权限ID：a234
	PrefixDomainID   = "t" // 租户/域ID：t1
	PrefixMenuID     = "m" // 菜单ID：m100

	// 特殊标识符
	RootFlag    = "root"    // 超级管理员标识
	DefaultFlag = "default" // 默认标识
)

// IDConverter 提供ID转换功能，参考旧项目auth-center的实现方式但简化版本
type IDConverter struct{}

// NewIDConverter 创建ID转换器实例
func NewIDConverter() *IDConverter {
	return &IDConverter{}
}

// ========== 角色ID转换 ==========

// ToDRoleID 转换角色ID为Casbin存储格式
func (c *IDConverter) ToDRoleID(roleID int64) string {
	return PrefixRoleID + fmt.Sprintf("%d", roleID)
}

// ToRoleID 转换Casbin存储的角色ID为正常格式
func (c *IDConverter) ToRoleID(role string) int64 {
	roleID, _ := strconv.ParseInt(strings.TrimPrefix(role, PrefixRoleID), 10, 64)
	return roleID
}

// ========== 租户/域ID转换 ==========

// ToDDomainID 转换租户ID为Casbin存储的域格式
func (c *IDConverter) ToDDomainID(domainID int64) string {
	return PrefixDomainID + fmt.Sprintf("%d", domainID)
}

// ToDomainID 转换Casbin存储的域ID为正常格式
func (c *IDConverter) ToDomainID(domain string) int64 {
	domainID, _ := strconv.ParseInt(strings.TrimPrefix(domain, PrefixDomainID), 10, 64)
	return domainID
}

// ========== 用户ID转换 ==========

// ToDUserID 转换用户ID为Casbin存储格式
func (c *IDConverter) ToDUserID(userID int64) string {
	return PrefixUserID + fmt.Sprintf("%d", userID)
}

// ToUserID 转换Casbin存储的用户ID为正常格式
func (c *IDConverter) ToUserID(user string) int64 {
	if len(user) == 0 {
		return 0
	}

	// 如果有用户前缀，去掉前缀
	if strings.HasPrefix(user, PrefixUserID) {
		userID, _ := strconv.ParseInt(strings.TrimPrefix(user, PrefixUserID), 10, 64)
		return userID
	}

	// 尝试直接解析为数字
	if id, err := strconv.ParseInt(user, 10, 64); err == nil {
		return id
	}

	return 0
}

// ========== 资源ID转换 ==========

// ToDResourceID 转换资源ID为Casbin存储格式
func (c *IDConverter) ToDResourceID(resourceID int64) string {
	return PrefixResourceID + fmt.Sprintf("%d", resourceID)
}

// ToResourceID 转换Casbin存储的资源ID为正常格式
func (c *IDConverter) ToResourceID(resource string) int64 {
	resourceID, _ := strconv.ParseInt(strings.TrimPrefix(resource, PrefixResourceID), 10, 64)
	return resourceID
}

// ========== 菜单ID转换 ==========

// ToDMenuID 转换菜单ID为Casbin存储格式
func (c *IDConverter) ToDMenuID(menuID int64) string {
	return PrefixMenuID + fmt.Sprintf("%d", menuID)
}

// ToMenuID 转换Casbin存储的菜单ID为正常格式
func (c *IDConverter) ToMenuID(menu string) int64 {
	menuID, _ := strconv.ParseInt(strings.TrimPrefix(menu, PrefixMenuID), 10, 64)
	return menuID
}

// ========== 权限ID转换（使用资源前缀） ==========

// ToDPermissionID 转换权限ID为Casbin存储格式（权限本质上是资源）
func (c *IDConverter) ToDPermissionID(permissionID int64) string {
	return PrefixResourceID + fmt.Sprintf("%d", permissionID)
}

// ToPermissionID 转换Casbin存储的权限ID为正常格式
func (c *IDConverter) ToPermissionID(permission string) int64 {
	permissionID, _ := strconv.ParseInt(strings.TrimPrefix(permission, PrefixResourceID), 10, 64)
	return permissionID
}

// ========== 通用转换方法 ==========

// ParseIDWithPrefix 解析带前缀的ID字符串，返回ID和前缀类型
func (c *IDConverter) ParseIDWithPrefix(idStr string) (int64, string, error) {
	if len(idStr) == 0 {
		return 0, "", fmt.Errorf("empty ID string")
	}

	// 如果没有前缀，尝试直接解析为数字
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		return id, "", nil
	}

	// 解析带前缀的ID
	if len(idStr) < 2 {
		return 0, "", fmt.Errorf("invalid ID format: %s", idStr)
	}

	prefix := idStr[0:1]
	idPart := idStr[1:]

	id, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid ID format: %s", idStr)
	}

	return id, prefix, nil
}

// FormatIDWithPrefix 使用指定前缀格式化ID
func (c *IDConverter) FormatIDWithPrefix(id int64, prefix string) string {
	return prefix + fmt.Sprintf("%d", id)
}

// ========== 批量转换方法 ==========

// ToDRoleIDs 批量转换角色ID列表
func (c *IDConverter) ToDRoleIDs(roleIDs []int64) []string {
	result := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		result[i] = c.ToDRoleID(id)
	}
	return result
}

// ToRoleIDs 批量转换Casbin格式的角色ID列表
func (c *IDConverter) ToRoleIDs(roles []string) []int64 {
	result := make([]int64, len(roles))
	for i, role := range roles {
		result[i] = c.ToRoleID(role)
	}
	return result
}

// ToDUserIDs 批量转换用户ID列表
func (c *IDConverter) ToDUserIDs(userIDs []int64) []string {
	result := make([]string, len(userIDs))
	for i, id := range userIDs {
		result[i] = c.ToDUserID(id)
	}
	return result
}

// ========== 验证方法 ==========

// IsValidPrefixedID 验证ID是否符合前缀格式
func (c *IDConverter) IsValidPrefixedID(idStr, expectedPrefix string) bool {
	if !strings.HasPrefix(idStr, expectedPrefix) {
		return false
	}

	idPart := strings.TrimPrefix(idStr, expectedPrefix)
	_, err := strconv.ParseInt(idPart, 10, 64)
	return err == nil
}

// GetIDType 根据前缀获取ID类型
func (c *IDConverter) GetIDType(idStr string) string {
	if len(idStr) == 0 {
		return "unknown"
	}

	prefix := idStr[0:1]
	switch prefix {
	case PrefixUserID:
		return "user"
	case PrefixRoleID:
		return "role"
	case PrefixResourceID:
		return "resource"
	case PrefixDomainID:
		return "domain"
	case PrefixMenuID:
		return "menu"
	default:
		return "unknown"
	}
}

// ========== 兼容性方法（与旧项目保持接口一致） ==========

// toDRoleID 兼容旧项目的方法命名
func (c *IDConverter) toDRoleID(role int64) string {
	return c.ToDRoleID(role)
}

// toRoleID 兼容旧项目的方法命名
func (c *IDConverter) toRoleID(role string) int64 {
	return c.ToRoleID(role)
}

// toDDomainID 兼容旧项目的方法命名
func (c *IDConverter) toDDomainID(domain int64) string {
	return c.ToDDomainID(domain)
}

// toDUserID 兼容旧项目的方法命名（简化版，去掉prefixType参数）
func (c *IDConverter) toDUserID(userId int64) string {
	return c.ToDUserID(userId)
}

// toUserID 兼容旧项目的方法命名（简化版，只返回ID）
func (c *IDConverter) toUserID(user string) int64 {
	return c.ToUserID(user)
}

// toDResourceID 兼容旧项目的方法命名
func (c *IDConverter) toDResourceID(resource int64) string {
	return c.ToDResourceID(resource)
}

// toResourceID 兼容旧项目的方法命名
func (c *IDConverter) toResourceID(resource string) int64 {
	return c.ToResourceID(resource)
}

// ========== 使用示例（注释说明） ==========

/*
使用示例（参考旧项目的用法）：

// 创建转换器
converter := NewIDConverter()

// 角色ID转换
roleID := converter.ToDRoleID(1)           // "r1"
originalRoleID := converter.ToRoleID("r1") // 1

// 用户ID转换
userID := converter.ToDUserID(123)           // "u123"
originalUserID := converter.ToUserID("u123") // 123

// 域/租户ID转换
domainID := converter.ToDDomainID(1)           // "t1"
originalDomainID := converter.ToDomainID("t1") // 1

// 资源/权限ID转换
resourceID := converter.ToDResourceID(456)           // "a456"
originalResourceID := converter.ToResourceID("a456") // 456

// 在Casbin中的使用方式（参考旧项目）：
// uc.ec.GetRolesForUser(converter.toDUserID(userInfo.UserId), converter.toDDomainID(userInfo.TenantId))
// uc.ec.AddRoleForUser(converter.toDUserID(req.UserId), converter.toDRoleID(rootRole.ID), converter.toDDomainID(userInfo.TenantId))
*/
