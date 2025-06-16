// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package tenant

//go:generate mockgen -destination mock_tenant.go -package tenant github.com/ashwinyue/one-auth/internal/apiserver/biz/v1/tenant TenantBiz

import (
	"context"
	"fmt"
	"strconv"

	"net/http"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authz"
	"github.com/ashwinyue/one-auth/pkg/errorsx"
	"github.com/ashwinyue/one-auth/pkg/store/where"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// 定义Tenant模块特有的错误码
var (
	// ErrTenantNotFound 表示租户未找到
	ErrTenantNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.TenantNotFound", Message: "Tenant not found."}

	// ErrTenantDisabled 表示租户已禁用
	ErrTenantDisabled = &errorsx.ErrorX{Code: http.StatusForbidden, Reason: "Forbidden.TenantDisabled", Message: "Tenant is disabled."}
)

// TenantBiz 定义处理租户相关请求所需的方法.
type TenantBiz interface {
	// 用户租户相关
	GetUserTenants(ctx context.Context, rq *apiv1.GetUserTenantsRequest) (*apiv1.GetUserTenantsResponse, error)
	SwitchTenant(ctx context.Context, rq *apiv1.SwitchTenantRequest) (*apiv1.SwitchTenantResponse, error)
	GetUserProfile(ctx context.Context, rq *apiv1.GetUserProfileRequest) (*apiv1.GetUserProfileResponse, error)

	// 租户管理
	ListTenants(ctx context.Context, rq *apiv1.ListTenantsRequest) (*apiv1.ListTenantsResponse, error)
}

// tenantBiz 是 TenantBiz 接口的实现.
type tenantBiz struct {
	store store.IStore
	authz *authz.Authz
}

// 确保 tenantBiz 实现了 TenantBiz 接口.
var _ TenantBiz = (*tenantBiz)(nil)

// New 创建一个新的 TenantBiz 实例.
func New(store store.IStore, authz *authz.Authz) *tenantBiz {
	return &tenantBiz{store: store, authz: authz}
}

// GetUserTenants 获取用户所属的租户列表
func (b *tenantBiz) GetUserTenants(ctx context.Context, rq *apiv1.GetUserTenantsRequest) (*apiv1.GetUserTenantsResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 使用tenant store的扩展方法获取用户租户
	tenants, err := b.store.Tenant().GetUserTenants(ctx, fmt.Sprintf("%d", userID))
	if err != nil {
		log.W(ctx).Errorw("Failed to get user tenants", "user_id", userID, "err", err)
		return nil, err
	}

	// 转换为响应格式
	var tenantList []*apiv1.Tenant
	for _, tenant := range tenants {
		tenantList = append(tenantList, convertTenantToAPI(tenant))
	}

	return &apiv1.GetUserTenantsResponse{Tenants: tenantList}, nil
}

// SwitchTenant 切换用户当前工作租户
func (b *tenantBiz) SwitchTenant(ctx context.Context, rq *apiv1.SwitchTenantRequest) (*apiv1.SwitchTenantResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 验证用户是否属于该租户
	hasAccess, err := b.store.Tenant().CheckUserTenant(ctx, fmt.Sprintf("%d", userID), rq.TenantId)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errno.ErrPermissionDenied.WithMessage("user does not belong to this tenant")
	}

	// 验证租户是否存在且状态有效
	tenant, err := b.store.Tenant().Get(ctx, where.F("id", rq.TenantId))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTenantNotFound.WithMessage("tenant not found")
		}
		return nil, err
	}

	if !tenant.Status {
		return nil, ErrTenantDisabled.WithMessage("tenant is disabled")
	}

	// 这里可以将租户ID存储到会话或缓存中
	// 暂时返回成功，实际实现中可能需要更新JWT token或会话信息
	log.W(ctx).Infow("User switched tenant successfully",
		"user_id", userID,
		"tenant_id", rq.TenantId,
		"tenant_name", tenant.Name)

	return &apiv1.SwitchTenantResponse{Success: true}, nil
}

// GetUserProfile 获取用户完整信息（包含当前租户、角色、权限）
func (b *tenantBiz) GetUserProfile(ctx context.Context, rq *apiv1.GetUserProfileRequest) (*apiv1.GetUserProfileResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == 0 {
		return nil, errno.ErrUnauthenticated.WithMessage("user not found in context")
	}

	// 获取当前租户ID（从请求参数或上下文中获取）
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

	// 获取用户基本信息
	userM, err := b.store.User().Get(ctx, where.F("id", userID))
	if err != nil {
		return nil, err
	}

	// 获取用户在当前租户下的角色
	userIdentifier := fmt.Sprintf("u%d", userID)
	tenantIdentifier := fmt.Sprintf("t%d", tenantID)
	roles, err := b.authz.GetRolesForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user roles", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		roles = []string{} // 如果获取失败，返回空角色列表
	}

	// 获取用户权限和菜单
	permissions, menus := b.getUserPermissionsAndMenus(ctx, userIdentifier, tenantIdentifier)

	// 获取当前租户信息
	tenantM, err := b.store.Tenant().Get(ctx, where.F("id", tenantID))
	if err != nil {
		log.W(ctx).Errorw("Failed to get tenant info", "tenant_id", tenantID, "err", err)
		// 使用默认租户信息
		tenantM = nil
	}

	var currentTenant *apiv1.Tenant
	if tenantM != nil {
		currentTenant = convertTenantToAPI(tenantM)
	} else {
		currentTenant = &apiv1.Tenant{
			Id:   tenantID,
			Name: "Default Tenant",
		}
	}

	return &apiv1.GetUserProfileResponse{
		User: &apiv1.UserProfile{
			Id:            userM.ID,
			UserId:        fmt.Sprintf("%d", userM.ID),
			Username:      userM.Username,
			Nickname:      userM.Nickname,
			Email:         userM.Email,
			Phone:         userM.Phone,
			CurrentTenant: currentTenant,
		},
		Roles:       roles,
		Permissions: permissions,
		Menus:       menus,
	}, nil
}

// getUserPermissionsAndMenus 获取用户权限和菜单
func (b *tenantBiz) getUserPermissionsAndMenus(ctx context.Context, userIdentifier, tenantIdentifier string) ([]*apiv1.Permission, []*apiv1.Menu) {
	// 从Casbin获取用户的所有权限（包括通过角色继承的权限）
	permissions, err := b.authz.GetImplicitPermissionsForUser(userIdentifier, tenantIdentifier)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user permissions", "user_id", userIdentifier, "tenant", tenantIdentifier, "err", err)
		return []*apiv1.Permission{}, []*apiv1.Menu{}
	}

	var permissionList []*apiv1.Permission
	var menuIDSet = make(map[int64]bool)
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

						// 通过menu_permissions关联表查询菜单ID
						_, menuPermissions, err := b.store.MenuPermission().List(ctx, where.F("permission_id", permissionID))
						if err == nil {
							for _, mp := range menuPermissions {
								menuIDSet[mp.MenuID] = true
							}
						}
					}
				}
			}
		}
	}

	// 查询菜单详情
	var menus []*apiv1.Menu
	for menuID := range menuIDSet {
		menuM, err := b.store.Menu().Get(ctx, where.F("id", menuID))
		if err == nil && menuM.Status {
			menus = append(menus, convertMenuToAPI(menuM))
		}
	}

	return permissionList, menus
}

// ListTenants 获取租户列表
func (b *tenantBiz) ListTenants(ctx context.Context, rq *apiv1.ListTenantsRequest) (*apiv1.ListTenantsResponse, error) {
	opts := where.NewWhere()

	// 添加分页
	if rq.Offset > 0 {
		opts = opts.O(int(rq.Offset))
	}
	if rq.Limit > 0 {
		opts = opts.L(int(rq.Limit))
	}

	count, tenants, err := b.store.Tenant().List(ctx, opts)
	if err != nil {
		log.W(ctx).Errorw("Failed to list tenants", "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 转换为响应格式
	var tenantList []*apiv1.Tenant
	for _, tenant := range tenants {
		tenantList = append(tenantList, convertTenantToAPI(tenant))
	}

	return &apiv1.ListTenantsResponse{
		Tenants:    tenantList,
		TotalCount: count,
	}, nil
}

// convertTenantToAPI 转换租户模型为API格式
func convertTenantToAPI(tenantM *model.TenantM) *apiv1.Tenant {
	status := int32(0)
	if tenantM.Status {
		status = 1
	}

	return &apiv1.Tenant{
		Id:          tenantM.ID,
		Name:        tenantM.Name,
		Description: derefString(tenantM.Description),
		Status:      status,
		CreatedAt:   timestamppb.New(tenantM.CreatedAt),
		UpdatedAt:   timestamppb.New(tenantM.UpdatedAt),
	}
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

// convertMenuToAPI 转换菜单模型为API格式
func convertMenuToAPI(menuM *model.MenuM) *apiv1.Menu {
	status := int32(0)
	if menuM.Status {
		status = 1
	}

	apiMenu := &apiv1.Menu{
		Id:        menuM.ID,
		TenantId:  menuM.TenantID,
		Title:     menuM.Title,
		MenuType:  menuM.MenuType,
		SortOrder: menuM.SortOrder,
		Visible:   menuM.Visible,
		Status:    status,
		CreatedAt: timestamppb.New(menuM.CreatedAt),
		UpdatedAt: timestamppb.New(menuM.UpdatedAt),
	}

	// 处理可选字段
	if menuM.ParentID != nil {
		apiMenu.ParentId = *menuM.ParentID
	}
	if menuM.RoutePath != nil {
		apiMenu.RoutePath = *menuM.RoutePath
	}
	if menuM.Component != nil {
		apiMenu.Component = *menuM.Component
	}
	if menuM.Icon != nil {
		apiMenu.Icon = *menuM.Icon
	}

	return apiMenu
}

// derefString 解引用字符串指针
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
