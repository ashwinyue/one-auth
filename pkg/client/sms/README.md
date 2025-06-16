# SMS Client Package

## 概述

`pkg/client/sms` 包提供了统一的短信客户端接口，支持多种短信服务提供商。

## 设计理念

- **客户端模式**：采用客户端设计模式，而非服务层设计
- **可插拔**：支持多种短信提供商，便于切换和扩展
- **配置化**：通过配置文件灵活配置提供商和模板
- **类型安全**：使用强类型定义提供商和验证码类型

## 包结构

```
pkg/client/sms/
├── client.go      # 主要的客户端接口和实现
├── providers.go   # 提供商类型定义和配置验证
└── README.md      # 包说明文档
```

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "github.com/ashwinyue/one-auth/pkg/client/sms"
)

func main() {
    // 使用默认配置（模拟提供商）
    client := sms.NewClient(nil)
    
    // 生成验证码
    code := client.GenerateCode()
    
    // 发送验证码
    err := client.SendVerifyCode(context.Background(), "13800138000", code, "login")
    if err != nil {
        // 处理错误
    }
    
    // 验证手机号格式
    if client.IsValidPhone("13800138000") {
        // 手机号格式正确
    }
}
```

### 自定义配置

```go
config := &sms.Config{
    Provider:    "aliyun",
    AccessKeyID: "your_access_key",
    SecretKey:   "your_secret_key",
    Region:      "cn-hangzhou",
    SignName:    "Your-App",
    Templates: map[string]string{
        "login":    "SMS_LOGIN_CODE",
        "register": "SMS_REGISTER_CODE",
    },
}

client := sms.NewClient(config)
```

## 支持的提供商

- **mock**: 模拟提供商（开发测试用）
- **aliyun**: 阿里云短信服务（待实现）
- **tencent**: 腾讯云短信服务（待实现）

## 验证码类型

- `login`: 登录验证码
- `register`: 注册验证码
- `reset_password`: 重置密码验证码
- `bind_phone`: 绑定手机号验证码

## 接口定义

```go
type Client interface {
    // SendVerifyCode 发送验证码
    SendVerifyCode(ctx context.Context, phone, code, template string) error
    // GenerateCode 生成验证码
    GenerateCode() string
    // IsValidPhone 验证手机号格式
    IsValidPhone(phone string) bool
}
```

## 配置说明

```go
type Config struct {
    Provider    string            // 短信服务提供商：aliyun, tencent, mock
    AccessKeyID string            // 访问密钥ID
    SecretKey   string            // 密钥
    Region      string            // 区域
    SignName    string            // 签名名称
    Templates   map[string]string // 模板配置
}
```

## 扩展指南

要添加新的短信提供商：

1. 在 `providers.go` 中添加新的提供商常量
2. 在 `client.go` 的 `SendVerifyCode` 方法中添加新的 case
3. 实现对应的发送方法（如 `sendNewProviderSMS`）

## 迁移指南

从旧的 `internal/apiserver/service/sms` 迁移到新的 `pkg/client/sms`：

1. 导入路径更改：
   ```go
   // 旧的
   import "github.com/ashwinyue/one-auth/internal/apiserver/service/sms"
   
   // 新的
   import "github.com/ashwinyue/one-auth/pkg/client/sms"
   ```

2. 接口名称更改：
   ```go
   // 旧的
   var smsService sms.SMSService
   
   // 新的
   var smsClient sms.Client
   ```

3. 构造函数更改：
   ```go
   // 旧的
   smsService := sms.NewSMSService(config)
   
   // 新的
   smsClient := sms.NewClient(config)
   ``` 