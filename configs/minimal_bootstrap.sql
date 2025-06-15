-- =====================================================
-- 最小化系统引导脚本
-- 只创建超级管理员，不预定义权限
-- =====================================================

-- 基础环境设置
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建基础租户
INSERT INTO `tenants` (`id`, `tenant_code`, `name`, `description`, `status`, `created_at`, `updated_at`) VALUES 
(1, 'default', '默认租户', '系统默认租户', 1, NOW(), NOW());

-- 创建角色层次结构
INSERT INTO `roles` (`id`, `tenant_id`, `role_code`, `name`, `description`, `status`, `created_at`, `updated_at`) VALUES 
(1, 1, 'super_admin', '超级管理员', '拥有系统所有权限的超级管理员角色，绕过权限检查', 1, NOW(), NOW()),
(2, 1, 'admin', '系统管理员', '拥有大部分管理权限的系统管理员角色，受权限控制', 1, NOW(), NOW()),
(3, 1, 'user', '普通用户', '普通业务用户角色，只有基础权限', 1, NOW(), NOW());

-- 创建超级管理员用户
INSERT INTO `user` (`id`, `username`, `nickname`, `email`, `phone`, `avatar`, `created_at`, `updated_at`) VALUES 
(1, 'admin', '超级管理员', 'admin@example.com', '', '', NOW(), NOW()),
(2, 'manager', '系统管理员', 'manager@example.com', '', '', NOW(), NOW());

-- 创建用户认证信息
INSERT INTO `user_status` (`user_id`, `auth_type`, `auth_id`, `credential`, `status`, `is_verified`, `is_primary`, `created_at`, `updated_at`) VALUES 
(1, 1, 'admin', '$2a$10$xQp7hMNdqGLJhd6p5bXN6ujgNtNqhq.AjQBiPO0aDJQj7tBqFPwX.', 1, 1, 1, NOW(), NOW()),
(2, 1, 'manager', '$2a$10$xQp7hMNdqGLJhd6p5bXN6ujgNtNqhq.AjQBiPO0aDJQj7tBqFPwX.', 1, 1, 1, NOW(), NOW());

-- 创建用户租户关联
INSERT INTO `user_tenants` (`user_id`, `tenant_id`, `created_at`, `updated_at`) VALUES 
(1, 1, NOW(), NOW()),
(2, 1, NOW(), NOW());

-- 分配用户角色
INSERT INTO `casbin_rule` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES 
('g', 'u1', 'r1', 't1', NULL, NULL, NULL),  -- admin -> super_admin
('g', 'u2', 'r2', 't1', NULL, NULL, NULL);  -- manager -> admin

-- =====================================================
-- 系统引导完成说明
-- =====================================================

SELECT 'Bootstrap completed successfully!' as message;

SELECT 
    '账号信息:' as info,
    'admin/admin123 - 超级管理员(免权限检查)' as super_admin,
    'manager/admin123 - 系统管理员(自动获得管理权限)' as admin,
    '新权限将自动分配给符合规则的管理员' as auto_assign;

-- =====================================================
-- 验证引导结果
-- =====================================================

-- 验证基础数据
SELECT 'System Bootstrap Verification' as step;

SELECT '租户数据' as check_type, COUNT(*) as count FROM tenants;
SELECT '角色数据' as check_type, COUNT(*) as count FROM roles; 
SELECT '用户数据' as check_type, COUNT(*) as count FROM user;
SELECT '用户状态数据' as check_type, COUNT(*) as count FROM user_status;
SELECT '用户租户关联' as check_type, COUNT(*) as count FROM user_tenants;
SELECT 'Casbin规则' as check_type, COUNT(*) as count FROM casbin_rule;

-- 显示角色层次结构
SELECT 
    r.id as role_id,
    r.role_code,
    r.name as role_name,
    r.description
FROM roles r
WHERE r.tenant_id = 1
ORDER BY r.id;

-- 显示用户角色分配
SELECT 
    u.id as user_id,
    u.username,
    u.nickname,
    r.role_code,
    r.name as role_name,
    CASE r.role_code
        WHEN 'super_admin' THEN '享受免权限检查'
        WHEN 'admin' THEN '自动获得管理权限'
        WHEN 'user' THEN '需要手动分配权限'
        ELSE '未知角色类型'
    END as privilege_type
FROM user u
JOIN casbin_rule cr ON cr.v0 = CONCAT('u', u.id) AND cr.ptype = 'g'
JOIN roles r ON cr.v1 = CONCAT('r', r.id) 
WHERE u.id IN (1, 2);

-- =====================================================
-- 动态权限创建示例
-- =====================================================

-- 当需要新功能时，可以动态创建权限
-- 例如：订单管理模块

/*
-- 动态创建订单管理权限（示例）
INSERT INTO `permissions` (`tenant_id`, `permission_code`, `name`, `description`, `resource_type`, `action`, `status`) VALUES
(1, 'order:view', '查看订单', '查看订单列表和详情', 'menu', 'view', 1),
(1, 'order:create', '创建订单', '创建新订单', 'menu', 'create', 1),
(1, 'order:update', '编辑订单', '修改订单信息', 'menu', 'update', 1),
(1, 'order:delete', '删除订单', '删除订单', 'menu', 'delete', 1),
(1, 'order:approve', '审批订单', '审批订单流程', 'feature', 'approve', 1);

-- 系统管理员会自动获得符合规则的权限：order:view, order:create, order:update, order:delete
-- 超级管理员自动拥有所有权限（包括order:approve）
-- 普通用户需要手动分配权限

-- 检查自动分配结果
SELECT 
    r.role_code,
    r.name as role_name,
    COUNT(cr.v1) as permission_count
FROM roles r
LEFT JOIN casbin_rule cr ON cr.v0 = CONCAT('r', r.id) AND cr.ptype = 'p'
WHERE r.tenant_id = 1
GROUP BY r.id, r.role_code, r.name
ORDER BY r.id;
*/

-- =====================================================
-- 权限自动分配规则说明
-- =====================================================

SELECT '
权限自动分配规则:

角色层次:
1. 超级管理员(super_admin) - 绕过所有权限检查，拥有无限权限
2. 系统管理员(admin) - 自动获得基础管理权限，受权限控制  
3. 普通用户(user) - 需要手动分配权限

自动分配规则:
- 超级管理员：无条件获得所有权限（代码层面免检）
- 系统管理员：自动获得匹配规则的权限
  * 包含模块：user, role, menu, permission
  * 包含操作：view, create, update, delete
  * 排除权限：system:config, system:backup, tenant:delete

使用场景:
1. 新增菜单 -> 系统管理员自动获得基础CRUD权限
2. 新增功能 -> 根据权限编码自动判断是否分配
3. 系统升级 -> 管理员权限自动同步

优势:
- 超级管理员永远可用，解决引导问题
- 系统管理员权限自动化，减少运维工作
- 普通用户权限精确控制，保证安全性
- 支持灵活的权限分配规则配置
' as rules_explanation;

SET FOREIGN_KEY_CHECKS = 1; 