// Package handler 提供 HTTP 请求处理
// Dashboard Handler 处理仪表盘统计相关请求
package handler

import (
	"net/http"

	"claw/internal/service"
	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 仪表盘 Handler
type DashboardHandler struct {
	dashboardService service.DashboardService
}

// NewDashboardHandler 创建 Dashboard Handler
func NewDashboardHandler(dashboardService service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetStats 获取仪表盘统计数据
// GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.dashboardService.GetStats(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取统计数据失败: "+err.Error())
		return
	}

	utils.SuccessWithData(c, stats)
}
