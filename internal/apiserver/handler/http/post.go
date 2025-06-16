// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package http

import (
	"github.com/gin-gonic/gin"

	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/core"
)

// CreatePost 创建博客文章.
func (h *Handler) CreatePost(c *gin.Context) {
	var req apiv1.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	if err := h.val.ValidateCreatePostRequest(c, &req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	resp, err := h.biz.PostV1().Create(c, &req)
	core.WriteResponse(c, resp, err)
}

// UpdatePost 更新博客文章.
func (h *Handler) UpdatePost(c *gin.Context) {
	var req apiv1.UpdatePostRequest
	if err := c.ShouldBindUri(&req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	if err := h.val.ValidateUpdatePostRequest(c, &req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	resp, err := h.biz.PostV1().Update(c, &req)
	core.WriteResponse(c, resp, err)
}

// DeletePost 删除博客文章.
func (h *Handler) DeletePost(c *gin.Context) {
	var req apiv1.DeletePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	if err := h.val.ValidateDeletePostRequest(c, &req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	resp, err := h.biz.PostV1().Delete(c, &req)
	core.WriteResponse(c, resp, err)
}

// GetPost 获取博客文章详情.
func (h *Handler) GetPost(c *gin.Context) {
	var req apiv1.GetPostRequest
	if err := c.ShouldBindUri(&req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	if err := h.val.ValidateGetPostRequest(c, &req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	resp, err := h.biz.PostV1().Get(c, &req)
	core.WriteResponse(c, resp, err)
}

// ListPost 获取博客文章列表.
func (h *Handler) ListPost(c *gin.Context) {
	var req apiv1.ListPostRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	if err := h.val.ValidateListPostRequest(c, &req); err != nil {
		core.WriteResponse(c, nil, err)
		return
	}

	resp, err := h.biz.PostV1().List(c, &req)
	core.WriteResponse(c, resp, err)
}
