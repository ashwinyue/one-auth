# Casbin 主键ID设计说明

## 🎯 设计目标

解决原有设计中ID映射不一致的问题，统一使用数据库主键ID进行关联，提升性能和数据一致性。

## 📊 修改前后对比

### ❌ 原有设计问题
```sql
-- 使用字符串代码映射，不一致且难维护
('g','admin','r1','t1',NULL,'',''),      -- 用户名 -> 角色代码 -> 租户代码
('p','r1','a1','t1',NULL,'',''),         -- 角色代码 -> 权限代码 -> 租户代码
```

**问题**：
1. 需要额外维护代码映射关系（如 `r1` -> `role.id=1`）
2. 用户标识不一致（有时用ID，有时用username）
3. 查询时需要额外JOIN操作获取真实ID
4. 代码可读性差，维护成本高

### ✅ 新设计优势
```sql
-- 直接使用数据库主键ID，简洁高效
('g','1','1','1',NULL,'',''),            -- user.id -> role.id -> tenant.id
('p','1','1','1',NULL,'',''),            -- role.id -> permission.id -> tenant.id
```

**优势**：
1. **性能更优**：直接使用主键，无需额外映射转换
2. **一致性强**：所有关联都使用主键ID
3. **维护简单**：无需维护额外的代码映射表
4. **可读性好**：直接对应数据库记录

## 🏗️ 新的ID映射关系

### Casbin 规则格式
```
g, user_id, role_id, tenant_id    - 用户角色分配
p, role_id, permission_id, tenant_id - 角色权限分配
```

### 实际数据示例

#### 用户角色分配 (g规则)
```sql
('g','1','1','1',NULL,'',''),   -- user.id=1(admin) -> role.id=1(super_admin) in tenant.id=1
('g','2','2','1',NULL,'',''),   -- user.id=2(user1) -> role.id=2(admin) in tenant.id=1  
('g','3','3','1',NULL,'',''),   -- user.id=3(user2) -> role.id=3(user) in tenant.id=1
```

#### 角色权限分配 (p规则)
```sql
('p','1','1','1',NULL,'',''),   -- role.id=1 -> permission.id=1(user:view) in tenant.id=1
('p','1','2','1',NULL,'',''),   -- role.id=1 -> permission.id=2(user:create) in tenant.id=1
('p','2','1','1',NULL,'',''),   -- role.id=2 -> permission.id=1(user:view) in tenant.id=1
```

## 📋 完整的主键映射表

### 用户表 (user)
| ID | userID | username | 说明 |
|----|--------|----------|------|
| 1  | admin  | admin    | 超级管理员 |
| 2  | user1  | user1    | 普通管理员 |
| 3  | user2  | user2    | 普通用户 |

### 角色表 (roles)
| ID | role_code   | name       | 说明 |
|----|-------------|------------|------|
| 1  | super_admin | 超级管理员 | 拥有所有权限 |
| 2  | admin       | 系统管理员 | 拥有部分权限 |
| 3  | user        | 普通用户   | 基础权限 |

### 租户表 (tenants)
| ID | tenant_code | name     | 说明 |
|----|-------------|----------|------|
| 1  | default     | 默认租户 | 系统默认租户 |
| 2  | demo        | 演示租户 | 演示用租户 |

### 权限表 (permissions) 
| ID | permission_code   | name         | resource_type | action |
|----|-------------------|--------------|---------------|---------|
| 1  | user:view         | 查看用户     | menu          | view    |
| 2  | user:create       | 创建用户     | menu          | create  |
| 3  | user:update       | 编辑用户     | menu          | update  |
| 4  | user:delete       | 删除用户     | menu          | delete  |
| 5  | user:export       | 导出用户     | feature       | export  |
| 6  | role:view         | 查看角色     | menu          | view    |
| 7  | role:create       | 创建角色     | menu          | create  |
| 8  | role:update       | 编辑角色     | menu          | update  |
| 9  | role:delete       | 删除角色     | menu          | delete  |
| 10 | role:assign       | 分配角色     | feature       | assign  |
| 11 | permission:view   | 查看权限     | menu          | view    |
| 12 | permission:assign | 分配权限     | feature       | assign  |
| 13 | menu:view         | 查看菜单     | menu          | view    |
| 14 | menu:create       | 创建菜单     | menu          | create  |
| 15 | menu:update       | 编辑菜单     | menu          | update  |
| 16 | menu:delete       | 删除菜单     | menu          | delete  |
| 17 | tenant:view       | 查看租户     | menu          | view    |
| 18 | tenant:switch     | 切换租户     | feature       | switch  |
| 19 | dashboard:view    | 查看仪表板   | menu          | view    |
| 20 | profile:view      | 查看个人资料 | menu          | view    |
| 21 | profile:update    | 编辑个人资料 | menu          | update  |

## 🚀 实施效果

### 数据验证结果
```
用户数量：3个
角色数量：5个 (含多租户)
权限数量：21个
Casbin规则：35条
菜单权限关联：7条
```

### 权限分配示例
- **超级管理员(role.id=1)**：拥有所有21个权限
- **系统管理员(role.id=2)**：拥有8个权限（用户管理、角色查看、菜单查看等）
- **普通用户(role.id=3)**：拥有3个基础权限（仪表板、个人资料）

## 🔧 技术实现要点

### 1. SQL字符集解决方案
```sql
-- 强制设置连接字符集为utf8mb4
SET character_set_client = utf8mb4;
SET character_set_connection = utf8mb4;
SET character_set_results = utf8mb4;
SET collation_connection = utf8mb4_unicode_ci;
```

**执行命令**：
```bash
docker exec -i miniblog-mysql mysql -u miniblog -pminiblog1234 --default-character-set=utf8mb4 miniblog < configs/miniblog.sql
```

### 2. 代码中的使用方式

#### Go代码示例
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

### 3. 查询优化

#### 获取用户权限（原设计）
```sql
-- 需要多次JOIN和映射转换
SELECT p.permission_code 
FROM casbin_rule cr
JOIN roles r ON cr.v1 = CONCAT('r', r.id)  -- 代码映射
JOIN permissions p ON cr.v2 = CONCAT('a', p.id)  -- 代码映射
WHERE cr.v0 = 'admin'  -- 用户名
```

#### 获取用户权限（新设计）
```sql
-- 直接使用主键ID，性能更好
SELECT p.permission_code 
FROM casbin_rule cr
JOIN permissions p ON cr.v1 = p.id  -- 直接主键关联
WHERE cr.v0 = '1'  -- 用户主键ID
```

## 📈 性能提升

1. **查询性能**：减少JOIN操作，直接使用主键索引
2. **存储空间**：数字ID比字符串代码占用空间更小
3. **维护成本**：无需维护额外的映射关系
4. **扩展性**：新增权限只需插入记录，无需更新映射

## ✅ 总结

通过统一使用数据库主键ID，我们实现了：

- **🎯 一致性**：所有关联关系都使用主键ID
- **⚡ 高性能**：直接使用主键索引，无额外映射开销  
- **🛠️ 易维护**：简化了权限管理的复杂度
- **🔧 标准化**：符合数据库设计最佳实践
- **🌐 国际化**：解决了中文注释乱码问题

新设计为后续的权限扩展和系统维护打下了坚实的基础。 