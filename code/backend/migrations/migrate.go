// Package migrations 提供数据库迁移功能
package migrations

import (
	"fmt"

	"claw/internal/logger"
	"claw/internal/model"

	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
// 使用 GORM AutoMigrate 自动同步表结构
func Migrate(db *gorm.DB) error {
	logger.Info("开始数据库迁移")

	// 自动迁移所有模型
	models := []interface{}{
		&model.Employee{},
		&model.Channel{},
		&model.ChannelMember{},
		&model.Message{},
		&model.Workflow{},
		&model.WorkflowExecution{},
		&model.Task{},
	}

	for _, mdl := range models {
		if err := db.AutoMigrate(mdl); err != nil {
			return fmt.Errorf("迁移 %T 失败: %w", mdl, err)
		}
	}

	// 修复：为已有数据设置默认 type
	if err := db.Exec("UPDATE employees SET type = 'human' WHERE type IS NULL OR type = ''").Error; err != nil {
		logger.Warn("修复员工类型失败", "error", err)
	}

	// 修复：为已有数据设置默认 skills
	if err := db.Exec("UPDATE employees SET skills = '[]' WHERE skills IS NULL").Error; err != nil {
		logger.Warn("修复员工技能失败", "error", err)
	}

	// 修复：为频道成员设置默认 role
	if err := db.Exec("UPDATE channel_members SET role = 'member' WHERE role IS NULL OR role = ''").Error; err != nil {
		logger.Warn("修复频道成员角色失败", "error", err)
	}

	// 插入默认管理员账号
	if err := createDefaultAdmin(db); err != nil {
		logger.Warn("创建默认管理员失败", "error", err)
		// 不返回错误，继续启动
	}

	logger.Info("数据库迁移完成")
	return nil
}

// createDefaultAdmin 创建默认管理员账号
func createDefaultAdmin(db *gorm.DB) error {
	// 检查是否已存在管理员
	var count int64
	if err := db.Model(&model.Employee{}).Where("type = ?", model.EmployeeTypeHuman).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info("人类员工已存在，跳过创建默认管理员")
		return nil
	}

	// 创建默认管理员
	// 密码: admin123 (bcrypt hash)
	admin := &model.Employee{
		Name:     "管理员",
		Type:     model.EmployeeTypeHuman,
		Email:    "admin@claw.local",
		Password: "JDJhJDEwJGdIVGRXUDhoQmZLYjVGWE9PUVB2YmVjcy41ekxqcWVuQTU5OVpycjhidTJzT3VydDYzbVoy",
		Skills:   `["系统管理"]`,
		Status:   model.EmployeeStatusActive,
	}

	if err := db.Create(admin).Error; err != nil {
		return err
	}

	logger.Info("创建默认管理员账号成功", "email", admin.Email, "id", admin.ID)
	return nil
}

// Reset 重置数据库（删除所有表）
func Reset(db *gorm.DB) error {
	logger.Warn("重置数据库，所有数据将被删除")

	// 删除所有表
	tables := []string{
		"tasks",
		"workflow_executions",
		"workflows",
		"messages",
		"channel_members",
		"channels",
		"employees",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			logger.Error("删除表失败", "table", table, "error", err)
		}
	}

	logger.Info("数据库已重置")
	return nil
}
