// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package sms provides SMS client implementations for different providers.
package sms

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// Client 短信客户端接口
type Client interface {
	// SendVerifyCode 发送验证码
	SendVerifyCode(ctx context.Context, phone, code, template string) error
	// GenerateCode 生成验证码
	GenerateCode() string
	// IsValidPhone 验证手机号格式
	IsValidPhone(phone string) bool
}

// Config 短信客户端配置
type Config struct {
	Provider    string            `yaml:"provider"`      // 短信服务提供商：aliyun, tencent, mock
	AccessKeyID string            `yaml:"access_key_id"` // 访问密钥ID
	SecretKey   string            `yaml:"secret_key"`    // 密钥
	Region      string            `yaml:"region"`        // 区域
	SignName    string            `yaml:"sign_name"`     // 签名名称
	Templates   map[string]string `yaml:"templates"`     // 模板配置
}

// client 短信客户端实现
type client struct {
	config *Config
}

// NewClient 创建短信客户端实例
func NewClient(config *Config) Client {
	if config == nil {
		config = &Config{
			Provider: "mock",
			SignName: "One-Auth",
			Templates: map[string]string{
				"login":          "LOGIN_VERIFY_CODE",
				"register":       "REGISTER_VERIFY_CODE",
				"reset_password": "RESET_PASSWORD_VERIFY_CODE",
				"bind_phone":     "BIND_PHONE_VERIFY_CODE",
			},
		}
	}
	return &client{config: config}
}

// SendVerifyCode 发送验证码
func (c *client) SendVerifyCode(ctx context.Context, phone, code, template string) error {
	// 根据配置的提供商发送短信
	switch c.config.Provider {
	case "aliyun":
		return c.sendAliyunSMS(ctx, phone, code, template)
	case "tencent":
		return c.sendTencentSMS(ctx, phone, code, template)
	default:
		// 模拟发送短信（开发环境）
		return c.mockSendSMS(ctx, phone, code, template)
	}
}

// GenerateCode 生成6位数字验证码
func (c *client) GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}

// IsValidPhone 验证中国手机号格式
func (c *client) IsValidPhone(phone string) bool {
	if len(phone) != 11 {
		return false
	}

	// 检查是否以1开头
	if phone[0] != '1' {
		return false
	}

	// 检查第二位是否为3-9
	secondDigit, err := strconv.Atoi(string(phone[1]))
	if err != nil || secondDigit < 3 || secondDigit > 9 {
		return false
	}

	// 检查剩余位数是否都是数字
	for i := 2; i < 11; i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return false
		}
	}

	return true
}

// mockSendSMS 模拟发送短信（用于开发环境）
func (c *client) mockSendSMS(ctx context.Context, phone, code, template string) error {
	templateName := c.config.Templates[template]
	if templateName == "" {
		templateName = "VERIFY_CODE"
	}

	log.Infow("模拟发送短信验证码",
		"phone", phone,
		"code", code,
		"template", templateName,
		"sign_name", c.config.SignName,
		"message", fmt.Sprintf("【%s】您的验证码是：%s，5分钟内有效，请勿泄露。", c.config.SignName, code),
	)

	// 模拟网络延迟
	time.Sleep(100 * time.Millisecond)

	return nil
}

// sendAliyunSMS 阿里云短信发送（待实现）
func (c *client) sendAliyunSMS(ctx context.Context, phone, code, template string) error {
	// TODO: 实现阿里云短信发送
	log.Warnw("阿里云短信发送功能待实现", "phone", phone, "template", template)
	return c.mockSendSMS(ctx, phone, code, template)
}

// sendTencentSMS 腾讯云短信发送（待实现）
func (c *client) sendTencentSMS(ctx context.Context, phone, code, template string) error {
	// TODO: 实现腾讯云短信发送
	log.Warnw("腾讯云短信发送功能待实现", "phone", phone, "template", template)
	return c.mockSendSMS(ctx, phone, code, template)
}
