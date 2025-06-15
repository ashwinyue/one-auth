// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package gin

import (
	"context"
	"net/http"
	"strings"

	"github.com/ashwinyue/one-auth/internal/apiserver/cache"
	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	"github.com/ashwinyue/one-auth/pkg/core"
	"github.com/ashwinyue/one-auth/pkg/token"
	"github.com/gin-gonic/gin"
)

// EnhancedUserRetriever 增强的用户检索器，支持会话管理
type EnhancedUserRetriever interface {
	// GetUser 根据用户ID获取用户信息
	GetUser(ctx context.Context, userID string) (*model.UserM, error)
	// GetSessionManager 获取会话管理器
	GetSessionManager() *cache.SessionManager
	// GetLoginSecurityManager 获取登录安全管理器
	GetLoginSecurityManager() *cache.LoginSecurityManager
}

// AuthenticationConfig 认证配置
type AuthenticationConfig struct {
	// SkipPaths 跳过认证的路径
	SkipPaths []string
	// RequireSession 是否要求会话验证
	RequireSession bool
	// AllowedClientTypes 允许的客户端类型
	AllowedClientTypes []cache.ClientType
}

// EnhancedAuthnMiddleware 增强的认证中间件
func EnhancedAuthnMiddleware(retriever EnhancedUserRetriever, config *AuthenticationConfig) gin.HandlerFunc {
	if config == nil {
		config = &AuthenticationConfig{
			RequireSession: true,
			AllowedClientTypes: []cache.ClientType{
				cache.ClientTypeWeb,
				cache.ClientTypeH5,
				cache.ClientTypeAndroid,
				cache.ClientTypeIOS,
			},
		}
	}

	return func(c *gin.Context) {
		// 检查是否为跳过路径
		if shouldSkipAuth(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// 获取IP地址
		clientIP := getClientIP(c)

		// 1. 基础token验证
		userID, err := token.ParseRequest(c)
		if err != nil {
			log.Debugw("Token parsing failed", "error", err, "path", c.Request.URL.Path)
			core.WriteResponse(c, nil, errno.ErrTokenInvalid.WithMessage(err.Error()))
			c.Abort()
			return
		}

		log.Debugw("Token parsing successful", "userID", userID)

		// 2. 获取用户信息
		user, err := retriever.GetUser(c, userID)
		if err != nil {
			log.Debugw("User retrieval failed", "userID", userID, "error", err)
			core.WriteResponse(c, nil, errno.ErrUnauthenticated.WithMessage(err.Error()))
			c.Abort()
			return
		}

		// 3. 会话验证（如果启用）
		if config.RequireSession {
			sessionManager := retriever.GetSessionManager()
			if sessionManager != nil {
				if err := validateSession(c, sessionManager, userID, clientIP); err != nil {
					log.Debugw("Session validation failed", "userID", userID, "error", err)
					core.WriteResponse(c, nil, errno.ErrUnauthenticated.WithMessage(err.Error()))
					c.Abort()
					return
				}
			}
		}

		// 4. 客户端类型验证
		if len(config.AllowedClientTypes) > 0 {
			clientType := getClientType(c)
			if !isClientTypeAllowed(clientType, config.AllowedClientTypes) {
				log.Debugw("Client type not allowed", "clientType", clientType)
				core.WriteResponse(c, nil, errno.ErrPermissionDenied.WithMessage("Client type not allowed"))
				c.Abort()
				return
			}
		}

		// 5. 设置用户上下文
		ctx := contextx.WithUserID(c.Request.Context(), user.UserID)
		ctx = contextx.WithUsername(ctx, user.Username)
		c.Request = c.Request.WithContext(ctx)

		// 6. 记录访问日志
		log.Debugw("Authentication successful",
			"userID", user.UserID,
			"username", user.Username,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"clientIP", clientIP)

		c.Next()
	}
}

// LoginMiddleware 登录专用中间件，处理登录安全策略
func LoginMiddleware(retriever EnhancedUserRetriever) gin.HandlerFunc {
	return func(c *gin.Context) {
		securityManager := retriever.GetLoginSecurityManager()
		if securityManager == nil {
			c.Next()
			return
		}

		// 从请求中提取用户标识（用户名、邮箱或手机号）
		identifier := extractUserIdentifier(c)
		if identifier == "" {
			c.Next()
			return
		}

		clientIP := getClientIP(c)

		// 检查登录尝试是否被锁定
		locked, reason, err := securityManager.CheckLoginAttempts(c, identifier, clientIP)
		if err != nil {
			log.Errorw("Failed to check login attempts", "error", err)
			core.WriteResponse(c, nil, errno.ErrInternal)
			c.Abort()
			return
		}

		if locked {
			log.Infow("Login attempt blocked", "identifier", identifier, "ip", clientIP, "reason", reason)
			core.WriteResponse(c, nil, errno.ErrUnauthenticated.WithMessage(reason))
			c.Abort()
			return
		}

		// 在上下文中设置登录信息，供后续处理使用
		c.Set("login_identifier", identifier)
		c.Set("client_ip", clientIP)

		c.Next()
	}
}

// shouldSkipAuth 检查是否应该跳过认证
func shouldSkipAuth(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getClientIP 获取客户端IP地址
func getClientIP(c *gin.Context) string {
	// 尝试从各种Header中获取真实IP
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip = c.GetHeader("X-Forwarded")
	if ip != "" {
		return ip
	}

	return c.ClientIP()
}

// getClientType 从请求头获取客户端类型
func getClientType(c *gin.Context) cache.ClientType {
	userAgent := c.GetHeader("User-Agent")
	clientTypeHeader := c.GetHeader("X-Client-Type")

	// 优先使用明确指定的客户端类型
	switch clientTypeHeader {
	case "web":
		return cache.ClientTypeWeb
	case "h5":
		return cache.ClientTypeH5
	case "android":
		return cache.ClientTypeAndroid
	case "ios":
		return cache.ClientTypeIOS
	case "miniprogram":
		return cache.ClientTypeMiniProgram
	case "op":
		return cache.ClientTypeOp
	}

	// 根据User-Agent推断客户端类型
	if strings.Contains(userAgent, "Android") {
		return cache.ClientTypeAndroid
	}
	if strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") {
		return cache.ClientTypeIOS
	}
	if strings.Contains(userAgent, "MicroMessenger") {
		return cache.ClientTypeMiniProgram
	}
	if strings.Contains(userAgent, "Mobile") {
		return cache.ClientTypeH5
	}

	return cache.ClientTypeWeb // 默认为Web
}

// isClientTypeAllowed 检查客户端类型是否允许
func isClientTypeAllowed(clientType cache.ClientType, allowedTypes []cache.ClientType) bool {
	for _, allowedType := range allowedTypes {
		if clientType == allowedType {
			return true
		}
	}
	return false
}

// validateSession 验证会话
func validateSession(c *gin.Context, sessionManager *cache.SessionManager, userID, clientIP string) error {
	// 从Header中获取SessionID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		// 如果没有SessionID，可能是新的登录，允许通过
		return nil
	}

	// 验证会话
	_, err := sessionManager.ValidateSession(c, sessionID)
	if err != nil {
		return err
	}

	return nil
}

// extractUserIdentifier 从请求中提取用户标识
func extractUserIdentifier(c *gin.Context) string {
	// 尝试从JSON body中提取
	var loginReq struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&loginReq); err == nil {
		if loginReq.Username != "" {
			return loginReq.Username
		}
		if loginReq.Email != "" {
			return loginReq.Email
		}
		if loginReq.Phone != "" {
			return loginReq.Phone
		}
	}

	// 重置body，使后续处理可以正常读取
	c.Request.Body = http.NoBody

	return ""
}

// RecordLoginAttempt 记录登录尝试结果的辅助函数
func RecordLoginAttempt(c *gin.Context, retriever EnhancedUserRetriever, success bool) {
	securityManager := retriever.GetLoginSecurityManager()
	if securityManager == nil {
		return
	}

	identifier, exists := c.Get("login_identifier")
	if !exists {
		return
	}

	clientIP, exists := c.Get("client_ip")
	if !exists {
		return
	}

	if err := securityManager.RecordLoginAttempt(c, identifier.(string), clientIP.(string), success); err != nil {
		log.Errorw("Failed to record login attempt", "error", err)
	}
}
