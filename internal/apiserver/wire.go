// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

//go:build wireinject
// +build wireinject

package apiserver

import (
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/google/wire"

	"github.com/ashwinyue/one-auth/internal/apiserver/biz"
	"github.com/ashwinyue/one-auth/internal/apiserver/cache"
	"github.com/ashwinyue/one-auth/internal/apiserver/pkg/validation"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/server"
)

func InitializeWebServer(*Config) (server.Server, error) {
	wire.Build(
		wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
		wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
		wire.NewSet(store.ProviderSet, biz.ProviderSet, cache.ProviderSet),
		ProvideDB,    // 提供数据库实例
		ProvideRedis, // 提供Redis实例
		validation.ProviderSet,
		authz.ProviderSet,
	)
	return nil, nil
}
