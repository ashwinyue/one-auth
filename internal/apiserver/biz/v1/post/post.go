// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package post

//go:generate mockgen -destination mock_post.go -package post github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/post PostBiz

import (
	"context"

	"github.com/ashwinyue/one-auth/pkg/store/where"
	"github.com/jinzhu/copier"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/pkg/conversion"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
)

// PostBiz 定义处理帖子请求所需的方法.
type PostBiz interface {
	Create(ctx context.Context, rq *apiv1.CreatePostRequest) (*apiv1.CreatePostResponse, error)
	Update(ctx context.Context, rq *apiv1.UpdatePostRequest) (*apiv1.UpdatePostResponse, error)
	Delete(ctx context.Context, rq *apiv1.DeletePostRequest) (*apiv1.DeletePostResponse, error)
	Get(ctx context.Context, rq *apiv1.GetPostRequest) (*apiv1.GetPostResponse, error)
	List(ctx context.Context, rq *apiv1.ListPostRequest) (*apiv1.ListPostResponse, error)

	PostExpansion
}

// PostExpansion 定义额外的帖子操作方法.
type PostExpansion interface{}

// postBiz 是 PostBiz 接口的实现.
type postBiz struct {
	store store.IStore
}

// 确保 postBiz 实现了 PostBiz 接口.
var _ PostBiz = (*postBiz)(nil)

// New 创建 postBiz 的实例.
func New(store store.IStore) *postBiz {
	return &postBiz{store: store}
}

// Create 实现 PostBiz 接口中的 Create 方法.
func (b *postBiz) Create(ctx context.Context, rq *apiv1.CreatePostRequest) (*apiv1.CreatePostResponse, error) {
	var postM model.PostM
	_ = copier.Copy(&postM, rq)

	// 设置当前用户ID（从认证中间件中获取）
	postM.UserID = contextx.UserID(ctx)

	if err := b.store.Post().Create(ctx, &postM); err != nil {
		log.W(ctx).Errorw("Failed to create post", "err", err, "user_id", postM.UserID, "title", postM.Title)
		return nil, err
	}

	log.W(ctx).Infow("Post created successfully", "post_id", postM.PostID, "user_id", postM.UserID, "title", postM.Title)

	return &apiv1.CreatePostResponse{PostID: postM.PostID}, nil
}

// Update 实现 PostBiz 接口中的 Update 方法.
func (b *postBiz) Update(ctx context.Context, rq *apiv1.UpdatePostRequest) (*apiv1.UpdatePostResponse, error) {
	// 使用tenant上下文和post_id查询，确保只能修改自己租户内的数据
	whr := where.T(ctx).F("post_id", rq.GetPostID())
	postM, err := b.store.Post().Get(ctx, whr)
	if err != nil {
		log.W(ctx).Errorw("Failed to get post for update", "err", err, "post_id", rq.GetPostID())
		return nil, err
	}

	// 验证当前用户是否有权限修改该文章（只能修改自己的文章）
	currentUserID := contextx.UserID(ctx)
	if postM.UserID != currentUserID {
		log.W(ctx).Warnw("User trying to update post of other user",
			"current_user_id", currentUserID,
			"post_owner_id", postM.UserID,
			"post_id", rq.GetPostID())
		return nil, errno.ErrPermissionDenied.WithMessage("You can only update your own posts")
	}

	// 更新字段
	updated := false
	if rq.Title != nil {
		postM.Title = rq.GetTitle()
		updated = true
	}

	if rq.Content != nil {
		postM.Content = rq.GetContent()
		updated = true
	}

	if !updated {
		log.W(ctx).Infow("No fields to update", "post_id", rq.GetPostID())
		return &apiv1.UpdatePostResponse{}, nil
	}

	if err := b.store.Post().Update(ctx, postM); err != nil {
		log.W(ctx).Errorw("Failed to update post", "err", err, "post_id", rq.GetPostID())
		return nil, err
	}

	log.W(ctx).Infow("Post updated successfully", "post_id", postM.PostID, "user_id", postM.UserID)

	return &apiv1.UpdatePostResponse{}, nil
}

// Delete 实现 PostBiz 接口中的 Delete 方法.
func (b *postBiz) Delete(ctx context.Context, rq *apiv1.DeletePostRequest) (*apiv1.DeletePostResponse, error) {
	currentUserID := contextx.UserID(ctx)
	deletedCount := 0

	// 批量删除多个帖子
	for _, postID := range rq.GetPostIDs() {
		// 先查询帖子确保存在且当前用户有权限删除
		whr := where.T(ctx).F("post_id", postID)
		postM, err := b.store.Post().Get(ctx, whr)
		if err != nil {
			log.W(ctx).Errorw("Failed to get post for deletion", "err", err, "post_id", postID)
			continue // 继续删除其他帖子
		}

		// 验证权限（只能删除自己的帖子）
		if postM.UserID != currentUserID {
			log.W(ctx).Warnw("User trying to delete post of other user",
				"current_user_id", currentUserID,
				"post_owner_id", postM.UserID,
				"post_id", postID)
			continue // 跳过无权限的帖子
		}

		// 执行删除
		if err := b.store.Post().Delete(ctx, whr); err != nil {
			log.W(ctx).Errorw("Failed to delete post", "err", err, "post_id", postID)
			continue
		}

		deletedCount++
		log.W(ctx).Infow("Post deleted successfully", "post_id", postID, "user_id", currentUserID)
	}

	log.W(ctx).Infow("Post deletion completed", "requested_count", len(rq.GetPostIDs()), "deleted_count", deletedCount)

	return &apiv1.DeletePostResponse{}, nil
}

// Get 实现 PostBiz 接口中的 Get 方法.
func (b *postBiz) Get(ctx context.Context, rq *apiv1.GetPostRequest) (*apiv1.GetPostResponse, error) {
	// 使用tenant上下文和post_id查询
	whr := where.T(ctx).F("post_id", rq.GetPostID())
	postM, err := b.store.Post().Get(ctx, whr)
	if err != nil {
		log.W(ctx).Errorw("Failed to get post", "err", err, "post_id", rq.GetPostID())
		return nil, err
	}

	log.W(ctx).Infow("Post retrieved successfully", "post_id", postM.PostID, "title", postM.Title)

	return &apiv1.GetPostResponse{Post: conversion.PostModelToPostV1(postM)}, nil
}

// List 实现 PostBiz 接口中的 List 方法.
func (b *postBiz) List(ctx context.Context, rq *apiv1.ListPostRequest) (*apiv1.ListPostResponse, error) {
	// 构建查询条件
	whr := where.T(ctx).P(int(rq.GetOffset()), int(rq.GetLimit()))

	// 如果有标题过滤条件
	if rq.Title != nil && *rq.Title != "" {
		whr = whr.F("title", "LIKE", "%"+*rq.Title+"%")
	}

	count, postList, err := b.store.Post().List(ctx, whr)
	if err != nil {
		log.W(ctx).Errorw("Failed to list posts", "err", err)
		return nil, err
	}

	posts := make([]*apiv1.Post, 0, len(postList))
	for _, post := range postList {
		converted := conversion.PostModelToPostV1(post)
		posts = append(posts, converted)
	}

	log.W(ctx).Infow("Posts listed successfully", "count", len(posts), "total", count)

	return &apiv1.ListPostResponse{TotalCount: count, Posts: posts}, nil
}
