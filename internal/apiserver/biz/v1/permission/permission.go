// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package permission

//go:generate mockgen -destination mock_permission.go -package permission github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/permission PermissionBiz

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PermissionBiz 定义处理权限相关请求所需的方法.
type PermissionBiz interface {
	// 权限检查相关
	GetUserPermissions(ctx context.Context, rq *apiv1.GetUserPermissionsRequest) (*apiv1.GetUserPermissionsResponse, error)
	CheckPermissions(ctx context.Context, rq *apiv1.CheckPermissionsRequest) (*apiv1.CheckPermissionsResponse, error)
	CheckAPIAccess(ctx context.Context, rq *apiv1.CheckAPIAccessRequest) (*apiv1.CheckAPIAccessResponse, error)
}

// permissionBiz 是 PermissionBiz 接口的实现.
type permissionBiz struct {
	store store.IStore
	authz *authz.Authz
}

// 确保 permissionBiz 实现了 PermissionBiz 接口.
var _ PermissionBiz = (*permissionBiz)(nil)

// New 创建一个新的 PermissionBiz 实例.
func New(store store.IStore, authz *authz.Authz) *permissionBiz {
	return &permissionBiz{store: store, authz: authz}
}

// GetUserPermissions 获取用户权限
func (b *permissionBiz) GetUserPermissions(ctx context.Context, rq *apiv1.GetUserPermissionsRequest) (*apiv1.GetUserPermissionsResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 获取租户信息
	tenantID := rq.TenantId
	if tenantID == 0 {
		contextTenantID := contextx.TenantID(ctx)
		if contextTenantID != "" {
			if tid, err := strconv.ParseInt(contextTenantID, 10, 64); err == nil {
				tenantID = tid
			}
		}
		if tenantID == 0 {
			tenantID = 1 // 默认租户
		}
	}

	// 构建用户标识符和租户标识符
	userIdentifier := fmt.Sprintf("u%d", userID)
	tenantIdentifier := fmt.Sprintf("t%d", tenantID)

	// 从Casbin获取用户的所有权限（包括通过角色继承的权限）
	permissions, err := b.authz.GetImplicitPermissionsForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user permissions", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		return nil, errno.ErrDBRead.WithMessage("Failed to get user permissions")
	}

	var permissionList []*apiv1.Permission
	permissionSet := make(map[int64]bool) // 用于去重

	// 解析权限并查询详细信息
	for _, perm := range permissions {
		if len(perm) >= 2 {
			// perm[1] 是权限标识符，格式为 a{id}
			permissionCode := perm[1]
			if len(permissionCode) > 1 && permissionCode[0] == 'a' {
				// 解析权限ID
				permissionIDStr := permissionCode[1:]
				if permissionID, err := strconv.ParseInt(permissionIDStr, 10, 64); err == nil {
					// 避免重复
					if permissionSet[permissionID] {
						continue
					}
					permissionSet[permissionID] = true

					// 从数据库查询权限详情
					permissionM, err := b.store.Permission().Get(ctx, where.F("id", permissionID))
					if err == nil {
						permissionList = append(permissionList, convertPermissionToAPI(permissionM))
					}
				}
			}
		}
	}

	return &apiv1.GetUserPermissionsResponse{
		Permissions: permissionList,
	}, nil
}

// CheckPermissions 批量检查权限
func (b *permissionBiz) CheckPermissions(ctx context.Context, rq *apiv1.CheckPermissionsRequest) (*apiv1.CheckPermissionsResponse, error) {
	// 获取当前用户
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("无法获取用户信息")
	}

	results := make(map[int64]bool)

	// 遍历需要检查的权限ID
	for _, permissionID := range rq.PermissionIds {
		// 直接使用权限ID进行权限检查
		hasPermission, err := b.authz.Enforce(fmt.Sprintf("u%d", userID), fmt.Sprintf("a%d", permissionID), "allow")
		if err != nil {
			log.W(ctx).Errorw("Failed to check permission", "user_id", userID, "permission_id", permissionID, "err", err)
			results[permissionID] = false
		} else {
			results[permissionID] = hasPermission
		}
	}

	return &apiv1.CheckPermissionsResponse{
		Results: results,
	}, nil
}

// CheckAPIAccess 检查API访问权限
func (b *permissionBiz) CheckAPIAccess(ctx context.Context, rq *apiv1.CheckAPIAccessRequest) (*apiv1.CheckAPIAccessResponse, error) {
	// 获取当前用户ID
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 获取租户信息
	tenantID := rq.TenantId
	if tenantID == 0 {
		contextTenantID := contextx.TenantID(ctx)
		if contextTenantID != "" {
			if tid, err := strconv.ParseInt(contextTenantID, 10, 64); err == nil {
				tenantID = tid
			}
		}
		if tenantID == 0 {
			tenantID = 1 // 默认租户
		}
	}

	// 构建用户标识符和租户标识符
	userIdentifier := fmt.Sprintf("u%d", userID)
	tenantIdentifier := fmt.Sprintf("t%d", tenantID)

	// 查询API对应的权限
	// 根据数据库模型，API路径存储在resource_path字段，HTTP方法存储在http_method字段
	whereCondition := where.F("resource_path", rq.Path).F("tenant_id", tenantID)
	if rq.Method != "" {
		whereCondition = whereCondition.F("http_method", rq.Method)
	}

	_, permissions, err := b.store.Permission().List(ctx, whereCondition)
	if err != nil {
		log.W(ctx).Errorw("Failed to get API permissions",
			"api_path", rq.Path,
			"http_method", rq.Method,
			"err", err)
		return &apiv1.CheckAPIAccessResponse{
			HasAccess: false,
		}, nil
	}

	// 如果没有找到对应的权限配置，默认允许访问
	if len(permissions) == 0 {
		return &apiv1.CheckAPIAccessResponse{
			HasAccess: true,
		}, nil
	}

	// 检查用户是否有任一权限
	for _, permission := range permissions {
		permissionIdentifier := fmt.Sprintf("a%d", permission.ID)

		hasPermission, err := b.authz.Enforce(userIdentifier, permissionIdentifier, tenantIdentifier)
		if err != nil {
			log.W(ctx).Errorw("Failed to check API permission",
				"user_id", userIdentifier,
				"permission_id", permission.ID,
				"err", err)
			continue
		}

		if hasPermission {
			return &apiv1.CheckAPIAccessResponse{
				HasAccess: true,
			}, nil
		}
	}

	return &apiv1.CheckAPIAccessResponse{
		HasAccess: false,
	}, nil
}

// derefString 解引用字符串指针，如果为nil则返回空字符串
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// convertPermissionToAPI 转换权限模型为API格式
func convertPermissionToAPI(permissionM *model.PermissionM) *apiv1.Permission {
	status := int32(0)
	if permissionM.Status {
		status = 1
	}

	return &apiv1.Permission{
		Id:          permissionM.ID,
		TenantId:    permissionM.TenantID,
		Name:        permissionM.Name,
		Description: derefString(permissionM.Description),
		Status:      status,
		CreatedAt:   timestamppb.New(permissionM.CreatedAt),
		UpdatedAt:   timestamppb.New(permissionM.UpdatedAt),
	}
}
