# One-Auth 数据库管理脚本

## 概述

本目录包含 One-Auth 用户认证系统的数据库管理脚本，支持完整的多租户RBAC权限管理系统。

## 数据库架构

### 核心表结构

| 表名 | 说明 | 主要功能 |
|------|------|----------|
| `casbin_rule` | 权限规则表 | Casbin权限管理规则存储 |
| `tenants` | 租户表 | 多租户系统的租户信息 |
| `roles` | 角色表 | 角色定义和管理 |
| `menus` | 菜单表 | 前端菜单和API路径管理 |
| `permissions` | 权限表 | 细粒度权限定义 |
| `user_tenants` | 用户租户关联表 | 用户与租户的多对多关系 |
| `role_permissions` | 角色权限关联表 | 角色与权限的多对多关系 |
| `user` | 用户表 | 用户基本信息（兼容原有结构） |
| `user_status` | 用户状态表 | 多认证方式和用户状态管理 |
| `post` | 博文表 | 博客文章管理 |

### 数据类型映射

#### 认证类型 (auth_type)
- `1` - username (用户名)
- `2` - email (邮箱)
- `3` - phone (手机号)
- `4` - wechat (微信)
- `5` - qq (QQ)
- `6` - github (Github)
- `7` - google (Google)
- `8` - apple (Apple)
- `9` - dingtalk (钉钉)
- `10` - feishu (飞书)

#### 用户状态 (status)
- `1` - active (活跃)
- `2` - inactive (未激活)
- `3` - locked (锁定)
- `4` - banned (封禁)

#### 菜单类型 (menu_type)
- `1` - menu (菜单)
- `2` - button (按钮)
- `3` - api (接口)

## 数据库初始化

### 快速开始

使用 `init-db.sh` 脚本进行数据库初始化：

```bash
# Docker方式（推荐）
./scripts/init-db.sh -d

# 直接连接方式
./scripts/init-db.sh -c

# 交互式客户端
./scripts/init-db.sh -i
```

### 执行方式详解

#### 1. Docker方式（推荐）
```bash
./scripts/init-db.sh -d
```
- 自动检测Docker容器状态
- 使用root用户执行SQL脚本
- 适合开发和测试环境

#### 2. 直接连接方式
```bash
./scripts/init-db.sh -c
```
- 直接连接MySQL服务器
- 使用miniblog用户执行
- 适合生产环境

#### 3. 交互式客户端
```bash
./scripts/init-db.sh -i
```
- 启动MySQL交互式客户端
- 手动执行SQL命令
- 适合调试和开发

### 手动执行

如果需要手动执行SQL脚本：

```bash
# Docker方式
docker exec -i miniblog-mysql mysql -u root -proot123456 < configs/miniblog.sql

# 直接连接方式
mysql -h 127.0.0.1 -P 3306 -u miniblog -pminiblog1234 < configs/miniblog.sql
```

## 初始化数据

### 默认租户
- `default` - 默认租户（系统默认租户）
- `demo` - 演示租户（演示用租户）

### 默认角色
- `admin` - 系统管理员（拥有系统所有权限）
- `user` - 普通用户（普通用户权限）

### 默认用户
| 用户名 | 密码 | 角色 | 邮箱 | 手机号 |
|--------|------|------|------|--------|
| admin | admin123 | 管理员 | admin@example.com | 13800138000 |
| user1 | user123 | 普通用户 | user1@example.com | 13800138001 |
| user2 | user123 | 普通用户 | user2@example.com | 13800138002 |

### 多认证方式支持
每个用户支持多种登录方式：
- 用户名登录
- 邮箱登录
- 手机号登录
- 第三方登录（预留接口）

## 权限管理

### Casbin集成
- 支持RBAC with Domains模型
- 多租户权限隔离
- 细粒度权限控制

### 菜单权限
- 前端路由权限控制
- API接口权限控制
- 按钮级权限控制

### 示例权限配置
```
管理员权限：
- dashboard:view - 查看仪表盘
- user:view - 查看用户
- user:create - 创建用户
- user:update - 更新用户
- user:delete - 删除用户

普通用户权限：
- dashboard:view - 查看仪表盘
- user:view - 查看用户（仅自己）
```

## 数据库连接配置

### Docker环境
```yaml
# docker-compose.yml
services:
  mysql:
    image: mysql:8.0
    container_name: miniblog-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root123456
      MYSQL_DATABASE: miniblog
      MYSQL_USER: miniblog
      MYSQL_PASSWORD: miniblog1234
    ports:
      - "3306:3306"
```

### 应用配置
```yaml
# configs/mb-apiserver.yaml
mysql:
  addr: 127.0.0.1:3306
  username: miniblog
  password: miniblog1234
  database: miniblog
  max-idle-connections: 100
  max-open-connections: 100
  max-connection-life-time: 10s
  log-level: 1
```

## 安全特性

### 用户安全
- 密码加密存储（bcrypt）
- 登录失败次数限制
- 账户锁定机制
- 多设备会话管理

### 数据安全
- 软删除支持
- 审计日志记录
- 数据完整性约束
- 索引优化

## 性能优化

### 索引设计
- 主键索引：所有表都有自增主键
- 唯一索引：用户名、邮箱、手机号等唯一字段
- 复合索引：多字段查询优化
- 外键索引：关联查询优化

### 查询优化
- 分页查询支持
- 条件查询优化
- 关联查询优化
- 缓存策略支持

## 故障排除

### 常见问题

1. **连接失败**
   ```bash
   # 检查MySQL服务状态
   docker ps | grep mysql
   
   # 检查端口占用
   netstat -tlnp | grep 3306
   ```

2. **权限错误**
   ```bash
   # 检查用户权限
   docker exec -it miniblog-mysql mysql -u root -proot123456 -e "SELECT user,host FROM mysql.user;"
   ```

3. **字符集问题**
   ```bash
   # 检查字符集设置
   docker exec -it miniblog-mysql mysql -u root -proot123456 -e "SHOW VARIABLES LIKE 'character%';"
   ```

### 日志查看
```bash
# 查看MySQL容器日志
docker logs miniblog-mysql

# 查看应用日志
tail -f logs/mb-apiserver.log
```

## 开发指南

### 模型生成
```bash
# 生成GORM模型
go run cmd/gen-gorm-model/gen_gorm_model.go -component mb
```

### 数据迁移
```bash
# 自动迁移（开发环境）
# 在应用启动时自动执行

# 手动迁移（生产环境）
# 使用SQL脚本手动执行
```

### 测试数据
```bash
# 重新初始化测试数据
./scripts/init-db.sh -d
```

## 版本历史

- **v1.0.0** - 基础用户认证系统
- **v2.0.0** - 多租户RBAC权限系统
- **v2.1.0** - 多认证方式支持
- **v2.2.0** - 完整数据库架构整合

## 贡献指南

1. 所有数据库变更必须通过SQL脚本管理
2. 新增表结构需要更新模型文件
3. 权限变更需要更新Casbin规则
4. 测试数据需要保持一致性

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../LICENSE) 文件。 