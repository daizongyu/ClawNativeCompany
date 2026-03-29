// Package database 提供数据库连接管理
// 使用 GORM + 纯 Go SQLite (支持 Windows/Linux/Mac 无需 GCC)
package database

import (
	"fmt"
	"time"

	"claw/internal/config"
	"claw/internal/logger"

	"github.com/glebarez/sqlite" // 纯 Go SQLite 实现，无需 CGO
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DB 全局数据库实例
var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.Config) error {
	var err error

	// 打开数据库连接
	DB, err = gorm.Open(sqlite.Open(cfg.DSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // 使用复数表名
		},
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	// 获取底层 SQL DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层 DB 失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("数据库连接成功",
		"path", cfg.Database.Path,
		"max_open", cfg.Database.MaxOpenConns,
		"max_idle", cfg.Database.MaxIdleConns,
	)

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	if DB == nil {
		panic("数据库未初始化，请先调用 Init")
	}
	return DB
}

// WithTx 执行事务
func WithTx(fn func(tx *gorm.DB) error) error {
	return GetDB().Transaction(fn)
}

// AutoMigrate 自动迁移模型
func AutoMigrate(models ...interface{}) error {
	return GetDB().AutoMigrate(models...)
}

// IsRecordNotFound 检查是否为记录不存在错误
func IsRecordNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// Paginate 分页查询
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
