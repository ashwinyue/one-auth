// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package validation

import (
	"context"
	"fmt"

	genericvalidation "github.com/ashwinyue/one-auth/pkg/validation"

	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
)

func (v *Validator) ValidateUserRules() genericvalidation.Rules {
	// 通用的密码校验函数
	validatePassword := func() genericvalidation.ValidatorFunc {
		return func(value any) error {
			return isValidPassword(value.(string))
		}
	}

	// 定义各字段的校验逻辑，通过一个 map 实现模块化和简化
	return genericvalidation.Rules{
		"Password":    validatePassword(),
		"OldPassword": validatePassword(),
		"NewPassword": validatePassword(),
		"UserID": func(value any) error {
			if value.(string) == "" {
				return errno.ErrInvalidArgument.WithMessage("userID cannot be empty")
			}
			return nil
		},
		"Username": func(value any) error {
			if !isValidUsername(value.(string)) {
				return errno.ErrUsernameInvalid
			}
			return nil
		},
		"Nickname": func(value any) error {
			if len(value.(string)) >= 30 {
				return errno.ErrInvalidArgument.WithMessage("nickname must be less than 30 characters")
			}
			return nil
		},
		"Email": func(value any) error {
			return isValidEmail(value.(string))
		},
		"Phone": func(value any) error {
			return isValidPhone(value.(string))
		},
		"Limit": func(value any) error {
			limit := value.(int64)
			if limit <= 0 {
				return errno.ErrInvalidArgument.WithMessage("limit must be greater than 0")
			}
			if limit > 1000 {
				return errno.ErrInvalidArgument.WithMessage("limit cannot exceed 1000")
			}
			return nil
		},
		"Offset": func(value any) error {
			return nil
		},
		// 新增登录相关字段验证
		"LoginType": func(value any) error {
			loginType := value.(string)
			if loginType == "" {
				return errno.ErrInvalidArgument.WithMessage("login_type cannot be empty")
			}
			validTypes := map[string]bool{
				"username": true,
				"email":    true,
				"phone":    true,
			}
			if !validTypes[loginType] {
				return errno.ErrInvalidArgument.WithMessage("invalid login_type, must be one of: username, email, phone")
			}
			return nil
		},
		"Identifier": func(value any) error {
			identifier := value.(string)
			if identifier == "" {
				return errno.ErrInvalidArgument.WithMessage("identifier cannot be empty")
			}
			return nil
		},
		"VerifyCode": func(value any) error {
			code := value.(string)
			if code != "" && len(code) != 6 {
				return errno.ErrInvalidArgument.WithMessage("verify_code must be 6 digits")
			}
			return nil
		},
		"ClientType": func(value any) error {
			clientType := value.(string)
			if clientType != "" {
				validTypes := map[string]bool{
					"web":          true,
					"h5":           true,
					"android":      true,
					"ios":          true,
					"mini_program": true,
					"op":           true,
				}
				if !validTypes[clientType] {
					return errno.ErrInvalidArgument.WithMessage("invalid client_type")
				}
			}
			return nil
		},
		"DeviceId": func(value any) error {
			// 设备ID可以为空，不为空时长度不超过128
			deviceId := value.(string)
			if deviceId != "" && len(deviceId) > 128 {
				return errno.ErrInvalidArgument.WithMessage("device_id must be less than 128 characters")
			}
			return nil
		},
		"Target": func(value any) error {
			target := value.(string)
			if target == "" {
				return errno.ErrInvalidArgument.WithMessage("target cannot be empty")
			}
			return nil
		},
		"CodeType": func(value any) error {
			codeType := value.(string)
			if codeType == "" {
				return errno.ErrInvalidArgument.WithMessage("code_type cannot be empty")
			}
			validTypes := map[string]bool{
				"login":          true,
				"register":       true,
				"reset_password": true,
			}
			if !validTypes[codeType] {
				return errno.ErrInvalidArgument.WithMessage("invalid code_type")
			}
			return nil
		},
		"TargetType": func(value any) error {
			targetType := value.(string)
			if targetType == "" {
				return errno.ErrInvalidArgument.WithMessage("target_type cannot be empty")
			}
			validTypes := map[string]bool{
				"email": true,
				"phone": true,
			}
			if !validTypes[targetType] {
				return errno.ErrInvalidArgument.WithMessage("invalid target_type")
			}
			return nil
		},
		"SessionId": func(value any) error {
			// 会话ID可以为空
			return nil
		},
		"LogoutAll": func(value any) error {
			// 布尔值，无需验证
			return nil
		},
	}
}

// ValidateLoginRequest 校验修改密码请求.
func (v *Validator) ValidateLoginRequest(ctx context.Context, rq *apiv1.LoginRequest) error {
	// 先执行基本字段校验
	if err := genericvalidation.ValidateAllFields(rq, v.ValidateUserRules()); err != nil {
		return err
	}

	// 业务规则校验：密码和验证码必须提供其中一个
	if rq.GetPassword() == "" && rq.GetVerifyCode() == "" {
		return errno.ErrInvalidArgument.WithMessage("Password or verify_code is required")
	}

	return nil
}

// ValidateChangePasswordRequest 校验 ChangePasswordRequest 结构体的有效性.
func (v *Validator) ValidateChangePasswordRequest(ctx context.Context, rq *apiv1.ChangePasswordRequest) error {
	contextUserID := contextx.UserID(ctx)
	if fmt.Sprintf("%d", contextUserID) != rq.GetUserID() {
		return errno.ErrPermissionDenied.WithMessage("The logged-in user `%d` does not match request user `%s`", contextUserID, rq.GetUserID())
	}
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateCreateUserRequest 校验 CreateUserRequest 结构体的有效性.
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *apiv1.CreateUserRequest) error {
	// 基本字段校验
	if err := genericvalidation.ValidateAllFields(rq, v.ValidateUserRules()); err != nil {
		return err
	}

	// 业务逻辑校验已经在ValidateUserRules()中实现了，不需要重复校验

	return nil
}

// ValidateUpdateUserRequest 校验更新用户请求.
func (v *Validator) ValidateUpdateUserRequest(ctx context.Context, rq *apiv1.UpdateUserRequest) error {
	contextUserID := contextx.UserID(ctx)
	if fmt.Sprintf("%d", contextUserID) != rq.GetUserID() {
		return errno.ErrPermissionDenied.WithMessage("The logged-in user `%d` does not match request user `%s`", contextUserID, rq.GetUserID())
	}
	return genericvalidation.ValidateSelectedFields(rq, v.ValidateUserRules(), "UserID")
}

// ValidateDeleteUserRequest 校验 DeleteUserRequest 结构体的有效性.
func (v *Validator) ValidateDeleteUserRequest(ctx context.Context, rq *apiv1.DeleteUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateGetUserRequest 校验 GetUserRequest 结构体的有效性.
func (v *Validator) ValidateGetUserRequest(ctx context.Context, rq *apiv1.GetUserRequest) error {
	contextUserID := contextx.UserID(ctx)
	if fmt.Sprintf("%d", contextUserID) != rq.GetUserID() {
		return errno.ErrPermissionDenied.WithMessage("The logged-in user `%d` does not match request user `%s`", contextUserID, rq.GetUserID())
	}
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateListUserRequest 校验 ListUserRequest 结构体的有效性.
func (v *Validator) ValidateListUserRequest(ctx context.Context, rq *apiv1.ListUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateSendVerifyCodeRequest 校验发送验证码请求.
func (v *Validator) ValidateSendVerifyCodeRequest(ctx context.Context, rq *apiv1.SendVerifyCodeRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateLogoutRequest 校验登出请求.
func (v *Validator) ValidateLogoutRequest(ctx context.Context, rq *apiv1.LogoutRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}
