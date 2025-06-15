// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package http

import (
	"time"

	"github.com/ashwinyue/one-auth/pkg/core"
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
)

// Healthz 服务健康检查.
func (h *Handler) Healthz(c *gin.Context) {
	log.W(c.Request.Context()).Infow("Healthz handler is called", "method", "Healthz", "status", "healthy")
	core.WriteResponse(c, apiv1.HealthzResponse{
		Status:    apiv1.ServiceStatus_Healthy,
		Timestamp: time.Now().Format(time.DateTime),
	}, nil)
}
