// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package permission

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"gorm.io/gorm"
)

// AutoAssignConfig 自动分配权限的配置
type AutoAssignConfig struct {
	// 启用自动分配
	Enabled bool `json:"enabled" yaml:"enabled"`

	// 超级管理员角色ID列表（享受免检权限）
	SuperAdminRoleIDs []int64 `json:"super_admin_role_ids" yaml:"super_admin_role_ids"`

	// 系统管理员角色ID列表（自动分配新权限）
	AdminRoleIDs []int64 `json:"admin_role_ids" yaml:"admin_role_ids"`

	// 需要自动分配权限的租户ID列表（空表示所有租户）
	TenantIDs []int64 `json:"tenant_ids" yaml:"tenant_ids"`

	// 管理员角色的权限过滤规则
	AdminPermissionRules []PermissionRule `json:"admin_permission_rules" yaml:"admin_permission_rules"`
}

// PermissionRule 权限分配规则
type PermissionRule struct {
	// 匹配的权限模块
	Modules []string `json:"modules" yaml:"modules"`

	// 匹配的权限操作
	Actions []string `json:"actions" yaml:"actions"`

	// 排除的权限编码
	ExcludeCodes []string `json:"exclude_codes" yaml:"exclude_codes"`

	// 是否包含该规则（true=包含，false=排除）
	Include bool `json:"include" yaml:"include"`
}

// DefaultAutoAssignConfig 返回默认配置
func DefaultAutoAssignConfig() *AutoAssignConfig {
	return &AutoAssignConfig{
		Enabled:           true,
		SuperAdminRoleIDs: []int64{1}, // 超级管理员角色ID为1
		AdminRoleIDs:      []int64{2}, // 系统管理员角色ID为2
		TenantIDs:         []int64{},  // 空表示所有租户
		AdminPermissionRules: []PermissionRule{
			{
				// 管理员拥有所有基础管理权限
				Modules: []string{"user", "role", "menu", "permission"},
				Actions: []string{"view", "create", "update", "delete"},
				Include: true,
			},
			{
				// 排除系统核心权限
				ExcludeCodes: []string{"system:config", "system:backup", "tenant:delete"},
				Include:      false,
			},
		},
	}
}

// AutoAssignPermissionsToRoles 自动为指定角色分配新权限
func (b *permissionBiz) AutoAssignPermissionsToRoles(ctx context.Context, permissionID int64, tenantID int64) error {
	config := DefaultAutoAssignConfig()

	if !config.Enabled {
		return nil
	}

	// 获取权限详情
	permission, err := b.store.Permission().Get(ctx, where.F("id", permissionID))
	if err != nil {
		return fmt.Errorf("failed to get permission details: %w", err)
	}

	// 为超级管理员分配权限（无条件）
	for _, roleID := range config.SuperAdminRoleIDs {
		err := b.assignPermissionToRole(ctx, roleID, permissionID, tenantID)
		if err != nil {
			log.W(ctx).Errorw("Failed to auto assign permission to super admin role",
				"role_id", roleID, "permission_id", permissionID, "err", err)
		}
	}

	// 为系统管理员分配权限（根据规则）
	if b.shouldAssignToAdmin(permission, config.AdminPermissionRules) {
		for _, roleID := range config.AdminRoleIDs {
			err := b.assignPermissionToRole(ctx, roleID, permissionID, tenantID)
			if err != nil {
				log.W(ctx).Errorw("Failed to auto assign permission to admin role",
					"role_id", roleID, "permission_id", permissionID, "err", err)
			} else {
				log.W(ctx).Infow("Auto assigned permission to admin role",
					"role_id", roleID, "permission_code", permission.PermissionCode)
			}
		}
	}

	return nil
}

// shouldAssignToAdmin 判断权限是否应该分配给管理员
func (b *permissionBiz) shouldAssignToAdmin(permission *model.PermissionM, rules []PermissionRule) bool {
	// 解析权限编码
	validator := &model.PermissionCodeValidator{}
	module := validator.GetModuleFromCode(permission.PermissionCode)
	action := validator.GetActionFromCode(permission.PermissionCode)

	for _, rule := range rules {
		matched := false

		// 检查模块匹配
		if len(rule.Modules) > 0 {
			for _, ruleModule := range rule.Modules {
				if module == ruleModule {
					matched = true
					break
				}
			}
		}

		// 检查操作匹配
		if !matched && len(rule.Actions) > 0 {
			for _, ruleAction := range rule.Actions {
				if action == ruleAction {
					matched = true
					break
				}
			}
		}

		// 检查排除列表
		if len(rule.ExcludeCodes) > 0 {
			for _, excludeCode := range rule.ExcludeCodes {
				if permission.PermissionCode == excludeCode {
					matched = true
					break
				}
			}
		}

		// 如果匹配到规则
		if matched {
			return rule.Include
		}
	}

	// 默认不分配
	return false
}

// CreateMenuWithAutoPermissions 创建菜单并自动生成权限
func (b *permissionBiz) CreateMenuWithAutoPermissions(ctx context.Context, menu *model.MenuM, actions []string) error {
	// 创建菜单
	err := b.store.Menu().Create(ctx, menu)
	if err != nil {
		return fmt.Errorf("failed to create menu: %w", err)
	}

	// 为菜单自动创建权限
	moduleCode := menu.MenuCode
	if moduleCode == "" {
		moduleCode = fmt.Sprintf("menu_%d", menu.ID)
	}

	// 如果没有指定操作，使用默认操作
	if len(actions) == 0 {
		actions = []string{"view", "create", "update", "delete"}
	}

	var createdPermissions []*model.PermissionM

	for _, action := range actions {
		permissionCode := moduleCode + ":" + action

		// 检查权限是否已存在
		exists, err := b.store.Permission().IsPermissionCodeExists(ctx, permissionCode, menu.TenantID)
		if err != nil {
			log.W(ctx).Errorw("Failed to check permission existence", "code", permissionCode, "err", err)
			continue
		}

		if exists {
			log.W(ctx).Warnw("Permission already exists", "code", permissionCode)
			continue
		}

		// 创建权限
		permission := &model.PermissionM{
			TenantID:       menu.TenantID,
			MenuID:         menu.ID,
			PermissionCode: permissionCode,
			Name:           getActionDisplayName(action) + menu.Title,
			Description:    stringPtr(getActionDisplayName(action) + menu.Title + "的权限"),
			Status:         true,
		}

		err = b.store.Permission().Create(ctx, permission)
		if err != nil {
			log.W(ctx).Errorw("Failed to create permission for menu", "menu_id", menu.ID, "action", action, "err", err)
			continue
		}

		createdPermissions = append(createdPermissions, permission)

		// 自动分配权限
		err = b.AutoAssignPermissionsToRoles(ctx, permission.ID, permission.TenantID)
		if err != nil {
			log.W(ctx).Errorw("Failed to auto assign permission", "permission_id", permission.ID, "err", err)
		}
	}

	log.W(ctx).Infow("Created menu with auto permissions",
		"menu_id", menu.ID,
		"menu_title", menu.Title,
		"permissions_created", len(createdPermissions))

	return nil
}

// getActionDisplayName 获取操作的显示名称
func getActionDisplayName(action string) string {
	actionNames := map[string]string{
		"view":   "查看",
		"create": "创建",
		"update": "编辑",
		"delete": "删除",
		"export": "导出",
		"import": "导入",
		"assign": "分配",
		"audit":  "审计",
	}

	if displayName, exists := actionNames[action]; exists {
		return displayName
	}

	return action
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// GetAdminPermissionPreview 预览管理员将获得的权限
func (b *permissionBiz) GetAdminPermissionPreview(ctx context.Context, permissionCode string) (bool, string) {
	config := DefaultAutoAssignConfig()

	// 模拟权限对象
	permission := &model.PermissionM{
		PermissionCode: permissionCode,
	}

	willAssign := b.shouldAssignToAdmin(permission, config.AdminPermissionRules)

	var reason string
	if willAssign {
		reason = "匹配管理员权限分配规则"
	} else {
		reason = "不匹配管理员权限分配规则或在排除列表中"
	}

	return willAssign, reason
}

// SyncAdminPermissions 同步管理员权限（补充遗漏的权限）
func (b *permissionBiz) SyncAdminPermissions(ctx context.Context, tenantID int64) (int, error) {
	config := DefaultAutoAssignConfig()

	if !config.Enabled || len(config.AdminRoleIDs) == 0 {
		return 0, nil
	}

	// 获取所有权限
	_, allPermissions, err := b.store.Permission().GetPermissionsByTenant(ctx, tenantID, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get permissions: %w", err)
	}

	var assignedCount int

	// 为每个管理员角色同步权限
	for _, roleID := range config.AdminRoleIDs {
		for _, permission := range allPermissions {
			// 检查是否应该分配给管理员
			if !b.shouldAssignToAdmin(permission, config.AdminPermissionRules) {
				continue
			}

			// 检查是否已经分配
			roleIdentifier := fmt.Sprintf("r%d", roleID)
			permissionIdentifier := fmt.Sprintf("p%d", permission.ID)
			tenantIdentifier := fmt.Sprintf("t%d", tenantID)

			hasPermission := b.authz.HasPermissionForUser(roleIdentifier, permissionIdentifier, tenantIdentifier)
			if hasPermission {
				continue
			}

			// 分配权限
			err := b.assignPermissionToRole(ctx, roleID, permission.ID, permission.TenantID)
			if err != nil {
				log.W(ctx).Errorw("Failed to sync permission for admin",
					"role_id", roleID,
					"permission_code", permission.PermissionCode,
					"err", err)
				continue
			}

			assignedCount++
		}
	}

	log.W(ctx).Infow("Synced admin permissions",
		"tenant_id", tenantID,
		"assigned_count", assignedCount)

	return assignedCount, nil
}

// AutoAssignPermissionsToSuperAdmin 自动为超级管理员分配新权限
func (b *permissionBiz) AutoAssignPermissionsToSuperAdmin(ctx context.Context, permissionID int64, tenantID int64) error {
	config := DefaultAutoAssignConfig()

	if !config.Enabled {
		return nil // 如果未启用自动分配，直接返回
	}

	// 检查是否需要为该租户自动分配权限
	if len(config.TenantIDs) > 0 {
		needAssign := false
		for _, tid := range config.TenantIDs {
			if tid == tenantID {
				needAssign = true
				break
			}
		}
		if !needAssign {
			return nil
		}
	}

	// 为所有超级管理员角色分配新权限
	for _, roleID := range config.SuperAdminRoleIDs {
		err := b.assignPermissionToRole(ctx, roleID, permissionID, tenantID)
		if err != nil {
			log.W(ctx).Errorw("Failed to auto assign permission to super admin role",
				"role_id", roleID,
				"permission_id", permissionID,
				"tenant_id", tenantID,
				"err", err)
			continue
		}

		log.W(ctx).Infow("Auto assigned permission to super admin role",
			"role_id", roleID,
			"permission_id", permissionID,
			"tenant_id", tenantID)
	}

	return nil
}

// assignPermissionToRole 为角色分配权限
func (b *permissionBiz) assignPermissionToRole(ctx context.Context, roleID, permissionID, tenantID int64) error {
	// 构建Casbin规则标识符
	roleIdentifier := fmt.Sprintf("r%d", roleID)
	permissionIdentifier := fmt.Sprintf("p%d", permissionID)
	tenantIdentifier := fmt.Sprintf("t%d", tenantID)

	// 检查权限是否已经分配
	hasPermission := b.authz.HasPermissionForUser(roleIdentifier, permissionIdentifier, tenantIdentifier)
	if hasPermission {
		return nil // 已经有该权限，无需重复分配
	}

	// 使用Casbin分配权限
	_, err := b.authz.AddPermissionForUser(roleIdentifier, permissionIdentifier, tenantIdentifier)
	if err != nil {
		return fmt.Errorf("failed to add permission to casbin: %w", err)
	}

	return nil
}

// SyncMissingPermissionsForSuperAdmin 同步超级管理员缺失的权限
func (b *permissionBiz) SyncMissingPermissionsForSuperAdmin(ctx context.Context, tenantID int64) error {
	config := DefaultAutoAssignConfig()

	if !config.Enabled {
		return nil
	}

	// 查询所有权限
	_, permissions, err := b.store.Permission().GetPermissionsByTenant(ctx, tenantID, nil)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	var assignedCount int

	// 使用数据库事务确保一致性
	err = b.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		for _, permission := range permissions {
			for _, roleID := range config.SuperAdminRoleIDs {
				err := b.assignPermissionToRole(ctx, roleID, permission.ID, permission.TenantID)
				if err != nil {
					log.W(ctx).Errorw("Failed to sync permission for super admin",
						"role_id", roleID,
						"permission_id", permission.ID,
						"permission_code", permission.PermissionCode,
						"err", err)
					continue
				}
				assignedCount++
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	log.W(ctx).Infow("Synced missing permissions for super admin",
		"tenant_id", tenantID,
		"total_permissions", len(permissions),
		"assigned_count", assignedCount)

	return nil
}

// CreatePermissionWithAutoAssign 创建权限并自动分配给超级管理员
func (b *permissionBiz) CreatePermissionWithAutoAssign(ctx context.Context, permission *model.PermissionM) error {
	// 创建权限
	err := b.store.Permission().Create(ctx, permission)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	// 自动分配给超级管理员
	err = b.AutoAssignPermissionsToSuperAdmin(ctx, permission.ID, permission.TenantID)
	if err != nil {
		log.W(ctx).Errorw("Failed to auto assign new permission to super admin",
			"permission_id", permission.ID,
			"permission_code", permission.PermissionCode,
			"err", err)
		// 不返回错误，因为权限已经创建成功
	}

	return nil
}

// GetSuperAdminMissingPermissions 获取超级管理员缺失的权限列表
func (b *permissionBiz) GetSuperAdminMissingPermissions(ctx context.Context, tenantID int64) ([]*model.PermissionM, error) {
	config := DefaultAutoAssignConfig()

	if len(config.SuperAdminRoleIDs) == 0 {
		return nil, nil
	}

	// 获取第一个超级管理员角色的权限（假设所有超级管理员角色权限相同）
	roleID := config.SuperAdminRoleIDs[0]
	roleIdentifier := fmt.Sprintf("r%d", roleID)
	tenantIdentifier := strconv.FormatInt(tenantID, 10)

	// 获取角色已有的权限
	existingPermissions := b.authz.GetPermissionsForUser(roleIdentifier, tenantIdentifier)
	existingPermissionMap := make(map[string]bool)

	for _, perm := range existingPermissions {
		if len(perm) >= 2 {
			existingPermissionMap[perm[1]] = true
		}
	}

	// 获取所有权限
	_, allPermissions, err := b.store.Permission().GetPermissionsByTenant(ctx, tenantID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions: %w", err)
	}

	// 找出缺失的权限
	var missingPermissions []*model.PermissionM
	for _, permission := range allPermissions {
		permissionIdentifier := fmt.Sprintf("p%d", permission.ID)
		if !existingPermissionMap[permissionIdentifier] {
			missingPermissions = append(missingPermissions, permission)
		}
	}

	return missingPermissions, nil
}
