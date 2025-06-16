# One-Auth 企业级认证权限系统

## 项目概述

One-Auth 是一个基于微服务架构的企业级用户认证与权限管理系统，专为现代企业应用设计。系统采用多租户架构，支持灵活的权限控制和安全的用户认证机制。

## 核心特性

- 🔐 **多种认证方式**：用户名密码、短信验证码、邮箱验证等
- 👥 **多租户架构**：完善的租户隔离和数据安全
- 🛡️ **基于RBAC的权限控制**：使用Casbin实现灵活的权限管理
- 📱 **短信验证系统**：集成多厂商SMS服务
- 🔑 **JWT Token管理**：安全的会话管理机制
- 🎯 **菜单权限分离**：功能权限与菜单权限独立管理
- 📊 **完善的日志审计**：操作追踪和安全审计

## 权限系统设计

### RBAC模型
系统采用基于角色的访问控制（RBAC）模型，支持以下层次结构：

```
租户(Tenant) → 用户(User) → 角色(Role) → 权限(Permission) → 资源(Resource)
```

### 主键ID优化
- **角色ID格式**：`tenant_{tenant_id}_role_{role_id}`
- **用户角色关联**：`user_{user_id}_tenant_{tenant_id}`
- **权限策略**：`tenant_{tenant_id}_resource_{resource_id}`

### 菜单权限分离设计

系统将功能权限和菜单权限进行分离管理：

#### 功能权限
- API接口访问控制
- 数据操作权限（增删改查）
- 资源级别的细粒度控制

#### 菜单权限
- 前端菜单显示控制
- 页面访问权限
- UI组件级别权限

 ## 短信验证系统

### 核心特性
- **多厂商支持**：阿里云、腾讯云、华为云等
- **验证码类型**：登录、注册、重置密码、绑定手机等
- **安全控制**：
  - 验证码有效期：10分钟
  - 发送冷却时间：1分钟
  - 一次性使用机制
  - 频率限制和防刷机制

### Redis缓存设计
```
验证码存储键：verify_code:{type}:{target}
冷却时间键：verify_cooldown:{type}:{target}
```

### 实现位置
短信验证码的缓存和验证逻辑主要在 `internal/apiserver/cache/login_security.go` 文件中实现。

## 技术栈

### 后端技术
- **语言**：Go 1.21+
- **框架**：Gin Web Framework
- **数据库**：MySQL 8.0+
- **缓存**：Redis 6.0+
- **权限**：Casbin v2
- **认证**：JWT-Go
- **配置**：Viper
- **日志**：Logrus

## 快速开始

### 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 安装步骤
```bash
# 克隆项目
git clone https://github.com/your-org/one-auth.git
cd one-auth

# 安装依赖
go mod download

# 配置数据库
# 编辑 configs/config.yaml

# 运行数据库迁移
make migrate

# 启动服务
make run
```


## 完整API接口列表

### 健康检查
```
GET    /healthz                       # 服务健康检查
```

### 认证接口
```
POST   /login                         # 用户登录
POST   /send-verify-code              # 发送短信验证码（无需认证）
PUT    /refresh-token                 # 刷新访问令牌
POST   /logout                        # 用户登出
```

### 用户管理
```
POST   /v1/users                      # 创建用户（无需认证）
PUT    /v1/users/:userID/change-password # 修改用户密码
PUT    /v1/users/:userID              # 更新用户信息
DELETE /v1/users/:userID              # 删除用户
GET    /v1/users/:userID              # 获取用户详情
GET    /v1/users                      # 获取用户列表
```
### 角色管理
```
GET    /v1/roles                      # 获取角色列表
POST   /v1/roles                      # 创建角色
PUT    /v1/roles/:roleID              # 更新角色
DELETE /v1/roles/:roleID              # 删除角色
```

### 权限管理
```
POST   /v1/permissions/check          # 批量检查权限
GET    /v1/user/permissions           # 获取当前用户权限
GET    /v1/api/check-access           # 检查API访问权限
```

### 租户管理
```
GET    /v1/user/tenants               # 获取用户所属租户列表
GET    /v1/user/profile               # 获取用户完整信息（含租户、角色、权限）
POST   /v1/tenant/switch              # 切换当前工作租户
GET    /v1/tenants                    # 获取租户列表
```
### 菜单管理
```
POST   /v1/menus                      # 创建菜单
GET    /v1/menus/:id                  # 获取菜单详情
PUT    /v1/menus/:id                  # 更新菜单
DELETE /v1/menus/:id                  # 删除菜单
GET    /v1/menus                      # 获取菜单列表
GET    /v1/menus/tree                 # 获取菜单树
GET    /v1/menus/user                 # 获取用户菜单
PUT    /v1/menus/sort                 # 批量更新菜单排序
POST   /v1/menus/copy                 # 复制菜单
PUT    /v1/menus/move                 # 移动菜单
```

### 博客管理（示例模块）
```
POST   /v1/posts                      # 创建博客
PUT    /v1/posts/:postID              # 更新博客
DELETE /v1/posts                      # 删除博客
GET    /v1/posts/:postID              # 获取博客详情
GET    /v1/posts                      # 获取博客列表
```
## 项目结构

```
├── cmd/                    # 应用程序入口
├── internal/               # 内部应用代码
│   ├── apiserver/         # API服务器
│   ├── authz/             # 权限控制
│   ├── pkg/               # 内部包
│   └── store/             # 数据存储
├── pkg/                   # 可重用的包
├── configs/               # 配置文件
├── scripts/               # 脚本工具
├── docs/                  # 文档
└── deployments/           # 部署配置
```

## 核心模块说明

### 短信验证码系统
- **位置**：`internal/apiserver/cache/login_security.go`
- **功能**：验证码生成、存储、验证和状态管理
- **缓存**：基于Redis的验证码缓存机制
- **安全**：冷却时间控制和频率限制

### 权限控制系统
- **位置**：`internal/authz/`
- **引擎**：基于Casbin的RBAC权限控制
- **特性**：多租户权限隔离、菜单权限分离

### Casbin权限配置详解

#### 策略配置模型
```ini
[request_definition]
r = sub, obj, act, tenant

[policy_definition]
p = sub, obj, act, tenant

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.tenant) && r.obj == p.obj && r.act == p.act && r.tenant == p.tenant
```

#### 数据权限策略
- **行级权限**：基于用户所属组织的数据隔离
- **字段级权限**：敏感字段的读写控制
- **租户隔离**：确保租户间数据完全隔离

### 系统架构设计

#### 整体架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   移动端应用    │    │   第三方应用    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
          ┌─────────────────────────────────────────────┐
          │              API Gateway                    │
          └─────────────────┬───────────────────────────┘
                            │
          ┌─────────────────────────────────────────────┐
          │             One-Auth 服务                   │
          │  ┌─────────────┐  ┌─────────────┐           │
          │  │ 认证服务    │  │ 权限服务    │           │
          │  └─────────────┘  └─────────────┘           │
          │  ┌─────────────┐  ┌─────────────┐           │
          │  │ 用户服务    │  │ 租户服务    │           │
          │  └─────────────┘  └─────────────┘           │
          └─────────────────┬───────────────────────────┘
                            │
          ┌─────────────────────────────────────────────┐
          │              数据层                         │
          │  ┌─────────────┐  ┌─────────────┐           │
          │  │   MySQL     │  │    Redis    │           │
          │  │   主数据    │  │   缓存/会话 │           │
          │  └─────────────┘  └─────────────┘           │
          └─────────────────────────────────────────────┘
```

#### 微服务模块架构
```
One-Auth 服务
├── 认证模块 (Authentication)
│   ├── JWT Token 管理
│   ├── 多因子认证
│   └── 会话管理
├── 授权模块 (Authorization)
│   ├── RBAC 权限控制
│   ├── 动态权限验证
│   └── 权限缓存
├── 用户管理模块 (User Management)
│   ├── 用户生命周期
│   ├── 密码策略
│   └── 用户画像
├── 租户管理模块 (Tenant Management)
│   ├── 多租户隔离
│   ├── 资源配额
│   └── 租户配置
├── 菜单管理模块 (Menu Management)
│   ├── 动态菜单
│   ├── 权限绑定
│   └── 菜单树构建
└── 审计模块 (Audit)
    ├── 操作日志
    ├── 安全审计
    └── 合规报告
```

## 多租户设计

### 数据隔离策略
- **数据库级隔离**：为每个租户分配独立的数据库schema
- **表级隔离**：通过tenant_id字段实现逻辑隔离
- **应用级隔离**：在应用层确保数据访问的租户边界

### 租户配置管理
```go
type TenantConfig struct {
    TenantID     string            `json:"tenant_id"`
    TenantName   string            `json:"tenant_name"`
    MaxUsers     int               `json:"max_users"`
    Features     []string          `json:"features"`
    CustomConfig map[string]interface{} `json:"custom_config"`
}
```

### 权限隔离机制
- **角色隔离**：租户间角色完全独立
- **权限隔离**：权限策略基于租户维度
- **资源隔离**：API访问自动注入租户上下文

## 用户认证系统

### 多种认证方式支持

#### 1. 用户名密码认证
- **密码策略**：支持复杂度要求、历史密码检查
- **密码加密**：使用bcrypt进行密码哈希
- **失败锁定**：连续失败自动锁定账户

#### 2. 短信验证码认证
- **验证码生成**：6位数字随机码
- **有效期控制**：10分钟有效期
- **频率限制**：1分钟冷却时间
- **防刷机制**：IP和手机号双重限制

#### 3. 邮箱验证认证
- **邮箱验证**：注册时邮箱验证
- **找回密码**：邮箱重置密码链接
- **安全通知**：重要操作邮箱通知

### JWT Token 设计

#### Token 结构
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "user_123",
    "tenant_id": "tenant_456", 
    "username": "john_doe",
    "roles": ["admin", "user"],
    "permissions": ["read", "write"],
    "session_id": "session_789",
    "iat": 1640995200,
    "exp": 1641081600
  }
}
```

#### 会话管理
- **单点登录**：支持跨应用的SSO
- **会话超时**：可配置的会话超时策略
- **令牌刷新**：Access Token + Refresh Token 机制
- **并发控制**：可限制用户并发登录数量

## 性能特性

### 缓存策略
- **多级缓存**：内存缓存 + Redis缓存
- **权限缓存**：用户权限信息缓存优化
- **会话缓存**：JWT Token黑名单缓存
- **菜单缓存**：用户菜单树缓存

### 性能优化
- **连接池**：数据库连接池管理
- **索引优化**：核心查询路径索引优化
- **异步处理**：耗时操作异步化
- **批量操作**：权限批量验证接口

### 监控指标
```
- 接口响应时间 (P95 < 200ms)
- 数据库查询时间 (P95 < 100ms)
- 缓存命中率 (> 90%)
- 并发处理能力 (> 1000 QPS)
```

## 安全特性

### 认证安全
- **密码策略**：强密码要求、密码历史检查
- **账户锁定**：失败次数限制、自动解锁
- **会话安全**：会话超时、并发登录控制
- **设备绑定**：可选的设备认证

### 数据安全
- **传输加密**：HTTPS/TLS 1.3强制加密
- **存储加密**：敏感数据AES-256加密
- **脱敏处理**：日志和响应数据脱敏
- **数据备份**：定期自动备份

### 访问控制
- **最小权限原则**：默认拒绝策略
- **权限时效性**：支持临时权限授予
- **IP白名单**：可配置IP访问限制
- **API限流**：防止恶意调用

### 审计安全
- **操作审计**：所有关键操作记录
- **登录审计**：登录行为完整记录
- **权限变更审计**：权限修改全程追踪
- **异常行为检测**：自动识别可疑操作

## 扩展性设计

### 水平扩展
- **无状态设计**：服务实例无状态，支持负载均衡
- **数据库分片**：支持数据库读写分离和分片
- **缓存集群**：Redis集群支持
- **微服务架构**：模块化设计，独立部署

### 插件化架构
```go
// 认证插件接口
type AuthPlugin interface {
    Name() string
    Authenticate(ctx context.Context, req *AuthRequest) (*AuthResponse, error)
}

// 权限插件接口
type AuthzPlugin interface {
    Name() string
    Authorize(ctx context.Context, req *AuthzRequest) bool
}
```

### 配置化管理
- **动态配置**：支持配置热更新
- **多环境配置**：开发、测试、生产环境隔离
- **特性开关**：功能特性开关控制
- **租户定制**：租户级别功能定制

## 开发指南

### 本地开发环境搭建

#### 1. 环境准备
```bash
# 安装Go 1.21+
go version

# 安装Docker和Docker Compose
docker --version
docker-compose --version

# 安装Make工具
make --version
```

#### 2. 依赖服务启动
```bash
# 启动MySQL和Redis
docker-compose up -d mysql redis

# 等待服务就绪
make wait-db
```

#### 3. 数据库初始化
```bash
# 运行数据库迁移
make migrate

# 初始化基础数据
make seed

# 创建超级管理员
make create-admin
```

#### 4. 服务启动
```bash
# 开发模式启动
make dev

# 或者使用air进行热重载
air
```

### 代码结构规范

#### 目录组织
```
internal/
├── apiserver/          # API服务层
│   ├── handler/       # HTTP处理器
│   ├── middleware/    # 中间件
│   ├── routes/        # 路由定义
│   └── validation/    # 请求验证
├── biz/               # 业务逻辑层
│   ├── auth/         # 认证业务
│   ├── user/         # 用户业务
│   └── permission/   # 权限业务
├── store/             # 数据访问层
│   ├── mysql/        # MySQL存储
│   └── redis/        # Redis存储
└── pkg/               # 内部工具包
    ├── middleware/   # 通用中间件
    ├── cache/       # 缓存工具
    └── utils/       # 工具函数
```

#### 命名规范
- **包名**：小写，简短，有意义
- **接口名**：动词 + er 形式
- **结构体**：大驼峰命名
- **方法名**：大驼峰命名
- **变量名**：小驼峰命名

### 测试指南

#### 单元测试
```bash
# 运行所有测试
make test

# 运行特定包测试
go test ./internal/biz/auth/...

# 生成测试覆盖率报告
make test-coverage
```

#### 集成测试
```bash
# 启动测试环境
make test-env

# 运行集成测试
make integration-test

# 清理测试环境
make clean-test-env
```

#### API测试
```bash
# 使用Postman集合
newman run docs/postman/one-auth.postman_collection.json

# 或使用httpie
http POST localhost:8080/login username=admin password=admin123
```

### 部署指南

#### Docker部署
```bash
# 构建镜像
make docker-build

# 运行容器
docker run -p 8080:8080 one-auth:latest
```

#### Kubernetes部署
```bash
# 应用配置
kubectl apply -f deployments/k8s/

# 检查部署状态
kubectl get pods -l app=one-auth
```

#### 生产环境配置
```yaml
# 生产环境配置示例
server:
  host: 0.0.0.0
  port: 8080
  
database:
  host: mysql.prod.local
  port: 3306
  username: one_auth
  password: ${DB_PASSWORD}
  database: one_auth_prod
  
redis:
  host: redis.prod.local
  port: 6379
  password: ${REDIS_PASSWORD}
  
jwt:
  secret: ${JWT_SECRET}
  expire: 24h
  
logging:
  level: info
  format: json
```

## 监控运维

### 健康检查
```bash
# 服务健康检查
curl http://localhost:8080/healthz

# 数据库连接检查
curl http://localhost:8080/healthz/db

# Redis连接检查  
curl http://localhost:8080/healthz/redis
```

### 日志管理
- **结构化日志**：JSON格式，便于分析
- **日志级别**：DEBUG、INFO、WARN、ERROR
- **日志轮转**：按大小和时间轮转
- **集中收集**：支持ELK、Fluentd等

### 指标监控
```
# Prometheus指标
- http_requests_total
- http_request_duration_seconds
- auth_attempts_total
- permission_check_duration_seconds
- active_sessions_total
```

### 告警配置
- **服务可用性**：服务下线告警
- **响应时间**：接口响应超时告警
- **错误率**：错误率超阈值告警
- **资源使用**：CPU、内存使用率告警

## 许可证

本项目采用 MIT License 开源许可证。
