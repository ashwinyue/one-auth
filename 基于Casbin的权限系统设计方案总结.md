# 基于Casbin的权限系统设计方案总结

## 1. 系统架构概览

系统采用了基于Casbin的多租户RBAC (Role-Based Access Control) 权限模型，实现了完整的权限管理功能。主要特点包括：

- **多租户隔离**：不同租户之间的权限完全隔离
- **角色管理**：支持在不同租户中为用户分配不同角色
- **权限继承**：用户通过角色获得权限，支持隐式权限查询
- **跨租户支持**：同一用户可以在不同租户中拥有不同角色和权限

## 2. 核心模型设计

### 2.1 Casbin RBAC with Domains 模型

系统采用了标准的Casbin RBAC with Domains模型，配置如下：

```conf
[request_definition]
r = sub, obj, dom

[policy_definition]
p = sub, obj, dom

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.obj == p.obj && r.dom == p.dom
```

### 2.2 主键ID设计优化

系统对Casbin的主键ID设计进行了优化，从原来的字符串代码映射改为直接使用数据库主键ID：

**原有设计问题**：
```sql
-- 使用字符串代码映射，不一致且难维护
('g','admin','r1','t1',NULL,'',''),      -- 用户名 -> 角色代码 -> 租户代码
('p','r1','a1','t1',NULL,'',''),         -- 角色代码 -> 权限代码 -> 租户代码
```

**新设计优势**：
```sql
-- 直接使用数据库主键ID，简洁高效
('g','1','1','1',NULL,'',''),            -- user.id -> role.id -> tenant.id
('p','1','1','1',NULL,'',''),            -- role.id -> permission.id -> tenant.id
```

优势：
1. **性能更优**：直接使用主键，无需额外映射转换
2. **一致性强**：所有关联都使用主键ID
3. **维护简单**：无需维护额外的代码映射表

## 3. 权限类型分类

系统将权限资源分为四种类型：

```go
type ResourceType string

const (
    ResourceTypeAPI     ResourceType = "api"     // API接口权限
    ResourceTypeMenu    ResourceType = "menu"    // 菜单权限
    ResourceTypeData    ResourceType = "data"    // 数据权限
    ResourceTypeFeature ResourceType = "feature" // 功能权限
)
```

## 4. 菜单权限实现方案

### 4.1 设计特点

- **需要关联表**：`menu_permissions`
- **多对多关系**：一个菜单多个权限，一个权限多个菜单
- **权限分级**：必需权限 vs 可选权限
- **动态UI控制**：根据权限显示/隐藏功能

### 4.2 菜单权限关联表

```sql
CREATE TABLE `menu_permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `menu_id` bigint NOT NULL COMMENT '菜单ID',
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `is_required` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为访问菜单的必需权限',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_menu_permission` (`menu_id`, `permission_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_menu_id` (`menu_id`),
  KEY `idx_permission_id` (`permission_id`),
  KEY `idx_required` (`is_required`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单权限关联表';
```

### 4.3 菜单权限矩阵

系统实现了菜单权限矩阵，用于管理菜单的必需权限和可选权限：

```go
// 菜单权限矩阵
type MenuPermissionMatrix struct {
    Menu                *MenuM            `json:"menu"`
    RequiredPermissions []*PermissionNewM `json:"required_permissions"`
    OptionalPermissions []*PermissionNewM `json:"optional_permissions"`
    AllPermissions      []*PermissionNewM `json:"all_permissions"`
}
```

## 5. 核心组件实现

### 5.1 授权器 (Authz)

系统实现了完整的授权器，提供了丰富的授权功能：

```go
type Authz struct {
    *casbin.SyncedCachedEnforcer                // 使用 Casbin 的同步缓存授权器
    tenantResolver               TenantResolver // 租户解析器
    idConverter                  *IDConverter   // ID转换器
}
```

主要方法：
- `AuthorizeWithDomain(sub, dom, obj, act)` - 多租户授权检查
- `AddRoleForUser(user, role, domain)` - 为用户添加角色
- `GetRolesForUser(user, domain)` - 获取用户角色
- `GetPermissionsForUser(user, domain)` - 获取用户权限

### 5.2 ID转换器 (IDConverter)

系统实现了ID转换器，用于处理不同类型ID的转换：

```go
type IDConverter struct{}

// 角色ID转换
func (c *IDConverter) ToDRoleID(roleID int64) string
func (c *IDConverter) ToRoleID(role string) int64

// 租户/域ID转换
func (c *IDConverter) ToDDomainID(domainID int64) string
func (c *IDConverter) ToDomainID(domain string) int64

// 用户ID转换
func (c *IDConverter) ToDUserID(userID int64) string
func (c *IDConverter) ToUserID(user string) int64

// 资源ID转换
func (c *IDConverter) ToDResourceID(resourceID int64) string
func (c *IDConverter) ToResourceID(resource string) int64
```

### 5.3 权限中间件 (AuthzMiddleware)

系统实现了Gin框架的权限中间件，用于HTTP请求的权限验证：

```go
func AuthzMiddleware(authorizer Authorizer) gin.HandlerFunc {
    return func(c *gin.Context) {
        subject := contextx.UserID(c.Request.Context())
        domain := contextx.TenantID(c.Request.Context()) // 获取租户ID作为domain
        object := c.Request.URL.Path
        action := c.Request.Method

        // 如果没有租户ID，使用默认租户
        if domain == "" {
            domain = "default"
        }

        // 权限检查
        allowed, err := authorizer.AuthorizeWithDomain(subject, domain, object, action)
        if err != nil || !allowed {
            // 返回权限拒绝错误
            c.Abort()
            return
        }

        c.Next() // 继续处理请求
    }
}
```

## 6. 权限标准化设计

### 6.1 权限编码规范

采用 `{module}:{action}` 格式：
- `user:view` - 查看用户
- `user:create` - 创建用户
- `role:assign` - 分配角色
- `dashboard:view` - 查看仪表板

### 6.2 操作类型标准化

- **view** - 查看/读取
- **create** - 创建/新增
- **update** - 编辑/修改
- **delete** - 删除/移除
- **export** - 导出数据
- **assign** - 分配关系

## 7. 数据库表结构

### 7.1 核心权限表

```sql
CREATE TABLE `permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '权限主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `permission_code` varchar(100) NOT NULL COMMENT '权限编码（格式：module:action）',
  `name` varchar(100) NOT NULL COMMENT '权限名称',
  `description` varchar(500) DEFAULT NULL COMMENT '权限描述',
  `resource_type` enum('api','menu','data','feature') NOT NULL DEFAULT 'menu' COMMENT '资源类型',
  `resource_path` varchar(255) DEFAULT NULL COMMENT 'API路径或资源标识',
  `http_method` varchar(20) DEFAULT NULL COMMENT 'HTTP方法：GET,POST,PUT,DELETE等',
  `action` varchar(50) DEFAULT NULL COMMENT '操作类型：view,create,update,delete,export等',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_permission_code_tenant` (`permission_code`, `tenant_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_resource_type` (`resource_type`),
  KEY `idx_action` (`action`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表（重构版-独立权限管理）';
```

### 7.2 菜单表（纯UI结构）

```sql
CREATE TABLE `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint NOT NULL,
  `parent_id` bigint DEFAULT NULL,
  `menu_code` varchar(50) NOT NULL,
  `title` varchar(100) NOT NULL,
  `menu_type` tinyint NOT NULL DEFAULT '1' COMMENT '1-目录，2-菜单，3-按钮，4-接口',
  `route_path` varchar(255) DEFAULT NULL,
  `component` varchar(255) DEFAULT NULL,
  `icon` varchar(50) DEFAULT NULL,
  `sort_order` int NOT NULL DEFAULT '0',
  `visible` tinyint(1) NOT NULL DEFAULT '1',
  `status` tinyint(1) NOT NULL DEFAULT '1',
  `remark` varchar(500) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表（重构版-纯UI结构）';
```

## 8. 系统优势

### 8.1 灵活性大幅提升
- 一个菜单可以关联多个权限
- 权限可以独立于菜单进行管理
- 支持细粒度的权限控制

### 8.2 可维护性增强
- 菜单与权限职责清晰分离
- 标准化的权限命名便于理解
- 模块化的代码结构易于扩展

### 8.3 性能优化
- 减少JOIN操作，直接使用主键索引
- 数字ID比字符串代码占用空间更小
- 使用缓存授权器提高性能

### 8.4 多租户支持
- 完整的租户隔离机制
- 跨租户的角色和权限管理
- 统一的授权接口

## 9. 实施效果

### 9.1 数据验证结果

```
用户数量：3个
角色数量：5个 (含多租户)
权限数量：21个
Casbin规则：35条
菜单权限关联：7条
```

### 9.2 权限分配示例
- **超级管理员(role.id=1)**：拥有所有21个权限
- **系统管理员(role.id=2)**：拥有8个权限（用户管理、角色查看、菜单查看等）
- **普通用户(role.id=3)**：拥有3个基础权限（仪表板、个人资料）

## 10. 技术实现要点

### 10.1 权限检查代码示例

```go
// 权限检查 - 直接使用主键ID
func (c *Casbin) HasPermission(userID, permissionID, tenantID int64) bool {
    return c.enforcer.Enforce(
        strconv.FormatInt(userID, 10),      // 用户主键ID
        strconv.FormatInt(permissionID, 10), // 权限主键ID  
        strconv.FormatInt(tenantID, 10),     // 租户主键ID
    )
}

// 角色分配 - 直接使用主键ID
func (c *Casbin) AssignRole(userID, roleID, tenantID int64) error {
    return c.enforcer.AddGroupingPolicy(
        strconv.FormatInt(userID, 10),   // 用户主键ID
        strconv.FormatInt(roleID, 10),   // 角色主键ID
        strconv.FormatInt(tenantID, 10), // 租户主键ID
    )
}
```

### 10.2 查询优化

```sql
-- 获取用户权限（新设计）
SELECT p.permission_code 
FROM casbin_rule cr
JOIN permissions p ON cr.v1 = p.id  -- 直接主键关联
WHERE cr.v0 = '1'  -- 用户主键ID
```

## 11. 总结

系统通过Casbin实现了一个完整的、灵活的、高性能的多租户RBAC权限管理系统。通过统一使用数据库主键ID，实现了权限管理的一致性、高性能和易维护性。同时，通过菜单权限矩阵和标准化的权限设计，提供了细粒度的权限控制和良好的用户体验。

重构后的系统在保持原有功能的基础上，大幅提升了权限管理的灵活性和可扩展性，为未来的功能扩展打下了坚实的基础。

---

*文档生成时间：2024年*
*基于 one-auth 项目权限系统分析*