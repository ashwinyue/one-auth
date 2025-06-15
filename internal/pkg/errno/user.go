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

var (
	// ErrUsernameInvalid 表示用户名不合法.
	ErrUsernameInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.UsernameInvalid",
		Message: "Invalid username: Username must consist of letters, digits, and underscores only, and its length must be between 3 and 20 characters.",
	}

	// ErrPasswordInvalid 表示密码不合法.
	ErrPasswordInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.PasswordInvalid",
		Message: "Password is incorrect.",
	}

	// ErrUserAlreadyExists 表示用户已存在.
	ErrUserAlreadyExists = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "AlreadyExist.UserAlreadyExists", Message: "User already exists."}

	// ErrUserNotFound 表示未找到指定用户.
	ErrUserNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.UserNotFound", Message: "User not found."}

	// ErrUserLocked 表示用户账户被锁定.
	ErrUserLocked = &errorsx.ErrorX{Code: http.StatusForbidden, Reason: "Forbidden.UserLocked", Message: "User account is locked"}
	// ErrUserInactive 表示用户账户未激活.
	ErrUserInactive = &errorsx.ErrorX{Code: http.StatusForbidden, Reason: "Forbidden.UserInactive", Message: "User account is inactive"}
	// ErrUserBanned 表示用户账户被封禁.
	ErrUserBanned = &errorsx.ErrorX{Code: http.StatusForbidden, Reason: "Forbidden.UserBanned", Message: "User account is banned"}
)
