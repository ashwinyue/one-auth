// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/one-auth. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/ashwinyue/one-auth/pkg/db"
	"github.com/spf13/pflag"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// 帮助信息文本.
const helpText = `Usage: main [flags] arg [arg...]

This is a pflag example.

Flags:
`

// Querier 定义了数据库查询接口.
type Querier interface {
	// FilterWithNameAndRole 按名称和角色查询记录
	FilterWithNameAndRole(name string) ([]gen.T, error)
}

// GenerateConfig 保存代码生成的配置.
type GenerateConfig struct {
	ModelPackagePath string
	GenerateFunc     func(g *gen.Generator)
}

// 预定义的生成配置.
var generateConfigs = map[string]GenerateConfig{
	"mb": {ModelPackagePath: "../../internal/apiserver/model", GenerateFunc: GenerateMiniBlogModels},
}

// 命令行参数.
var (
	addr       = pflag.StringP("addr", "a", "127.0.0.1:3306", "MySQL host address.")
	username   = pflag.StringP("username", "u", "miniblog", "Username to connect to the database.")
	password   = pflag.StringP("password", "p", "miniblog1234", "Password to use when connecting to the database.")
	database   = pflag.StringP("db", "d", "miniblog", "Database name to connect to.")
	modelPath  = pflag.String("model-pkg-path", "", "Generated model code's package name.")
	components = pflag.StringSlice("component", []string{"mb"}, "Generated model code's for specified component.")
	help       = pflag.BoolP("help", "h", false, "Show this help message.")
)

func main() {
	// 设置自定义的使用说明函数
	pflag.Usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
	pflag.Parse()

	// 如果设置了帮助标志，则显示帮助信息并退出
	if *help {
		pflag.Usage()
		return
	}

	// 初始化数据库连接
	dbInstance, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 处理组件并生成代码
	for _, component := range *components {
		processComponent(component, dbInstance)
	}
}

// initializeDatabase 创建并返回一个数据库连接.
func initializeDatabase() (*gorm.DB, error) {
	dbOptions := &db.MySQLOptions{
		Addr:     *addr,
		Username: *username,
		Password: *password,
		Database: *database,
	}

	// 创建并返回数据库连接
	return db.NewMySQL(dbOptions)
}

// processComponent 处理单个组件以生成代码.
func processComponent(component string, dbInstance *gorm.DB) {
	config, ok := generateConfigs[component]
	if !ok {
		log.Printf("Component '%s' not found in configuration. Skipping.", component)
		return
	}

	// 解析模型包路径
	modelPkgPath := resolveModelPackagePath(config.ModelPackagePath)

	// 创建生成器实例
	generator := createGenerator(modelPkgPath)
	generator.UseDB(dbInstance)

	// 应用自定义生成器选项
	applyGeneratorOptions(generator)

	// 使用指定的函数生成模型
	config.GenerateFunc(generator)

	// 执行代码生成
	generator.Execute()
}

// resolveModelPackagePath 确定模型生成的包路径.
func resolveModelPackagePath(defaultPath string) string {
	if *modelPath != "" {
		return *modelPath
	}
	absPath, err := filepath.Abs(defaultPath)
	if err != nil {
		log.Printf("Error resolving path: %v", err)
		return defaultPath
	}
	return absPath
}

// createGenerator 初始化并返回一个新的生成器实例.
func createGenerator(packagePath string) *gen.Generator {
	return gen.NewGenerator(gen.Config{
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		ModelPkgPath:      packagePath,
		WithUnitTest:      true,
		FieldNullable:     true,  // 对于数据库中可空的字段，使用指针类型。
		FieldSignable:     false, // 禁用无符号属性以提高兼容性。
		FieldWithIndexTag: false, // 不包含 GORM 的索引标签。
		FieldWithTypeTag:  false, // 不包含 GORM 的类型标签。
	})
}

// applyGeneratorOptions 设置自定义生成器选项.
func applyGeneratorOptions(g *gen.Generator) {
	// 为特定字段自定义 GORM 标签
	g.WithOpts(
		gen.FieldGORMTag("created_at", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
		gen.FieldGORMTag("updated_at", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
	)
}

// GenerateMiniBlogModels 为 miniblog 组件生成模型.
func GenerateMiniBlogModels(g *gen.Generator) {
	// 基础表
	g.GenerateModelAs(
		"user",
		"UserM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("username", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_username")
			return tag
		}),
		gen.FieldGORMTag("phone", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_phone")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	g.GenerateModelAs(
		"post",
		"PostM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("post_id", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_post_post_id")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	// 租户管理表
	g.GenerateModelAs(
		"tenants",
		"TenantM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("tenant_code", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_tenant_code")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	// 角色管理表
	g.GenerateModelAs(
		"roles",
		"RoleM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("role_code", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_role_code_tenant")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	// 权限管理表（重构版）
	g.GenerateModelAs(
		"permissions",
		"PermissionM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("permission_code", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_permission_code_tenant")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	// 菜单管理表（重构版-纯UI结构）
	g.GenerateModelAs(
		"menus",
		"MenuM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("menu_code", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_menu_code_tenant")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	// 用户状态表（支持多认证方式）
	g.GenerateModelAs(
		"user_status",
		"UserStatusM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("auth_id", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_auth_id_type")
			return tag
		}),
		gen.FieldGORMTag("auth_type", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_auth_id_type")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	// 关联表
	g.GenerateModelAs(
		"user_tenants",
		"UserTenantM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("user_id", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_tenant")
			return tag
		}),
		gen.FieldGORMTag("tenant_id", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_tenant")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "")
			return tag
		}),
	)

	g.GenerateModelAs(
		"menu_permissions",
		"MenuPermissionM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("menu_id", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_menu_permission")
			return tag
		}),
		gen.FieldGORMTag("permission_id", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_menu_permission")
			return tag
		}),
	)
}
