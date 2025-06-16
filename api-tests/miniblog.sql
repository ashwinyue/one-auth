/*
 Navicat Premium Data Transfer

 Source Server         : docker
 Source Server Type    : MySQL
 Source Server Version : 80042
 Source Host           : localhost:3306
 Source Schema         : miniblog

 Target Server Type    : MySQL
 Target Server Version : 80042
 File Encoding         : 65001

 Date: 17/06/2025 00:26:06
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for casbin_rule
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule`;
CREATE TABLE `casbin_rule`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `v0` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `v1` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `v2` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `v3` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `v4` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `v5` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_casbin_rule`(`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 36 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '权限管理规则表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of casbin_rule
-- ----------------------------
INSERT INTO `casbin_rule` VALUES (1, 'g', 'u1', 'r1', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (2, 'g', 'u2', 'r2', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (3, 'g', 'u3', 'r3', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (4, 'p', 'r1', 'p1', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (13, 'p', 'r1', 'p10', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (14, 'p', 'r1', 'p11', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (15, 'p', 'r1', 'p12', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (16, 'p', 'r1', 'p13', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (17, 'p', 'r1', 'p14', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (18, 'p', 'r1', 'p15', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (19, 'p', 'r1', 'p16', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (20, 'p', 'r1', 'p17', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (21, 'p', 'r1', 'p18', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (22, 'p', 'r1', 'p19', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (5, 'p', 'r1', 'p2', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (23, 'p', 'r1', 'p20', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (24, 'p', 'r1', 'p21', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (6, 'p', 'r1', 'p3', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (7, 'p', 'r1', 'p4', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (8, 'p', 'r1', 'p5', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (9, 'p', 'r1', 'p6', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (10, 'p', 'r1', 'p7', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (11, 'p', 'r1', 'p8', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (12, 'p', 'r1', 'p9', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (25, 'p', 'r2', 'p1', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (29, 'p', 'r2', 'p13', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (30, 'p', 'r2', 'p19', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (26, 'p', 'r2', 'p2', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (31, 'p', 'r2', 'p20', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (32, 'p', 'r2', 'p21', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (27, 'p', 'r2', 'p3', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (28, 'p', 'r2', 'p6', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (33, 'p', 'r3', 'p19', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (34, 'p', 'r3', 'p20', 't1', NULL, '', '');
INSERT INTO `casbin_rule` VALUES (35, 'p', 'r3', 'p21', 't1', NULL, '', '');

-- ----------------------------
-- Table structure for menu_permissions
-- ----------------------------
DROP TABLE IF EXISTS `menu_permissions`;
CREATE TABLE `menu_permissions`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `menu_id` bigint NOT NULL COMMENT '菜单ID',
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `is_required` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为访问菜单的必需权限',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_menu_permission`(`menu_id`, `permission_id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id`) USING BTREE,
  INDEX `idx_menu_id`(`menu_id`) USING BTREE,
  INDEX `idx_permission_id`(`permission_id`) USING BTREE,
  INDEX `idx_required`(`is_required`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 8 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '菜单权限关联表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of menu_permissions
-- ----------------------------
INSERT INTO `menu_permissions` VALUES (1, 1, 1, 19, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40');
INSERT INTO `menu_permissions` VALUES (2, 1, 4, 13, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40');
INSERT INTO `menu_permissions` VALUES (3, 1, 5, 6, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40');
INSERT INTO `menu_permissions` VALUES (4, 1, 6, 11, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40');
INSERT INTO `menu_permissions` VALUES (5, 1, 7, 17, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40');
INSERT INTO `menu_permissions` VALUES (6, 1, 8, 1, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40');
INSERT INTO `menu_permissions` VALUES (7, 1, 9, 20, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40');

-- ----------------------------
-- Table structure for menus
-- ----------------------------
DROP TABLE IF EXISTS `menus`;
CREATE TABLE `menus`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '菜单主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `parent_id` bigint NULL DEFAULT NULL COMMENT '父菜单ID',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单标题',
  `menu_type` tinyint NOT NULL DEFAULT 1 COMMENT '菜单类型：1-目录，2-菜单，3-按钮，4-接口',
  `route_path` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '前端路由路径',
  `component` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '前端组件路径',
  `icon` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '图标',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序',
  `visible` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否可见',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id`) USING BTREE,
  INDEX `idx_parent_id`(`parent_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_menu_type`(`menu_type`) USING BTREE,
  INDEX `idx_visible`(`visible`) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '菜单表（重构版-纯UI结构）' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of menus
-- ----------------------------
INSERT INTO `menus` VALUES (1, 1, NULL, '仪表板', 2, '/dashboard', 'Dashboard', 'dashboard', 1, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (2, 1, NULL, '系统管理', 1, '/system', 'Layout', 'setting', 100, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (3, 1, NULL, '用户管理', 1, '/user', 'Layout', 'user', 200, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (4, 1, 2, '菜单管理', 2, '/system/menu', 'system/Menu', 'menu', 101, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (5, 1, 2, '角色管理', 2, '/system/role', 'system/Role', 'peoples', 102, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (6, 1, 2, '权限管理', 2, '/system/permission', 'system/Permission', 'lock', 103, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (7, 1, 2, '租户管理', 2, '/system/tenant', 'system/Tenant', 'office-building', 104, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (8, 1, 3, '用户列表', 2, '/user/list', 'user/List', 'user', 201, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `menus` VALUES (9, 1, 3, '个人资料', 2, '/user/profile', 'user/Profile', 'user-filled', 202, 1, 1, NULL, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);

-- ----------------------------
-- Table structure for permissions
-- ----------------------------
DROP TABLE IF EXISTS `permissions`;
CREATE TABLE `permissions`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '权限主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '权限名称',
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '权限描述',
  `resource_type` enum('api','menu','data','feature') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'menu' COMMENT '资源类型',
  `resource_path` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'API路径或资源标识',
  `http_method` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'HTTP方法：GET,POST,PUT,DELETE等',
  `action` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '操作类型：view,create,update,delete,export等',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id`) USING BTREE,
  INDEX `idx_resource_type`(`resource_type`) USING BTREE,
  INDEX `idx_action`(`action`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 22 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '权限表（重构版-独立权限管理）' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of permissions
-- ----------------------------
INSERT INTO `permissions` VALUES (1, 1, '查看用户', '查看用户列表和详情', 'menu', NULL, NULL, 'view', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (2, 1, '创建用户', '创建新用户', 'menu', NULL, NULL, 'create', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (3, 1, '编辑用户', '修改用户信息', 'menu', NULL, NULL, 'update', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (4, 1, '删除用户', '删除用户账号', 'menu', NULL, NULL, 'delete', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (5, 1, '导出用户', '导出用户数据', 'feature', NULL, NULL, 'export', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (6, 1, '查看角色', '查看角色列表和详情', 'menu', NULL, NULL, 'view', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (7, 1, '创建角色', '创建新角色', 'menu', NULL, NULL, 'create', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (8, 1, '编辑角色', '修改角色信息', 'menu', NULL, NULL, 'update', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (9, 1, '删除角色', '删除角色', 'menu', NULL, NULL, 'delete', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (10, 1, '分配角色', '为用户分配角色', 'feature', NULL, NULL, 'assign', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (11, 1, '查看权限', '查看权限列表', 'menu', NULL, NULL, 'view', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (12, 1, '分配权限', '为角色分配权限', 'feature', NULL, NULL, 'assign', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (13, 1, '查看菜单', '查看菜单列表', 'menu', NULL, NULL, 'view', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (14, 1, '创建菜单', '创建新菜单', 'menu', NULL, NULL, 'create', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (15, 1, '编辑菜单', '修改菜单信息', 'menu', NULL, NULL, 'update', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (16, 1, '删除菜单', '删除菜单', 'menu', NULL, NULL, 'delete', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (17, 1, '查看租户', '查看租户信息', 'menu', NULL, NULL, 'view', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (18, 1, '切换租户', '切换工作租户', 'feature', NULL, NULL, 'switch', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (19, 1, '查看仪表板', '访问系统仪表板', 'menu', NULL, NULL, 'view', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (20, 1, '查看个人资料', '查看个人信息', 'menu', NULL, NULL, 'view', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `permissions` VALUES (21, 1, '编辑个人资料', '修改个人信息', 'menu', NULL, NULL, 'update', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);

-- ----------------------------
-- Table structure for post
-- ----------------------------
DROP TABLE IF EXISTS `post`;
CREATE TABLE `post`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '博文主键ID',
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID（关联user表）',
  `post_id` varchar(35) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '博文唯一标识',
  `title` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '博文标题',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '博文内容',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '博文创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '博文最后修改时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_post_post_id`(`post_id`) USING BTREE,
  INDEX `idx_post_user_id`(`user_id`) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE,
  CONSTRAINT `post_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '博文表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of post
-- ----------------------------

-- ----------------------------
-- Table structure for roles
-- ----------------------------
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '描述',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_name_tenant`(`name`, `tenant_id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '角色表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of roles
-- ----------------------------
INSERT INTO `roles` VALUES (1, 1, '超级管理员', '超级管理员角色（不可删除）', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `roles` VALUES (2, 1, '系统管理员', '拥有系统所有权限', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `roles` VALUES (3, 1, '普通用户', '普通用户权限', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `roles` VALUES (4, 2, '超级管理员', '演示租户超级管理员', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `roles` VALUES (5, 2, '租户管理员', '租户管理员权限', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);

-- ----------------------------
-- Table structure for tenants
-- ----------------------------
DROP TABLE IF EXISTS `tenants`;
CREATE TABLE `tenants`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '租户主键ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '租户名称',
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '描述',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '租户表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tenants
-- ----------------------------
INSERT INTO `tenants` VALUES (1, '默认租户', '系统默认租户', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `tenants` VALUES (2, '演示租户', '演示用租户', 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户主键ID',
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户名（唯一）',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户密码（加密后）',
  `nickname` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户昵称',
  `email` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户电子邮箱地址',
  `phone` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户手机号',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '用户创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '用户最后修改时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_user_username`(`username`) USING BTREE,
  UNIQUE INDEX `idx_user_phone`(`phone`) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO `user` VALUES (1, 'admin', '$2a$10$ctsFXEUAMd7rXXpmccNlO.ZRiYGYz0eOfj8EicPGWqiz64YBBgR1y', '管理员', 'admin@example.com', '13800138000', '2024-12-12 03:55:25', '2024-12-12 03:55:25', NULL);
INSERT INTO `user` VALUES (2, 'user1', '$2a$10$ctsFXEUAMd7rXXpmccNlO.ZRiYGYz0eOfj8EicPGWqiz64YBBgR1y', '用户1', 'user1@example.com', '13800138001', '2024-12-12 03:55:25', '2024-12-12 03:55:25', NULL);
INSERT INTO `user` VALUES (3, 'user2', '$2a$10$ctsFXEUAMd7rXXpmccNlO.ZRiYGYz0eOfj8EicPGWqiz64YBBgR1y', '用户2', 'user2@example.com', '13800138002', '2024-12-12 03:55:25', '2024-12-12 03:55:25', NULL);

-- ----------------------------
-- Table structure for user_status
-- ----------------------------
DROP TABLE IF EXISTS `user_status`;
CREATE TABLE `user_status`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `auth_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '认证标识符（邮箱、手机号、用户名等）',
  `auth_type` tinyint NOT NULL COMMENT '认证类型：1-username,2-email,3-phone,4-wechat,5-qq,6-github,7-google,8-apple,9-dingtalk,10-feishu',
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID（关联user表的id）',
  `tenant_id` bigint NOT NULL DEFAULT 1 COMMENT '租户ID',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '用户状态：1-active,2-inactive,3-locked,4-banned',
  `lock_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '锁定原因',
  `locked_until` timestamp NULL DEFAULT NULL COMMENT '锁定到期时间',
  `last_login_time` timestamp NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '最后登录IP',
  `last_login_device` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '最后登录设备',
  `login_count` int NOT NULL DEFAULT 0 COMMENT '登录次数',
  `failed_login_attempts` int NOT NULL DEFAULT 0 COMMENT '累计登录失败次数',
  `last_failed_login` timestamp NULL DEFAULT NULL COMMENT '最后一次登录失败时间',
  `password_changed_at` timestamp NULL DEFAULT NULL COMMENT '密码最后修改时间',
  `is_verified` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已验证',
  `verified_at` timestamp NULL DEFAULT NULL COMMENT '验证时间',
  `is_primary` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为主要认证方式',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_auth_id_type`(`auth_id`, `auth_type`) USING BTREE COMMENT '认证标识符+类型唯一',
  INDEX `idx_user_id`(`user_id`) USING BTREE COMMENT '用户ID索引（非唯一，一个用户多个认证方式）',
  INDEX `idx_auth_id_type_tenant`(`auth_id`, `auth_type`, `tenant_id`) USING BTREE COMMENT '认证查询索引',
  INDEX `idx_user_tenant`(`user_id`, `tenant_id`) USING BTREE COMMENT '用户租户索引',
  INDEX `idx_status`(`status`) USING BTREE COMMENT '状态索引',
  INDEX `idx_auth_type`(`auth_type`) USING BTREE COMMENT '认证类型索引',
  INDEX `idx_primary`(`is_primary`) USING BTREE COMMENT '主要认证方式索引',
  INDEX `idx_verified`(`is_verified`) USING BTREE COMMENT '验证状态索引',
  INDEX `idx_last_login`(`last_login_time`) USING BTREE COMMENT '最后登录时间索引',
  INDEX `idx_tenant_id`(`tenant_id`) USING BTREE COMMENT '租户ID索引',
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE COMMENT '软删除索引'
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户状态表-支持多种认证方式' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_status
-- ----------------------------
INSERT INTO `user_status` VALUES (1, 'admin', 1, 1, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 1, NULL, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (2, 'admin@example.com', 2, 1, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 1, NULL, 0, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (3, '13800138000', 3, 1, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 1, NULL, 0, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (4, 'user1', 1, 2, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 1, NULL, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (5, 'user1@example.com', 2, 2, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 1, NULL, 0, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (6, '13800138001', 3, 2, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 1, NULL, 0, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (7, 'user2', 1, 3, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 1, NULL, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (8, 'user2@example.com', 2, 3, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 0, NULL, 0, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_status` VALUES (9, '13800138002', 3, 3, 1, 1, NULL, NULL, NULL, NULL, NULL, 0, 0, NULL, NULL, 0, NULL, 0, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);

-- ----------------------------
-- Table structure for user_tenants
-- ----------------------------
DROP TABLE IF EXISTS `user_tenants`;
CREATE TABLE `user_tenants`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID（关联user表的id字段）',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_user_tenant`(`user_id`, `tenant_id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 5 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户租户关联表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_tenants
-- ----------------------------
INSERT INTO `user_tenants` VALUES (1, 1, 1, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_tenants` VALUES (2, 2, 1, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_tenants` VALUES (3, 3, 1, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);
INSERT INTO `user_tenants` VALUES (4, 1, 2, 1, '2025-06-15 15:34:40', '2025-06-15 15:34:40', NULL);

SET FOREIGN_KEY_CHECKS = 1;
