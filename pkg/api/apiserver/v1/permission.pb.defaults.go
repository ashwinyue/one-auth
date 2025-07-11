// 权限管理 API 定义

// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

// Code generated by protoc-gen-defaults. DO NOT EDIT.

package v1

import (
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	_ *timestamppb.Timestamp
	_ *durationpb.Duration
	_ *wrapperspb.BoolValue
)

func (x *Permission) Default() {
}

func (x *GetUserPermissionsRequest) Default() {
}

func (x *GetUserPermissionsResponse) Default() {
}

func (x *CheckPermissionsRequest) Default() {
}

func (x *CheckPermissionsResponse) Default() {
}

func (x *CheckAPIAccessRequest) Default() {
}

func (x *CheckAPIAccessResponse) Default() {
}
