// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	IP           string    `json:"ip"`
	AttemptCount int       `json:"attempt_count"`
	LastAttempt  time.Time `json:"last_attempt"`
	LockedUntil  time.Time `json:"locked_until,omitempty"`
	IsLocked     bool      `json:"is_locked"`
}

// VerifyCode 验证码信息
type VerifyCode struct {
	Code      string    `json:"code"`
	Type      string    `json:"type"`   // login, register, reset_password
	Target    string    `json:"target"` // phone or email
	CreatedAt time.Time `json:"created_at"`
	UsedAt    time.Time `json:"used_at,omitempty"`
	IsUsed    bool      `json:"is_used"`
}

// LoginSecurityManager 登录安全管理器
type LoginSecurityManager struct {
	cache ICache
}

// NewLoginSecurityManager 创建登录安全管理器
func NewLoginSecurityManager(cache ICache) *LoginSecurityManager {
	return &LoginSecurityManager{cache: cache}
}

// 常量定义
const (
	MaxLoginAttempts     = 5                // 最大登录尝试次数
	LoginLockDuration    = 30 * time.Minute // 锁定时长
	VerifyCodeExpiration = 10 * time.Minute // 验证码有效期
	VerifyCodeCooldown   = 1 * time.Minute  // 验证码发送冷却时间
)

// loginAttemptKey 生成登录尝试缓存key
func (lsm *LoginSecurityManager) loginAttemptKey(identifier string) string {
	return fmt.Sprintf("login_attempt:%s", identifier)
}

// ipAttemptKey 生成IP登录尝试缓存key
func (lsm *LoginSecurityManager) ipAttemptKey(ip string) string {
	return fmt.Sprintf("ip_attempt:%s", ip)
}

// verifyCodeKey 生成验证码缓存key
func (lsm *LoginSecurityManager) verifyCodeKey(target, codeType string) string {
	return fmt.Sprintf("verify_code:%s:%s", codeType, target)
}

// verifyCodeCooldownKey 生成验证码冷却缓存key
func (lsm *LoginSecurityManager) verifyCodeCooldownKey(target, codeType string) string {
	return fmt.Sprintf("verify_cooldown:%s:%s", codeType, target)
}

// RecordLoginAttempt 记录登录尝试
func (lsm *LoginSecurityManager) RecordLoginAttempt(ctx context.Context, identifier, ip string, success bool) error {
	if success {
		// 登录成功，清除尝试记录
		return lsm.ClearLoginAttempts(ctx, identifier, ip)
	}

	// 记录用户名/邮箱的登录尝试
	if err := lsm.recordAttempt(ctx, lsm.loginAttemptKey(identifier), identifier, ip); err != nil {
		return err
	}

	// 记录IP的登录尝试
	return lsm.recordAttempt(ctx, lsm.ipAttemptKey(ip), identifier, ip)
}

// recordAttempt 记录具体的登录尝试
func (lsm *LoginSecurityManager) recordAttempt(ctx context.Context, key, identifier, ip string) error {
	// 获取现有记录
	data, err := lsm.cache.Get(ctx, key)
	var attempt LoginAttempt

	if err != nil {
		// 首次尝试
		attempt = LoginAttempt{
			UserID:       identifier,
			Username:     identifier,
			IP:           ip,
			AttemptCount: 1,
			LastAttempt:  time.Now(),
			IsLocked:     false,
		}
	} else {
		// 解析现有记录
		if err := json.Unmarshal([]byte(data), &attempt); err != nil {
			return fmt.Errorf("failed to parse login attempt: %w", err)
		}

		// 增加尝试次数
		attempt.AttemptCount++
		attempt.LastAttempt = time.Now()

		// 检查是否需要锁定
		if attempt.AttemptCount >= MaxLoginAttempts {
			attempt.IsLocked = true
			attempt.LockedUntil = time.Now().Add(LoginLockDuration)
		}
	}

	// 存储记录
	expiration := LoginLockDuration
	if attempt.IsLocked {
		expiration = time.Until(attempt.LockedUntil)
	}

	return lsm.cache.Set(ctx, key, attempt, expiration)
}

// CheckLoginAttempts 检查登录尝试是否被锁定
func (lsm *LoginSecurityManager) CheckLoginAttempts(ctx context.Context, identifier, ip string) (bool, string, error) {
	// 检查用户名锁定
	if locked, reason, err := lsm.checkAttemptLock(ctx, lsm.loginAttemptKey(identifier)); err != nil {
		return false, "", err
	} else if locked {
		return true, fmt.Sprintf("账户被锁定: %s", reason), nil
	}

	// 检查IP锁定
	if locked, reason, err := lsm.checkAttemptLock(ctx, lsm.ipAttemptKey(ip)); err != nil {
		return false, "", err
	} else if locked {
		return true, fmt.Sprintf("IP被锁定: %s", reason), nil
	}

	return false, "", nil
}

// checkAttemptLock 检查具体的锁定状态
func (lsm *LoginSecurityManager) checkAttemptLock(ctx context.Context, key string) (bool, string, error) {
	data, err := lsm.cache.Get(ctx, key)
	if err != nil {
		return false, "", nil // 没有记录，未锁定
	}

	var attempt LoginAttempt
	if err := json.Unmarshal([]byte(data), &attempt); err != nil {
		return false, "", fmt.Errorf("failed to parse login attempt: %w", err)
	}

	if attempt.IsLocked {
		if time.Now().Before(attempt.LockedUntil) {
			remaining := time.Until(attempt.LockedUntil)
			return true, fmt.Sprintf("剩余锁定时间: %v", remaining.Round(time.Minute)), nil
		}
		// 锁定已过期，清除记录
		_ = lsm.cache.Del(ctx, key)
	}

	return false, "", nil
}

// ClearLoginAttempts 清除登录尝试记录
func (lsm *LoginSecurityManager) ClearLoginAttempts(ctx context.Context, identifier, ip string) error {
	// 清除用户记录
	_ = lsm.cache.Del(ctx, lsm.loginAttemptKey(identifier))
	// 清除IP记录
	_ = lsm.cache.Del(ctx, lsm.ipAttemptKey(ip))
	return nil
}

// StoreVerifyCode 存储验证码
func (lsm *LoginSecurityManager) StoreVerifyCode(ctx context.Context, target, codeType, code string) error {
	// 检查冷却时间
	cooldownKey := lsm.verifyCodeCooldownKey(target, codeType)
	if exists, _ := lsm.cache.Exists(ctx, cooldownKey); exists {
		return fmt.Errorf("验证码发送过于频繁，请稍后再试")
	}

	// 存储验证码
	verifyCode := VerifyCode{
		Code:      code,
		Type:      codeType,
		Target:    target,
		CreatedAt: time.Now(),
		IsUsed:    false,
	}

	codeKey := lsm.verifyCodeKey(target, codeType)
	if err := lsm.cache.Set(ctx, codeKey, verifyCode, VerifyCodeExpiration); err != nil {
		return fmt.Errorf("failed to store verify code: %w", err)
	}

	// 设置冷却时间
	return lsm.cache.Set(ctx, cooldownKey, "1", VerifyCodeCooldown)
}

// ValidateVerifyCode 验证验证码
func (lsm *LoginSecurityManager) ValidateVerifyCode(ctx context.Context, target, codeType, inputCode string) error {
	codeKey := lsm.verifyCodeKey(target, codeType)
	data, err := lsm.cache.Get(ctx, codeKey)
	if err != nil {
		return fmt.Errorf("验证码不存在或已过期")
	}

	var verifyCode VerifyCode
	if err := json.Unmarshal([]byte(data), &verifyCode); err != nil {
		return fmt.Errorf("验证码数据异常")
	}

	if verifyCode.IsUsed {
		return fmt.Errorf("验证码已使用")
	}

	if verifyCode.Code != inputCode {
		return fmt.Errorf("验证码错误")
	}

	// 标记为已使用
	verifyCode.IsUsed = true
	verifyCode.UsedAt = time.Now()

	// 更新状态（短时间保留，防止重复使用）
	_ = lsm.cache.Set(ctx, codeKey, verifyCode, 5*time.Minute)

	return nil
}

// GetLoginAttemptCount 获取登录尝试次数
func (lsm *LoginSecurityManager) GetLoginAttemptCount(ctx context.Context, identifier string) (int, error) {
	data, err := lsm.cache.Get(ctx, lsm.loginAttemptKey(identifier))
	if err != nil {
		return 0, nil // 没有记录
	}

	var attempt LoginAttempt
	if err := json.Unmarshal([]byte(data), &attempt); err != nil {
		return 0, err
	}

	return attempt.AttemptCount, nil
}

// IsAccountLocked 检查账户是否被锁定
func (lsm *LoginSecurityManager) IsAccountLocked(ctx context.Context, identifier string) (bool, time.Time, error) {
	data, err := lsm.cache.Get(ctx, lsm.loginAttemptKey(identifier))
	if err != nil {
		return false, time.Time{}, nil // 没有记录，未锁定
	}

	var attempt LoginAttempt
	if err := json.Unmarshal([]byte(data), &attempt); err != nil {
		return false, time.Time{}, err
	}

	if attempt.IsLocked && time.Now().Before(attempt.LockedUntil) {
		return true, attempt.LockedUntil, nil
	}

	return false, time.Time{}, nil
}

// UnlockAccount 手动解锁账户（管理员功能）
func (lsm *LoginSecurityManager) UnlockAccount(ctx context.Context, identifier string) error {
	return lsm.cache.Del(ctx, lsm.loginAttemptKey(identifier))
}

// GetLoginSecurityStats 获取登录安全统计信息
func (lsm *LoginSecurityManager) GetLoginSecurityStats(ctx context.Context, identifier string) (map[string]interface{}, error) {
	attemptCount, _ := lsm.GetLoginAttemptCount(ctx, identifier)
	isLocked, lockedUntil, _ := lsm.IsAccountLocked(ctx, identifier)

	stats := map[string]interface{}{
		"attempt_count":      attemptCount,
		"max_attempts":       MaxLoginAttempts,
		"is_locked":          isLocked,
		"lock_duration":      LoginLockDuration.String(),
		"remaining_attempts": MaxLoginAttempts - attemptCount,
	}

	if isLocked {
		stats["locked_until"] = lockedUntil
		stats["remaining_lock_time"] = time.Until(lockedUntil).String()
	}

	return stats, nil
}
