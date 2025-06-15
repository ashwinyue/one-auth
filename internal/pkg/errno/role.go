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
	// ErrRoleAlreadyExists 表示角色已存在.
	ErrRoleAlreadyExists = &errorsx.ErrorX{
		Code:    http.StatusConflict,
		Reason:  "Conflict.RoleAlreadyExists",
		Message: "Role already exists.",
	}

	// ErrRoleNotFound 表示角色未找到.
	ErrRoleNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.RoleNotFound",
		Message: "Role not found.",
	}

	// ErrRoleCannotDelete 表示角色不能删除.
	ErrRoleCannotDelete = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "BadRequest.RoleCannotDelete",
		Message: "Role cannot be deleted.",
	}

	// ErrRoleInUse 表示角色正在使用中.
	ErrRoleInUse = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "BadRequest.RoleInUse",
		Message: "Role is in use and cannot be modified.",
	}

	// ErrInvalidRoleCode 表示角色编码无效.
	ErrInvalidRoleCode = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "BadRequest.InvalidRoleCode",
		Message: "Invalid role code.",
	}
)
