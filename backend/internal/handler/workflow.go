// Package handler 提供 HTTP 请求处理
// 工作流 Handler 处理工作流管理相关请求
package handler

import (
	"net/http"
	"strconv"

	"claw/internal/service"
	"claw/pkg/utils"
	"claw/pkg/validator"

	"github.com/gin-gonic/gin"
)

// WorkflowHandler 工作流 Handler
type WorkflowHandler struct {
	workflowService *service.WorkflowService
}

// NewWorkflowHandler 创建工作流 Handler
func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: service.NewWorkflowService(),
	}
}

// Create 创建工作流
// POST /api/v1/workflows
func (h *WorkflowHandler) Create(c *gin.Context) {
	var req service.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	createdBy := utils.GetEmployeeID(c)
	if createdBy == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := h.workflowService.Create(c.Request.Context(), req, createdBy)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// Get 获取工作流详情
// GET /api/v1/workflows/:id
func (h *WorkflowHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "工作流ID不能为空")
		return
	}

	resp, err := h.workflowService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrWorkflowNotFound {
			utils.Error(c, http.StatusNotFound, "工作流不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// Update 更新工作流
// PUT /api/v1/workflows/:id
func (h *WorkflowHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "工作流ID不能为空")
		return
	}

	var req service.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	resp, err := h.workflowService.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == service.ErrWorkflowNotFound {
			utils.Error(c, http.StatusNotFound, "工作流不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// Delete 删除工作流
// DELETE /api/v1/workflows/:id
func (h *WorkflowHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "工作流ID不能为空")
		return
	}

	if err := h.workflowService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrWorkflowNotFound {
			utils.Error(c, http.StatusNotFound, "工作流不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "删除成功"})
}

// List 获取工作流列表
// GET /api/v1/workflows
func (h *WorkflowHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, total, err := h.workflowService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{
		"list":      resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Search 搜索工作流
// GET /api/v1/workflows/search
func (h *WorkflowHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, total, err := h.workflowService.Search(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{
		"list":      resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateStatus 更新工作流状态
// PUT /api/v1/workflows/:id/status
func (h *WorkflowHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "工作流ID不能为空")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.workflowService.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		if err == service.ErrWorkflowNotFound {
			utils.Error(c, http.StatusNotFound, "工作流不存在")
			return
		}
		if err == service.ErrInvalidStatus {
			utils.ValidationError(c, "无效的状态值")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// Trigger 触发工作流
// POST /api/v1/workflows/:id/trigger
func (h *WorkflowHandler) Trigger(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "工作流ID不能为空")
		return
	}

	var req struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空数据
		req.Data = nil
	}

	triggeredBy := utils.GetEmployeeID(c)
	if triggeredBy == "" {
		triggeredBy = "system"
	}

	resp, err := h.workflowService.TriggerWorkflow(c.Request.Context(), id, triggeredBy, req.Data)
	if err != nil {
		if err == service.ErrWorkflowNotFound {
			utils.Error(c, http.StatusNotFound, "工作流不存在")
			return
		}
		if err == service.ErrWorkflowInactive {
			utils.Error(c, http.StatusBadRequest, "工作流未激活")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// GetExecution 获取工作流执行详情
// GET /api/v1/workflow-executions/:id
func (h *WorkflowHandler) GetExecution(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "执行ID不能为空")
		return
	}

	resp, err := h.workflowService.GetExecution(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrExecutionNotFound {
			utils.Error(c, http.StatusNotFound, "执行不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// ListExecutions 获取工作流执行列表
// GET /api/v1/workflows/:id/executions
func (h *WorkflowHandler) ListExecutions(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		utils.ValidationError(c, "工作流ID不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, total, err := h.workflowService.ListExecutions(c.Request.Context(), workflowID, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{
		"list":      resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListActiveExecutions 获取活跃执行列表
// GET /api/v1/workflow-executions/active
func (h *WorkflowHandler) ListActiveExecutions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, total, err := h.workflowService.ListActiveExecutions(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{
		"list":      resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CancelExecution 取消执行
// PUT /api/v1/workflow-executions/:id/cancel
func (h *WorkflowHandler) CancelExecution(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "执行ID不能为空")
		return
	}

	if err := h.workflowService.CancelExecution(c.Request.Context(), id); err != nil {
		if err == service.ErrExecutionNotFound {
			utils.Error(c, http.StatusNotFound, "执行不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "已取消"})
}

// RegisterRoutes 注册路由
func (h *WorkflowHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 工作流管理
	workflows := r.Group("/workflows")
	{
		workflows.POST("", h.Create)
		workflows.GET("", h.List)
		workflows.GET("/search", h.Search)
		workflows.GET("/:id", h.Get)
		workflows.PUT("/:id", h.Update)
		workflows.DELETE("/:id", h.Delete)
		workflows.PUT("/:id/status", h.UpdateStatus)
		workflows.POST("/:id/trigger", h.Trigger)
		workflows.GET("/:id/executions", h.ListExecutions)
	}

	// 执行管理
	executions := r.Group("/workflow-executions")
	{
		executions.GET("/active", h.ListActiveExecutions)
		executions.GET("/:id", h.GetExecution)
		executions.PUT("/:id/cancel", h.CancelExecution)
	}
}
