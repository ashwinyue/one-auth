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
	// 菜单相关错误

	// ErrMenuNotFound 表示菜单未找到.
	ErrMenuNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.MenuNotFound", Message: "Menu not found."}

	// ErrMenuCodeExists 表示菜单编码已存在.
	ErrMenuCodeExists = &errorsx.ErrorX{Code: http.StatusConflict, Reason: "Conflict.MenuCodeExists", Message: "Menu code already exists."}

	// ErrMenuPermissionConfiguration 表示菜单权限配置失败.
	ErrMenuPermissionConfiguration = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.MenuPermissionConfiguration", Message: "Menu permission configuration failed."}

	// ErrMenuPermissionNotFound 表示菜单权限关联未找到.
	ErrMenuPermissionNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.MenuPermissionNotFound", Message: "Menu permission association not found."}

	// ErrMenuPermissionExists 表示菜单权限关联已存在.
	ErrMenuPermissionExists = &errorsx.ErrorX{Code: http.StatusConflict, Reason: "Conflict.MenuPermissionExists", Message: "Menu permission association already exists."}

	// ErrMenuAccessDenied 表示菜单访问被拒绝.
	ErrMenuAccessDenied = &errorsx.ErrorX{Code: http.StatusForbidden, Reason: "Forbidden.MenuAccessDenied", Message: "Menu access denied."}

	// ErrMenuInvalidType 表示菜单类型无效.
	ErrMenuInvalidType = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "BadRequest.MenuInvalidType", Message: "Invalid menu type."}

	// ErrMenuHierarchyLoop 表示菜单层级循环引用.
	ErrMenuHierarchyLoop = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "BadRequest.MenuHierarchyLoop", Message: "Menu hierarchy contains circular reference."}

	// ErrMenuBatchConfigurationFailed 表示菜单批量配置失败.
	ErrMenuBatchConfigurationFailed = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.MenuBatchConfigurationFailed", Message: "Menu batch configuration failed."}

	// ErrMenuPermissionMatrixGeneration 表示菜单权限矩阵生成失败.
	ErrMenuPermissionMatrixGeneration = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.MenuPermissionMatrixGeneration", Message: "Menu permission matrix generation failed."}
)
