// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package http

import (
	"github.com/ashwinyue/one-auth/internal/apiserver/biz"
	"github.com/ashwinyue/one-auth/internal/apiserver/pkg/validation"
)

// Handler 处理博客模块的请求.
type Handler struct {
	biz biz.IBiz
	val *validation.Validator
}

// NewHandler 创建新的 Handler 实例.
func NewHandler(biz biz.IBiz, val *validation.Validator) *Handler {
	return &Handler{
		biz: biz,
		val: val,
	}
}

// 为了保持代码结构的一致性，这里声明Role相关的方法
// 实际实现可以通过嵌入RoleHandler或直接调用来完成
// 这些方法将在单独的role_handler.go文件中实现
