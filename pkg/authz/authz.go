// Package authz 提供基于 Casbin 的多租户 RBAC 授权功能
//
// 使用示例:
//
//	// 使用默认RBAC模型
//	authz, err := NewAuthz(db)
//
//	// 使用自定义RBAC模型
//	authz, err := NewAuthz(db, WithRBACModel(customModel))
//
//	// 使用旧的ACL模型（已废弃）
//	authz, err := NewAuthz(db, WithAclModel(aclModel))
//
// 参考官方文档: https://casbin.org/zh/docs/rbac-with-domains
package authz

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/google/wire"
	"gorm.io/gorm"
)

const (
	// 默认的 Casbin RBAC with Domains 访问控制模型，支持多租户.
	// 参考官方文档: https://casbin.org/zh/docs/rbac-with-domains
	// 修正为与旧项目一致的标准格式：obj 在 dom 之前
	defaultRBACWithDomainsModel = `[request_definition]
r = sub, obj, dom

[policy_definition]
p = sub, obj, dom

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.obj == p.obj && r.dom == p.dom`
)

// TenantResolver 定义租户解析器接口，用于在租户标识符和租户ID之间转换
type TenantResolver interface {
	// GetTenantID 根据租户标识符获取租户ID
	GetTenantID(tenantIdentifier string) (int64, error)
	// GetTenantIdentifier 根据租户ID获取租户标识符
	GetTenantIdentifier(tenantID int64) (string, error)
	// GetRoleID 根据角色标识符和租户获取角色ID
	GetRoleID(roleIdentifier, tenantIdentifier string) (int64, error)
	// GetRoleIdentifier 根据角色ID获取角色标识符
	GetRoleIdentifier(roleID int64) (string, error)
	// GetPermissionID 根据权限标识符和租户获取权限ID
	GetPermissionID(permissionIdentifier, tenantIdentifier string) (int64, error)
	// GetPermissionIdentifier 根据权限ID获取权限标识符
	GetPermissionIdentifier(permissionID int64) (string, error)
}

// DefaultTenantResolver 默认租户解析器实现
type DefaultTenantResolver struct {
	db *gorm.DB
}

// NewDefaultTenantResolver 创建默认租户解析器
func NewDefaultTenantResolver(db *gorm.DB) *DefaultTenantResolver {
	return &DefaultTenantResolver{db: db}
}

// GetTenantID 根据租户标识符获取租户ID
func (r *DefaultTenantResolver) GetTenantID(tenantIdentifier string) (int64, error) {
	// 如果是"default"，返回ID 1
	if tenantIdentifier == "default" || tenantIdentifier == "" {
		return 1, nil
	}

	// 如果是带前缀的格式 t{id}，直接解析
	if strings.HasPrefix(tenantIdentifier, "t") {
		if id, err := strconv.ParseInt(tenantIdentifier[1:], 10, 64); err == nil {
			return id, nil
		}
	}

	// 尝试直接解析为数字
	if id, err := strconv.ParseInt(tenantIdentifier, 10, 64); err == nil {
		return id, nil
	}

	// 从数据库查询
	var tenant struct {
		ID int64 `gorm:"column:id"`
	}

	err := r.db.Table("tenants").
		Select("id").
		Where("tenant_code = ? AND deleted_at IS NULL", tenantIdentifier).
		First(&tenant).Error

	if err != nil {
		// 如果找不到，返回默认租户ID
		return 1, err
	}

	return tenant.ID, nil
}

// GetTenantIdentifier 根据租户ID获取租户标识符
func (r *DefaultTenantResolver) GetTenantIdentifier(tenantID int64) (string, error) {
	// 如果是ID 1，返回"default"
	if tenantID == 1 {
		return "default", nil
	}

	// 直接使用租户ID作为标识符，格式：t{id}
	return fmt.Sprintf("t%d", tenantID), nil
}

// GetRoleID 根据角色标识符和租户获取角色ID
func (r *DefaultTenantResolver) GetRoleID(roleIdentifier, tenantIdentifier string) (int64, error) {
	// 如果是带前缀的格式 r{id}，直接解析
	if strings.HasPrefix(roleIdentifier, "r") {
		if id, err := strconv.ParseInt(roleIdentifier[1:], 10, 64); err == nil {
			return id, nil
		}
	}

	// 尝试直接解析为数字
	if id, err := strconv.ParseInt(roleIdentifier, 10, 64); err == nil {
		return id, nil
	}

	// 获取租户ID
	tenantID, err := r.GetTenantID(tenantIdentifier)
	if err != nil {
		tenantID = 1 // 使用默认租户
	}

	// 从数据库查询，使用角色名称而不是role_code
	var role struct {
		ID int64 `gorm:"column:id"`
	}

	err = r.db.Table("roles").
		Select("id").
		Where("name = ? AND tenant_id = ? AND deleted_at IS NULL",
			roleIdentifier, tenantID).
		First(&role).Error

	if err != nil {
		return 0, err
	}

	return role.ID, nil
}

// GetRoleIdentifier 根据角色ID获取角色标识符
func (r *DefaultTenantResolver) GetRoleIdentifier(roleID int64) (string, error) {
	var role struct {
		Name string `gorm:"column:name"`
	}

	err := r.db.Table("roles").
		Select("name").
		Where("id = ? AND deleted_at IS NULL", roleID).
		First(&role).Error

	if err != nil {
		return "", err
	}

	return role.Name, nil
}

// GetPermissionID 根据权限标识符和租户获取权限ID
func (r *DefaultTenantResolver) GetPermissionID(permissionIdentifier, tenantIdentifier string) (int64, error) {
	// 如果是带前缀的格式 a{id}，直接解析
	if strings.HasPrefix(permissionIdentifier, "a") {
		if id, err := strconv.ParseInt(permissionIdentifier[1:], 10, 64); err == nil {
			return id, nil
		}
	}

	// 尝试直接解析为数字
	if id, err := strconv.ParseInt(permissionIdentifier, 10, 64); err == nil {
		return id, nil
	}

	// 获取租户ID
	tenantID, err := r.GetTenantID(tenantIdentifier)
	if err != nil {
		tenantID = 1 // 使用默认租户
	}

	// 从数据库查询，使用权限名称而不是permission_code
	var permission struct {
		ID int64 `gorm:"column:id"`
	}

	err = r.db.Table("permissions").
		Select("id").
		Where("name = ? AND tenant_id = ? AND deleted_at IS NULL",
			permissionIdentifier, tenantID).
		First(&permission).Error

	if err != nil {
		return 0, err
	}

	return permission.ID, nil
}

// GetPermissionIdentifier 根据权限ID获取权限标识符
func (r *DefaultTenantResolver) GetPermissionIdentifier(permissionID int64) (string, error) {
	var permission struct {
		Name string `gorm:"column:name"`
	}

	err := r.db.Table("permissions").
		Select("name").
		Where("id = ? AND deleted_at IS NULL", permissionID).
		First(&permission).Error

	if err != nil {
		return "", err
	}

	return permission.Name, nil
}

// Authz 定义了一个授权器，提供授权功能.
type Authz struct {
	*casbin.SyncedCachedEnforcer                // 使用 Casbin 的同步缓存授权器（参考旧项目）
	tenantResolver               TenantResolver // 租户解析器
	idConverter                  *IDConverter   // ID转换器（参考旧项目实现，统一使用）
}

// Option 定义了一个函数选项类型，用于自定义 NewAuthz 的行为.
type Option func(*authzConfig)

// authzConfig 是授权器的配置结构.
type authzConfig struct {
	model              string        // Casbin 的模型字符串（统一字段名）
	autoLoadPolicyTime time.Duration // 自动加载策略的时间间隔
}

// ProviderSet 是一个 Wire 的 Provider 集合，用于声明依赖注入的规则。
// 包含 NewAuthz 构造函数，用于生成 Authz 实例。
var ProviderSet = wire.NewSet(NewAuthz, DefaultOptions)

// defaultAuthzConfig 返回默认的授权器配置，使用RBAC with Domains模型
func defaultAuthzConfig() *authzConfig {
	return &authzConfig{
		model:              defaultRBACWithDomainsModel,
		autoLoadPolicyTime: 10 * time.Second,
	}
}

// DefaultOptions 返回默认选项
func DefaultOptions() []Option {
	return []Option{
		WithRBACModel(defaultRBACWithDomainsModel),
		WithAutoLoadPolicyTime(10 * time.Second),
	}
}

// WithAclModel 允许通过选项自定义 ACL 模型（已废弃，建议使用WithRBACModel）
// Deprecated: 使用 WithRBACModel 替代
func WithAclModel(model string) Option {
	return func(cfg *authzConfig) {
		cfg.model = model
	}
}

// WithRBACModel 允许通过选项自定义 RBAC 模型
func WithRBACModel(model string) Option {
	return func(cfg *authzConfig) {
		cfg.model = model
	}
}

// WithAutoLoadPolicyTime 允许通过选项自定义自动加载策略的时间间隔.
func WithAutoLoadPolicyTime(interval time.Duration) Option {
	return func(cfg *authzConfig) {
		cfg.autoLoadPolicyTime = interval
	}
}

// NewAuthz 创建一个使用 Casbin 完成授权的授权器，通过函数选项模式支持自定义配置.
func NewAuthz(db *gorm.DB, opts ...Option) (*Authz, error) {
	// 初始化默认配置
	cfg := defaultAuthzConfig()

	// 应用所有选项
	for _, opt := range opts {
		opt(cfg)
	}

	// 初始化 Gorm 适配器并用于 Casbin 授权器
	adapter, err := adapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err // 返回错误
	}

	// 从配置中加载 Casbin 模型
	m, _ := model.NewModelFromString(cfg.model)

	// 初始化授权器（使用缓存版本，参考旧项目）
	enforcer, err := casbin.NewSyncedCachedEnforcer(m, adapter)
	if err != nil {
		return nil, err // 返回错误
	}

	// 从数据库加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err // 返回错误
	}

	// 启动自动加载策略，使用配置的时间间隔
	enforcer.StartAutoLoadPolicy(cfg.autoLoadPolicyTime)

	// 创建默认租户解析器和ID转换器
	tenantResolver := NewDefaultTenantResolver(db)
	idConverter := NewIDConverter()

	// 返回新的授权器实例
	return &Authz{
		SyncedCachedEnforcer: enforcer,
		tenantResolver:       tenantResolver,
		idConverter:          idConverter,
	}, nil
}

// resolveTenantDomain 将租户标识符转换为租户ID字符串用于Casbin domain
func (a *Authz) resolveTenantDomain(tenantIdentifier string) string {
	tenantID, err := a.tenantResolver.GetTenantID(tenantIdentifier)
	if err != nil {
		// 如果解析失败，使用默认租户ID
		return "t1"
	}
	return "t" + strconv.FormatInt(tenantID, 10)
}

// resolveRoleSubject 将角色标识符转换为角色ID字符串用于Casbin subject
func (a *Authz) resolveRoleSubject(roleIdentifier, tenantIdentifier string) string {
	roleID, err := a.tenantResolver.GetRoleID(roleIdentifier, tenantIdentifier)
	if err != nil {
		// 如果解析失败，返回原始标识符
		return roleIdentifier
	}
	return "r" + strconv.FormatInt(roleID, 10)
}

// Authorize 使用domain进行授权检查（原Authorize方法已废弃，统一使用domain方式）
func (a *Authz) Authorize(sub, obj, act string) (bool, error) {
	// 使用默认domain进行授权检查
	return a.AuthorizeWithDomain(sub, "default", obj, act)
}

// AuthorizeWithDomain 使用domain进行授权检查
func (a *Authz) AuthorizeWithDomain(sub, tenantIdentifier, obj, act string) (bool, error) {
	// 将租户标识符转换为租户ID
	domain := a.resolveTenantDomain(tenantIdentifier)

	// 根据用户名查找实际的用户ID
	var result struct {
		ID int64 `gorm:"column:id"`
	}
	err := a.tenantResolver.(*DefaultTenantResolver).db.Table("user").
		Select("id").
		Where("username = ? AND deleted_at IS NULL", sub).
		First(&result).Error

	userID := result.ID

	if err != nil {
		// 如果找不到用户，尝试解析为数字ID
		if id, parseErr := strconv.ParseInt(sub, 10, 64); parseErr == nil {
			userID = id
		} else {
			return false, fmt.Errorf("user not found: %s", sub)
		}
	}

	// 格式化用户ID
	formattedUserID := a.idConverter.ToDUserID(userID)
	// 调用 Enforce 方法进行授权检查，修正参数顺序为：sub, obj, dom
	return a.Enforce(formattedUserID, obj, domain)
}

// AddRoleForUser 为用户在指定domain中添加角色
func (a *Authz) AddRoleForUser(user, roleIdentifier, tenantIdentifier string) (bool, error) {
	domain := a.resolveTenantDomain(tenantIdentifier)

	// 获取实际的用户ID
	var userResult struct {
		ID int64 `gorm:"column:id"`
	}
	err := a.tenantResolver.(*DefaultTenantResolver).db.Table("user").
		Select("id").
		Where("username = ? AND deleted_at IS NULL", user).
		First(&userResult).Error

	userID := userResult.ID

	if err != nil {
		// 如果找不到用户，尝试解析为数字ID
		if id, parseErr := strconv.ParseInt(user, 10, 64); parseErr == nil {
			userID = id
		} else {
			return false, fmt.Errorf("user not found: %s", user)
		}
	}

	// 获取实际的角色ID
	roleID, err := a.tenantResolver.GetRoleID(roleIdentifier, tenantIdentifier)
	if err != nil {
		return false, fmt.Errorf("role not found: %s", roleIdentifier)
	}

	// 格式化ID
	formattedUserID := a.idConverter.ToDUserID(userID)
	formattedRoleID := a.idConverter.ToDRoleID(roleID)

	return a.AddGroupingPolicy(formattedUserID, formattedRoleID, domain)
}

// DeleteRoleForUser 移除用户在指定domain中的角色
func (a *Authz) DeleteRoleForUser(user, roleIdentifier, tenantIdentifier string) (bool, error) {
	domain := a.resolveTenantDomain(tenantIdentifier)
	role := a.resolveRoleSubject(roleIdentifier, tenantIdentifier)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))
	return a.RemoveGroupingPolicy(userID, role, domain)
}

// GetRolesForUser 获取用户在指定domain中的角色
func (a *Authz) GetRolesForUser(user, tenantIdentifier string) ([]string, error) {
	domain := a.resolveTenantDomain(tenantIdentifier)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))
	roleIDs, err := a.SyncedCachedEnforcer.GetRolesForUser(userID, domain)
	if err != nil {
		return nil, err
	}

	// 将角色ID转换为角色标识符
	var roles []string
	for _, roleIDStr := range roleIDs {
		// 如果是带前缀的格式 r{id}，解析并转换
		if strings.HasPrefix(roleIDStr, "r") {
			if roleID, err := strconv.ParseInt(roleIDStr[1:], 10, 64); err == nil {
				if roleIdentifier, err := a.tenantResolver.GetRoleIdentifier(roleID); err == nil {
					roles = append(roles, roleIdentifier)
				} else {
					roles = append(roles, roleIDStr) // 如果转换失败，使用原始值
				}
			} else {
				roles = append(roles, roleIDStr) // 如果解析失败，使用原始值
			}
		} else if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
			// 兼容纯数字格式
			if roleIdentifier, err := a.tenantResolver.GetRoleIdentifier(roleID); err == nil {
				roles = append(roles, roleIdentifier)
			} else {
				roles = append(roles, roleIDStr) // 如果转换失败，使用原始值
			}
		} else {
			roles = append(roles, roleIDStr) // 如果不是数字，使用原始值
		}
	}

	return roles, nil
}

// GetUsersForRole 获取在指定domain中拥有指定角色的用户
func (a *Authz) GetUsersForRole(roleIdentifier, tenantIdentifier string) ([]string, error) {
	domain := a.resolveTenantDomain(tenantIdentifier)
	role := a.resolveRoleSubject(roleIdentifier, tenantIdentifier)
	userIDs, err := a.SyncedCachedEnforcer.GetUsersForRole(role, domain)
	if err != nil {
		return nil, err
	}

	// 将用户ID转换为用户标识符（如果需要的话）
	var users []string
	for _, userIDStr := range userIDs {
		// 如果是带前缀的格式 u{id}，解析并转换
		if strings.HasPrefix(userIDStr, "u") {
			if userID, err := strconv.ParseInt(userIDStr[1:], 10, 64); err == nil {
				// 这里可以添加用户ID到用户标识符的转换逻辑
				// 暂时直接使用用户ID
				users = append(users, strconv.FormatInt(userID, 10))
			} else {
				users = append(users, userIDStr) // 如果解析失败，使用原始值
			}
		} else {
			users = append(users, userIDStr) // 如果不是带前缀格式，使用原始值
		}
	}

	return users, nil
}

// GetAllUsersByDomain 获取指定domain中的所有用户
func (a *Authz) GetAllUsersByDomain(tenantIdentifier string) ([]string, error) {
	domain := a.resolveTenantDomain(tenantIdentifier)
	userIDs, err := a.SyncedCachedEnforcer.GetAllUsersByDomain(domain)
	if err != nil {
		return nil, err
	}

	// 将用户ID转换为用户标识符（如果需要的话）
	var users []string
	for _, userIDStr := range userIDs {
		// 如果是带前缀的格式 u{id}，解析并转换
		if strings.HasPrefix(userIDStr, "u") {
			if userID, err := strconv.ParseInt(userIDStr[1:], 10, 64); err == nil {
				// 这里可以添加用户ID到用户标识符的转换逻辑
				// 暂时直接使用用户ID
				users = append(users, strconv.FormatInt(userID, 10))
			} else {
				users = append(users, userIDStr) // 如果解析失败，使用原始值
			}
		} else {
			users = append(users, userIDStr) // 如果不是带前缀格式，使用原始值
		}
	}

	return users, nil
}

// DeleteAllRolesForUser 删除用户在指定domain中的所有角色
func (a *Authz) DeleteAllRolesForUser(user, tenantIdentifier string) bool {
	domain := a.resolveTenantDomain(tenantIdentifier)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))
	result, _ := a.RemoveFilteredGroupingPolicy(0, userID, "", domain)
	return result
}

// DeleteRole 删除指定domain中的角色
func (a *Authz) DeleteRole(roleIdentifier, tenantIdentifier string) bool {
	domain := a.resolveTenantDomain(tenantIdentifier)
	role := a.resolveRoleSubject(roleIdentifier, tenantIdentifier)
	// 删除角色的所有权限
	res1, _ := a.RemoveFilteredPolicy(0, role, domain)
	// 删除所有用户与该角色的关联
	res2, _ := a.RemoveFilteredGroupingPolicy(1, role, domain)
	return res1 || res2
}

// AddPermissionForUser 为用户在指定租户下添加权限（参考旧项目API）
func (a *Authz) AddPermissionForUser(user, permission, tenant string) (bool, error) {
	domain := a.resolveTenantDomain(tenant)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))
	permissionID := a.idConverter.ToDResourceID(a.parseStringToInt64(permission))

	return a.AddPolicy(userID, permissionID, domain)
}

// DeletePermissionForUser 删除用户在指定租户下的权限（参考旧项目API）
func (a *Authz) DeletePermissionForUser(user, permission, tenant string) (bool, error) {
	domain := a.resolveTenantDomain(tenant)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))
	permissionID := a.idConverter.ToDResourceID(a.parseStringToInt64(permission))

	return a.RemovePolicy(userID, permissionID, domain)
}

// GetPermissionsForUser 获取用户在指定租户下的权限列表
func (a *Authz) GetPermissionsForUser(user, tenant string) [][]string {
	domain := a.resolveTenantDomain(tenant)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))

	policies, _ := a.GetFilteredPolicy(0, userID, domain)
	return policies
}

// HasPermissionForUser 检查用户在指定租户中是否有权限（参考旧项目API）
func (a *Authz) HasPermissionForUser(user, permission, tenant string) bool {
	domain := a.resolveTenantDomain(tenant)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))
	permissionID := a.idConverter.ToDResourceID(a.parseStringToInt64(permission))

	has, _ := a.HasPolicy(userID, permissionID, domain)
	return has
}

// GetImplicitRolesForUser 获取用户在指定domain中的所有角色（包括继承的角色）
func (a *Authz) GetImplicitRolesForUser(user, tenantIdentifier string) []string {
	domain := a.resolveTenantDomain(tenantIdentifier)
	roleIDs, _ := a.SyncedCachedEnforcer.GetImplicitRolesForUser(a.idConverter.ToDUserID(a.parseStringToInt64(user)), domain)

	// 将角色ID转换为角色标识符
	var roles []string
	for _, roleIDStr := range roleIDs {
		// 如果是带前缀的格式 r{id}，解析并转换
		if strings.HasPrefix(roleIDStr, "r") {
			if roleID, err := strconv.ParseInt(roleIDStr[1:], 10, 64); err == nil {
				if roleIdentifier, err := a.tenantResolver.GetRoleIdentifier(roleID); err == nil {
					roles = append(roles, roleIdentifier)
				} else {
					roles = append(roles, roleIDStr) // 如果转换失败，使用原始值
				}
			} else {
				roles = append(roles, roleIDStr) // 如果解析失败，使用原始值
			}
		} else if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
			// 兼容纯数字格式
			if roleIdentifier, err := a.tenantResolver.GetRoleIdentifier(roleID); err == nil {
				roles = append(roles, roleIdentifier)
			} else {
				roles = append(roles, roleIDStr) // 如果转换失败，使用原始值
			}
		} else {
			roles = append(roles, roleIDStr) // 如果不是数字，使用原始值
		}
	}

	return roles
}

// GetImplicitPermissionsForUser 获取用户在指定租户下的隐式权限（包括通过角色继承的权限）
func (a *Authz) GetImplicitPermissionsForUser(user, tenant string) ([][]string, error) {
	domain := a.resolveTenantDomain(tenant)
	userID := a.idConverter.ToDUserID(a.parseStringToInt64(user))

	return a.SyncedCachedEnforcer.GetImplicitPermissionsForUser(userID, domain)
}

// GetPermissionsForUserInDomain 获取用户在指定域中的权限（参考旧项目API）
func (a *Authz) GetPermissionsForUserInDomain(user, domain string) [][]string {
	policies, _ := a.SyncedCachedEnforcer.GetPermissionsForUser(user, domain)
	return policies
}

// parseStringToInt64 辅助方法：将字符串转换为int64，如果失败则返回0
func (a *Authz) parseStringToInt64(s string) int64 {
	if id, err := strconv.ParseInt(s, 10, 64); err == nil {
		return id
	}
	return 0
}

// CheckPermission 检查用户是否具有指定权限
func (a *Authz) CheckPermission(userID, tenantIdentifier, permissionCode string) (bool, error) {
	// 首先检查用户是否为超级管理员
	isSuperAdmin, err := a.isSuperAdmin(userID, tenantIdentifier)
	if err != nil {
		return false, err
	}

	// 超级管理员拥有所有权限，直接返回true
	if isSuperAdmin {
		return true, nil
	}

	// 普通用户需要检查具体权限
	return a.checkSpecificPermission(userID, tenantIdentifier, permissionCode)
}

// checkSpecificPermission 检查普通用户的具体权限
func (a *Authz) checkSpecificPermission(userID, tenantIdentifier, permissionCode string) (bool, error) {
	// 将租户标识符转换为租户ID
	tenantID, err := a.tenantResolver.GetTenantID(tenantIdentifier)
	if err != nil {
		return false, err
	}

	// 查询权限ID
	permissionID, err := a.tenantResolver.GetPermissionID(permissionCode, tenantIdentifier)
	if err != nil {
		// 如果权限不存在，普通用户没有该权限
		return false, nil
	}

	// 构建用户、权限和租户标识符
	userIdentifier := fmt.Sprintf("u%s", userID)
	permissionIdentifier := fmt.Sprintf("p%d", permissionID)
	domain := fmt.Sprintf("t%d", tenantID)

	// 使用Casbin进行权限检查
	return a.Enforce(userIdentifier, permissionIdentifier, domain)
}

// isSuperAdmin 检查用户是否为超级管理员
func (a *Authz) isSuperAdmin(userID, tenantIdentifier string) (bool, error) {
	// 获取用户在指定租户下的角色
	roles, err := a.GetRolesForUser(userID, tenantIdentifier)
	if err != nil {
		return false, err
	}

	// 检查是否包含超级管理员角色
	for _, role := range roles {
		if role == "super_admin" || role == "r1" {
			return true, nil
		}
	}

	return false, nil
}

// CheckMenuAccess 检查菜单访问权限（也支持超级管理员免检）
func (a *Authz) CheckMenuAccess(userID, tenantIdentifier string, menuID int64) (bool, error) {
	// 超级管理员可以访问所有菜单
	if isSuperAdmin, _ := a.isSuperAdmin(userID, tenantIdentifier); isSuperAdmin {
		return true, nil
	}

	// 普通用户检查菜单权限
	return a.checkMenuAccessForRegularUser(userID, tenantIdentifier, menuID)
}

// checkMenuAccessForRegularUser 检查普通用户的菜单访问权限
func (a *Authz) checkMenuAccessForRegularUser(userID, tenantIdentifier string, menuID int64) (bool, error) {
	// 查询菜单相关的权限
	var results []struct {
		PermissionCode string `gorm:"column:permission_code"`
		IsRequired     bool   `gorm:"column:is_required"`
	}

	err := a.tenantResolver.(*DefaultTenantResolver).db.Table("menu_permissions mp").
		Select("p.permission_code, mp.is_required").
		Joins("JOIN permissions p ON p.id = mp.permission_id").
		Where("mp.menu_id = ? AND p.deleted_at IS NULL", menuID).
		Find(&results).Error

	if err != nil {
		return false, err
	}

	// 如果菜单没有配置权限，默认允许访问
	if len(results) == 0 {
		return true, nil
	}

	// 检查必需权限
	for _, result := range results {
		if result.IsRequired {
			hasPermission, err := a.CheckPermission(userID, tenantIdentifier, result.PermissionCode)
			if err != nil || !hasPermission {
				return false, err
			}
		}
	}

	return true, nil
}

// CheckAPIPermission 检查API访问权限（也支持超级管理员免检）
func (a *Authz) CheckAPIPermission(userID, tenantIdentifier, accessPath, httpMethod string) (bool, error) {
	// 超级管理员可以访问所有API
	if isSuperAdmin, _ := a.isSuperAdmin(userID, tenantIdentifier); isSuperAdmin {
		return true, nil
	}

	// 普通用户检查API权限
	return a.checkAPIAccessForRegularUser(userID, tenantIdentifier, accessPath, httpMethod)
}

// CheckAPIAccess 实现APIAuthorizer接口（适配中间件）
func (a *Authz) CheckAPIAccess(subject, domain, object, action string) (bool, error) {
	// 从domain中解析租户标识符
	tenantIdentifier := domain
	if domain == "default" || domain == "" {
		tenantIdentifier = "t1" // 默认租户
	}

	// 调用原有的API权限检查方法
	return a.CheckAPIPermission(subject, tenantIdentifier, object, action)
}

// checkAPIAccessForRegularUser 检查普通用户的API访问权限
func (a *Authz) checkAPIAccessForRegularUser(userID, tenantIdentifier, accessPath, httpMethod string) (bool, error) {
	// 查询API对应的权限
	var permissions []struct {
		PermissionCode string `gorm:"column:permission_code"`
	}

	err := a.tenantResolver.(*DefaultTenantResolver).db.Table("permissions").
		Select("permission_code").
		Where("resource_type = 'api' AND resource_path = ? AND http_method = ? AND deleted_at IS NULL",
			accessPath, httpMethod).
		Find(&permissions).Error

	if err != nil {
		return false, err
	}

	// 如果API没有配置权限，根据默认策略决定
	if len(permissions) == 0 {
		// 可以配置为默认拒绝或默认允许
		return false, nil // 默认拒绝未配置权限的API
	}

	// 检查用户是否拥有其中任一权限
	for _, perm := range permissions {
		hasPermission, err := a.CheckPermission(userID, tenantIdentifier, perm.PermissionCode)
		if err != nil {
			continue
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// CheckMenuPermissionByCode 通过权限编码检查用户权限
func (a *Authz) CheckMenuPermissionByCode(userID, tenantIdentifier, permissionCode string) (bool, error) {
	// 解析租户域
	domain := a.resolveTenantDomain(tenantIdentifier)

	// 获取权限ID（通过权限码）
	permissionID, err := a.tenantResolver.GetPermissionID(permissionCode, tenantIdentifier)
	if err != nil {
		return false, fmt.Errorf("无法获取权限ID: %w", err)
	}

	// 格式化用户ID和权限ID
	formattedUserID := a.idConverter.ToDUserID(a.parseStringToInt64(userID))
	permissionIDStr := a.idConverter.ToDResourceID(permissionID)

	// 检查权限
	return a.Enforce(formattedUserID, permissionIDStr, domain)
}

// GetUserAccessibleMenus 获取用户可访问的菜单列表
func (a *Authz) GetUserAccessibleMenus(userID, tenantIdentifier string) ([]map[string]interface{}, error) {
	// 解析租户ID
	tenantID, err := a.tenantResolver.GetTenantID(tenantIdentifier)
	if err != nil {
		return nil, fmt.Errorf("无法解析租户ID: %w", err)
	}

	formattedUserID := a.idConverter.ToDUserID(a.parseStringToInt64(userID))
	formattedTenantID := a.idConverter.ToDDomainID(tenantID)

	// 查询用户可访问的菜单
	var menus []map[string]interface{}
	err = a.tenantResolver.(*DefaultTenantResolver).db.Table("v_user_accessible_menus").
		Where("user_id = ? AND tenant_id = ?", formattedUserID, formattedTenantID).
		Find(&menus).Error

	return menus, err
}

// GetMenuPermissions 获取菜单的所有权限
func (a *Authz) GetMenuPermissions(menuID int64) ([]map[string]interface{}, error) {
	var permissions []map[string]interface{}
	err := a.tenantResolver.(*DefaultTenantResolver).db.Table("permissions").
		Where("menu_id = ? AND status = 1 AND deleted_at IS NULL", menuID).
		Find(&permissions).Error

	return permissions, err
}

// AddMenuPermission 为菜单添加权限
func (a *Authz) AddMenuPermission(menuID int64, actionType, permissionName, description string) error {
	// 首先获取菜单的租户ID
	var tenantID int64
	err := a.tenantResolver.(*DefaultTenantResolver).db.Table("menus").
		Select("tenant_id").
		Where("id = ? AND deleted_at IS NULL", menuID).
		First(&tenantID).Error

	if err != nil {
		return fmt.Errorf("failed to get menu tenant: %w", err)
	}

	permissionCode := fmt.Sprintf("menu_%d_%s", menuID, actionType)

	permission := map[string]interface{}{
		"tenant_id":       tenantID,
		"menu_id":         menuID,
		"permission_code": permissionCode,
		"name":            permissionName,
		"description":     description,
		"status":          1,
	}

	return a.tenantResolver.(*DefaultTenantResolver).db.Table("permissions").Create(permission).Error
}

// RemoveMenuPermission 移除菜单权限
func (a *Authz) RemoveMenuPermission(menuID int64, actionType string) error {
	permissionCode := fmt.Sprintf("menu_%d_%s", menuID, actionType)
	return a.tenantResolver.(*DefaultTenantResolver).db.Table("permissions").
		Where("menu_id = ? AND permission_code = ?", menuID, permissionCode).
		Update("deleted_at", time.Now()).Error
}

// UpdateMenuAccessConfig 更新菜单访问配置
func (a *Authz) UpdateMenuAccessConfig(menuID int64, apiPath, httpMethods string, requireAuth bool) error {
	updates := map[string]interface{}{
		"api_path":     apiPath,
		"http_methods": httpMethods,
		"require_auth": requireAuth,
	}

	return a.tenantResolver.(*DefaultTenantResolver).db.Table("menus").
		Where("id = ? AND deleted_at IS NULL", menuID).
		Updates(updates).Error
}
