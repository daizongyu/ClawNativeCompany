// Package migrations 提供数据库迁移功能
package migrations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
		&model.GatewayConfig{}, // 新增 GatewayConfig 表
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

	// 【用户模块优化】迁移：为已有员工生成 username
	if err := migrateEmployeeUsername(db); err != nil {
		logger.Warn("迁移员工用户名失败", "error", err)
	}

	// 【用户模块优化】迁移：为已有员工设置 display_name
	if err := migrateEmployeeDisplayName(db); err != nil {
		logger.Warn("迁移员工显示名失败", "error", err)
	}

	// 【用户模块优化】迁移：为已有员工设置默认通知偏好
	if err := migrateEmployeeNotificationPrefs(db); err != nil {
		logger.Warn("迁移员工通知偏好失败", "error", err)
	}

	// 插入默认管理员账号
	if err := createDefaultAdmin(db); err != nil {
		logger.Warn("创建默认管理员失败", "error", err)
		// 不返回错误，继续启动
	}

	logger.Info("数据库迁移完成")
	return nil
}

// migrateEmployeeUsername 为已有员工生成 username
func migrateEmployeeUsername(db *gorm.DB) error {
	// 检查是否已有 username
	var count int64
	if err := db.Model(&model.Employee{}).Where("username IS NOT NULL AND username != ''").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info("员工 username 已存在，跳过迁移")
		return nil
	}

	// 为所有员工生成 username（基于 email 前缀）
	var employees []model.Employee
	if err := db.Find(&employees).Error; err != nil {
		return err
	}

	for _, emp := range employees {
		username := generateUsernameFromEmail(emp.Email)
		// 确保唯一性
		var existingCount int64
		db.Model(&model.Employee{}).Where("username = ?", username).Count(&existingCount)
		if existingCount > 0 {
			username = fmt.Sprintf("%s_%s", username, emp.ID[:4])
		}

		if err := db.Model(&emp).Update("username", username).Error; err != nil {
			logger.Warn("更新员工 username 失败", "employee_id", emp.ID, "error", err)
		}
	}

	logger.Info("员工 username 迁移完成", "count", len(employees))
	return nil
}

// generateUsernameFromEmail 从 email 生成 username
func generateUsernameFromEmail(email string) string {
	// 提取 @ 前面的部分
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		username := parts[0]
		// 只保留字母数字和下划线
		username = strings.ToLower(username)
		var result strings.Builder
		for _, ch := range username {
			if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
				result.WriteRune(ch)
			}
		}
		if result.Len() > 0 {
			return result.String()
		}
	}
	return "user_" + generateRandomString(6)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// migrateEmployeeDisplayName 为已有员工设置 display_name
func migrateEmployeeDisplayName(db *gorm.DB) error {
	// 检查是否已有 display_name
	var count int64
	if err := db.Model(&model.Employee{}).Where("display_name IS NOT NULL AND display_name != ''").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info("员工 display_name 已存在，跳过迁移")
		return nil
	}

	// 将 name 复制到 display_name
	if err := db.Exec("UPDATE employees SET display_name = name WHERE display_name IS NULL OR display_name = ''").Error; err != nil {
		return err
	}

	logger.Info("员工 display_name 迁移完成")
	return nil
}

// migrateEmployeeNotificationPrefs 为已有员工设置默认通知偏好
func migrateEmployeeNotificationPrefs(db *gorm.DB) error {
	// 检查是否已有 notification_prefs
	var count int64
	if err := db.Model(&model.Employee{}).Where("notification_prefs IS NOT NULL AND notification_prefs != ''").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info("员工 notification_prefs 已存在，跳过迁移")
		return nil
	}

	// 设置默认通知偏好
	defaultPrefs := model.DefaultNotificationPreferences()
	prefsJSON, _ := json.Marshal(defaultPrefs)

	if err := db.Exec("UPDATE employees SET notification_prefs = ? WHERE notification_prefs IS NULL OR notification_prefs = ''", string(prefsJSON)).Error; err != nil {
		return err
	}

	logger.Info("员工 notification_prefs 迁移完成")
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
