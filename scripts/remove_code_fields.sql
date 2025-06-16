-- =======================================================
-- 移除编码字段的数据库迁移脚本
-- =======================================================

-- 1. 移除菜单表的 menu_code 字段和相关索引
ALTER TABLE menus DROP INDEX idx_menu_code_tenant;
ALTER TABLE menus DROP COLUMN menu_code;

-- 2. 移除权限表的 permission_code 字段和相关索引
ALTER TABLE permissions DROP INDEX idx_permission_code_tenant;
ALTER TABLE permissions DROP COLUMN permission_code;

-- 3. 移除角色表的 role_code 字段和相关索引
ALTER TABLE roles DROP INDEX idx_role_code_tenant;
ALTER TABLE roles DROP COLUMN role_code;

-- 4. 移除租户表的 tenant_code 字段和相关索引（如果需要的话）
ALTER TABLE tenants DROP INDEX idx_tenant_code;
ALTER TABLE tenants DROP COLUMN tenant_code;

-- 5. 为角色表的 name 字段添加租户内唯一约束
ALTER TABLE roles ADD UNIQUE KEY idx_name_tenant (name, tenant_id);

-- 6. 更新 Casbin 规则中的格式（如果有使用编码的话，现在统一使用ID）
-- 这部分数据已经是使用ID格式，无需修改

-- 7. 验证表结构
SHOW CREATE TABLE menus;
SHOW CREATE TABLE permissions;
SHOW CREATE TABLE roles;

-- 8. 检查数据完整性
SELECT COUNT(*) as menu_count FROM menus;
SELECT COUNT(*) as permission_count FROM permissions;
SELECT COUNT(*) as role_count FROM roles;
SELECT COUNT(*) as casbin_rule_count FROM casbin_rule; 