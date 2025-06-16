// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package menu_permission

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	"github.com/ashwinyue/one-auth/internal/pkg/contextx"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
	v1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/store/where"
)

// MenuPermissionBiz 定义了菜单权限相关的业务逻辑接口
type MenuPermissionBiz interface {
	// 配置菜单权限
	ConfigureMenuPermissions(ctx context.Context, r *v1.ConfigureMenuPermissionsRequest) (*v1.ConfigureMenuPermissionsResponse, error)

	// 获取菜单权限
	GetMenuPermissions(ctx context.Context, r *v1.GetMenuPermissionsRequest) (*v1.GetMenuPermissionsResponse, error)

	// 获取权限关联的菜单
	GetPermissionMenus(ctx context.Context, r *v1.GetPermissionMenusRequest) (*v1.GetPermissionMenusResponse, error)

	// 获取用户菜单权限
	GetUserMenuPermissions(ctx context.Context, r *v1.GetUserMenuPermissionsRequest) (*v1.GetUserMenuPermissionsResponse, error)

	// 验证菜单访问权限
	ValidateMenuAccess(ctx context.Context, r *v1.ValidateMenuAccessRequest) (*v1.ValidateMenuAccessResponse, error)

	// 批量配置菜单权限
	BatchConfigureMenuPermissions(ctx context.Context, r *v1.BatchConfigureMenuPermissionsRequest) (*v1.BatchConfigureMenuPermissionsResponse, error)

	// 获取菜单权限矩阵
	GetMenuPermissionMatrix(ctx context.Context, r *v1.GetMenuPermissionMatrixRequest) (*v1.GetMenuPermissionMatrixResponse, error)

	// 同步菜单权限到Casbin
	SyncMenuPermissionsToCasbin(ctx context.Context, tenantID int64) error

	// 权限校验辅助方法
	CheckMenuPermission(ctx context.Context, userID string, menuID int64, action string) (bool, error)
}

// menuPermissionBiz 是 MenuPermissionBiz 接口的实现
type menuPermissionBiz struct {
	ds store.IStore
}

// 确保 menuPermissionBiz 实现了 MenuPermissionBiz 接口
var _ MenuPermissionBiz = (*menuPermissionBiz)(nil)

// NewMenuPermissionBiz 创建一个新的菜单权限业务逻辑实例
func NewMenuPermissionBiz(ds store.IStore) *menuPermissionBiz {
	return &menuPermissionBiz{ds: ds}
}

// ConfigureMenuPermissions 配置菜单权限
func (b *menuPermissionBiz) ConfigureMenuPermissions(ctx context.Context, r *v1.ConfigureMenuPermissionsRequest) (*v1.ConfigureMenuPermissionsResponse, error) {
	// 验证菜单存在
	_, err := b.ds.Menu().Get(ctx, where.F("id", r.MenuId))
	if err != nil {
		log.W(ctx).Errorw("Menu not found", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrMenuNotFound.WithMessage(fmt.Sprintf("菜单ID %d 不存在", r.MenuId))
	}

	// 转换权限配置
	var permissions []model.MenuPermissionConfig
	for _, perm := range r.Permissions {
		permissions = append(permissions, model.MenuPermissionConfig{
			PermissionID:   perm.PermissionId,
			PermissionName: perm.PermissionName,
			IsRequired:     perm.IsRequired,
			AutoCreate:     perm.AutoCreate,
		})
	}

	// 配置菜单权限
	err = b.ds.MenuPermission().ConfigureMenuPermissions(ctx, r.MenuId, permissions)
	if err != nil {
		log.W(ctx).Errorw("Failed to configure menu permissions", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrMenuPermissionConfiguration.WithMessage(err.Error())
	}

	log.W(ctx).Infow("Menu permissions configured successfully", "menu_id", r.MenuId, "permission_count", len(permissions))

	return &v1.ConfigureMenuPermissionsResponse{
		Success:       true,
		Message:       "菜单权限配置成功",
		AffectedCount: int32(len(permissions)),
	}, nil
}

// GetMenuPermissions 获取菜单权限
func (b *menuPermissionBiz) GetMenuPermissions(ctx context.Context, r *v1.GetMenuPermissionsRequest) (*v1.GetMenuPermissionsResponse, error) {
	// 获取菜单的所有权限
	allPermissions, err := b.ds.MenuPermission().GetMenuPermissions(ctx, r.MenuId)
	if err != nil {
		log.W(ctx).Errorw("Failed to get menu permissions", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 获取必需权限
	requiredPermissions, err := b.ds.MenuPermission().GetMenuRequiredPermissions(ctx, r.MenuId)
	if err != nil {
		log.W(ctx).Errorw("Failed to get menu required permissions", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 计算可选权限
	requiredMap := make(map[int64]bool)
	for _, req := range requiredPermissions {
		requiredMap[req.ID] = true
	}

	var optionalPermissions []*model.PermissionNewM
	for _, perm := range allPermissions {
		if !requiredMap[perm.ID] {
			optionalPermissions = append(optionalPermissions, perm)
		}
	}

	// 转换为API格式
	return &v1.GetMenuPermissionsResponse{
		MenuId:              r.MenuId,
		AllPermissions:      convertPermissionsToAPI(allPermissions),
		RequiredPermissions: convertPermissionsToAPI(requiredPermissions),
		OptionalPermissions: convertPermissionsToAPI(optionalPermissions),
	}, nil
}

// GetPermissionMenus 获取权限关联的菜单
func (b *menuPermissionBiz) GetPermissionMenus(ctx context.Context, r *v1.GetPermissionMenusRequest) (*v1.GetPermissionMenusResponse, error) {
	// 获取权限关联的菜单
	menus, err := b.ds.MenuPermission().GetPermissionMenus(ctx, r.PermissionId)
	if err != nil {
		log.W(ctx).Errorw("Failed to get permission menus", "permission_id", r.PermissionId, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 为每个菜单获取权限信息
	var menuWithPermissions []*v1.MenuWithPermissions
	for _, menu := range menus {
		permissions, _ := b.ds.MenuPermission().GetMenuPermissions(ctx, menu.ID)
		requiredPerms, _ := b.ds.MenuPermission().GetMenuRequiredPermissions(ctx, menu.ID)

		menuWithPerms := &v1.MenuWithPermissions{
			MenuId:              menu.ID,
			Title:               menu.Title,
			Permissions:         convertPermissionsToAPI(permissions),
			RequiredPermissions: convertPermissionsToAPI(requiredPerms),
		}
		menuWithPermissions = append(menuWithPermissions, menuWithPerms)
	}

	return &v1.GetPermissionMenusResponse{
		PermissionId: r.PermissionId,
		Menus:        menuWithPermissions,
	}, nil
}

// GetUserMenuPermissions 获取用户菜单权限
func (b *menuPermissionBiz) GetUserMenuPermissions(ctx context.Context, r *v1.GetUserMenuPermissionsRequest) (*v1.GetUserMenuPermissionsResponse, error) {
	userID := r.UserId
	if userID == "" {
		// 从上下文获取当前用户ID
		if contextUserID := contextx.UserID(ctx); contextUserID != 0 {
			userID = fmt.Sprintf("u%d", contextUserID)
		} else {
			return nil, errno.ErrUnauthenticated.WithMessage("无法获取用户信息")
		}
	}

	// 获取用户可访问的菜单
	accessibleMenus, err := b.ds.MenuPermission().GetUserAccessibleMenus(ctx, userID, r.TenantId)
	if err != nil {
		log.W(ctx).Errorw("Failed to get user accessible menus", "user_id", userID, "tenant_id", r.TenantId, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 获取总菜单数
	totalCount, _, err := b.ds.Menu().List(ctx, where.F("tenant_id", r.TenantId, "status", true))
	if err != nil {
		totalCount = 0
	}

	// 转换为API格式
	var apiMenus []*v1.MenuWithPermissions
	for _, menu := range accessibleMenus {
		apiMenu := &v1.MenuWithPermissions{
			MenuId:      menu.ID,
			Title:       menu.Title,
			Permissions: convertPermissionsToAPI(menu.Permissions),
		}

		// 获取必需权限
		requiredPerms, _ := b.ds.MenuPermission().GetMenuRequiredPermissions(ctx, menu.ID)
		apiMenu.RequiredPermissions = convertPermissionsToAPI(requiredPerms)

		// 如果需要包含可执行操作
		if r.IncludeActions {
			// 获取用户在此菜单的可执行操作
			apiMenu.AvailableActions = []string{"view", "access"}
		}

		apiMenus = append(apiMenus, apiMenu)
	}

	return &v1.GetUserMenuPermissionsResponse{
		UserId:              userID,
		TenantId:            r.TenantId,
		AccessibleMenus:     apiMenus,
		TotalMenuCount:      int32(totalCount),
		AccessibleMenuCount: int32(len(apiMenus)),
	}, nil
}

// ValidateMenuAccess 验证菜单访问权限
func (b *menuPermissionBiz) ValidateMenuAccess(ctx context.Context, r *v1.ValidateMenuAccessRequest) (*v1.ValidateMenuAccessResponse, error) {
	userID := r.UserId
	if userID == "" {
		// 从上下文获取当前用户ID
		if contextUserID := contextx.UserID(ctx); contextUserID != 0 {
			userID = fmt.Sprintf("u%d", contextUserID)
		} else {
			return nil, errno.ErrUnauthenticated.WithMessage("无法获取用户信息")
		}
	}

	// 获取菜单必需权限
	requiredPermissions, err := b.ds.MenuPermission().GetMenuRequiredPermissions(ctx, r.MenuId)
	if err != nil {
		log.W(ctx).Errorw("Failed to get menu required permissions", "menu_id", r.MenuId, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 获取用户权限（这里简化实现，实际应该通过authz获取）
	userPermissions := []string{} // 从authz获取用户权限

	// 检查权限
	var missingPermissions []string
	for _, reqPerm := range requiredPermissions {
		if !contains(userPermissions, reqPerm.Name) {
			missingPermissions = append(missingPermissions, reqPerm.Name)
		}
	}

	hasAccess := len(missingPermissions) == 0

	// 获取可执行操作（基于用户权限）
	availableActions := []string{}
	if hasAccess {
		availableActions = append(availableActions, "view", "access")
	}

	message := "权限验证成功"
	if !hasAccess {
		message = fmt.Sprintf("缺少必需权限: %s", strings.Join(missingPermissions, ", "))
	}

	return &v1.ValidateMenuAccessResponse{
		HasAccess:          hasAccess,
		MissingPermissions: missingPermissions,
		AvailableActions:   availableActions,
		Message:            message,
	}, nil
}

// contains 检查字符串切片是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// BatchConfigureMenuPermissions 批量配置菜单权限
func (b *menuPermissionBiz) BatchConfigureMenuPermissions(ctx context.Context, r *v1.BatchConfigureMenuPermissionsRequest) (*v1.BatchConfigureMenuPermissionsResponse, error) {
	var processedCount int32
	var errorCount int32
	var errors []string

	for _, config := range r.Configurations {
		// 转换权限配置
		var permissions []model.MenuPermissionConfig
		for _, perm := range config.Permissions {
			permissions = append(permissions, model.MenuPermissionConfig{
				PermissionID:   perm.PermissionId,
				PermissionName: perm.PermissionName,
				IsRequired:     perm.IsRequired,
				AutoCreate:     perm.AutoCreate,
			})
		}

		// 配置菜单权限
		err := b.ds.MenuPermission().ConfigureMenuPermissions(ctx, config.MenuId, permissions)
		if err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("菜单ID %d: %v", config.MenuId, err))
			log.W(ctx).Errorw("Failed to configure menu permissions in batch", "menu_id", config.MenuId, "err", err)
		} else {
			processedCount++
		}
	}

	success := errorCount == 0
	message := "批量配置完成"
	if errorCount > 0 {
		message = fmt.Sprintf("批量配置完成，%d个成功，%d个失败", processedCount, errorCount)
	}

	return &v1.BatchConfigureMenuPermissionsResponse{
		Success:        success,
		Message:        message,
		ProcessedCount: processedCount,
		ErrorCount:     errorCount,
		Errors:         errors,
	}, nil
}

// GetMenuPermissionMatrix 获取菜单权限矩阵
func (b *menuPermissionBiz) GetMenuPermissionMatrix(ctx context.Context, r *v1.GetMenuPermissionMatrixRequest) (*v1.GetMenuPermissionMatrixResponse, error) {
	// 获取菜单权限矩阵
	matrices, err := b.ds.MenuPermission().GetMenuPermissionMatrix(ctx, r.TenantId)
	if err != nil {
		log.W(ctx).Errorw("Failed to get menu permission matrix", "tenant_id", r.TenantId, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	// 转换为API格式
	var apiMatrices []*v1.MenuPermissionMatrix
	totalPermissions := 0

	for _, matrix := range matrices {
		// 菜单类型过滤
		if len(r.MenuTypes) > 0 {
			found := false
			for _, menuType := range r.MenuTypes {
				if int32(matrix.Menu.MenuType) == menuType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		apiMatrix := &v1.MenuPermissionMatrix{
			MenuId:              matrix.Menu.ID,
			MenuCode:            matrix.Menu.MenuCode,
			MenuTitle:           matrix.Menu.Title,
			RequiredPermissions: convertPermissionsToAPI(matrix.RequiredPermissions),
			OptionalPermissions: convertPermissionsToAPI(matrix.OptionalPermissions),
			PermissionCount:     int32(len(matrix.AllPermissions)),
		}

		apiMatrices = append(apiMatrices, apiMatrix)
		totalPermissions += len(matrix.AllPermissions)
	}

	return &v1.GetMenuPermissionMatrixResponse{
		TenantId:         r.TenantId,
		Matrix:           apiMatrices,
		TotalMenus:       int32(len(apiMatrices)),
		TotalPermissions: int32(totalPermissions),
	}, nil
}

// SyncMenuPermissionsToCasbin 同步菜单权限到Casbin
func (b *menuPermissionBiz) SyncMenuPermissionsToCasbin(ctx context.Context, tenantID int64) error {
	// 实现同步逻辑
	log.W(ctx).Infow("Syncing menu permissions to Casbin", "tenant_id", tenantID)
	return nil
}

// CheckMenuPermission 检查菜单权限
func (b *menuPermissionBiz) CheckMenuPermission(ctx context.Context, userID string, menuID int64, action string) (bool, error) {
	// 实现权限检查逻辑
	return true, nil
}

// convertPermissionsToAPI 转换权限模型为API格式
func convertPermissionsToAPI(permissions []*model.PermissionNewM) []*v1.Permission {
	var result []*v1.Permission
	for _, perm := range permissions {
		result = append(result, &v1.Permission{
			Id:       perm.ID,
			TenantId: perm.TenantID,
			Name:     perm.Name,
		})
	}
	return result
}
