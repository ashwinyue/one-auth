// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package role

//go:generate mockgen -destination mock_role.go -package role github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/role RoleBiz

import (
	"context"

	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authz"
)

// RoleBiz 定义处理角色相关请求所需的方法.
type RoleBiz interface {
	// 角色CRUD
	ListRoles(ctx context.Context, rq *apiv1.ListRolesRequest) (*apiv1.ListRolesResponse, error)
	CreateRole(ctx context.Context, rq *apiv1.CreateRoleRequest) (*apiv1.CreateRoleResponse, error)
	UpdateRole(ctx context.Context, rq *apiv1.UpdateRoleRequest) (*apiv1.UpdateRoleResponse, error)
	DeleteRole(ctx context.Context, rq *apiv1.DeleteRoleRequest) (*apiv1.DeleteRoleResponse, error)
	CheckDeleteRole(ctx context.Context, rq *apiv1.CheckDeleteRoleRequest) (*apiv1.CheckDeleteRoleResponse, error)

	// 角色权限管理
	GetRolePermissions(ctx context.Context, rq *apiv1.GetRolePermissionsRequest) (*apiv1.GetRolePermissionsResponse, error)
	AssignRolePermissions(ctx context.Context, rq *apiv1.AssignRolePermissionsRequest) (*apiv1.AssignRolePermissionsResponse, error)
	GetRoleMenus(ctx context.Context, rq *apiv1.GetRoleMenusRequest) (*apiv1.GetRoleMenusResponse, error)
	UpdateRoleMenus(ctx context.Context, rq *apiv1.UpdateRoleMenusRequest) (*apiv1.UpdateRoleMenusResponse, error)

	// 用户角色管理
	GetUserRoles(ctx context.Context, rq *apiv1.GetUserRolesRequest) (*apiv1.GetUserRolesResponse, error)
	AssignUserRoles(ctx context.Context, rq *apiv1.AssignUserRolesRequest) (*apiv1.AssignUserRolesResponse, error)
	GetRolesByUser(ctx context.Context, rq *apiv1.GetRolesByUserRequest) (*apiv1.GetRolesByUserResponse, error)

	// 权限数据刷新
	RefreshPrivilegeData(ctx context.Context, rq *apiv1.RefreshPrivilegeDataRequest) (*apiv1.RefreshPrivilegeDataResponse, error)
}

// roleBiz 是 RoleBiz 接口的实现.
type roleBiz struct {
	store       store.IStore
	authz       *authz.Authz
	idConverter *authz.IDConverter
}

// 确保 roleBiz 实现了 RoleBiz 接口.
var _ RoleBiz = (*roleBiz)(nil)

// New 创建一个新的 RoleBiz 实例.
func New(store store.IStore, authorizer *authz.Authz) *roleBiz {
	return &roleBiz{
		store:       store,
		authz:       authorizer,
		idConverter: authz.NewIDConverter(),
	}
}
