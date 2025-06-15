// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package validation

import (
	"context"

	genericvalidation "github.com/ashwinyue/one-auth/pkg/validation"

	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
)

func (v *Validator) ValidatePermissionRules() genericvalidation.Rules {
	return genericvalidation.Rules{
		"TenantId": func(value any) error {
			if value.(int64) <= 0 {
				return errno.ErrInvalidArgument.WithMessage("tenant_id must be greater than 0")
			}
			return nil
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
			if value.(int64) < 0 {
				return errno.ErrInvalidArgument.WithMessage("offset must be greater than or equal to 0")
			}
			return nil
		},
		"Path": func(value any) error {
			if value.(string) == "" {
				return errno.ErrInvalidArgument.WithMessage("path cannot be empty")
			}
			return nil
		},
		"Method": func(value any) error {
			method := value.(string)
			if method == "" {
				return errno.ErrInvalidArgument.WithMessage("method cannot be empty")
			}
			// 验证HTTP方法
			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true,
				"PATCH": true, "HEAD": true, "OPTIONS": true,
			}
			if !validMethods[method] {
				return errno.ErrInvalidArgument.WithMessage("invalid HTTP method")
			}
			return nil
		},
		"Permissions": func(value any) error {
			permissions := value.([]string)
			if len(permissions) == 0 {
				return errno.ErrInvalidArgument.WithMessage("permissions cannot be empty")
			}
			for _, perm := range permissions {
				if perm == "" {
					return errno.ErrInvalidArgument.WithMessage("permission code cannot be empty")
				}
			}
			return nil
		},
	}
}

// ValidateGetUserPermissionsRequest 校验获取用户权限请求
func (v *Validator) ValidateGetUserPermissionsRequest(ctx context.Context, rq *apiv1.GetUserPermissionsRequest) error {
	// tenant_id 是可选的，如果提供则需要验证
	if rq.TenantId > 0 {
		return genericvalidation.ValidateAllFields(rq, v.ValidatePermissionRules())
	}
	return nil
}

// ValidateCheckPermissionsRequest 校验批量检查权限请求
func (v *Validator) ValidateCheckPermissionsRequest(ctx context.Context, rq *apiv1.CheckPermissionsRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidatePermissionRules())
}

// ValidateCheckAPIAccessRequest 校验检查API访问权限请求
func (v *Validator) ValidateCheckAPIAccessRequest(ctx context.Context, rq *apiv1.CheckAPIAccessRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidatePermissionRules())
}
