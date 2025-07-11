// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameRoleM = "roles"

// RoleM mapped from table <roles>
type RoleM struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement:true;comment:角色主键ID" json:"id"`                    // 角色主键ID
	TenantID    int64          `gorm:"column:tenant_id;not null;comment:租户ID" json:"tenant_id"`                             // 租户ID
	Name        string         `gorm:"column:name;not null;uniqueIndex:idx_name_tenant;comment:角色名称" json:"name"`           // 角色名称
	Description *string        `gorm:"column:description;comment:描述" json:"description"`                                    // 描述
	Status      bool           `gorm:"column:status;not null;default:1;comment:状态：1-启用，0-禁用" json:"status"`                 // 状态：1-启用，0-禁用
	CreatedAt   time.Time      `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null;default:current_timestamp;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间（软删除）" json:"deleted_at"`                         // 删除时间（软删除）
}

// TableName RoleM's table name
func (*RoleM) TableName() string {
	return TableNameRoleM
}
