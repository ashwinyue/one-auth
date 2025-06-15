// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package model

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ResourceType 资源类型枚举
type ResourceType string

const (
	ResourceTypeAPI     ResourceType = "api"
	ResourceTypeMenu    ResourceType = "menu"
	ResourceTypeData    ResourceType = "data"
	ResourceTypeFeature ResourceType = "feature"
)

// PermissionNewM 新的权限模型（重构版）
type PermissionNewM struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement:true;comment:权限主键ID" json:"id"`
	TenantID       int64          `gorm:"column:tenant_id;not null;comment:租户ID" json:"tenant_id"`
	PermissionCode string         `gorm:"column:permission_code;not null;uniqueIndex:idx_permission_code_tenant;comment:权限编码" json:"permission_code"`
	Name           string         `gorm:"column:name;not null;comment:权限名称" json:"name"`
	Description    *string        `gorm:"column:description;comment:权限描述" json:"description"`
	ResourceType   ResourceType   `gorm:"column:resource_type;not null;default:menu;comment:资源类型" json:"resource_type"`
	ResourcePath   *string        `gorm:"column:resource_path;comment:API路径或资源标识" json:"resource_path"`
	HTTPMethod     *string        `gorm:"column:http_method;comment:HTTP方法" json:"http_method"`
	Action         *string        `gorm:"column:action;comment:操作类型" json:"action"`
	Status         bool           `gorm:"column:status;not null;default:1;comment:状态：1-启用，0-禁用" json:"status"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间（软删除）" json:"deleted_at"`
}

// TableName 表名
func (*PermissionNewM) TableName() string {
	return "permissions"
}

// IsAPIPermission 判断是否为API权限
func (p *PermissionNewM) IsAPIPermission() bool {
	return p.ResourceType == ResourceTypeAPI
}

// IsMenuPermission 判断是否为菜单权限
func (p *PermissionNewM) IsMenuPermission() bool {
	return p.ResourceType == ResourceTypeMenu
}

// GetActionString 获取操作类型字符串
func (p *PermissionNewM) GetActionString() string {
	if p.Action != nil {
		return *p.Action
	}
	return ""
}

// GetResourcePathString 获取资源路径字符串
func (p *PermissionNewM) GetResourcePathString() string {
	if p.ResourcePath != nil {
		return *p.ResourcePath
	}
	return ""
}

// GetHTTPMethodString 获取HTTP方法字符串
func (p *PermissionNewM) GetHTTPMethodString() string {
	if p.HTTPMethod != nil {
		return *p.HTTPMethod
	}
	return ""
}

// PermissionGroup 权限分组
type PermissionGroup struct {
	Module      string            `json:"module"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Permissions []*PermissionNewM `json:"permissions"`
}

// UserPermissionSummary 用户权限汇总
type UserPermissionSummary struct {
	UserID      string              `json:"user_id"`
	TenantID    int64               `json:"tenant_id"`
	Roles       []string            `json:"roles"`
	Permissions []*PermissionNewM   `json:"permissions"`
	MenuAccess  map[int64][]string  `json:"menu_access"` // 菜单ID -> 可执行操作列表
	APIAccess   map[string][]string `json:"api_access"`  // API路径 -> HTTP方法列表
}

// PermissionTemplate 权限模板
type PermissionTemplate struct {
	Code         string       `json:"code"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Action       string       `json:"action"`
	ResourceType ResourceType `json:"resource_type"`
}

// CreatePermission 从模板创建权限
func (template PermissionTemplate) CreatePermission(tenantID int64) *PermissionNewM {
	return &PermissionNewM{
		TenantID:       tenantID,
		PermissionCode: template.Code,
		Name:           template.Name,
		Description:    &template.Description,
		ResourceType:   template.ResourceType,
		Action:         &template.Action,
		Status:         true,
	}
}

// GenerateStandardPermissionsForModule 为模块动态生成标准权限模板
func GenerateStandardPermissionsForModule(module string) []PermissionTemplate {
	// 标准CRUD操作
	standardActions := []struct {
		action      string
		namePrefix  string
		description string
	}{
		{"view", "查看", "查看%s列表和详情"},
		{"create", "创建", "创建新%s"},
		{"update", "编辑", "修改%s信息"},
		{"delete", "删除", "删除%s"},
		{"export", "导出", "导出%s数据"},
	}

	var templates []PermissionTemplate

	for _, stdAction := range standardActions {
		template := PermissionTemplate{
			Code:         module + ":" + stdAction.action,
			Name:         stdAction.namePrefix + getModuleDisplayName(module),
			Description:  fmt.Sprintf(stdAction.description, getModuleDisplayName(module)),
			Action:       stdAction.action,
			ResourceType: ResourceTypeMenu,
		}
		templates = append(templates, template)
	}

	return templates
}

// getModuleDisplayName 获取模块的显示名称
func getModuleDisplayName(module string) string {
	moduleNames := map[string]string{
		"user":       "用户",
		"role":       "角色",
		"permission": "权限",
		"menu":       "菜单",
		"tenant":     "租户",
		"order":      "订单",
		"product":    "商品",
		"customer":   "客户",
		"report":     "报表",
		"system":     "系统",
	}

	if displayName, exists := moduleNames[module]; exists {
		return displayName
	}

	return module // 如果没有映射，返回原始模块名
}

// CreatePermissionsForNewModule 为新模块创建完整的权限集
func CreatePermissionsForNewModule(module string, tenantID int64, customActions ...string) []*PermissionNewM {
	var permissions []*PermissionNewM

	// 添加标准权限
	standardTemplates := GenerateStandardPermissionsForModule(module)
	for _, template := range standardTemplates {
		permissions = append(permissions, template.CreatePermission(tenantID))
	}

	// 添加自定义权限
	for _, action := range customActions {
		template := PermissionTemplate{
			Code:         module + ":" + action,
			Name:         action + getModuleDisplayName(module),
			Description:  getModuleDisplayName(module) + "的" + action + "操作",
			Action:       action,
			ResourceType: ResourceTypeFeature,
		}
		permissions = append(permissions, template.CreatePermission(tenantID))
	}

	return permissions
}

// PermissionCodeValidator 权限编码验证器
type PermissionCodeValidator struct{}

// IsValidPermissionCode 验证权限编码格式
func (v *PermissionCodeValidator) IsValidPermissionCode(code string) bool {
	// 标准格式: {module}:{action}
	parts := strings.Split(code, ":")
	if len(parts) != 2 {
		return false
	}

	module := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])

	return module != "" && action != ""
}

// GetModuleFromCode 从权限编码中提取模块名
func (v *PermissionCodeValidator) GetModuleFromCode(code string) string {
	parts := strings.Split(code, ":")
	if len(parts) >= 1 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

// GetActionFromCode 从权限编码中提取操作名
func (v *PermissionCodeValidator) GetActionFromCode(code string) string {
	parts := strings.Split(code, ":")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
