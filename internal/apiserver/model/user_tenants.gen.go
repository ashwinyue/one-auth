// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameUserTenantM = "user_tenants"

// UserTenantM mapped from table <user_tenants>
type UserTenantM struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement:true;comment:主键ID" json:"id"`                                // 主键ID
	UserID    int64          `gorm:"column:user_id;not null;uniqueIndex:idx_user_tenant;comment:用户ID（关联user表的id字段）" json:"user_id"` // 用户ID（关联user表的id字段）
	TenantID  int64          `gorm:"column:tenant_id;not null;uniqueIndex:idx_user_tenant;comment:租户ID" json:"tenant_id"`           // 租户ID
	Status    bool           `gorm:"column:status;not null;default:1;comment:状态：1-启用，0-禁用" json:"status"`                           // 状态：1-启用，0-禁用
	CreatedAt time.Time      `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"`           // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;default:current_timestamp;comment:更新时间" json:"updated_at"`           // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间（软删除）" json:"deleted_at"`                                   // 删除时间（软删除）
}

// TableName UserTenantM's table name
func (*UserTenantM) TableName() string {
	return TableNameUserTenantM
}
