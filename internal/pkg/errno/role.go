// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package errno

import (
	"net/http"

	"github.com/ashwinyue/one-auth/pkg/errorsx"
)

// 角色相关错误定义
var (
	// ErrRoleNotFound 表示角色未找到.
	ErrRoleNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.RoleNotFound",
		Message: "Role not found.",
	}

	// ErrRoleNameExists 表示角色名称已存在.
	ErrRoleNameExists = &errorsx.ErrorX{
		Code:    http.StatusConflict,
		Reason:  "Conflict.RoleNameExists",
		Message: "Role name already exists.",
	}

	// ErrRoleAlreadyExists 表示角色已存在.
	ErrRoleAlreadyExists = &errorsx.ErrorX{
		Code:    http.StatusConflict,
		Reason:  "Conflict.RoleAlreadyExists",
		Message: "Role already exists.",
	}

	// ErrRoleDeleteWithUsers 表示角色有关联用户无法删除.
	ErrRoleDeleteWithUsers = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "BadRequest.RoleDeleteWithUsers",
		Message: "Role has associated users and cannot be deleted.",
	}

	// ErrRolePermissionConfiguration 表示角色权限配置失败.
	ErrRolePermissionConfiguration = &errorsx.ErrorX{
		Code:    http.StatusInternalServerError,
		Reason:  "InternalError.RolePermissionConfiguration",
		Message: "Role permission configuration failed.",
	}

	// ErrRolePermissionNotFound 表示角色权限关联未找到.
	ErrRolePermissionNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.RolePermissionNotFound",
		Message: "Role permission association not found.",
	}

	// ErrRolePermissionExists 表示角色权限关联已存在.
	ErrRolePermissionExists = &errorsx.ErrorX{
		Code:    http.StatusConflict,
		Reason:  "Conflict.RolePermissionExists",
		Message: "Role permission association already exists.",
	}

	// ErrRoleAccessDenied 表示角色访问被拒绝.
	ErrRoleAccessDenied = &errorsx.ErrorX{
		Code:    http.StatusForbidden,
		Reason:  "Forbidden.RoleAccessDenied",
		Message: "Role access denied.",
	}

	// ErrRoleInvalidStatus 表示角色状态无效.
	ErrRoleInvalidStatus = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "BadRequest.RoleInvalidStatus",
		Message: "Invalid role status.",
	}

	// ErrRoleBatchOperationFailed 表示角色批量操作失败.
	ErrRoleBatchOperationFailed = &errorsx.ErrorX{
		Code:    http.StatusInternalServerError,
		Reason:  "InternalError.RoleBatchOperationFailed",
		Message: "Role batch operation failed.",
	}

	// ErrRoleHierarchyLoop 表示角色层级循环引用.
	ErrRoleHierarchyLoop = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "BadRequest.RoleHierarchyLoop",
		Message: "Role hierarchy contains circular reference.",
	}
)
