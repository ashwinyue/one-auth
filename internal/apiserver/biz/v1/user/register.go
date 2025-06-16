// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package user

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authn"
	"github.com/ashwinyue/one-auth/pkg/store/where"
)

// Register 用户注册
func (b *userBiz) Register(ctx context.Context, rq *apiv1.RegisterRequest) (*apiv1.RegisterResponse, error) {
	// 验证手机号格式
	if !b.smsClient.IsValidPhone(rq.GetPhone()) {
		return nil, errno.ErrInvalidArgument.WithMessage("Invalid phone number format")
	}

	// 验证短信验证码
	if b.loginSecurity != nil {
		if err := b.loginSecurity.ValidateVerifyCode(ctx, rq.GetPhone(), "register", rq.GetVerifyCode()); err != nil {
			log.W(ctx).Errorw("Failed to validate verify code", "phone", rq.GetPhone(), "err", err)
			return nil, errno.ErrPasswordInvalid.WithMessage("验证码错误或已过期")
		}
	}

	// 检查手机号是否已注册
	existingStatus, err := b.store.UserStatus().Get(ctx, where.F("auth_id", rq.GetPhone(), "auth_type", int32(model.AuthTypePhone)))
	if err == nil && existingStatus != nil {
		return nil, errno.ErrUserAlreadyExists.WithMessage("Phone number already registered")
	}

	// 检查用户名是否已存在
	if rq.GetUsername() != "" {
		existingUser, err := b.store.User().Get(ctx, where.F("username", rq.GetUsername()))
		if err == nil && existingUser != nil {
			return nil, errno.ErrUserAlreadyExists.WithMessage("Username already exists")
		}
	}

	// 加密密码
	encryptedPassword, err := authn.Encrypt(rq.GetPassword())
	if err != nil {
		log.W(ctx).Errorw("Failed to encrypt password", "err", err)
		return nil, errno.ErrInternal.WithMessage("Failed to encrypt password")
	}

	// 使用事务确保数据一致性
	var responseData *apiv1.RegisterResponse

	err = b.store.TX(ctx, func(txCtx context.Context) error {
		// 创建用户基本信息
		userM := &model.UserM{
			Username:  rq.GetUsername(),
			Password:  encryptedPassword,
			Nickname:  rq.GetNickname(),
			Email:     rq.GetEmail(),
			Phone:     rq.GetPhone(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 如果用户名为空，使用手机号作为用户名
		if userM.Username == "" {
			userM.Username = rq.GetPhone()
		}

		// 如果昵称为空，使用手机号的部分作为昵称
		if userM.Nickname == "" {
			userM.Nickname = fmt.Sprintf("用户%s", rq.GetPhone()[len(rq.GetPhone())-4:])
		}

		// 保存用户基本信息
		if err := b.store.User().Create(txCtx, userM); err != nil {
			log.W(txCtx).Errorw("Failed to create user", "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to create user")
		}

		// 创建手机号认证状态记录
		phoneStatus := &model.UserStatusM{
			AuthID:              rq.GetPhone(),
			AuthType:            int32(model.AuthTypePhone),
			UserID:              userM.ID,
			TenantID:            1, // 默认租户ID
			Status:              int32(model.UserStatusActive),
			LoginCount:          0,
			FailedLoginAttempts: 0,
			IsVerified:          true, // 注册时手机号已验证
			IsPrimary:           true, // 手机号为主要认证方式
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// 设置可选字段的指针值
		now := time.Now()
		phoneStatus.PasswordChangedAt = &now

		if err := b.store.UserStatus().Create(txCtx, phoneStatus); err != nil {
			log.W(txCtx).Errorw("Failed to create phone status", "user_id", userM.ID, "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to create user status")
		}

		// 如果提供了邮箱，也创建一个邮箱认证记录
		if rq.GetEmail() != "" {
			emailStatus := &model.UserStatusM{
				AuthID:              rq.GetEmail(),
				AuthType:            int32(model.AuthTypeEmail),
				UserID:              userM.ID,
				TenantID:            1, // 默认租户ID
				Status:              int32(model.UserStatusActive),
				LoginCount:          0,
				FailedLoginAttempts: 0,
				IsVerified:          false, // 邮箱需要验证
				IsPrimary:           false, // 非主要认证方式
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}
			emailStatus.PasswordChangedAt = &now

			if err := b.store.UserStatus().Create(txCtx, emailStatus); err != nil {
				log.W(txCtx).Errorw("Failed to create email status", "user_id", userM.ID, "email", rq.GetEmail(), "err", err)
				// 邮箱状态创建失败不影响注册，继续
			}
		}

		// 如果用户名与手机号不同，也创建一个用户名认证记录
		if rq.GetUsername() != "" && rq.GetUsername() != rq.GetPhone() {
			usernameStatus := &model.UserStatusM{
				AuthID:              rq.GetUsername(),
				AuthType:            int32(model.AuthTypeUsername),
				UserID:              userM.ID,
				TenantID:            1, // 默认租户ID
				Status:              int32(model.UserStatusActive),
				LoginCount:          0,
				FailedLoginAttempts: 0,
				IsVerified:          true,  // 用户名无需验证
				IsPrimary:           false, // 非主要认证方式
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}
			usernameStatus.PasswordChangedAt = &now

			if err := b.store.UserStatus().Create(txCtx, usernameStatus); err != nil {
				log.W(txCtx).Errorw("Failed to create username status", "user_id", userM.ID, "username", rq.GetUsername(), "err", err)
				// 用户名状态创建失败不影响注册，继续
			}
		}

		log.Infow("用户注册成功",
			"user_id", userM.ID,
			"username", userM.Username,
			"phone", rq.GetPhone(),
			"email", rq.GetEmail())

		responseData = &apiv1.RegisterResponse{
			UserId:  strconv.FormatInt(userM.ID, 10),
			Success: true,
			Message: "注册成功",
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return responseData, nil
}

// BindPhone 绑定手机号
func (b *userBiz) BindPhone(ctx context.Context, rq *apiv1.BindPhoneRequest) (*apiv1.BindPhoneResponse, error) {
	// 验证手机号格式
	if !b.smsClient.IsValidPhone(rq.GetPhone()) {
		return nil, errno.ErrInvalidArgument.WithMessage("Invalid phone number format")
	}

	// 验证短信验证码
	if b.loginSecurity != nil {
		if err := b.loginSecurity.ValidateVerifyCode(ctx, rq.GetPhone(), "bind_phone", rq.GetVerifyCode()); err != nil {
			log.W(ctx).Errorw("Failed to validate verify code", "phone", rq.GetPhone(), "err", err)
			return nil, errno.ErrPasswordInvalid.WithMessage("验证码错误或已过期")
		}
	}

	// 检查手机号是否已被其他用户绑定
	existingStatus, err := b.store.UserStatus().Get(ctx, where.F("auth_id", rq.GetPhone(), "auth_type", int32(model.AuthTypePhone)))
	if err == nil && existingStatus != nil {
		return nil, errno.ErrUserAlreadyExists.WithMessage("Phone number already bound to another user")
	}

	// 获取当前用户信息（从上下文获取用户ID）
	currentUserID := contextx.UserID(ctx)
	if currentUserID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("User not authenticated")
	}

	userM, err := b.store.User().Get(ctx, where.F("id", currentUserID))
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// 使用事务确保数据一致性
	var responseData *apiv1.BindPhoneResponse

	err = b.store.TX(ctx, func(txCtx context.Context) error {
		// 更新用户基本信息中的手机号
		userM.Phone = rq.GetPhone()
		userM.UpdatedAt = time.Now()
		if err := b.store.User().Update(txCtx, userM); err != nil {
			log.W(txCtx).Errorw("Failed to update user phone", "user_id", currentUserID, "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to update user phone")
		}

		// 创建或更新手机号认证状态
		phoneStatus := &model.UserStatusM{
			AuthID:              rq.GetPhone(),
			AuthType:            int32(model.AuthTypePhone),
			UserID:              userM.ID,
			TenantID:            1, // 默认租户ID
			Status:              int32(model.UserStatusActive),
			LoginCount:          0,
			FailedLoginAttempts: 0,
			IsVerified:          true,  // 绑定时手机号已验证
			IsPrimary:           false, // 非主要认证方式
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// 设置可选字段的指针值
		now := time.Now()
		phoneStatus.PasswordChangedAt = &now

		if err := b.store.UserStatus().Create(txCtx, phoneStatus); err != nil {
			log.W(txCtx).Errorw("Failed to create phone status", "user_id", userM.ID, "phone", rq.GetPhone(), "err", err)
			return errno.ErrDBWrite.WithMessage("Failed to bind phone")
		}

		log.Infow("手机号绑定成功",
			"user_id", userM.ID,
			"phone", rq.GetPhone())

		responseData = &apiv1.BindPhoneResponse{
			Success: true,
			Message: "手机号绑定成功",
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return responseData, nil
}

// CheckPhoneAvailable 检查手机号是否可用
func (b *userBiz) CheckPhoneAvailable(ctx context.Context, rq *apiv1.CheckPhoneAvailableRequest) (*apiv1.CheckPhoneAvailableResponse, error) {
	// 验证手机号格式
	if !b.smsClient.IsValidPhone(rq.GetPhone()) {
		return &apiv1.CheckPhoneAvailableResponse{
			Available: false,
			Message:   "Invalid phone number format",
		}, nil
	}

	// 检查手机号是否已被注册
	_, err := b.store.UserStatus().Get(ctx, where.F("auth_id", rq.GetPhone(), "auth_type", int32(model.AuthTypePhone)))
	if err != nil && err != gorm.ErrRecordNotFound {
		log.W(ctx).Errorw("Failed to check phone availability", "phone", rq.GetPhone(), "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to check phone availability")
	}

	available := err == gorm.ErrRecordNotFound
	message := "Phone number is available"
	if !available {
		message = "Phone number is already registered"
	}

	return &apiv1.CheckPhoneAvailableResponse{
		Available: available,
		Message:   message,
	}, nil
}
