// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package user

//go:generate mockgen -destination mock_user.go -package user github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/user UserBiz

import (
	"context"
	"time"

	"github.com/ashwinyue/one-auth/internal/apiserver/cache"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/ashwinyue/one-auth/pkg/client/sms"
)

// UserBiz 定义了 user 模块在 biz 层所实现的方法.
type UserBiz interface {
	Create(ctx context.Context, rq *apiv1.CreateUserRequest) (*apiv1.CreateUserResponse, error)
	Update(ctx context.Context, rq *apiv1.UpdateUserRequest) (*apiv1.UpdateUserResponse, error)
	Delete(ctx context.Context, rq *apiv1.DeleteUserRequest) (*apiv1.DeleteUserResponse, error)
	Get(ctx context.Context, rq *apiv1.GetUserRequest) (*apiv1.GetUserResponse, error)
	List(ctx context.Context, rq *apiv1.ListUserRequest) (*apiv1.ListUserResponse, error)
	UserExpansion
}

// UserExpansion 定义了 user 模块在 biz 层所实现的扩展方法.
type UserExpansion interface {
	Login(ctx context.Context, rq *apiv1.LoginRequest) (*apiv1.LoginResponse, error)
	RefreshToken(ctx context.Context, rq *apiv1.RefreshTokenRequest) (*apiv1.RefreshTokenResponse, error)
	ChangePassword(ctx context.Context, rq *apiv1.ChangePasswordRequest) (*apiv1.ChangePasswordResponse, error)
	ListWithBadPerformance(ctx context.Context, rq *apiv1.ListUserRequest) (*apiv1.ListUserResponse, error)
	SendVerifyCode(ctx context.Context, rq *apiv1.SendVerifyCodeRequest) (*apiv1.SendVerifyCodeResponse, error)
	Logout(ctx context.Context, rq *apiv1.LogoutRequest) (*apiv1.LogoutResponse, error)
	Register(ctx context.Context, rq *apiv1.RegisterRequest) (*apiv1.RegisterResponse, error)
	BindPhone(ctx context.Context, rq *apiv1.BindPhoneRequest) (*apiv1.BindPhoneResponse, error)
	CheckPhoneAvailable(ctx context.Context, rq *apiv1.CheckPhoneAvailableRequest) (*apiv1.CheckPhoneAvailableResponse, error)
}

// userBiz 是 UserBiz 接口的实现.
type userBiz struct {
	store          store.IStore
	authz          *authz.Authz
	loginSecurity  *cache.LoginSecurityManager
	sessionManager *cache.SessionManager
	smsClient      sms.Client
}

// 确保 userBiz 实现了 UserBiz 接口.
var _ UserBiz = (*userBiz)(nil)

// New 创建一个 UserBiz 实例.
func New(store store.IStore, authz *authz.Authz, sessionManager *cache.SessionManager, loginSecurity *cache.LoginSecurityManager, smsClient sms.Client) *userBiz {
	return &userBiz{
		store:          store,
		authz:          authz,
		loginSecurity:  loginSecurity,
		sessionManager: sessionManager,
		smsClient:      smsClient,
	}
}

// SessionInfo 会话信息结构体
type SessionInfo struct {
	UserID     string
	Username   string
	ClientIP   string
	UserAgent  string
	ClientType string
	DeviceID   string
	LoginTime  time.Time
}

// 认证相关方法已移至 auth.go 文件
// CRUD相关方法已移至 crud.go 文件
// 注册相关方法已移至 register.go 文件
