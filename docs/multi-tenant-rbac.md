# 多租户RBAC with Domains实现说明

## 概述

本项目已成功从ACL模式升级为多租户RBAC with Domains模式，支持以下特性：

- **多租户隔离**：不同租户之间的权限完全隔离
- **角色管理**：支持在不同租户中为用户分配不同角色
- **权限继承**：用户通过角色获得权限，支持隐式权限查询
- **跨租户支持**：同一用户可以在不同租户中拥有不同角色和权限

## 设计原则

本实现严格遵循[Casbin RBAC with Domains官方文档](https://casbin.org/zh/docs/rbac-with-domains)的设计规范：

1. **角色定义**：使用三元组 `g = _, _, _`，第三个参数表示域/租户
2. **策略格式**：采用 `p, sub, dom, obj, act, eft` 格式
3. **匹配器规则**：使用 `g(r.sub, p.sub, r.dom) && r.dom == p.dom` 确保域隔离
4. **令牌约定**：域令牌名称使用标准的 `dom`，位于第二个位置

## 核心变更

### 1. Casbin模型更新

从简单的ACL模型升级为RBAC with Domains模型，严格遵循[Casbin官方规范](https://casbin.org/zh/docs/rbac-with-domains)：

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _

[policy_effect]
e = !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch(r.obj, p.obj) && r.act == p.act
```

### 2. 数据库表结构

新增以下表支持多租户RBAC：

- `tenants` - 租户表
- `roles` - 角色表（支持多租户）
- `permissions` - 权限表（支持多租户）
- `menus` - 菜单表（支持多租户）
- `user_tenants` - 用户租户关联表

### 3. API接口更新

#### 授权接口
- `AuthorizeWithDomain(sub, dom, obj, act)` - 多租户授权检查
- `Authorize(sub, obj, act)` - 兼容接口（使用默认租户）

#### 角色管理接口
- `AddRoleForUser(user, role, domain)` - 为用户添加角色
- `DeleteRoleForUser(user, role, domain)` - 删除用户角色
- `GetRolesForUser(user, domain)` - 获取用户角色
- `GetUsersForRole(role, domain)` - 获取角色用户
- `DeleteAllRolesForUser(user, domain)` - 删除用户所有角色
- `DeleteRole(role, domain)` - 删除角色

#### 权限管理接口
- `AddPermissionForUser(user, domain, obj, act)` - 为用户添加权限
- `DeletePermissionForUser(user, domain, obj, act)` - 删除用户权限
- `GetPermissionsForUser(user, domain)` - 获取用户权限
- `HasPermissionForUser(user, domain, obj, act)` - 检查用户权限

#### 高级查询接口
- `GetImplicitRolesForUser(user, domain)` - 获取隐式角色
- `GetImplicitPermissionsForUser(user, domain)` - 获取隐式权限
- `GetAllUsersByDomain(domain)` - 获取租户所有用户

### 4. 中间件更新

#### Gin中间件
```go
func AuthzMiddleware(authorizer Authorizer) gin.HandlerFunc {
    return func(c *gin.Context) {
        subject := contextx.UserID(c.Request.Context())
        domain := contextx.TenantID(c.Request.Context()) // 获取租户ID
        object := c.Request.URL.Path
        action := c.Request.Method

        // 如果没有租户ID，使用默认租户
        if domain == "" {
            domain = "default"
        }

        allowed, err := authorizer.AuthorizeWithDomain(subject, domain, object, action)
        // ... 权限检查逻辑
    }
}
```

#### gRPC中间件
类似的多租户支持逻辑。

### 5. 上下文支持

新增租户ID上下文支持：

```go
// 设置租户ID到上下文
func WithTenantID(ctx context.Context, tenantID string) context.Context

// 从上下文获取租户ID
func TenantID(ctx context.Context) string
```

## 使用示例

### 基本用法

```go
// 创建授权器
authz, err := authz.NewAuthz(db)

// 为用户在指定租户中添加角色
success, err := authz.AddRoleForUser("user123", "admin", "tenant1")

// 检查权限
allowed, err := authz.AuthorizeWithDomain("user123", "tenant1", "/v1/posts/1", "GET")

// 获取用户在租户中的所有权限
permissions := authz.GetImplicitPermissionsForUser("user123", "tenant1")
```

### 跨租户场景

```go
// 同一用户在不同租户中可以有不同角色
authz.AddRoleForUser("user123", "admin", "tenant1")    // 在tenant1中是管理员
authz.AddRoleForUser("user123", "user", "tenant2")     // 在tenant2中是普通用户

// 权限检查会根据租户进行隔离
allowed1, _ := authz.AuthorizeWithDomain("user123", "tenant1", "/v1/users/*", "DELETE") // true
allowed2, _ := authz.AuthorizeWithDomain("user123", "tenant2", "/v1/users/*", "DELETE") // false
```

## 数据库策略示例

### 角色权限策略
```sql
-- 管理员在默认租户的权限
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4) VALUES
('p', 'admin', 'default', '/v1/posts/*', 'GET', 'allow'),
('p', 'admin', 'default', '/v1/posts/*', 'POST', 'allow'),
('p', 'admin', 'default', '/v1/posts/*', 'PUT', 'allow'),
('p', 'admin', 'default', '/v1/posts/*', 'DELETE', 'allow');
```

### 用户角色分配
```sql
-- 用户在指定租户中的角色
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('g', 'user123', 'admin', 'default'),
('g', 'user456', 'user', 'default');
```

## 测试验证

运行测试脚本验证功能：

```bash
go run scripts/test-rbac.go
```

测试覆盖：
- ✅ 用户角色管理
- ✅ 权限检查
- ✅ 跨租户隔离
- ✅ 隐式权限查询
- ✅ 租户用户列表

## 迁移说明

### 从ACL到RBAC的迁移步骤

1. **执行数据库迁移**
   ```bash
   mysql -u root -p < scripts/create-rbac-tables.sql
   mysql -u root -p < scripts/init-casbin-policies.sql
   ```

2. **清理旧格式策略**
   ```sql
   DELETE FROM casbin_rule WHERE ptype = 'p' AND (v3 IS NULL OR v4 IS NULL);
   DELETE FROM casbin_rule WHERE ptype = 'g' AND v2 IS NULL;
   ```

3. **更新应用代码**
   - 使用新的授权接口
   - 在请求上下文中设置租户ID
   - 更新中间件配置

## 注意事项

1. **向后兼容性**：保留了原有的`Authorize`方法，使用默认租户
2. **性能考虑**：建议为`casbin_rule`表添加适当的索引
3. **租户隔离**：确保在请求处理中正确设置租户ID
4. **错误处理**：权限检查失败时返回适当的错误信息

## 参考文档

- [Casbin RBAC with Domains官方文档](https://casbin.org/zh/docs/rbac-with-domains) - 本实现严格遵循此官方规范
- [Casbin Go API文档](https://casbin.org/docs/management-api)
- [Casbin RBAC with Domains英文文档](https://casbin.org/docs/rbac-with-domains) 