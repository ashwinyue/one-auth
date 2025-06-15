// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package model

import (
	"time"

	"gorm.io/gorm"
)

// AuthType 认证类型
type AuthType int8

const (
	AuthTypeUsername AuthType = 1  // 用户名
	AuthTypeEmail    AuthType = 2  // 邮箱
	AuthTypePhone    AuthType = 3  // 手机号
	AuthTypeWechat   AuthType = 4  // 微信
	AuthTypeQQ       AuthType = 5  // QQ
	AuthTypeGithub   AuthType = 6  // Github
	AuthTypeGoogle   AuthType = 7  // Google
	AuthTypeApple    AuthType = 8  // Apple
	AuthTypeDingtalk AuthType = 9  // 钉钉
	AuthTypeFeishu   AuthType = 10 // 飞书
)

// AuthTypeMap 认证类型映射
var AuthTypeMap = map[string]AuthType{
	"username": AuthTypeUsername,
	"email":    AuthTypeEmail,
	"phone":    AuthTypePhone,
	"wechat":   AuthTypeWechat,
	"qq":       AuthTypeQQ,
	"github":   AuthTypeGithub,
	"google":   AuthTypeGoogle,
	"apple":    AuthTypeApple,
	"dingtalk": AuthTypeDingtalk,
	"feishu":   AuthTypeFeishu,
}

// AuthTypeStringMap 认证类型反向映射
var AuthTypeStringMap = map[AuthType]string{
	AuthTypeUsername: "username",
	AuthTypeEmail:    "email",
	AuthTypePhone:    "phone",
	AuthTypeWechat:   "wechat",
	AuthTypeQQ:       "qq",
	AuthTypeGithub:   "github",
	AuthTypeGoogle:   "google",
	AuthTypeApple:    "apple",
	AuthTypeDingtalk: "dingtalk",
	AuthTypeFeishu:   "feishu",
}

// UserStatus 用户状态类型
type UserStatus int8

const (
	UserStatusActive   UserStatus = 1 // 活跃
	UserStatusInactive UserStatus = 2 // 未激活
	UserStatusLocked   UserStatus = 3 // 锁定
	UserStatusBanned   UserStatus = 4 // 封禁
)

// UserStatusMap 用户状态映射
var UserStatusMap = map[string]UserStatus{
	"active":   UserStatusActive,
	"inactive": UserStatusInactive,
	"locked":   UserStatusLocked,
	"banned":   UserStatusBanned,
}

// UserStatusStringMap 用户状态反向映射
var UserStatusStringMap = map[UserStatus]string{
	UserStatusActive:   "active",
	UserStatusInactive: "inactive",
	UserStatusLocked:   "locked",
	UserStatusBanned:   "banned",
}

// UserStatusM 用户状态模型
type UserStatusM struct {
	ID       int64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AuthID   string   `gorm:"column:auth_id;type:varchar(255);not null" json:"auth_id"`
	AuthType AuthType `gorm:"column:auth_type;type:tinyint;not null" json:"auth_type"`
	UserID   int64    `gorm:"column:user_id;type:bigint unsigned;not null" json:"user_id"`
	TenantID int64    `gorm:"column:tenant_id;not null;default:1" json:"tenant_id"`

	// 用户状态
	Status      UserStatus `gorm:"column:status;type:tinyint;not null;default:1" json:"status"`
	LockReason  *string    `gorm:"column:lock_reason;type:varchar(255)" json:"lock_reason,omitempty"`
	LockedUntil *time.Time `gorm:"column:locked_until" json:"locked_until,omitempty"`

	// 登录信息
	LastLoginTime   *time.Time `gorm:"column:last_login_time" json:"last_login_time,omitempty"`
	LastLoginIP     *string    `gorm:"column:last_login_ip;type:varchar(45)" json:"last_login_ip,omitempty"`
	LastLoginDevice *string    `gorm:"column:last_login_device;type:varchar(128)" json:"last_login_device,omitempty"`
	LoginCount      int        `gorm:"column:login_count;not null;default:0" json:"login_count"`

	// 安全信息
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts;not null;default:0" json:"failed_login_attempts"`
	LastFailedLogin     *time.Time `gorm:"column:last_failed_login" json:"last_failed_login,omitempty"`
	PasswordChangedAt   *time.Time `gorm:"column:password_changed_at" json:"password_changed_at,omitempty"`

	// 验证状态
	IsVerified bool       `gorm:"column:is_verified;not null;default:0" json:"is_verified"`
	VerifiedAt *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`
	IsPrimary  bool       `gorm:"column:is_primary;not null;default:0" json:"is_primary"`

	// 时间戳
	CreatedAt time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (u *UserStatusM) TableName() string {
	return "user_status"
}

// GetAuthTypeString 获取认证类型字符串
func (u *UserStatusM) GetAuthTypeString() string {
	return AuthTypeStringMap[u.AuthType]
}

// GetStatusString 获取状态字符串
func (u *UserStatusM) GetStatusString() string {
	return UserStatusStringMap[u.Status]
}

// IsActive 检查用户是否为活跃状态
func (u *UserStatusM) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsLocked 检查用户是否被锁定
func (u *UserStatusM) IsLocked() bool {
	if u.Status == UserStatusLocked {
		// 检查锁定是否已过期
		if u.LockedUntil != nil && time.Now().After(*u.LockedUntil) {
			return false
		}
		return true
	}
	return false
}

// CanLogin 检查用户是否可以登录
func (u *UserStatusM) CanLogin() bool {
	return u.IsActive() && !u.IsLocked()
}

// GetUserAuthMethods 获取用户的所有认证方式
func GetUserAuthMethods(db *gorm.DB, userID int64) ([]UserStatusM, error) {
	var authMethods []UserStatusM
	err := db.Where("user_id = ?", userID).Find(&authMethods).Error
	return authMethods, err
}

// GetUserByAuthID 根据认证标识符查找用户
func GetUserByAuthID(db *gorm.DB, authID string, authType AuthType) (*UserStatusM, error) {
	var userStatus UserStatusM
	err := db.Where("auth_id = ? AND auth_type = ?", authID, authType).First(&userStatus).Error
	if err != nil {
		return nil, err
	}
	return &userStatus, nil
}

// GetPrimaryAuthMethod 获取用户的主要认证方式
func GetPrimaryAuthMethod(db *gorm.DB, userID int64) (*UserStatusM, error) {
	var userStatus UserStatusM
	err := db.Where("user_id = ? AND is_primary = ?", userID, true).First(&userStatus).Error
	if err != nil {
		return nil, err
	}
	return &userStatus, nil
}

// GetPrimaryAuthForUser 获取用户的主要认证方式
func GetPrimaryAuthForUser(authMethods []UserStatusM) *UserStatusM {
	for _, auth := range authMethods {
		if auth.IsPrimary {
			return &auth
		}
	}
	// 如果没有主要认证方式，返回第一个用户名类型的
	for _, auth := range authMethods {
		if auth.AuthType == AuthTypeUsername {
			return &auth
		}
	}
	// 如果都没有，返回第一个
	if len(authMethods) > 0 {
		return &authMethods[0]
	}
	return nil
}

// StringToAuthType 字符串转认证类型
func StringToAuthType(s string) AuthType {
	if authType, ok := AuthTypeMap[s]; ok {
		return authType
	}
	return AuthTypeUsername // 默认返回用户名类型
}

// StringToUserStatus 字符串转用户状态
func StringToUserStatus(s string) UserStatus {
	if status, ok := UserStatusMap[s]; ok {
		return status
	}
	return UserStatusActive // 默认返回活跃状态
}
