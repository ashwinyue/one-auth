// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameUserStatusM = "user_status"

// UserStatusM mapped from table <user_status>
type UserStatusM struct {
	ID                  int64          `gorm:"column:id;primaryKey;autoIncrement:true;comment:主键ID" json:"id"`                                                                                                               // 主键ID
	AuthID              string         `gorm:"column:auth_id;not null;uniqueIndex:idx_auth_id_type;comment:认证标识符（邮箱、手机号、用户名等）" json:"auth_id"`                                                                               // 认证标识符（邮箱、手机号、用户名等）
	AuthType            int32          `gorm:"column:auth_type;not null;uniqueIndex:idx_auth_id_type;comment:认证类型：1-username,2-email,3-phone,4-wechat,5-qq,6-github,7-google,8-apple,9-dingtalk,10-feishu" json:"auth_type"` // 认证类型：1-username,2-email,3-phone,4-wechat,5-qq,6-github,7-google,8-apple,9-dingtalk,10-feishu
	UserID              int64          `gorm:"column:user_id;not null;comment:用户ID（关联user表的id）" json:"user_id"`                                                                                                              // 用户ID（关联user表的id）
	TenantID            int64          `gorm:"column:tenant_id;not null;default:1;comment:租户ID" json:"tenant_id"`                                                                                                            // 租户ID
	Status              int32          `gorm:"column:status;not null;default:1;comment:用户状态：1-active,2-inactive,3-locked,4-banned" json:"status"`                                                                            // 用户状态：1-active,2-inactive,3-locked,4-banned
	LockReason          *string        `gorm:"column:lock_reason;comment:锁定原因" json:"lock_reason"`                                                                                                                           // 锁定原因
	LockedUntil         *time.Time     `gorm:"column:locked_until;comment:锁定到期时间" json:"locked_until"`                                                                                                                       // 锁定到期时间
	LastLoginTime       *time.Time     `gorm:"column:last_login_time;comment:最后登录时间" json:"last_login_time"`                                                                                                                 // 最后登录时间
	LastLoginIP         *string        `gorm:"column:last_login_ip;comment:最后登录IP" json:"last_login_ip"`                                                                                                                     // 最后登录IP
	LastLoginDevice     *string        `gorm:"column:last_login_device;comment:最后登录设备" json:"last_login_device"`                                                                                                             // 最后登录设备
	LoginCount          int32          `gorm:"column:login_count;not null;comment:登录次数" json:"login_count"`                                                                                                                  // 登录次数
	FailedLoginAttempts int32          `gorm:"column:failed_login_attempts;not null;comment:累计登录失败次数" json:"failed_login_attempts"`                                                                                          // 累计登录失败次数
	LastFailedLogin     *time.Time     `gorm:"column:last_failed_login;comment:最后一次登录失败时间" json:"last_failed_login"`                                                                                                         // 最后一次登录失败时间
	PasswordChangedAt   *time.Time     `gorm:"column:password_changed_at;comment:密码最后修改时间" json:"password_changed_at"`                                                                                                       // 密码最后修改时间
	IsVerified          bool           `gorm:"column:is_verified;not null;comment:是否已验证" json:"is_verified"`                                                                                                                 // 是否已验证
	VerifiedAt          *time.Time     `gorm:"column:verified_at;comment:验证时间" json:"verified_at"`                                                                                                                           // 验证时间
	IsPrimary           bool           `gorm:"column:is_primary;not null;comment:是否为主要认证方式" json:"is_primary"`                                                                                                               // 是否为主要认证方式
	CreatedAt           time.Time      `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"`                                                                                          // 创建时间
	UpdatedAt           time.Time      `gorm:"column:updated_at;not null;default:current_timestamp;comment:更新时间" json:"updated_at"`                                                                                          // 更新时间
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间（软删除）" json:"deleted_at"`                                                                                                                  // 删除时间（软删除）
}

// TableName UserStatusM's table name
func (*UserStatusM) TableName() string {
	return TableNameUserStatusM
}
