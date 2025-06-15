// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package apiserver

import (
	"context"

	"github.com/ashwinyue/one-auth/pkg/core"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	handler "github.com/ashwinyue/one-auth/internal/apiserver/handler/http"
	"github.com/ashwinyue/one-auth/internal/apiserver/routes"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	mw "github.com/ashwinyue/one-auth/internal/pkg/middleware/gin"
	"github.com/ashwinyue/one-auth/internal/pkg/server"
)

// ginServer 定义一个使用 Gin 框架开发的 HTTP 服务器.
type ginServer struct {
	srv server.Server
}

// 确保 *ginServer 实现了 server.Server 接口.
var _ server.Server = (*ginServer)(nil)

// NewGinServer 初始化一个新的 Gin 服务器实例.
func (c *ServerConfig) NewGinServer() server.Server {
	// 创建 Gin 引擎
	engine := gin.New()

	// 注册全局中间件，用于恢复 panic、设置 HTTP 头、添加请求 ID 等
	engine.Use(gin.Recovery(), mw.NoCache, mw.Cors, mw.Secure, mw.RequestIDMiddleware())

	// 注册 REST API 路由
	c.InstallRESTAPI(engine)

	httpsrv := server.NewHTTPServer(c.cfg.HTTPOptions, c.cfg.TLSOptions, engine)

	return &ginServer{srv: httpsrv}
}

// InstallRESTAPI 注册 API 路由。路由的路径和 HTTP 方法，严格遵循 REST 规范.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	// 注册业务无关的 API 接口
	InstallGenericAPI(engine)

	// 创建核心业务处理器
	h := handler.NewHandler(c.biz, c.val)

	// 注册健康检查接口
	engine.GET("/healthz", h.Healthz)

	// 注册用户登录和令牌刷新接口。这2个接口比较简单，所以没有 API 版本
	engine.POST("/login", h.Login)
	engine.POST("/send-verify-code", h.SendVerifyCode) // 发送验证码不需要认证
	// 注意：认证中间件要在 handler.RefreshToken 之前加载
	engine.PUT("/refresh-token", mw.AuthnMiddleware(c.retriever), h.RefreshToken)
	engine.POST("/logout", mw.AuthnMiddleware(c.retriever), h.Logout) // 登出需要认证

	// 认证和授权中间件
	authMiddlewares := []gin.HandlerFunc{mw.AuthnMiddleware(c.retriever), mw.AuthzMiddleware(c.authz)}

	// 注册 v1 版本 API 路由分组
	v1 := engine.Group("/v1")

	// 按模块安装路由
	routes.InstallUserRoutes(v1, h, authMiddlewares...)
	routes.InstallTenantRoutes(v1, h, authMiddlewares...)
	routes.InstallRoleRoutes(v1, h, authMiddlewares...)
	routes.InstallPermissionRoutes(v1, h, authMiddlewares...)
	routes.InstallMenuRoutes(v1, h, authMiddlewares...)
	routes.InstallPostRoutes(v1, h, authMiddlewares...)
}

// InstallGenericAPI 注册业务无关的路由，例如 pprof、404 处理等.
func InstallGenericAPI(engine *gin.Engine) {
	// 注册 pprof 路由
	pprof.Register(engine)

	// 注册 404 路由处理
	engine.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrPageNotFound, nil)
	})
}

// RunOrDie 启动 Gin 服务器，出错则程序崩溃退出.
func (s *ginServer) RunOrDie() {
	s.srv.RunOrDie()
}

// GracefulStop 优雅停止服务器.
func (s *ginServer) GracefulStop(ctx context.Context) {
	s.srv.GracefulStop(ctx)
}
