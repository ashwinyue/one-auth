// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package store

import (
	"context"

	genericstore "github.com/ashwinyue/one-auth/pkg/store"
	"github.com/ashwinyue/one-auth/pkg/store/where"

	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/pkg/errno"
	"github.com/ashwinyue/one-auth/internal/pkg/log"
)

// MenuPermissionStore 定义了菜单权限关联的存储层方法
type MenuPermissionStore interface {
	Create(ctx context.Context, obj *model.MenuPermissionM) error
	Update(ctx context.Context, obj *model.MenuPermissionM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.MenuPermissionM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.MenuPermissionM, error)

	MenuPermissionExpansion
}

// MenuPermissionExpansion 定义了菜单权限关联的扩展方法
type MenuPermissionExpansion interface {
	// 批量配置菜单权限
	ConfigureMenuPermissions(ctx context.Context, menuID int64, permissions []model.MenuPermissionConfig) error

	// 获取菜单的所有权限
	GetMenuPermissions(ctx context.Context, menuID int64) ([]*model.PermissionNewM, error)

	// 获取菜单的必需权限
	GetMenuRequiredPermissions(ctx context.Context, menuID int64) ([]*model.PermissionNewM, error)

	// 获取权限关联的菜单
	GetPermissionMenus(ctx context.Context, permissionID int64) ([]*model.MenuM, error)

	// 检查菜单是否具有指定权限
	HasMenuPermission(ctx context.Context, menuID, permissionID int64) (bool, error)

	// 获取菜单权限矩阵
	GetMenuPermissionMatrix(ctx context.Context, tenantID int64) ([]*model.MenuPermissionMatrix, error)

	// 获取用户可访问的菜单
	GetUserAccessibleMenus(ctx context.Context, userID string, tenantID int64) ([]*model.MenuWithPermissions, error)

	// 清除菜单的所有权限关联
	ClearMenuPermissions(ctx context.Context, menuID int64) error

	// 批量删除权限关联
	BatchDeleteByPermissionIDs(ctx context.Context, permissionIDs []int64) error
}

// menuPermissionStore 是 MenuPermissionStore 接口的实现
type menuPermissionStore struct {
	*genericstore.Store[model.MenuPermissionM]
	ds *datastore
}

// 确保 menuPermissionStore 实现了 MenuPermissionStore 接口
var _ MenuPermissionStore = (*menuPermissionStore)(nil)

// newMenuPermissionStore 创建 menuPermissionStore 的实例
func newMenuPermissionStore(store *datastore) *menuPermissionStore {
	return &menuPermissionStore{
		Store: genericstore.NewStore[model.MenuPermissionM](store, NewLogger()),
		ds:    store,
	}
}

// ConfigureMenuPermissions 批量配置菜单权限
func (s *menuPermissionStore) ConfigureMenuPermissions(ctx context.Context, menuID int64, permissions []model.MenuPermissionConfig) error {
	return s.ds.TX(ctx, func(ctx context.Context) error {
		// 1. 清除现有关联
		if err := s.ClearMenuPermissions(ctx, menuID); err != nil {
			return err
		}

		// 2. 添加新的权限关联
		for _, permConfig := range permissions {
			// 查找权限ID
			permission, err := s.ds.Permission().Get(ctx, where.F("permission_code", permConfig.PermissionCode))
			if err != nil {
				if permConfig.AutoCreate {
					// 自动创建权限
					newPerm := &model.PermissionM{
						PermissionCode: permConfig.PermissionCode,
						Name:           permConfig.PermissionCode,
						ResourceType:   "menu",
						Status:         true,
					}
					if err := s.ds.DB(ctx).Create(newPerm).Error; err != nil {
						log.W(ctx).Errorw("Failed to auto create permission", "code", permConfig.PermissionCode, "err", err)
						continue
					}
					permission = newPerm
				} else {
					log.W(ctx).Errorw("Permission not found", "code", permConfig.PermissionCode, "err", err)
					continue
				}
			}

			// 创建关联
			menuPerm := &model.MenuPermissionM{
				MenuID:       menuID,
				PermissionID: permission.ID,
				IsRequired:   permConfig.IsRequired,
			}

			if err := s.Create(ctx, menuPerm); err != nil {
				log.W(ctx).Errorw("Failed to create menu permission", "menu_id", menuID, "permission_id", permission.ID, "err", err)
				return err
			}
		}

		return nil
	})
}

// GetMenuPermissions 获取菜单的所有权限
func (s *menuPermissionStore) GetMenuPermissions(ctx context.Context, menuID int64) ([]*model.PermissionNewM, error) {
	var permissions []*model.PermissionNewM

	err := s.ds.DB(ctx).Table("permissions p").
		Select("p.*").
		Joins("JOIN menu_permissions mp ON mp.permission_id = p.id").
		Where("mp.menu_id = ? AND p.deleted_at IS NULL", menuID).
		Find(&permissions).Error

	if err != nil {
		log.W(ctx).Errorw("Failed to get menu permissions", "menu_id", menuID, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	return permissions, nil
}

// GetMenuRequiredPermissions 获取菜单的必需权限
func (s *menuPermissionStore) GetMenuRequiredPermissions(ctx context.Context, menuID int64) ([]*model.PermissionNewM, error) {
	var permissions []*model.PermissionNewM

	err := s.ds.DB(ctx).Table("permissions p").
		Select("p.*").
		Joins("JOIN menu_permissions mp ON mp.permission_id = p.id").
		Where("mp.menu_id = ? AND mp.is_required = 1 AND p.deleted_at IS NULL", menuID).
		Find(&permissions).Error

	if err != nil {
		log.W(ctx).Errorw("Failed to get menu required permissions", "menu_id", menuID, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	return permissions, nil
}

// GetPermissionMenus 获取权限关联的菜单
func (s *menuPermissionStore) GetPermissionMenus(ctx context.Context, permissionID int64) ([]*model.MenuM, error) {
	var menus []*model.MenuM

	err := s.ds.DB(ctx).Table("menus m").
		Select("m.*").
		Joins("JOIN menu_permissions mp ON mp.menu_id = m.id").
		Where("mp.permission_id = ? AND m.deleted_at IS NULL", permissionID).
		Find(&menus).Error

	if err != nil {
		log.W(ctx).Errorw("Failed to get permission menus", "permission_id", permissionID, "err", err)
		return nil, errno.ErrDBRead.WithMessage(err.Error())
	}

	return menus, nil
}

// HasMenuPermission 检查菜单是否具有指定权限
func (s *menuPermissionStore) HasMenuPermission(ctx context.Context, menuID, permissionID int64) (bool, error) {
	_, err := s.Get(ctx, where.F("menu_id", menuID, "permission_id", permissionID))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetMenuPermissionMatrix 获取菜单权限矩阵
func (s *menuPermissionStore) GetMenuPermissionMatrix(ctx context.Context, tenantID int64) ([]*model.MenuPermissionMatrix, error) {
	// 获取租户下的所有菜单
	_, menus, err := s.ds.Menu().List(ctx, where.F("tenant_id", tenantID, "status", true))
	if err != nil {
		return nil, err
	}

	var matrices []*model.MenuPermissionMatrix
	for _, menu := range menus {
		// 获取菜单的所有权限
		allPermissions, err := s.GetMenuPermissions(ctx, menu.ID)
		if err != nil {
			log.W(ctx).Errorw("Failed to get menu permissions", "menu_id", menu.ID, "err", err)
			continue
		}

		// 获取必需权限
		requiredPermissions, err := s.GetMenuRequiredPermissions(ctx, menu.ID)
		if err != nil {
			log.W(ctx).Errorw("Failed to get menu required permissions", "menu_id", menu.ID, "err", err)
			continue
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

		matrix := &model.MenuPermissionMatrix{
			Menu:                menu,
			RequiredPermissions: requiredPermissions,
			OptionalPermissions: optionalPermissions,
			AllPermissions:      allPermissions,
		}

		matrices = append(matrices, matrix)
	}

	return matrices, nil
}

// GetUserAccessibleMenus 获取用户可访问的菜单
func (s *menuPermissionStore) GetUserAccessibleMenus(ctx context.Context, userID string, tenantID int64) ([]*model.MenuWithPermissions, error) {
	// 获取用户权限列表（这里简化实现，实际应该通过authz获取）
	userPermissions := []string{} // 从authz获取用户权限

	// 获取菜单权限矩阵
	matrices, err := s.GetMenuPermissionMatrix(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var accessibleMenus []*model.MenuWithPermissions
	for _, matrix := range matrices {
		// 检查用户是否有权限访问此菜单
		if matrix.HasRequiredPermissions(userPermissions) {
			menuWithPerms := &model.MenuWithPermissions{
				MenuM:       *matrix.Menu,
				Permissions: matrix.AllPermissions,
			}
			accessibleMenus = append(accessibleMenus, menuWithPerms)
		}
	}

	return accessibleMenus, nil
}

// ClearMenuPermissions 清除菜单的所有权限关联
func (s *menuPermissionStore) ClearMenuPermissions(ctx context.Context, menuID int64) error {
	return s.Delete(ctx, where.F("menu_id", menuID))
}

// BatchDeleteByPermissionIDs 批量删除权限关联
func (s *menuPermissionStore) BatchDeleteByPermissionIDs(ctx context.Context, permissionIDs []int64) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	// 转换为interface{}切片
	ids := make([]interface{}, len(permissionIDs))
	for i, id := range permissionIDs {
		ids[i] = id
	}

	return s.ds.DB(ctx).Where("permission_id IN (?)", ids).Delete(&model.MenuPermissionM{}).Error
}
