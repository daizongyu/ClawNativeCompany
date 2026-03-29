// Package handler 提供 HTTP 请求处理
package handler

import (
	"runtime"
	"time"

	"claw/internal/database"
	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Timestamp int64                  `json:"timestamp"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

// HealthCheck 健康检查
// GET /api/v1/health
func HealthCheck(c *gin.Context) {
	// 执行各项检查
	checks := make(map[string]interface{})

	// 1. 数据库连接检查
	dbStatus := "ok"
	if sqlDB, err := database.GetDB().DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			dbStatus = "error: " + err.Error()
		}
	} else {
		dbStatus = "error: " + err.Error()
	}
	checks["database"] = dbStatus

	// 2. 系统资源检查
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	checks["memory"] = map[string]interface{}{
		"alloc_mb":      m.Alloc / 1024 / 1024,
		"sys_mb":        m.Sys / 1024 / 1024,
		"num_gc":        m.NumGC,
		"goroutines":    runtime.NumGoroutine(),
	}

	// 确定整体状态
	status := "healthy"
	if dbStatus != "ok" {
		status = "unhealthy"
	}

	utils.SuccessWithData(c, HealthResponse{
		Status:    status,
		Version:   "1.0.0",
		Timestamp: time.Now().Unix(),
		Checks:    checks,
	})
}

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", HealthCheck)
	r.GET("/ready", ReadinessCheck)
	r.GET("/live", LivenessCheck)
}

// ReadinessCheck 就绪检查（包含数据库连接检查）
// GET /api/v1/ready
func ReadinessCheck(c *gin.Context) {
	// 检查数据库连接
	db, err := database.GetDB().DB()
	if err != nil {
		c.JSON(503, gin.H{
			"ready":   false,
			"reason":  "database connection failed",
			"error":   err.Error(),
		})
		return
	}
	if err := db.Ping(); err != nil {
		c.JSON(503, gin.H{
			"ready":   false,
			"reason":  "database connection failed",
			"error":   err.Error(),
		})
		return
	}

	// 检查数据库连接池状态
	stats := db.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections {
		c.JSON(503, gin.H{
			"ready":   false,
			"reason":  "database connection pool exhausted",
			"stats":   stats,
		})
		return
	}

	utils.SuccessWithData(c, gin.H{
		"ready": true,
		"database": map[string]interface{}{
			"open_connections":    stats.OpenConnections,
			"in_use":              stats.InUse,
			"idle":                stats.Idle,
			"wait_count":          stats.WaitCount,
			"max_open_connections": stats.MaxOpenConnections,
		},
	})
}

// LivenessCheck 存活检查
// GET /api/v1/live
func LivenessCheck(c *gin.Context) {
	// 基本存活检查 - 只要能响应就说明还活着
	utils.SuccessWithData(c, gin.H{
		"alive":     true,
		"timestamp": time.Now().Unix(),
	})
}
