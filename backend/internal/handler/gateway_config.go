package handler

import (
	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/service"
	"claw/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GatewayConfigHandler Gateway 配置处理器
type GatewayConfigHandler struct {
	gatewayConfigService *service.GatewayConfigService
}

// NewGatewayConfigHandler 创建 Gateway 配置处理器
func NewGatewayConfigHandler(gatewayConfigService *service.GatewayConfigService) *GatewayConfigHandler {
	return &GatewayConfigHandler{
		gatewayConfigService: gatewayConfigService,
	}
}

// RegisterRoutes 注册路由
func (h *GatewayConfigHandler) RegisterRoutes(r *gin.RouterGroup) {
	gatewayConfigs := r.Group("/gateway-configs")
	{
		gatewayConfigs.POST("", h.Create)
		gatewayConfigs.GET("", h.List)
		gatewayConfigs.GET("/:id", h.GetByID)
		gatewayConfigs.PUT("/:id", h.Update)
		gatewayConfigs.DELETE("/:id", h.Delete)
		gatewayConfigs.POST("/:id/verify", h.Verify)
		gatewayConfigs.POST("/:id/test", h.SendTestMessage)
		gatewayConfigs.POST("/:id/default", h.SetDefault)
	}
}

// Create 创建 Gateway 配置
func (h *GatewayConfigHandler) Create(c *gin.Context) {
	employeeID, exists := c.Get("employee_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req service.CreateGatewayConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	config, err := h.gatewayConfigService.Create(c.Request.Context(), employeeID.(string), req)
	if err != nil {
		logger.Error("创建 Gateway 配置失败", "error", err)
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, config.ToResponse())
}

// List 列出 Gateway 配置
func (h *GatewayConfigHandler) List(c *gin.Context) {
	employeeID, exists := c.Get("employee_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取查询参数
	gatewayType := c.Query("type")
	status := c.Query("status")
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	var gatewayTypeEnum model.GatewayType
	if gatewayType != "" {
		gatewayTypeEnum = model.GatewayType(gatewayType)
	}

	var statusEnum model.GatewayStatus
	if status != "" {
		statusEnum = model.GatewayStatus(status)
	}

	configs, total, err := h.gatewayConfigService.List(
		employeeID.(string),
		gatewayTypeEnum,
		statusEnum,
		page,
		pageSize,
	)
	if err != nil {
		logger.Error("获取 Gateway 配置列表失败", "error", err)
		utils.Error(c, http.StatusInternalServerError, "获取列表失败")
		return
	}

	// 转换为响应格式
	configResponses := make([]model.GatewayConfigResponse, len(configs))
	for i, config := range configs {
		configResponses[i] = config.ToResponse()
	}

	utils.SuccessWithData(c, gin.H{
		"list": configResponses,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetByID 根据 ID 获取 Gateway 配置
func (h *GatewayConfigHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "配置ID不能为空")
		return
	}

	config, err := h.gatewayConfigService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, http.StatusNotFound, "配置不存在")
		return
	}

	utils.SuccessWithData(c, config.ToResponse())
}

// Update 更新 Gateway 配置
func (h *GatewayConfigHandler) Update(c *gin.Context) {
	employeeID, exists := c.Get("employee_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "配置ID不能为空")
		return
	}

	var req service.UpdateGatewayConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	config, err := h.gatewayConfigService.Update(c.Request.Context(), id, employeeID.(string), req)
	if err != nil {
		logger.Error("更新 Gateway 配置失败", "error", err)
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, config.ToResponse())
}

// Delete 删除 Gateway 配置
func (h *GatewayConfigHandler) Delete(c *gin.Context) {
	employeeID, exists := c.Get("employee_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "配置ID不能为空")
		return
	}

	if err := h.gatewayConfigService.Delete(c.Request.Context(), id, employeeID.(string)); err != nil {
		logger.Error("删除 Gateway 配置失败", "error", err)
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "删除成功"})
}

// Verify 验证 Gateway 配置
func (h *GatewayConfigHandler) Verify(c *gin.Context) {
	employeeID, exists := c.Get("employee_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "配置ID不能为空")
		return
	}

	if err := h.gatewayConfigService.Verify(c.Request.Context(), id, employeeID.(string)); err != nil {
		logger.Error("验证 Gateway 配置失败", "error", err)
		utils.ValidationError(c, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "验证成功"})
}

// SendTestMessage 发送测试消息
func (h *GatewayConfigHandler) SendTestMessage(c *gin.Context) {
	employeeID, exists := c.Get("employee_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "配置ID不能为空")
		return
	}

	if err := h.gatewayConfigService.SendTestMessage(c.Request.Context(), id, employeeID.(string)); err != nil {
		logger.Error("发送测试消息失败", "error", err)
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "测试消息已发送"})
}

// SetDefault 设置默认 Gateway 配置
func (h *GatewayConfigHandler) SetDefault(c *gin.Context) {
	employeeID, exists := c.Get("employee_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "配置ID不能为空")
		return
	}

	if err := h.gatewayConfigService.SetDefault(c.Request.Context(), id, employeeID.(string)); err != nil {
		logger.Error("设置默认 Gateway 配置失败", "error", err)
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "已设为默认配置"})
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var result int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, nil
		}
		result = result*10 + int(ch-'0')
	}
	return result, nil
}
