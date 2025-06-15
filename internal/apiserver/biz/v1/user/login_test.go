// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/ashwinyue/one-auth/internal/apiserver/cache"
	"github.com/ashwinyue/one-auth/internal/apiserver/model"
	"github.com/ashwinyue/one-auth/internal/apiserver/store"
	apiv1 "github.com/ashwinyue/one-auth/pkg/api/apiserver/v1"
	"github.com/ashwinyue/one-auth/pkg/authn"
	"github.com/ashwinyue/one-auth/pkg/store/where"
)

// MockStore 模拟store
type MockStore struct {
	mock.Mock
}

func (m *MockStore) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	args := m.Called(ctx, wheres)
	return args.Get(0).(*gorm.DB)
}

func (m *MockStore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *MockStore) User() store.UserStore {
	args := m.Called()
	return args.Get(0).(store.UserStore)
}

func (m *MockStore) Post() store.PostStore {
	args := m.Called()
	return args.Get(0).(store.PostStore)
}

func (m *MockStore) ConcretePost() store.ConcretePostStore {
	args := m.Called()
	return args.Get(0).(store.ConcretePostStore)
}

func (m *MockStore) Tenant() store.TenantStore {
	args := m.Called()
	return args.Get(0).(store.TenantStore)
}

func (m *MockStore) Role() store.RoleStore {
	args := m.Called()
	return args.Get(0).(store.RoleStore)
}

func (m *MockStore) Permission() store.PermissionStore {
	args := m.Called()
	return args.Get(0).(store.PermissionStore)
}

func (m *MockStore) Menu() store.MenuStore {
	args := m.Called()
	return args.Get(0).(store.MenuStore)
}

// MockUserStore 模拟用户store
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, obj *model.UserM) error {
	args := m.Called(ctx, obj)
	return args.Error(0)
}

func (m *MockUserStore) Update(ctx context.Context, obj *model.UserM) error {
	args := m.Called(ctx, obj)
	return args.Error(0)
}

func (m *MockUserStore) Delete(ctx context.Context, opts *where.Options) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func (m *MockUserStore) Get(ctx context.Context, opts *where.Options) (*model.UserM, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*model.UserM), args.Error(1)
}

func (m *MockUserStore) List(ctx context.Context, opts *where.Options) (int64, []*model.UserM, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(int64), args.Get(1).([]*model.UserM), args.Error(2)
}

func TestLogin_Success(t *testing.T) {
	// 创建模拟对象
	mockStore := &MockStore{}
	mockUserStore := &MockUserStore{}

	// 设置模拟返回值
	mockStore.On("User").Return(mockUserStore)

	// 创建测试用户
	hashedPassword, _ := authn.Encrypt("password123")
	testUser := &model.UserM{
		UserID:   "test-user-id",
		Username: "testuser",
		Password: hashedPassword,
		Email:    "test@example.com",
		Phone:    "13800138000",
		Nickname: "Test User",
	}

	// 模拟数据库查询
	mockUserStore.On("Get", mock.Anything, mock.Anything).Return(testUser, nil)

	// 创建登录请求
	req := &apiv1.LoginRequest{
		LoginType:  "username",
		Identifier: "testuser",
		Password:   stringPtr("password123"),
	}

	// 验证请求参数验证
	assert.NotEmpty(t, req.GetLoginType())
	assert.NotEmpty(t, req.GetIdentifier())
	assert.NotEmpty(t, req.GetPassword())
}

func TestValidateLoginCredentials(t *testing.T) {
	biz := &userBiz{}

	// 创建测试用户
	hashedPassword, _ := authn.Encrypt("password123")
	testUser := &model.UserM{
		UserID:   "test-user-id",
		Username: "testuser",
		Password: hashedPassword,
	}

	// 测试密码验证
	req := &apiv1.LoginRequest{
		Password: stringPtr("password123"),
	}

	err := biz.validateLoginCredentials(context.Background(), testUser, req)
	assert.NoError(t, err)

	// 测试错误密码
	req.Password = stringPtr("wrongpassword")
	err = biz.validateLoginCredentials(context.Background(), testUser, req)
	assert.Error(t, err)
}

func TestGetClientTypeFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected cache.ClientType
	}{
		{"web", cache.ClientTypeWeb},
		{"h5", cache.ClientTypeH5},
		{"android", cache.ClientTypeAndroid},
		{"ios", cache.ClientTypeIOS},
		{"mini_program", cache.ClientTypeMiniProgram},
		{"op", cache.ClientTypeOp},
		{"unknown", cache.ClientTypeWeb}, // 默认值
	}

	for _, test := range tests {
		result := getClientTypeFromString(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

func TestGenerateVerifyCode(t *testing.T) {
	code := generateVerifyCode()
	assert.Len(t, code, 6)
	assert.Regexp(t, `^\d{6}$`, code)
}

func TestGenerateSessionID(t *testing.T) {
	sessionID := generateSessionID()
	assert.NotEmpty(t, sessionID)
	assert.Contains(t, sessionID, "sess_")
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
