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
	"strconv"
	"time"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
)

// ClientType 客户端类型
type ClientType int32

const (
	ClientTypeWeb         ClientType = 1 // PC端
	ClientTypeH5          ClientType = 2 // H5端
	ClientTypeAndroid     ClientType = 3 // Android端
	ClientTypeIOS         ClientType = 4 // iOS端
	ClientTypeMiniProgram ClientType = 5 // 小程序
	ClientTypeOp          ClientType = 6 // 运营端
)

// SessionValidDuration 不同客户端的会话有效期
var SessionValidDuration = map[ClientType]time.Duration{
	ClientTypeWeb:         7 * 24 * time.Hour,       // PC端 7天
	ClientTypeH5:          3 * 30 * 24 * time.Hour,  // H5端 3个月
	ClientTypeAndroid:     12 * 30 * 24 * time.Hour, // Android 12个月
	ClientTypeIOS:         12 * 30 * 24 * time.Hour, // iOS 12个月
	ClientTypeMiniProgram: 30 * 24 * time.Hour,      // 小程序 30天
	ClientTypeOp:          12 * time.Hour,           // 运营端 12小时
}

// UserSession 用户会话信息
type UserSession struct {
	UserID     string     `json:"user_id"`
	Username   string     `json:"username"`
	TenantID   string     `json:"tenant_id,omitempty"`
	OrgID      string     `json:"org_id,omitempty"`
	SessionID  string     `json:"session_id"`
	ClientType ClientType `json:"client_type"`
	DeviceID   string     `json:"device_id,omitempty"`
	LoginIP    string     `json:"login_ip,omitempty"`
	LoginTime  int64      `json:"login_time"`
	ExpiredAt  int64      `json:"expired_at"`
	LastActive int64      `json:"last_active"`
}

// SessionManager 会话管理器
type SessionManager struct {
	cache ICache
}

// NewSessionManager 创建会话管理器
func NewSessionManager(cache ICache) *SessionManager {
	return &SessionManager{cache: cache}
}

// sessionKey 生成会话缓存key
func (sm *SessionManager) sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// userSessionKey 生成用户会话索引key
func (sm *SessionManager) userSessionKey(userID string, clientType ClientType) string {
	return fmt.Sprintf("user_session:%s:%d", userID, clientType)
}

// deviceSessionKey 生成设备会话索引key
func (sm *SessionManager) deviceSessionKey(deviceID string) string {
	return fmt.Sprintf("device_session:%s", deviceID)
}

// CreateSession 创建用户会话
func (sm *SessionManager) CreateSession(ctx context.Context, session *UserSession) error {
	// 设置会话过期时间
	duration := SessionValidDuration[session.ClientType]
	if duration == 0 {
		duration = 12 * time.Hour // 默认12小时
	}

	session.ExpiredAt = time.Now().Add(duration).Unix()
	session.LastActive = time.Now().Unix()

	// 存储会话信息
	sessionKey := sm.sessionKey(session.SessionID)
	if err := sm.cache.Set(ctx, sessionKey, session, duration); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// 建立用户到会话的索引
	userSessionKey := sm.userSessionKey(session.UserID, session.ClientType)
	if err := sm.cache.Set(ctx, userSessionKey, session.SessionID, duration); err != nil {
		return fmt.Errorf("failed to store user session index: %w", err)
	}

	// 如果有设备ID，建立设备到会话的索引
	if session.DeviceID != "" {
		deviceSessionKey := sm.deviceSessionKey(session.DeviceID)
		if err := sm.cache.Set(ctx, deviceSessionKey, session.SessionID, duration); err != nil {
			return fmt.Errorf("failed to store device session index: %w", err)
		}
	}

	return nil
}

// GetSession 获取会话信息
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*UserSession, error) {
	sessionKey := sm.sessionKey(sessionID)
	data, err := sm.cache.Get(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var session UserSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	// 检查会话是否过期
	if session.ExpiredAt < time.Now().Unix() {
		// 清理过期会话
		_ = sm.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// RefreshSession 刷新会话活跃时间
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActive = time.Now().Unix()

	// 更新会话信息
	sessionKey := sm.sessionKey(sessionID)
	duration := time.Unix(session.ExpiredAt, 0).Sub(time.Now())
	return sm.cache.Set(ctx, sessionKey, session, duration)
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	// 先获取会话信息，用于清理索引
	session, err := sm.GetSession(ctx, sessionID)

	// 删除会话主体（即使获取会话信息失败也要尝试删除）
	sessionKey := sm.sessionKey(sessionID)
	if delErr := sm.cache.Del(ctx, sessionKey); delErr != nil {
		return fmt.Errorf("failed to delete session: %w", delErr)
	}

	// 如果成功获取到会话信息，清理相关索引
	if err == nil && session != nil {
		// 清理用户会话索引
		userSessionKey := sm.userSessionKey(session.UserID, session.ClientType)
		_ = sm.cache.Del(ctx, userSessionKey)

		// 清理设备会话索引
		if session.DeviceID != "" {
			deviceSessionKey := sm.deviceSessionKey(session.DeviceID)
			_ = sm.cache.Del(ctx, deviceSessionKey)
		}
	}

	return nil
}

// GetUserSession 获取用户在指定客户端的会话
func (sm *SessionManager) GetUserSession(ctx context.Context, userID string, clientType ClientType) (*UserSession, error) {
	userSessionKey := sm.userSessionKey(userID, clientType)
	sessionIDData, err := sm.cache.Get(ctx, userSessionKey)
	if err != nil {
		return nil, fmt.Errorf("user session not found: %w", err)
	}

	return sm.GetSession(ctx, sessionIDData)
}

// KickUserSession 踢出用户在指定客户端的会话
func (sm *SessionManager) KickUserSession(ctx context.Context, userID string, clientType ClientType) error {
	session, err := sm.GetUserSession(ctx, userID, clientType)
	if err != nil {
		return err // 会话不存在，认为已经被踢出
	}

	return sm.DeleteSession(ctx, session.SessionID)
}

// ListUserSessions 列出用户的所有活跃会话
func (sm *SessionManager) ListUserSessions(ctx context.Context, userID string) ([]*UserSession, error) {
	var sessions []*UserSession

	for clientType := range SessionValidDuration {
		session, err := sm.GetUserSession(ctx, userID, clientType)
		if err == nil {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// ValidateSession 验证会话并返回用户信息
func (sm *SessionManager) ValidateSession(ctx context.Context, sessionID string) (*model.UserM, error) {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 刷新会话活跃时间
	_ = sm.RefreshSession(ctx, sessionID)

	// 将字符串用户ID转换为数字
	userID, err := strconv.ParseInt(session.UserID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in session: %w", err)
	}

	// 构建用户信息
	user := &model.UserM{
		ID:       userID,
		Username: session.Username,
		// 其他字段可以根据需要从数据库获取
	}

	return user, nil
}

// CleanExpiredSessions 清理过期会话 (可以通过定时任务调用)
func (sm *SessionManager) CleanExpiredSessions(ctx context.Context) error {
	// 注意：这是一个简化的实现
	// 在生产环境中，建议使用Redis的SCAN命令来批量处理
	// 或者依赖Redis的TTL自动过期机制

	// 由于Redis会自动清理过期的key，这里主要是清理可能残留的索引
	// 实际生产中可以通过以下方式优化：
	// 1. 使用Redis的SCAN命令扫描session:*模式的key
	// 2. 检查每个session是否过期
	// 3. 清理过期session及其相关索引

	// 目前依赖Redis的TTL机制自动清理
	return nil
}
