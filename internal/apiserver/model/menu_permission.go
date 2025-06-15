// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package model

import (
	"time"
)

const TableNameMenuPermissionM = "menu_permissions"

// MenuPermissionM 菜单权限关联模型
type MenuPermissionM struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:主键ID" json:"id"`
	TenantID     int64     `gorm:"column:tenant_id;not null;comment:租户ID" json:"tenant_id"`
	MenuID       int64     `gorm:"column:menu_id;not null;comment:菜单ID" json:"menu_id"`
	PermissionID int64     `gorm:"column:permission_id;not null;comment:权限ID" json:"permission_id"`
	IsRequired   bool      `gorm:"column:is_required;not null;default:0;comment:是否为访问菜单的必需权限" json:"is_required"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// TableName MenuPermissionM's table name
func (*MenuPermissionM) TableName() string {
	return TableNameMenuPermissionM
}

// MenuWithPermissions 带权限的菜单结构
type MenuWithPermissions struct {
	MenuM
	Permissions         []*PermissionNewM `json:"permissions,omitempty"`
	RequiredPermissions []*PermissionNewM `json:"required_permissions,omitempty"`
}

// PermissionWithMenus 带菜单的权限结构
type PermissionWithMenus struct {
	PermissionNewM
	Menus []*MenuM `json:"menus,omitempty"`
}

// MenuPermissionConfig 菜单权限配置
type MenuPermissionConfig struct {
	PermissionCode string `json:"permission_code" validate:"required"`
	IsRequired     bool   `json:"is_required"`
	AutoCreate     bool   `json:"auto_create"` // 如果权限不存在是否自动创建
}

// MenuPermissionMatrix 菜单权限矩阵（用于权限检查）
type MenuPermissionMatrix struct {
	Menu                *MenuM            `json:"menu"`
	RequiredPermissions []*PermissionNewM `json:"required_permissions"`
	OptionalPermissions []*PermissionNewM `json:"optional_permissions"`
	AllPermissions      []*PermissionNewM `json:"all_permissions"`
}

// HasRequiredPermissions 检查是否具有必需权限
func (matrix *MenuPermissionMatrix) HasRequiredPermissions(userPermissions []string) bool {
	if len(matrix.RequiredPermissions) == 0 {
		return true // 没有必需权限，允许访问
	}

	permissionMap := make(map[string]bool)
	for _, perm := range userPermissions {
		permissionMap[perm] = true
	}

	for _, reqPerm := range matrix.RequiredPermissions {
		if !permissionMap[reqPerm.PermissionCode] {
			return false
		}
	}

	return true
}

// GetAvailableActions 获取用户在此菜单可执行的操作
func (matrix *MenuPermissionMatrix) GetAvailableActions(userPermissions []string) []string {
	permissionMap := make(map[string]bool)
	for _, perm := range userPermissions {
		permissionMap[perm] = true
	}

	var actions []string
	for _, perm := range matrix.AllPermissions {
		if permissionMap[perm.PermissionCode] && perm.Action != nil {
			actions = append(actions, *perm.Action)
		}
	}

	return actions
}
