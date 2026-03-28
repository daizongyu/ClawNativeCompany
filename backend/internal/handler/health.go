// Package handler 提供 HTTP 请求处理
package handler

import (
	"time"

	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

// HealthCheck 健康检查
// GET /api/v1/health
func HealthCheck(c *gin.Context) {
	utils.SuccessWithData(c, HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: time.Now().Unix(),
	})
}

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", HealthCheck)
}

// ReadinessCheck 就绪检查（包含数据库连接检查）
// GET /api/v1/ready
func ReadinessCheck(c *gin.Context) {
	// TODO: 检查数据库连接
	utils.SuccessWithData(c, gin.H{
		"ready": true,
	})
}

// LivenessCheck 存活检查
// GET /api/v1/live
func LivenessCheck(c *gin.Context) {
	utils.SuccessWithData(c, gin.H{
		"alive": true,
	})
}
