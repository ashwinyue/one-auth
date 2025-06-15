-- =====================================================
-- One-Auth 用户认证系统完整数据库初始化脚本
-- =====================================================
-- 
-- 执行方式：
-- 1. Docker方式（推荐）：
--    docker exec -i miniblog-mysql mysql -u miniblog -pminiblog1234 --default-character-set=utf8mb4 miniblog < configs/miniblog.sql
-- 
-- 2. 直接连接方式：
--    mysql -h 127.0.0.1 -P 3306 -u miniblog -pminiblog1234 --default-character-set=utf8mb4 miniblog < configs/miniblog.sql
-- 
-- 注意：请确保MySQL服务已启动：docker compose up -d mysql
-- =====================================================

-- 设置字符集和环境
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

-- 强制设置连接字符集为utf8mb4
SET character_set_client = utf8mb4;
SET character_set_connection = utf8mb4;
SET character_set_results = utf8mb4;
SET collation_connection = utf8mb4_unicode_ci;

-- =====================================================
-- 创建数据库
-- =====================================================

DROP DATABASE IF EXISTS `miniblog`;
CREATE DATABASE `miniblog` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `miniblog`;

-- =====================================================
-- 权限管理表 (casbin_rule)
-- =====================================================

DROP TABLE IF EXISTS `casbin_rule`;
CREATE TABLE `casbin_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) DEFAULT NULL,
  `v0` varchar(100) DEFAULT NULL,
  `v1` varchar(100) DEFAULT NULL,
  `v2` varchar(100) DEFAULT NULL,
  `v3` varchar(100) DEFAULT NULL,
  `v4` varchar(100) DEFAULT NULL,
  `v5` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限管理规则表';

-- =====================================================
-- 租户表 (tenants)
-- =====================================================

DROP TABLE IF EXISTS `tenants`;
CREATE TABLE `tenants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '租户主键ID',
  `tenant_code` varchar(50) NOT NULL COMMENT '租户编码',
  `name` varchar(100) NOT NULL COMMENT '租户名称',
  `description` varchar(500) DEFAULT NULL COMMENT '描述',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tenant_code` (`tenant_code`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- =====================================================
-- 角色表 (roles)
-- =====================================================

DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '角色主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `role_code` varchar(50) NOT NULL COMMENT '角色编码',
  `name` varchar(100) NOT NULL COMMENT '角色名称',
  `description` varchar(500) DEFAULT NULL COMMENT '描述',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_role_code_tenant` (`role_code`, `tenant_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- =====================================================
-- 菜单表 (menus) - 重构版：纯UI结构
-- =====================================================

DROP TABLE IF EXISTS `menus`;
CREATE TABLE `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '菜单主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `parent_id` bigint DEFAULT NULL COMMENT '父菜单ID',
  `menu_code` varchar(50) NOT NULL COMMENT '菜单编码',
  `title` varchar(100) NOT NULL COMMENT '菜单标题',
  `menu_type` tinyint NOT NULL DEFAULT '1' COMMENT '菜单类型：1-目录，2-菜单，3-按钮，4-接口',
  `route_path` varchar(255) DEFAULT NULL COMMENT '前端路由路径',
  `component` varchar(255) DEFAULT NULL COMMENT '前端组件路径',
  `icon` varchar(50) DEFAULT NULL COMMENT '图标',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序',
  `visible` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否可见',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1-启用，0-禁用',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_menu_code_tenant` (`menu_code`, `tenant_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_status` (`status`),
  KEY `idx_menu_type` (`menu_type`),
  KEY `idx_visible` (`visible`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表（重构版-纯UI结构）';

-- =====================================================
-- 权限表 (permissions) - 重构版：独立权限管理
-- =====================================================

DROP TABLE IF EXISTS `permissions`;
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

-- =====================================================
-- 菜单权限关联表 (menu_permissions)
-- =====================================================

DROP TABLE IF EXISTS `menu_permissions`;
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

-- =====================================================
-- 用户租户关联表 (user_tenants)
-- =====================================================

DROP TABLE IF EXISTS `user_tenants`;
CREATE TABLE `user_tenants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID（关联user表的id字段）',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_tenant` (`user_id`, `tenant_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户租户关联表';

-- =====================================================
-- 角色权限关联表 (role_permissions)
-- =====================================================

DROP TABLE IF EXISTS `role_permissions`;
CREATE TABLE `role_permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_role_permission_tenant` (`role_id`, `permission_id`, `tenant_id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_permission_id` (`permission_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- =====================================================
-- 用户表 (user) - 兼容原有结构
-- =====================================================

DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户主键ID',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名（唯一）',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码（加密后）',
  `nickname` varchar(30) NOT NULL DEFAULT '' COMMENT '用户昵称',
  `email` varchar(256) NOT NULL DEFAULT '' COMMENT '用户电子邮箱地址',
  `phone` varchar(16) NOT NULL DEFAULT '' COMMENT '用户手机号',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '用户创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '用户最后修改时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_username` (`username`),
  UNIQUE KEY `idx_user_phone` (`phone`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- =====================================================
-- 用户状态表 (user_status) - 新增多认证方式支持
-- =====================================================

DROP TABLE IF EXISTS `user_status`;
CREATE TABLE `user_status` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `auth_id` varchar(255) NOT NULL COMMENT '认证标识符（邮箱、手机号、用户名等）',
  `auth_type` tinyint NOT NULL COMMENT '认证类型：1-username,2-email,3-phone,4-wechat,5-qq,6-github,7-google,8-apple,9-dingtalk,10-feishu',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID（关联user表的id）',
  `tenant_id` bigint NOT NULL DEFAULT '1' COMMENT '租户ID',
  
  -- 用户状态
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '用户状态：1-active,2-inactive,3-locked,4-banned',
  `lock_reason` varchar(255) DEFAULT NULL COMMENT '锁定原因',
  `locked_until` timestamp NULL DEFAULT NULL COMMENT '锁定到期时间',
  
  -- 登录信息
  `last_login_time` timestamp NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) DEFAULT NULL COMMENT '最后登录IP',
  `last_login_device` varchar(128) DEFAULT NULL COMMENT '最后登录设备',
  `login_count` int NOT NULL DEFAULT '0' COMMENT '登录次数',
  
  -- 安全信息（持久化记录，配合Redis实时防护）
  `failed_login_attempts` int NOT NULL DEFAULT '0' COMMENT '累计登录失败次数',
  `last_failed_login` timestamp NULL DEFAULT NULL COMMENT '最后一次登录失败时间',
  `password_changed_at` timestamp NULL DEFAULT NULL COMMENT '密码最后修改时间',
  
  -- 验证状态
  `is_verified` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已验证',
  `verified_at` timestamp NULL DEFAULT NULL COMMENT '验证时间',
  `is_primary` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为主要认证方式',
  
  -- 时间戳
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  
  -- 索引设计
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_auth_id_type` (`auth_id`, `auth_type`) COMMENT '认证标识符+类型唯一',
  KEY `idx_user_id` (`user_id`) COMMENT '用户ID索引（非唯一，一个用户多个认证方式）',
  KEY `idx_auth_id_type_tenant` (`auth_id`, `auth_type`, `tenant_id`) COMMENT '认证查询索引',
  KEY `idx_user_tenant` (`user_id`, `tenant_id`) COMMENT '用户租户索引',
  KEY `idx_status` (`status`) COMMENT '状态索引',
  KEY `idx_auth_type` (`auth_type`) COMMENT '认证类型索引',
  KEY `idx_primary` (`is_primary`) COMMENT '主要认证方式索引',
  KEY `idx_verified` (`is_verified`) COMMENT '验证状态索引',
  KEY `idx_last_login` (`last_login_time`) COMMENT '最后登录时间索引',
  KEY `idx_tenant_id` (`tenant_id`) COMMENT '租户ID索引',
  KEY `idx_deleted_at` (`deleted_at`) COMMENT '软删除索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户状态表-支持多种认证方式';

-- =====================================================
-- 博文表 (post)
-- =====================================================

DROP TABLE IF EXISTS `post`;
CREATE TABLE `post` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '博文主键ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID（关联user表）',
  `post_id` varchar(35) NOT NULL DEFAULT '' COMMENT '博文唯一标识',
  `title` varchar(256) NOT NULL DEFAULT '' COMMENT '博文标题',
  `content` longtext NOT NULL COMMENT '博文内容',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '博文创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '博文最后修改时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_post_post_id` (`post_id`),
  KEY `idx_post_user_id` (`user_id`),
  KEY `idx_deleted_at` (`deleted_at`),
  FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博文表';

-- =====================================================
-- 插入初始数据
-- =====================================================



-- 插入默认租户
INSERT INTO `tenants` (`tenant_code`, `name`, `description`, `status`) VALUES
('default', '默认租户', '系统默认租户', 1),
('demo', '演示租户', '演示用租户', 1);

-- 插入默认角色
INSERT INTO `roles` (`tenant_id`, `role_code`, `name`, `description`, `status`) VALUES
(1, 'super_admin', '超级管理员', '超级管理员角色（不可删除）', 1),
(1, 'admin', '系统管理员', '拥有系统所有权限', 1),
(1, 'user', '普通用户', '普通用户权限', 1),
(2, 'super_admin', '超级管理员', '演示租户超级管理员', 1),
(2, 'admin', '租户管理员', '租户管理员权限', 1);

-- 插入Casbin权限规则数据（使用前缀+ID格式提高可读性）
INSERT INTO `casbin_rule` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES
-- 用户角色分配（格式：g, u{user_id}, r{role_id}, t{tenant_id}）
('g','u1','r1','t1',NULL,'',''),   -- user.id=1(admin) -> role.id=1(super_admin) in tenant.id=1
('g','u2','r2','t1',NULL,'',''),   -- user.id=2(user1) -> role.id=2(admin) in tenant.id=1  
('g','u3','r3','t1',NULL,'',''),   -- user.id=3(user2) -> role.id=3(user) in tenant.id=1

-- 角色权限分配（格式：p, r{role_id}, p{permission_id}, t{tenant_id}）
-- 超级管理员(role.id=1)拥有所有权限
('p','r1','p1','t1',NULL,'',''),   -- role.id=1 -> permission.id=1(user:view) in tenant.id=1
('p','r1','p2','t1',NULL,'',''),   -- role.id=1 -> permission.id=2(user:create) in tenant.id=1
('p','r1','p3','t1',NULL,'',''),   -- role.id=1 -> permission.id=3(user:update) in tenant.id=1
('p','r1','p4','t1',NULL,'',''),   -- role.id=1 -> permission.id=4(user:delete) in tenant.id=1
('p','r1','p5','t1',NULL,'',''),   -- role.id=1 -> permission.id=5(user:export) in tenant.id=1
('p','r1','p6','t1',NULL,'',''),   -- role.id=1 -> permission.id=6(role:view) in tenant.id=1
('p','r1','p7','t1',NULL,'',''),   -- role.id=1 -> permission.id=7(role:create) in tenant.id=1
('p','r1','p8','t1',NULL,'',''),   -- role.id=1 -> permission.id=8(role:update) in tenant.id=1
('p','r1','p9','t1',NULL,'',''),   -- role.id=1 -> permission.id=9(role:delete) in tenant.id=1
('p','r1','p10','t1',NULL,'',''),  -- role.id=1 -> permission.id=10(role:assign) in tenant.id=1
('p','r1','p11','t1',NULL,'',''),  -- role.id=1 -> permission.id=11(permission:view) in tenant.id=1
('p','r1','p12','t1',NULL,'',''),  -- role.id=1 -> permission.id=12(permission:assign) in tenant.id=1
('p','r1','p13','t1',NULL,'',''),  -- role.id=1 -> permission.id=13(menu:view) in tenant.id=1
('p','r1','p14','t1',NULL,'',''),  -- role.id=1 -> permission.id=14(menu:create) in tenant.id=1
('p','r1','p15','t1',NULL,'',''),  -- role.id=1 -> permission.id=15(menu:update) in tenant.id=1
('p','r1','p16','t1',NULL,'',''),  -- role.id=1 -> permission.id=16(menu:delete) in tenant.id=1
('p','r1','p17','t1',NULL,'',''),  -- role.id=1 -> permission.id=17(tenant:view) in tenant.id=1
('p','r1','p18','t1',NULL,'',''),  -- role.id=1 -> permission.id=18(tenant:switch) in tenant.id=1
('p','r1','p19','t1',NULL,'',''),  -- role.id=1 -> permission.id=19(dashboard:view) in tenant.id=1
('p','r1','p20','t1',NULL,'',''),  -- role.id=1 -> permission.id=20(profile:view) in tenant.id=1
('p','r1','p21','t1',NULL,'',''),  -- role.id=1 -> permission.id=21(profile:update) in tenant.id=1

-- 系统管理员(role.id=2)拥有部分权限
('p','r2','p1','t1',NULL,'',''),   -- role.id=2 -> permission.id=1(user:view) in tenant.id=1
('p','r2','p2','t1',NULL,'',''),   -- role.id=2 -> permission.id=2(user:create) in tenant.id=1
('p','r2','p3','t1',NULL,'',''),   -- role.id=2 -> permission.id=3(user:update) in tenant.id=1
('p','r2','p6','t1',NULL,'',''),   -- role.id=2 -> permission.id=6(role:view) in tenant.id=1
('p','r2','p13','t1',NULL,'',''),  -- role.id=2 -> permission.id=13(menu:view) in tenant.id=1
('p','r2','p19','t1',NULL,'',''),  -- role.id=2 -> permission.id=19(dashboard:view) in tenant.id=1
('p','r2','p20','t1',NULL,'',''),  -- role.id=2 -> permission.id=20(profile:view) in tenant.id=1
('p','r2','p21','t1',NULL,'',''),  -- role.id=2 -> permission.id=21(profile:update) in tenant.id=1

-- 普通用户(role.id=3)拥有基础权限
('p','r3','p19','t1',NULL,'',''),  -- role.id=3 -> permission.id=19(dashboard:view) in tenant.id=1
('p','r3','p20','t1',NULL,'',''),  -- role.id=3 -> permission.id=20(profile:view) in tenant.id=1
('p','r3','p21','t1',NULL,'','');  -- role.id=3 -> permission.id=21(profile:update) in tenant.id=1

-- 插入默认用户数据
INSERT INTO `user` VALUES
(1,'admin','$2a$10$ctsFXEUAMd7rXXpmccNlO.ZRiYGYz0eOfj8EicPGWqiz64YBBgR1y','管理员','admin@example.com','13800138000','2024-12-12 03:55:25','2024-12-12 03:55:25',NULL),
(2,'user1','$2a$10$ctsFXEUAMd7rXXpmccNlO.ZRiYGYz0eOfj8EicPGWqiz64YBBgR1y','用户1','user1@example.com','13800138001','2024-12-12 03:55:25','2024-12-12 03:55:25',NULL),
(3,'user2','$2a$10$ctsFXEUAMd7rXXpmccNlO.ZRiYGYz0eOfj8EicPGWqiz64YBBgR1y','用户2','user2@example.com','13800138002','2024-12-12 03:55:25','2024-12-12 03:55:25',NULL);

-- 插入用户状态数据（与user表对应）
INSERT INTO `user_status` (`auth_id`, `auth_type`, `user_id`, `tenant_id`, `status`, `is_verified`, `is_primary`) VALUES
-- admin 用户(id=1)的多种认证方式
('admin', 1, 1, 1, 1, 1, 1),                         -- username, active, verified, primary
('admin@example.com', 2, 1, 1, 1, 1, 0),             -- email, active, verified, not primary
('13800138000', 3, 1, 1, 1, 1, 0),                   -- phone, active, verified, not primary

-- user1 用户(id=2)的多种认证方式
('user1', 1, 2, 1, 1, 1, 1),                         -- username, active, verified, primary
('user1@example.com', 2, 2, 1, 1, 1, 0),            -- email, active, verified, not primary
('13800138001', 3, 2, 1, 1, 1, 0),                   -- phone, active, verified, not primary

-- user2 用户(id=3)的多种认证方式
('user2', 1, 3, 1, 1, 1, 1),                         -- username, active, verified, primary
('user2@example.com', 2, 3, 1, 1, 0, 0),            -- email, active, not verified, not primary
('13800138002', 3, 3, 1, 1, 0, 0);                   -- phone, active, not verified, not primary

-- 插入用户租户关联
INSERT INTO `user_tenants` (`user_id`, `tenant_id`, `status`) VALUES
(1, 1, 1),  -- admin(id=1) 属于默认租户
(2, 1, 1),  -- user1(id=2) 属于默认租户
(3, 1, 1),  -- user2(id=3) 属于默认租户
(1, 2, 1);  -- admin(id=1) 也属于演示租户

-- 插入标准菜单
INSERT INTO `menus` (`tenant_id`, `menu_code`, `title`, `menu_type`, `route_path`, `component`, `icon`, `sort_order`, `visible`, `status`) VALUES
-- 一级菜单
(1, 'dashboard', '仪表板', 2, '/dashboard', 'Dashboard', 'dashboard', 1, 1, 1),
(1, 'system', '系统管理', 1, '/system', 'Layout', 'setting', 100, 1, 1),
(1, 'user-manage', '用户管理', 1, '/user', 'Layout', 'user', 200, 1, 1),

-- 系统管理子菜单
(1, 'menu-manage', '菜单管理', 2, '/system/menu', 'system/Menu', 'menu', 101, 1, 1),
(1, 'role-manage', '角色管理', 2, '/system/role', 'system/Role', 'peoples', 102, 1, 1),
(1, 'permission-manage', '权限管理', 2, '/system/permission', 'system/Permission', 'lock', 103, 1, 1),
(1, 'tenant-manage', '租户管理', 2, '/system/tenant', 'system/Tenant', 'office-building', 104, 1, 1),

-- 用户管理子菜单
(1, 'user-list', '用户列表', 2, '/user/list', 'user/List', 'user', 201, 1, 1),
(1, 'user-profile', '个人资料', 2, '/user/profile', 'user/Profile', 'user-filled', 202, 1, 1);

-- 更新菜单层级关系
UPDATE `menus` SET `parent_id` = (SELECT id FROM (SELECT id FROM menus WHERE menu_code = 'system' LIMIT 1) t) WHERE menu_code IN ('menu-manage', 'role-manage', 'permission-manage', 'tenant-manage');
UPDATE `menus` SET `parent_id` = (SELECT id FROM (SELECT id FROM menus WHERE menu_code = 'user-manage' LIMIT 1) t) WHERE menu_code IN ('user-list', 'user-profile');

-- 插入标准权限数据
INSERT INTO `permissions` (`tenant_id`, `permission_code`, `name`, `description`, `resource_type`, `action`, `status`) VALUES
-- 用户管理权限
(1, 'user:view', '查看用户', '查看用户列表和详情', 'menu', 'view', 1),
(1, 'user:create', '创建用户', '创建新用户', 'menu', 'create', 1),
(1, 'user:update', '编辑用户', '修改用户信息', 'menu', 'update', 1),
(1, 'user:delete', '删除用户', '删除用户账号', 'menu', 'delete', 1),
(1, 'user:export', '导出用户', '导出用户数据', 'feature', 'export', 1),

-- 角色管理权限
(1, 'role:view', '查看角色', '查看角色列表和详情', 'menu', 'view', 1),
(1, 'role:create', '创建角色', '创建新角色', 'menu', 'create', 1),
(1, 'role:update', '编辑角色', '修改角色信息', 'menu', 'update', 1),
(1, 'role:delete', '删除角色', '删除角色', 'menu', 'delete', 1),
(1, 'role:assign', '分配角色', '为用户分配角色', 'feature', 'assign', 1),

-- 权限管理
(1, 'permission:view', '查看权限', '查看权限列表', 'menu', 'view', 1),
(1, 'permission:assign', '分配权限', '为角色分配权限', 'feature', 'assign', 1),

-- 菜单管理权限
(1, 'menu:view', '查看菜单', '查看菜单列表', 'menu', 'view', 1),
(1, 'menu:create', '创建菜单', '创建新菜单', 'menu', 'create', 1),
(1, 'menu:update', '编辑菜单', '修改菜单信息', 'menu', 'update', 1),
(1, 'menu:delete', '删除菜单', '删除菜单', 'menu', 'delete', 1),

-- 租户管理权限
(1, 'tenant:view', '查看租户', '查看租户信息', 'menu', 'view', 1),
(1, 'tenant:switch', '切换租户', '切换工作租户', 'feature', 'switch', 1),

-- 系统管理权限
(1, 'dashboard:view', '查看仪表板', '访问系统仪表板', 'menu', 'view', 1),
(1, 'profile:view', '查看个人资料', '查看个人信息', 'menu', 'view', 1),
(1, 'profile:update', '编辑个人资料', '修改个人信息', 'menu', 'update', 1);

-- 配置菜单权限关联（示例）
INSERT INTO `menu_permissions` (`tenant_id`, `menu_id`, `permission_id`, `is_required`)
SELECT 
    1 as tenant_id,
    m.id as menu_id,
    p.id as permission_id,
    1 as is_required
FROM `menus` m
JOIN `permissions` p ON (
    (m.menu_code = 'dashboard' AND p.permission_code = 'dashboard:view') OR
    (m.menu_code = 'menu-manage' AND p.permission_code = 'menu:view') OR
    (m.menu_code = 'role-manage' AND p.permission_code = 'role:view') OR
    (m.menu_code = 'permission-manage' AND p.permission_code = 'permission:view') OR
    (m.menu_code = 'tenant-manage' AND p.permission_code = 'tenant:view') OR
    (m.menu_code = 'user-list' AND p.permission_code = 'user:view') OR
    (m.menu_code = 'user-profile' AND p.permission_code = 'profile:view')
)
WHERE m.deleted_at IS NULL AND p.deleted_at IS NULL;

-- 为超级管理员分配所有权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`, `tenant_id`)
SELECT 1, p.id, 1 FROM `permissions` p WHERE p.tenant_id = 1;

-- 为普通管理员分配部分权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`, `tenant_id`)
SELECT 2, p.id, 1 FROM `permissions` p 
WHERE p.tenant_id = 1 AND p.permission_code IN (
  'dashboard:view', 'user:view', 'user:create', 'user:update', 
  'role:view', 'menu:view', 'profile:view', 'profile:update'
);

-- 为普通用户分配基础权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`, `tenant_id`)
SELECT 3, p.id, 1 FROM `permissions` p 
WHERE p.tenant_id = 1 AND p.permission_code IN (
  'dashboard:view', 'profile:view', 'profile:update'
);

-- =====================================================
-- 恢复环境设置
-- =====================================================

/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- =====================================================
-- 数据类型映射说明
-- =====================================================
-- 
-- auth_type 认证类型映射：
-- 1  - username   (用户名)
-- 2  - email      (邮箱)
-- 3  - phone      (手机号)
-- 4  - wechat     (微信)
-- 5  - qq         (QQ)
-- 6  - github     (Github)
-- 7  - google     (Google)
-- 8  - apple      (Apple)
-- 9  - dingtalk   (钉钉)
-- 10 - feishu     (飞书)
-- 
-- status 用户状态映射：
-- 1 - active      (活跃)
-- 2 - inactive    (未激活)
-- 3 - locked      (锁定)
-- 4 - banned      (封禁)
-- 
-- menu_type 菜单类型映射：
-- 1 - menu        (菜单)
-- 2 - button      (按钮)
-- 3 - api         (接口)
-- 
-- =====================================================
-- Casbin ID映射说明（使用前缀+ID格式）
-- =====================================================
-- 
-- Casbin规则说明：
-- g, u{user_id}, r{role_id}, t{tenant_id}        : 用户角色分配
-- p, r{role_id}, p{permission_id}, t{tenant_id}  : 角色权限分配
-- 
-- 前缀含义：
-- u{id} - 用户ID (user.id)
-- r{id} - 角色ID (role.id)  
-- p{id} - 权限ID (permission.id)
-- t{id} - 租户ID (tenant.id)
-- 
-- 使用前缀+ID格式的优势：
-- 1. 提高可读性：一眼就能看出是什么类型的资源
-- 2. 便于调试：规则含义清晰明了
-- 3. 减少错误：避免混淆不同类型的ID
-- 4. 便于维护：新人容易理解系统逻辑
-- 
-- 所有ID都对应数据库表的主键，无需额外映射：
-- 
-- 用户表 (user):
-- 1 - admin (id=1, username='admin')
-- 2 - user1 (id=2, username='user1') 
-- 3 - user2 (id=3, username='user2')
-- 
-- 角色表 (roles):
-- 1 - super_admin (id=1, role_code='super_admin', 超级管理员)
-- 2 - admin       (id=2, role_code='admin', 系统管理员)
-- 3 - user        (id=3, role_code='user', 普通用户)
-- 
-- 租户表 (tenants):
-- 1 - default (id=1, tenant_code='default', 默认租户)
-- 2 - demo    (id=2, tenant_code='demo', 演示租户)
-- 
-- 权限表 (permissions)，权限ID直接对应permission_code：
-- 1  - user:view        (查看用户)
-- 2  - user:create      (创建用户)
-- 3  - user:update      (编辑用户)
-- 4  - user:delete      (删除用户)
-- 5  - user:export      (导出用户)
-- 6  - role:view        (查看角色)
-- 7  - role:create      (创建角色)
-- 8  - role:update      (编辑角色)
-- 9  - role:delete      (删除角色)
-- 10 - role:assign      (分配角色)
-- 11 - permission:view  (查看权限)
-- 12 - permission:assign(分配权限)
-- 13 - menu:view        (查看菜单)
-- 14 - menu:create      (创建菜单)
-- 15 - menu:update      (编辑菜单)
-- 16 - menu:delete      (删除菜单)
-- 17 - tenant:view      (查看租户)
-- 18 - tenant:switch    (切换租户)
-- 19 - dashboard:view   (查看仪表板)
-- 20 - profile:view     (查看个人资料)
-- 21 - profile:update   (编辑个人资料)
-- 
-- 优势：
-- 1. 前缀格式增强可读性和可维护性
-- 2. 底层仍使用主键ID，性能优异
-- 3. 无需维护额外的映射关系
-- 4. 数据一致性更强
-- 5. 调试和管理更加便捷
-- 
-- =====================================================
-- 验证数据
-- =====================================================

-- 查看租户数据
SELECT 'Tenant Table Data' as info;
SELECT id, tenant_code, name, description, status, created_at FROM tenants;

-- 查看角色数据
SELECT 'Role Table Data' as info;
SELECT id, tenant_id, role_code, name, description, status FROM roles;

-- 查看用户表数据
SELECT 'User Table Data' as info;
SELECT id, username, nickname, email, phone, created_at FROM user;

-- 查看用户状态表数据
SELECT 'User Status Table Data' as info;
SELECT 
    user_id,
    auth_id,
    CASE auth_type 
        WHEN 1 THEN 'username'
        WHEN 2 THEN 'email' 
        WHEN 3 THEN 'phone'
        ELSE CONCAT('type_', auth_type)
    END as auth_type_name,
    CASE status
        WHEN 1 THEN 'active'
        WHEN 2 THEN 'inactive'
        WHEN 3 THEN 'locked'
        WHEN 4 THEN 'banned'
    END as status_name,
    is_verified,
    is_primary,
    created_at
FROM user_status 
ORDER BY user_id, auth_type;

-- 查看菜单数据
SELECT 'Menu Table Data' as info;
SELECT id, tenant_id, menu_code, title, route_path, menu_type, visible, status FROM menus;

-- 查看权限数据
SELECT 'Permission Table Data' as info;
SELECT id, tenant_id, permission_code, name, resource_type, action, status FROM permissions;

-- 统计信息
SELECT 'Statistics' as info;
SELECT 
    (SELECT COUNT(*) FROM tenants) as total_tenants,
    (SELECT COUNT(*) FROM roles) as total_roles,
    (SELECT COUNT(*) FROM user) as total_users,
    (SELECT COUNT(*) FROM user_status) as total_auth_methods,
    (SELECT COUNT(DISTINCT user_id) FROM user_status) as users_with_auth,
    (SELECT COUNT(*) FROM user_status WHERE is_primary = 1) as primary_methods,
    (SELECT COUNT(*) FROM user_status WHERE is_verified = 1) as verified_methods,
    (SELECT COUNT(*) FROM menus) as total_menus,
    (SELECT COUNT(*) FROM permissions) as total_permissions,
    (SELECT COUNT(*) FROM role_permissions) as total_role_permissions,
    (SELECT COUNT(*) FROM user_tenants) as total_user_tenants,
    (SELECT COUNT(*) FROM casbin_rule) as total_casbin_rules;

-- 数据库初始化完成
SELECT 'Database initialization completed successfully!' as result;
