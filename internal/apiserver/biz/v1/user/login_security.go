// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package user

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/ashwinyue/one-auth/internal/apiserver/cache"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
)

// LoginSecurityConfig 登录安全配置
type LoginSecurityConfig struct {
	// 是否启用IP白名单
	EnableIPWhitelist bool
	// IP白名单
	IPWhitelist []string
	// 是否启用地理位置检查
	EnableGeoCheck bool
	// 允许的国家/地区代码
	AllowedCountries []string
	// 是否启用设备指纹检查
	EnableDeviceFingerprint bool
	// 是否启用异常登录检测
	EnableAnomalyDetection bool
}

// LoginSecurityEnhancer 登录安全增强器
type LoginSecurityEnhancer struct {
	config         *LoginSecurityConfig
	loginSecurity  *cache.LoginSecurityManager
	sessionManager *cache.SessionManager
}

// NewLoginSecurityEnhancer 创建登录安全增强器
func NewLoginSecurityEnhancer(
	config *LoginSecurityConfig,
	loginSecurity *cache.LoginSecurityManager,
	sessionManager *cache.SessionManager,
) *LoginSecurityEnhancer {
	if config == nil {
		config = &LoginSecurityConfig{
			EnableIPWhitelist:       false,
			EnableGeoCheck:          false,
			EnableDeviceFingerprint: false,
			EnableAnomalyDetection:  true, // 默认启用异常检测
		}
	}

	return &LoginSecurityEnhancer{
		config:         config,
		loginSecurity:  loginSecurity,
		sessionManager: sessionManager,
	}
}

// ValidateLoginSecurity 验证登录安全性
func (lse *LoginSecurityEnhancer) ValidateLoginSecurity(ctx context.Context, userID string, rq *apiv1.LoginRequest) error {
	clientIP := getClientIP(ctx)

	// 1. IP白名单检查
	if lse.config.EnableIPWhitelist {
		if err := lse.checkIPWhitelist(clientIP); err != nil {
			log.W(ctx).Warnw("IP whitelist check failed", "user_id", userID, "ip", clientIP, "err", err)
			return err
		}
	}

	// 2. 地理位置检查
	if lse.config.EnableGeoCheck {
		if err := lse.checkGeoLocation(ctx, clientIP); err != nil {
			log.W(ctx).Warnw("Geo location check failed", "user_id", userID, "ip", clientIP, "err", err)
			return err
		}
	}

	// 3. 设备指纹检查
	if lse.config.EnableDeviceFingerprint {
		if err := lse.checkDeviceFingerprint(ctx, userID, rq); err != nil {
			log.W(ctx).Warnw("Device fingerprint check failed", "user_id", userID, "device_id", rq.GetDeviceId(), "err", err)
			return err
		}
	}

	// 4. 异常登录检测
	if lse.config.EnableAnomalyDetection {
		if err := lse.detectAnomalousLogin(ctx, userID, rq); err != nil {
			log.W(ctx).Warnw("Anomalous login detected", "user_id", userID, "ip", clientIP, "err", err)
			return err
		}
	}

	return nil
}

// checkIPWhitelist 检查IP白名单
func (lse *LoginSecurityEnhancer) checkIPWhitelist(clientIP string) error {
	if len(lse.config.IPWhitelist) == 0 {
		return nil
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return errno.ErrInvalidArgument.WithMessage("Invalid IP address")
	}

	for _, allowedIP := range lse.config.IPWhitelist {
		// 支持CIDR格式
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err != nil {
				continue
			}
			if ipNet.Contains(ip) {
				return nil
			}
		} else {
			// 精确匹配
			if clientIP == allowedIP {
				return nil
			}
		}
	}

	return errno.ErrPermissionDenied.WithMessage("IP address not in whitelist")
}

// checkGeoLocation 检查地理位置
func (lse *LoginSecurityEnhancer) checkGeoLocation(ctx context.Context, clientIP string) error {
	if len(lse.config.AllowedCountries) == 0 {
		return nil
	}

	// 这里应该调用地理位置服务API来获取IP的地理位置
	// 暂时返回nil，实际实现需要集成第三方地理位置服务
	log.Infow("Geo location check skipped (not implemented)", "ip", clientIP)
	return nil
}

// checkDeviceFingerprint 检查设备指纹
func (lse *LoginSecurityEnhancer) checkDeviceFingerprint(ctx context.Context, userID string, rq *apiv1.LoginRequest) error {
	if rq.GetDeviceId() == "" {
		// 如果没有设备ID，可以选择是否允许登录
		return nil
	}

	// 检查设备是否已注册
	if lse.sessionManager != nil {
		// 这里可以实现设备注册和验证逻辑
		log.Infow("Device fingerprint check", "user_id", userID, "device_id", rq.GetDeviceId())
	}

	return nil
}

// detectAnomalousLogin 检测异常登录
func (lse *LoginSecurityEnhancer) detectAnomalousLogin(ctx context.Context, userID string, rq *apiv1.LoginRequest) error {
	if lse.sessionManager == nil {
		return nil
	}

	// 获取用户的历史会话
	sessions, err := lse.sessionManager.ListUserSessions(ctx, userID)
	if err != nil {
		// 如果获取失败，不阻止登录，只记录日志
		log.W(ctx).Warnw("Failed to get user sessions for anomaly detection", "user_id", userID, "err", err)
		return nil
	}

	currentIP := getClientIP(ctx)
	currentTime := time.Now()

	// 检查是否有异常模式
	for _, session := range sessions {
		// 1. 检查短时间内多地登录
		if session.LoginIP != currentIP {
			timeDiff := currentTime.Sub(time.Unix(session.LoginTime, 0))
			if timeDiff < 30*time.Minute { // 30分钟内从不同IP登录
				return errno.ErrUnauthenticated.WithMessage("Suspicious login detected: multiple locations in short time")
			}
		}

		// 2. 检查异常时间登录（可以根据用户历史行为模式判断）
		hour := currentTime.Hour()
		if hour < 6 || hour > 23 { // 凌晨登录可能异常
			log.W(ctx).Warnw("Unusual login time detected", "user_id", userID, "hour", hour)
			// 这里可以选择是否阻止登录或要求额外验证
		}
	}

	return nil
}

// GenerateSecurityReport 生成安全报告
func (lse *LoginSecurityEnhancer) GenerateSecurityReport(ctx context.Context, userID string) (map[string]interface{}, error) {
	report := make(map[string]interface{})

	// 获取登录统计
	if lse.loginSecurity != nil {
		stats, err := lse.loginSecurity.GetLoginSecurityStats(ctx, userID)
		if err == nil {
			report["login_security"] = stats
		}
	}

	// 获取会话信息
	if lse.sessionManager != nil {
		sessions, err := lse.sessionManager.ListUserSessions(ctx, userID)
		if err == nil {
			sessionInfo := make([]map[string]interface{}, 0, len(sessions))
			for _, session := range sessions {
				sessionInfo = append(sessionInfo, map[string]interface{}{
					"session_id":  session.SessionID,
					"client_type": session.ClientType,
					"login_ip":    session.LoginIP,
					"login_time":  time.Unix(session.LoginTime, 0),
					"last_active": time.Unix(session.LastActive, 0),
					"device_id":   session.DeviceID,
				})
			}
			report["active_sessions"] = sessionInfo
		}
	}

	// 安全配置信息
	report["security_config"] = map[string]interface{}{
		"ip_whitelist_enabled":       lse.config.EnableIPWhitelist,
		"geo_check_enabled":          lse.config.EnableGeoCheck,
		"device_fingerprint_enabled": lse.config.EnableDeviceFingerprint,
		"anomaly_detection_enabled":  lse.config.EnableAnomalyDetection,
	}

	return report, nil
}

// NotifySecurityEvent 通知安全事件
func (lse *LoginSecurityEnhancer) NotifySecurityEvent(ctx context.Context, userID, eventType, description string) {
	// 这里可以实现安全事件通知逻辑
	// 例如：发送邮件、短信、推送通知等
	log.Infow("Security event",
		"user_id", userID,
		"event_type", eventType,
		"description", description,
		"ip", getClientIP(ctx),
		"timestamp", time.Now(),
	)

	// 可以将安全事件存储到数据库或发送到安全监控系统
}
