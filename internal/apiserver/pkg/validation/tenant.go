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

func (v *Validator) ValidateTenantRules() genericvalidation.Rules {
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
	}
}

// ValidateGetUserTenantsRequest 校验获取用户租户列表请求
func (v *Validator) ValidateGetUserTenantsRequest(ctx context.Context, rq *apiv1.GetUserTenantsRequest) error {
	// 该请求没有参数需要验证
	return nil
}

// ValidateSwitchTenantRequest 校验切换租户请求
func (v *Validator) ValidateSwitchTenantRequest(ctx context.Context, rq *apiv1.SwitchTenantRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateTenantRules())
}

// ValidateGetUserProfileRequest 校验获取用户完整信息请求
func (v *Validator) ValidateGetUserProfileRequest(ctx context.Context, rq *apiv1.GetUserProfileRequest) error {
	// tenant_id 是可选的，如果提供则需要验证
	if rq.TenantId > 0 {
		return genericvalidation.ValidateAllFields(rq, v.ValidateTenantRules())
	}
	return nil
}

// ValidateListTenantsRequest 校验获取租户列表请求
func (v *Validator) ValidateListTenantsRequest(ctx context.Context, rq *apiv1.ListTenantsRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateTenantRules())
}
