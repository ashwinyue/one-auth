package model

// MenuPermissionConfig 菜单权限配置结构
type MenuPermissionConfig struct {
	PermissionCode string `json:"permission_code"` // 权限编码
	IsRequired     bool   `json:"is_required"`     // 是否为必需权限
	AutoCreate     bool   `json:"auto_create"`     // 如果权限不存在是否自动创建
}

// MenuPermissionMatrix 菜单权限矩阵
type MenuPermissionMatrix struct {
	Menu                *MenuM         `json:"menu"`                 // 菜单信息
	AllPermissions      []*PermissionM `json:"all_permissions"`      // 所有权限
	RequiredPermissions []*PermissionM `json:"required_permissions"` // 必需权限
	OptionalPermissions []*PermissionM `json:"optional_permissions"` // 可选权限
}

// HasRequiredPermissions 检查用户是否拥有必需权限
func (m *MenuPermissionMatrix) HasRequiredPermissions(userPermissions []string) bool {
	if len(m.RequiredPermissions) == 0 {
		return true // 如果没有必需权限，则允许访问
	}

	// 创建用户权限映射
	userPermMap := make(map[string]bool)
	for _, perm := range userPermissions {
		userPermMap[perm] = true
	}

	// 检查是否拥有所有必需权限
	for _, reqPerm := range m.RequiredPermissions {
		if !userPermMap[reqPerm.PermissionCode] {
			return false
		}
	}

	return true
}

// MenuWithPermissions 带权限的菜单结构
type MenuWithPermissions struct {
	MenuM                      // 菜单基本信息（内嵌而不是指针）
	Permissions []*PermissionM `json:"permissions"` // 关联的权限列表
}

// PermissionNewM 权限模型的别名，用于兼容性
type PermissionNewM = PermissionM

// AuthType 认证类型枚举
type AuthType int32

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

// UserStatus 用户状态枚举
type UserStatus int32

const (
	UserStatusActive   UserStatus = 1 // 活跃
	UserStatusInactive UserStatus = 2 // 未激活
	UserStatusLocked   UserStatus = 3 // 锁定
	UserStatusBanned   UserStatus = 4 // 封禁
)

// StringToAuthType 将字符串转换为认证类型
func StringToAuthType(s string) AuthType {
	switch s {
	case "username":
		return AuthTypeUsername
	case "email":
		return AuthTypeEmail
	case "phone":
		return AuthTypePhone
	case "wechat":
		return AuthTypeWechat
	case "qq":
		return AuthTypeQQ
	case "github":
		return AuthTypeGithub
	case "google":
		return AuthTypeGoogle
	case "apple":
		return AuthTypeApple
	case "dingtalk":
		return AuthTypeDingtalk
	case "feishu":
		return AuthTypeFeishu
	default:
		return AuthTypeUsername
	}
}

// GetUserByAuthID 根据认证ID获取用户（临时实现）
func GetUserByAuthID(authID string, authType AuthType) (*UserM, error) {
	// 这是一个占位函数，实际应该从数据库查询
	return nil, nil
}

// PermissionCodeValidator 权限代码验证器（临时实现）
func PermissionCodeValidator(code string) bool {
	// 简单的权限代码验证
	return len(code) > 0
}

// CanLogin 检查用户状态是否可以登录
func (u *UserStatusM) CanLogin() bool {
	return u.Status == int32(UserStatusActive)
}

// IsLocked 检查用户是否被锁定
func (u *UserStatusM) IsLocked() bool {
	return u.Status == int32(UserStatusLocked)
}

// GetStatusString 获取状态字符串
func (u *UserStatusM) GetStatusString() string {
	switch UserStatus(u.Status) {
	case UserStatusActive:
		return "active"
	case UserStatusInactive:
		return "inactive"
	case UserStatusLocked:
		return "locked"
	case UserStatusBanned:
		return "banned"
	default:
		return "unknown"
	}
}
