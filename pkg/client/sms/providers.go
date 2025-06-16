// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package sms

// Provider 定义短信提供商类型
type Provider string

const (
	// ProviderMock 模拟提供商（开发测试用）
	ProviderMock Provider = "mock"
	// ProviderAliyun 阿里云短信
	ProviderAliyun Provider = "aliyun"
	// ProviderTencent 腾讯云短信
	ProviderTencent Provider = "tencent"
)

// CodeType 定义验证码类型
type CodeType string

const (
	// CodeTypeLogin 登录验证码
	CodeTypeLogin CodeType = "login"
	// CodeTypeRegister 注册验证码
	CodeTypeRegister CodeType = "register"
	// CodeTypeResetPassword 重置密码验证码
	CodeTypeResetPassword CodeType = "reset_password"
	// CodeTypeBindPhone 绑定手机号验证码
	CodeTypeBindPhone CodeType = "bind_phone"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Provider: string(ProviderMock),
		SignName: "One-Auth",
		Templates: map[string]string{
			string(CodeTypeLogin):         "LOGIN_VERIFY_CODE",
			string(CodeTypeRegister):      "REGISTER_VERIFY_CODE",
			string(CodeTypeResetPassword): "RESET_PASSWORD_VERIFY_CODE",
			string(CodeTypeBindPhone):     "BIND_PHONE_VERIFY_CODE",
		},
	}
}

// ValidateConfig 验证配置
func ValidateConfig(config *Config) error {
	if config == nil {
		return nil // 使用默认配置
	}

	// 验证提供商
	provider := Provider(config.Provider)
	switch provider {
	case ProviderMock, ProviderAliyun, ProviderTencent:
		// 合法的提供商
	default:
		config.Provider = string(ProviderMock)
	}

	// 验证签名名称
	if config.SignName == "" {
		config.SignName = "One-Auth"
	}

	// 确保模板映射存在
	if config.Templates == nil {
		config.Templates = make(map[string]string)
	}

	return nil
}
