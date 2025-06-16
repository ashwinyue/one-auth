// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/pkg/conversion"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/store/where"
)

// Create 创建用户.
func (b *userBiz) Create(ctx context.Context, rq *apiv1.CreateUserRequest) (*apiv1.CreateUserResponse, error) {
	// 使用事务确保数据一致性
	var userID int64

	err := b.store.TX(ctx, func(ctx context.Context) error {
		// 1. 创建用户基本信息
		var userM model.UserM
		_ = copier.Copy(&userM, rq)

		if err := b.store.User().Create(ctx, &userM); err != nil {
			return err
		}
		userID = userM.ID // 使用数字主键ID

		// 2. 创建用户状态记录（支持多种认证方式）
		tenantIDStr := contextx.TenantID(ctx)
		var tenantID int64 = 1 // 默认租户
		if tenantIDStr != "" {
			if tid, err := strconv.ParseInt(tenantIDStr, 10, 64); err == nil {
				tenantID = tid
			}
		}

		// 创建用户名认证方式（主要认证方式）
		userStatusUsername := &model.UserStatusM{
			AuthID:     userM.Username,
			AuthType:   int32(model.AuthTypeUsername),
			UserID:     userM.ID, // 使用数字主键ID
			TenantID:   tenantID,
			Status:     int32(model.UserStatusActive),
			IsPrimary:  true,
			IsVerified: true,
		}
		if err := b.store.DB(ctx).Create(userStatusUsername).Error; err != nil {
			return fmt.Errorf("failed to create username auth: %w", err)
		}

		// 创建邮箱认证方式
		if userM.Email != "" {
			userStatusEmail := &model.UserStatusM{
				AuthID:     userM.Email,
				AuthType:   int32(model.AuthTypeEmail),
				UserID:     userM.ID, // 使用数字主键ID
				TenantID:   tenantID,
				Status:     int32(model.UserStatusActive),
				IsPrimary:  false,
				IsVerified: false, // 邮箱需要验证
			}
			if err := b.store.DB(ctx).Create(userStatusEmail).Error; err != nil {
				return fmt.Errorf("failed to create email auth: %w", err)
			}
		}

		// 创建手机号认证方式
		if userM.Phone != "" {
			userStatusPhone := &model.UserStatusM{
				AuthID:     userM.Phone,
				AuthType:   int32(model.AuthTypePhone),
				UserID:     userM.ID, // 使用数字主键ID
				TenantID:   tenantID,
				Status:     int32(model.UserStatusActive),
				IsPrimary:  false,
				IsVerified: false, // 手机号需要验证
			}
			if err := b.store.DB(ctx).Create(userStatusPhone).Error; err != nil {
				return fmt.Errorf("failed to create phone auth: %w", err)
			}
		}

		// 3. 创建用户租户关联
		userTenant := &model.UserTenantM{
			UserID:   userM.ID, // 使用数字主键ID
			TenantID: tenantID,
			Status:   true,
		}
		if err := b.store.DB(ctx).Create(userTenant).Error; err != nil {
			return fmt.Errorf("failed to create user tenant relation: %w", err)
		}

		log.W(ctx).Infow("User created successfully", "user_id", userM.ID, "username", userM.Username, "tenant_id", tenantID)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &apiv1.CreateUserResponse{UserID: fmt.Sprintf("%d", userID)}, nil // 暂时返回字符串格式以保持API兼容性
}

// Update 更新用户信息.
func (b *userBiz) Update(ctx context.Context, rq *apiv1.UpdateUserRequest) (*apiv1.UpdateUserResponse, error) {
	// 将字符串ID转换为数字ID查询
	userIDInt, err := strconv.ParseInt(rq.GetUserID(), 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid user_id format")
	}

	userM, err := b.store.User().Get(ctx, where.F("id", userIDInt))
	if err != nil {
		return nil, err
	}

	if rq.GetNickname() != "" {
		userM.Nickname = rq.GetNickname()
	}
	if rq.GetEmail() != "" {
		userM.Email = rq.GetEmail()
	}
	if rq.GetPhone() != "" {
		userM.Phone = rq.GetPhone()
	}

	if err := b.store.User().Update(ctx, userM); err != nil {
		return nil, err
	}

	return &apiv1.UpdateUserResponse{}, nil
}

// Delete 删除用户.
func (b *userBiz) Delete(ctx context.Context, rq *apiv1.DeleteUserRequest) (*apiv1.DeleteUserResponse, error) {
	// 将字符串ID转换为数字ID
	userIDInt, err := strconv.ParseInt(rq.GetUserID(), 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid user_id format")
	}

	if err := b.store.User().Delete(ctx, where.F("id", userIDInt)); err != nil {
		return nil, err
	}

	return &apiv1.DeleteUserResponse{}, nil
}

// Get 获取用户详情.
func (b *userBiz) Get(ctx context.Context, rq *apiv1.GetUserRequest) (*apiv1.GetUserResponse, error) {
	// 将字符串ID转换为数字ID
	userIDInt, err := strconv.ParseInt(rq.GetUserID(), 10, 64)
	if err != nil {
		return nil, errno.ErrInvalidArgument.WithMessage("invalid user_id format")
	}

	userM, err := b.store.User().Get(ctx, where.F("id", userIDInt))
	if err != nil {
		return nil, err
	}

	user := conversion.UserModelToUserV1(userM)

	return &apiv1.GetUserResponse{User: user}, nil
}

// List 返回用户列表.
func (b *userBiz) List(ctx context.Context, rq *apiv1.ListUserRequest) (*apiv1.ListUserResponse, error) {
	count, list, err := b.store.User().List(ctx, where.P(int(rq.GetOffset()), int(rq.GetLimit())))
	if err != nil {
		log.W(ctx).Errorw("Failed to list users from storage", "err", err)
		return nil, err
	}

	users := make([]*apiv1.User, 0, len(list))
	for _, item := range list {
		users = append(users, conversion.UserModelToUserV1(item))
	}

	log.W(ctx).Infow("Get users from backend storage", "count", len(users))

	return &apiv1.ListUserResponse{TotalCount: count, Users: users}, nil
}

// ListWithBadPerformance 返回用户列表（性能较差的实现，用于演示）.
func (b *userBiz) ListWithBadPerformance(ctx context.Context, rq *apiv1.ListUserRequest) (*apiv1.ListUserResponse, error) {
	// 这里故意使用低效的实现方式，用于演示性能对比
	count, list, err := b.store.User().List(ctx, where.P(int(rq.GetOffset()), int(rq.GetLimit())))
	if err != nil {
		log.W(ctx).Errorw("Failed to list users from storage", "err", err)
		return nil, err
	}

	users := make([]*apiv1.User, 0, len(list))
	for _, item := range list {
		// 模拟额外的数据库查询或处理，降低性能
		userInfo := conversion.UserModelToUserV1(item)

		// 添加一些额外的处理时间
		if userInfo.CreatedAt != nil {
			userInfo.CreatedAt = timestamppb.New(userInfo.CreatedAt.AsTime())
		}

		users = append(users, userInfo)
	}

	log.W(ctx).Infow("Get users from backend storage (bad performance)", "count", len(users))

	return &apiv1.ListUserResponse{TotalCount: count, Users: users}, nil
}

// ValidateUserRequest 展示biz层简单参数校验的示例方法
func (b *userBiz) ValidateUserRequest(ctx context.Context, rq *apiv1.UpdateUserRequest) error {
	// 1. 身份认证检查
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return errno.ErrUnauthenticated
	}

	// 2. 权限检查：用户只能修改自己的信息
	reqUserIDInt, err := strconv.ParseInt(rq.UserID, 10, 64)
	if err != nil {
		return errno.ErrInvalidArgument.WithMessage("invalid user_id format")
	}
	if reqUserIDInt != userID {
		return errno.ErrPermissionDenied.WithMessage("cannot modify other user's information")
	}

	// 3. 业务状态检查示例
	userM, err := b.store.User().Get(ctx, where.F("id", userID))
	if err != nil {
		return errno.ErrUserNotFound
	}

	// 简单的业务状态检查示例（这里只是示例，检查用户是否有效）
	if userM.ID == 0 {
		return errno.ErrPermissionDenied.WithMessage("invalid user")
	}

	return nil
}
