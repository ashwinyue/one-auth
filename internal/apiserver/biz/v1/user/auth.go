// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package user

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/ashwinyue/one-auth/internal/apiserver/cache"
	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authn"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"github.com/ashwinyue/one-auth/pkg/token"
)

// Login 实现 UserBiz 接口中的 Login 方法.
func (b *userBiz) Login(ctx context.Context, rq *apiv1.LoginRequest) (*apiv1.LoginResponse, error) {
	// 检查登录安全限制
	if b.loginSecurity != nil {
		clientIP := getClientIP(ctx)
		locked, reason, err := b.loginSecurity.CheckLoginAttempts(ctx, rq.GetIdentifier(), clientIP)
		if err != nil {
			log.W(ctx).Errorw("Failed to check login attempts", "err", err)
		} else if locked {
			return nil, errno.ErrUserLocked.WithMessage(reason)
		}
	}

	// 根据标识符类型查找用户
	var userM *model.UserM
	var userStatus *model.UserStatusM
	var err error

	// 通过标识符查找用户
	userM, userStatus, err = b.findUserByIdentifier(ctx, rq.GetIdentifier(), rq.GetLoginType())
	if err != nil {
		// 记录登录失败（用户不存在）
		b.recordLoginAttempt(ctx, rq.GetIdentifier(), false)
		return nil, err
	}

	// 检查用户状态
	if !userStatus.CanLogin() {
		// 记录登录失败（用户状态异常）
		b.recordLoginAttempt(ctx, strconv.FormatInt(userM.ID, 10), false)
		if userStatus.IsLocked() {
			return nil, errno.ErrUserLocked.WithMessage("User account is locked")
		}
		return nil, errno.ErrUserInactive.WithMessage("User account is inactive")
	}

	// 验证登录凭证
	if err := b.validateLoginCredentials(ctx, userM, rq); err != nil {
		// 记录登录失败
		b.recordLoginAttempt(ctx, strconv.FormatInt(userM.ID, 10), false)
		return nil, err
	}

	// 登录成功，记录成功尝试
	b.recordLoginAttempt(ctx, strconv.FormatInt(userM.ID, 10), true)

	// 更新用户状态
	if err := b.updateLoginSuccess(ctx, strconv.FormatInt(userM.ID, 10), rq); err != nil {
		log.W(ctx).Errorw("Failed to update login success info", "user_id", userM.ID, "err", err)
	}

	// 生成令牌
	tokenStr, expireAt, err := token.Sign(strconv.FormatInt(userM.ID, 10))
	if err != nil {
		log.W(ctx).Errorw("Failed to sign token", "err", err)
		return nil, errno.ErrSignToken
	}

	// 生成刷新令牌（使用更长的过期时间）
	refreshToken, _, err := token.SignWithExpiration(strconv.FormatInt(userM.ID, 10), 7*24*time.Hour) // 7天
	if err != nil {
		log.W(ctx).Errorw("Failed to sign refresh token", "err", err)
		return nil, errno.ErrSignToken
	}

	// 创建会话
	sessionID, err := b.createUserSession(ctx, userM, rq)
	if err != nil {
		log.W(ctx).Errorw("Failed to create user session", "user_id", userM.ID, "err", err)
		// 会话创建失败不影响登录，继续返回token
	}

	// 构建用户信息
	userInfo := &apiv1.UserInfo{
		UserId:   strconv.FormatInt(userM.ID, 10),
		Username: userM.Username,
		Nickname: userM.Nickname,
		Email:    userM.Email,
		Phone:    userM.Phone,
		Status:   userStatus.GetStatusString(),
	}
	if userStatus.LastLoginTime != nil {
		userInfo.LastLoginTime = timestamppb.New(*userStatus.LastLoginTime)
	}

	return &apiv1.LoginResponse{
		Token:        tokenStr,
		ExpireAt:     timestamppb.New(expireAt),
		RefreshToken: refreshToken,
		UserInfo:     userInfo,
		SessionId:    sessionID,
	}, nil
}

// RefreshToken 用于刷新用户的身份验证令牌.
func (b *userBiz) RefreshToken(ctx context.Context, rq *apiv1.RefreshTokenRequest) (*apiv1.RefreshTokenResponse, error) {
	tokenStr, expireAt, err := token.Sign(strconv.FormatInt(contextx.UserID(ctx), 10))
	if err != nil {
		log.W(ctx).Errorw("Failed to sign token", "err", err)
		return nil, errno.ErrSignToken
	}

	return &apiv1.RefreshTokenResponse{Token: tokenStr, ExpireAt: timestamppb.New(expireAt)}, nil
}

// ChangePassword 实现 UserBiz 接口中的 ChangePassword 方法.
func (b *userBiz) ChangePassword(ctx context.Context, rq *apiv1.ChangePasswordRequest) (*apiv1.ChangePasswordResponse, error) {
	userM, err := b.store.User().Get(ctx, where.T(ctx))
	if err != nil {
		return nil, err
	}

	if err := authn.Compare(userM.Password, rq.GetOldPassword()); err != nil {
		log.W(ctx).Errorw("Failed to compare password", "err", err)
		return nil, errno.ErrPasswordInvalid
	}

	userM.Password, _ = authn.Encrypt(rq.GetNewPassword())
	if err := b.store.User().Update(ctx, userM); err != nil {
		return nil, err
	}

	return &apiv1.ChangePasswordResponse{}, nil
}

// SendVerifyCode 发送验证码
func (b *userBiz) SendVerifyCode(ctx context.Context, rq *apiv1.SendVerifyCodeRequest) (*apiv1.SendVerifyCodeResponse, error) {
	// 验证目标类型和格式
	if rq.GetTargetType() == "phone" {
		if !b.smsClient.IsValidPhone(rq.GetTarget()) {
			return nil, errno.ErrInvalidArgument.WithMessage("Invalid phone number format")
		}
	}

	// 生成验证码
	code := b.smsClient.GenerateCode()

	// 存储验证码到缓存
	if b.loginSecurity != nil {
		if err := b.loginSecurity.StoreVerifyCode(ctx, rq.GetTarget(), rq.GetCodeType(), code); err != nil {
			log.W(ctx).Errorw("Failed to store verify code", "target", rq.GetTarget(), "err", err)
			return nil, errno.ErrOperationFailed.WithMessage(err.Error())
		}
	}

	// 发送验证码
	var err error
	switch rq.GetTargetType() {
	case "phone":
		err = b.smsClient.SendVerifyCode(ctx, rq.GetTarget(), code, rq.GetCodeType())
	case "email":
		// TODO: 实现邮件发送服务
		log.Infow("邮件验证码发送功能待实现", "email", rq.GetTarget(), "code", code)
	default:
		return nil, errno.ErrInvalidArgument.WithMessage("Unsupported target type")
	}

	if err != nil {
		log.W(ctx).Errorw("Failed to send verify code", "target", rq.GetTarget(), "type", rq.GetTargetType(), "err", err)
		return nil, errno.ErrOperationFailed.WithMessage("Failed to send verify code")
	}

	log.Infow("验证码发送成功",
		"target", rq.GetTarget(),
		"code_type", rq.GetCodeType(),
		"target_type", rq.GetTargetType(),
	)

	return &apiv1.SendVerifyCodeResponse{
		Success:         true,
		Message:         "验证码发送成功",
		CooldownSeconds: 60,
	}, nil
}

// Logout 用户登出
func (b *userBiz) Logout(ctx context.Context, rq *apiv1.LogoutRequest) (*apiv1.LogoutResponse, error) {
	userID := contextx.UserID(ctx)

	// 如果指定了session_id，则只登出指定会话
	if rq.GetSessionId() != "" {
		if err := b.logoutSession(ctx, userID, rq.GetSessionId()); err != nil {
			log.W(ctx).Errorw("Failed to logout session", "session_id", rq.GetSessionId(), "err", err)
			return nil, errno.ErrOperationFailed.WithMessage("Failed to logout session")
		}
		return &apiv1.LogoutResponse{
			Success: true,
			Message: "Session logged out successfully",
		}, nil
	}

	// 如果指定logout_all，则登出所有设备
	if rq.GetLogoutAll() {
		if err := b.logoutAllSessions(ctx, userID); err != nil {
			log.W(ctx).Errorw("Failed to logout all sessions", "user_id", userID, "err", err)
			return nil, errno.ErrOperationFailed.WithMessage("Failed to logout all sessions")
		}
		return &apiv1.LogoutResponse{
			Success: true,
			Message: "All sessions logged out successfully",
		}, nil
	}

	// 默认登出当前会话
	sessionID := getSessionIDFromContext(ctx)
	if sessionID != "" {
		if err := b.logoutSession(ctx, userID, sessionID); err != nil {
			log.W(ctx).Errorw("Failed to logout current session", "session_id", sessionID, "err", err)
		}
	}

	return &apiv1.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// logoutSession 登出指定会话
func (b *userBiz) logoutSession(ctx context.Context, userID int64, sessionID string) error {
	if b.sessionManager == nil {
		return nil
	}

	return b.sessionManager.DeleteSession(ctx, sessionID)
}

// logoutAllSessions 登出用户的所有会话
func (b *userBiz) logoutAllSessions(ctx context.Context, userID int64) error {
	if b.sessionManager == nil {
		return nil
	}

	userIDStr := strconv.FormatInt(userID, 10)

	// 遍历所有客户端类型，登出对应的会话
	for clientType := range cache.SessionValidDuration {
		if err := b.sessionManager.KickUserSession(ctx, userIDStr, clientType); err != nil {
			log.W(ctx).Errorw("Failed to kick user session",
				"user_id", userIDStr,
				"client_type", clientType,
				"err", err)
		}
	}

	return nil
}

// getSessionIDFromContext 从上下文中获取会话ID
func getSessionIDFromContext(ctx context.Context) string {
	// 尝试从 Gin 上下文中获取会话ID
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if sessionID := ginCtx.GetHeader("X-Session-ID"); sessionID != "" {
			return sessionID
		}
		if sessionID := ginCtx.GetHeader("Session-ID"); sessionID != "" {
			return sessionID
		}
	}

	// 从 Context Value 中获取
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}

	return ""
}

// 以下是辅助方法

// findUserByIdentifier 根据认证标识符查找用户（全局查找，自动确定租户）
func (b *userBiz) findUserByIdentifier(ctx context.Context, authID, authType string) (*model.UserM, *model.UserStatusM, error) {
	// 根据认证标识符查找用户状态
	authTypeEnum := model.StringToAuthType(authType)
	userStatus, err := b.store.UserStatus().Get(ctx, where.F("auth_id", authID, "auth_type", int32(authTypeEnum)))
	if err != nil {
		return nil, nil, errno.ErrUserNotFound.WithMessage("Invalid login credentials")
	}

	// 查找用户基本信息
	userM, err := b.store.User().Get(ctx, where.F("id", userStatus.UserID))
	if err != nil {
		return nil, nil, errno.ErrUserNotFound
	}

	return userM, userStatus, nil
}

// validateLoginCredentials 验证登录凭证
func (b *userBiz) validateLoginCredentials(ctx context.Context, userM *model.UserM, rq *apiv1.LoginRequest) error {
	// 密码登录
	if rq.GetPassword() != "" {
		if err := authn.Compare(userM.Password, rq.GetPassword()); err != nil {
			log.W(ctx).Errorw("Failed to compare password", "err", err)
			return errno.ErrPasswordInvalid
		}
		return nil
	}

	// 验证码登录
	if rq.GetVerifyCode() != "" {
		if b.loginSecurity == nil {
			return errno.ErrInternal.WithMessage("Login security manager not available")
		}

		// 验证验证码
		if err := b.loginSecurity.ValidateVerifyCode(ctx, rq.GetIdentifier(), "login", rq.GetVerifyCode()); err != nil {
			log.W(ctx).Errorw("Failed to validate verify code", "err", err)
			return errno.ErrPasswordInvalid.WithMessage(err.Error())
		}
		return nil
	}

	return errno.ErrPasswordInvalid.WithMessage("No valid login credentials provided")
}

// recordLoginAttempt 记录登录尝试
func (b *userBiz) recordLoginAttempt(ctx context.Context, identifier string, success bool) {
	if b.loginSecurity == nil {
		return
	}

	clientIP := getClientIP(ctx)
	if err := b.loginSecurity.RecordLoginAttempt(ctx, identifier, clientIP, success); err != nil {
		log.W(ctx).Errorw("Failed to record login attempt", "identifier", identifier, "success", success, "err", err)
	}
}

// updateLoginSuccess 更新登录成功信息
func (b *userBiz) updateLoginSuccess(ctx context.Context, userID string, rq *apiv1.LoginRequest) error {
	now := time.Now()
	clientIP := getClientIP(ctx)

	// 更新用户状态表
	updates := map[string]interface{}{
		"last_login_time":       now,
		"last_login_ip":         clientIP,
		"login_count":           gorm.Expr("login_count + 1"),
		"failed_login_attempts": 0,
		"last_failed_login":     nil,
	}

	if rq.GetDeviceId() != "" {
		updates["last_login_device"] = rq.GetDeviceId()
	}

	return b.store.DB(ctx).Model(&model.UserStatusM{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// createUserSession 创建用户会话
func (b *userBiz) createUserSession(ctx context.Context, userM *model.UserM, rq *apiv1.LoginRequest) (string, error) {
	if b.sessionManager == nil {
		return "", nil // 会话管理器不可用，返回空字符串
	}

	clientIP := getClientIP(ctx)
	_ = getUserAgent(ctx) // 暂时不使用，避免编译错误

	// 创建会话信息
	sessionInfo := &cache.UserSession{
		UserID:     strconv.FormatInt(userM.ID, 10), // 转换为字符串以保持接口兼容性
		Username:   userM.Username,
		LoginIP:    clientIP,
		DeviceID:   rq.GetDeviceId(),
		LoginTime:  time.Now().Unix(),
		ClientType: getClientTypeFromString(rq.GetClientType()),
	}

	// 生成会话ID
	sessionID := generateSessionID()
	sessionInfo.SessionID = sessionID

	if err := b.sessionManager.CreateSession(ctx, sessionInfo); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return fmt.Sprintf("sess_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// getClientTypeFromString 将字符串转换为ClientType
func getClientTypeFromString(clientType string) cache.ClientType {
	switch clientType {
	case "web":
		return cache.ClientTypeWeb
	case "h5":
		return cache.ClientTypeH5
	case "android":
		return cache.ClientTypeAndroid
	case "ios":
		return cache.ClientTypeIOS
	case "mini_program":
		return cache.ClientTypeMiniProgram
	case "op":
		return cache.ClientTypeOp
	default:
		return cache.ClientTypeWeb // 默认为web
	}
}

// getClientIP 从上下文中获取客户端IP
func getClientIP(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if clientIP := ginCtx.ClientIP(); clientIP != "" {
			return clientIP
		}
	}
	return ""
}

// getUserAgent 从上下文中获取客户端UserAgent
func getUserAgent(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if userAgent := ginCtx.Request.UserAgent(); userAgent != "" {
			return userAgent
		}
	}
	return ""
}
