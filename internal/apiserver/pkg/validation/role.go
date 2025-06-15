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

func (v *Validator) ValidateRoleRules() genericvalidation.Rules {
	return genericvalidation.Rules{
		"TenantId": func(value any) error {
			if value.(int64) <= 0 {
				return errno.ErrInvalidArgument.WithMessage("tenant_id must be greater than 0")
			}
			return nil
		},
		"RoleId": func(value any) error {
			if value.(int64) <= 0 {
				return errno.ErrInvalidArgument.WithMessage("role_id must be greater than 0")
			}
			return nil
		},
		"UserId": func(value any) error {
			if value.(string) == "" {
				return errno.ErrInvalidArgument.WithMessage("user_id cannot be empty")
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
		"RoleIds": func(value any) error {
			roleIds := value.([]int64)
			if len(roleIds) == 0 {
				return errno.ErrInvalidArgument.WithMessage("role_ids cannot be empty")
			}
			for _, id := range roleIds {
				if id <= 0 {
					return errno.ErrInvalidArgument.WithMessage("all role_ids must be greater than 0")
				}
			}
			return nil
		},
		"PermissionIds": func(value any) error {
			permissionIds := value.([]int64)
			if len(permissionIds) == 0 {
				return errno.ErrInvalidArgument.WithMessage("permission_ids cannot be empty")
			}
			for _, id := range permissionIds {
				if id <= 0 {
					return errno.ErrInvalidArgument.WithMessage("all permission_ids must be greater than 0")
				}
			}
			return nil
		},
	}
}

// ValidateListRolesRequest 校验获取角色列表请求
func (v *Validator) ValidateListRolesRequest(ctx context.Context, rq *apiv1.ListRolesRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
}

// ValidateGetRolePermissionsRequest 校验获取角色权限请求
func (v *Validator) ValidateGetRolePermissionsRequest(ctx context.Context, rq *apiv1.GetRolePermissionsRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
}

// ValidateAssignRolePermissionsRequest 校验分配角色权限请求
func (v *Validator) ValidateAssignRolePermissionsRequest(ctx context.Context, rq *apiv1.AssignRolePermissionsRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
}

// ValidateGetUserRolesRequest 校验获取用户角色请求
func (v *Validator) ValidateGetUserRolesRequest(ctx context.Context, rq *apiv1.GetUserRolesRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
}

// ValidateAssignUserRolesRequest 校验分配用户角色请求
func (v *Validator) ValidateAssignUserRolesRequest(ctx context.Context, rq *apiv1.AssignUserRolesRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
}

// ValidateCreateRoleRequest 校验创建角色请求
func (v *Validator) ValidateCreateRoleRequest(ctx context.Context, rq *apiv1.CreateRoleRequest) error {
	rules := v.ValidateRoleRules()
	rules["RoleCode"] = func(value any) error {
		if value.(string) == "" {
			return errno.ErrInvalidArgument.WithMessage("role_code cannot be empty")
		}
		return nil
	}
	rules["Name"] = func(value any) error {
		if value.(string) == "" {
			return errno.ErrInvalidArgument.WithMessage("name cannot be empty")
		}
		return nil
	}
	return genericvalidation.ValidateAllFields(rq, rules)
}

// ValidateUpdateRoleRequest 校验更新角色请求
func (v *Validator) ValidateUpdateRoleRequest(ctx context.Context, rq *apiv1.UpdateRoleRequest) error {
	rules := v.ValidateRoleRules()
	rules["Name"] = func(value any) error {
		if value.(string) == "" {
			return errno.ErrInvalidArgument.WithMessage("name cannot be empty")
		}
		return nil
	}
	return genericvalidation.ValidateAllFields(rq, rules)
}

// ValidateDeleteRoleRequest 校验删除角色请求
func (v *Validator) ValidateDeleteRoleRequest(ctx context.Context, rq *apiv1.DeleteRoleRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
}

// ValidateCheckDeleteRoleRequest 校验检查删除角色请求
func (v *Validator) ValidateCheckDeleteRoleRequest(ctx context.Context, rq *apiv1.CheckDeleteRoleRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
}

// ValidateGetRoleMenusRequest 校验获取角色菜单请求
func (v *Validator) ValidateGetRoleMenusRequest(ctx context.Context, rq *apiv1.GetRoleMenusRequest) error {
	rules := v.ValidateRoleRules()
	rules["Platform"] = func(value any) error {
		platform := value.(int32)
		if platform != 0 && platform != 1 && platform != 2 {
			return errno.ErrInvalidArgument.WithMessage("platform must be 0, 1, or 2")
		}
		return nil
	}
	return genericvalidation.ValidateAllFields(rq, rules)
}

// ValidateUpdateRoleMenusRequest 校验更新角色菜单请求
func (v *Validator) ValidateUpdateRoleMenusRequest(ctx context.Context, rq *apiv1.UpdateRoleMenusRequest) error {
	rules := v.ValidateRoleRules()
	rules["MenuIds"] = func(value any) error {
		menuIds := value.([]int64)
		for _, id := range menuIds {
			if id <= 0 {
				return errno.ErrInvalidArgument.WithMessage("all menu_ids must be greater than 0")
			}
		}
		return nil
	}
	return genericvalidation.ValidateAllFields(rq, rules)
}

// ValidateGetRolesByUserRequest 校验获取用户角色请求
func (v *Validator) ValidateGetRolesByUserRequest(ctx context.Context, rq *apiv1.GetRolesByUserRequest) error {
	// tenant_id 是可选的，如果提供则需要验证
	if rq.TenantId > 0 {
		return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
	}
	return nil
}

// ValidateRefreshPrivilegeDataRequest 校验刷新权限数据请求
func (v *Validator) ValidateRefreshPrivilegeDataRequest(ctx context.Context, rq *apiv1.RefreshPrivilegeDataRequest) error {
	// tenant_id 是可选的，如果提供则需要验证
	if rq.TenantId > 0 {
		return genericvalidation.ValidateAllFields(rq, v.ValidateRoleRules())
	}
	return nil
}
