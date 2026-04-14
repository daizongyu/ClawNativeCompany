// Package handler 提供 HTTP 请求处理
// 任务 Handler 处理任务管理相关请求
package handler

import (
	"context"
	"net/http"
	"strconv"

	"claw/internal/service"
	"claw/pkg/utils"
	"claw/pkg/validator"

	"github.com/gin-gonic/gin"
)

// TaskHandler 任务 Handler
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler 创建任务 Handler
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		taskService: service.NewTaskService(),
	}
}

// Create 创建任务
// POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	// 获取当前用户 ID
	createdBy := utils.GetEmployeeID(c)
	if createdBy == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 执行创建
	resp, err := h.taskService.Create(c.Request.Context(), req, createdBy)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// Get 获取任务详情
// GET /api/v1/tasks/:id
func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	resp, err := h.taskService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrTaskNotFound {
			utils.Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// List 获取任务列表
// GET /api/v1/tasks
func (h *TaskHandler) List(c *gin.Context) {
	// 获取查询参数
	page := utils.GetIntQuery(c, "page", 1)
	pageSize := utils.GetIntQuery(c, "page_size", 20)
	status := c.Query("status")
	priority := c.Query("priority")
	keyword := c.Query("keyword")
	mine := c.Query("mine") == "true"
	unclaimed := c.Query("unclaimed") == "true"

	// 将当前用户ID放入context（用于mine筛选）
	ctx := c.Request.Context()
	if userID, exists := c.Get("employee_id"); exists {
		if id, ok := userID.(string); ok {
			ctx = context.WithValue(ctx, "employee_id", id)
		}
	}

	req := service.ListTaskRequest{
		Page:      page,
		PageSize:  pageSize,
		Status:    status,
		Priority:  priority,
		Keyword:   keyword,
		Mine:      mine,
		Unclaimed: unclaimed,
	}

	resp, err := h.taskService.List(ctx, req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// Update 更新任务
// PUT /api/v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	resp, err := h.taskService.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == service.ErrTaskNotFound {
			utils.Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// Delete 删除任务
// DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	if err := h.taskService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrTaskNotFound {
			utils.Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "删除成功"})
}

// Search 搜索任务
// GET /api/v1/tasks/search
func (h *TaskHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	priority := c.Query("priority")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var statusPtr, priorityPtr *string
	if status != "" {
		statusPtr = &status
	}
	if priority != "" {
		priorityPtr = &priority
	}

	resp, total, err := h.taskService.Search(c.Request.Context(), keyword, statusPtr, priorityPtr, page, pageSize)
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

// GetMyTasks 获取我的任务
// GET /api/v1/tasks/my
func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	employeeID := utils.GetEmployeeID(c)
	if employeeID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
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

	resp, total, err := h.taskService.GetMyTasks(c.Request.Context(), employeeID, page, pageSize)
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

// GetUnclaimedTasks 获取未认领任务池
// GET /api/v1/tasks/unclaimed
func (h *TaskHandler) GetUnclaimedTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, total, err := h.taskService.ListUnclaimed(c.Request.Context(), page, pageSize)
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

// ClaimTask 认领任务
// POST /api/v1/tasks/:id/claim
func (h *TaskHandler) ClaimTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	employeeID := utils.GetEmployeeID(c)
	if employeeID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := h.taskService.ClaimTask(c.Request.Context(), taskID, employeeID)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			utils.Error(c, http.StatusNotFound, "任务不存在")
		case service.ErrTaskAlreadyClaimed:
			utils.Error(c, http.StatusConflict, "任务已被认领")
		case service.ErrTaskNotClaimable:
			utils.Error(c, http.StatusBadRequest, "任务不可认领")
		default:
			utils.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// AssignTask 指派任务
// POST /api/v1/tasks/:id/assign
func (h *TaskHandler) AssignTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.taskService.AssignTask(c.Request.Context(), taskID, req.AssigneeID)
	if err != nil {
		if err == service.ErrTaskNotFound {
			utils.Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// CompleteTask 完成任务
// POST /api/v1/tasks/:id/complete
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	employeeID := utils.GetEmployeeID(c)
	if employeeID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空结果
		req.Result = nil
	}

	resp, err := h.taskService.CompleteTask(c.Request.Context(), taskID, employeeID, req.Result)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			utils.Error(c, http.StatusNotFound, "任务不存在")
		case service.ErrNotTaskAssignee:
			utils.Error(c, http.StatusForbidden, "不是任务负责人")
		case service.ErrTaskAlreadyComplete:
			utils.Error(c, http.StatusBadRequest, "任务已完成")
		default:
			utils.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// CancelTask 取消任务
// POST /api/v1/tasks/:id/cancel
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	resp, err := h.taskService.CancelTask(c.Request.Context(), taskID)
	if err != nil {
		if err == service.ErrTaskNotFound {
			utils.Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		if err == service.ErrTaskAlreadyComplete {
			utils.Error(c, http.StatusBadRequest, "任务已完成，无法取消")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, resp)
}

// GetTaskStats 获取任务统计
// GET /api/v1/tasks/stats
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	employeeID := utils.GetEmployeeID(c)
	if employeeID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	stats, err := h.taskService.GetTaskStats(c.Request.Context(), employeeID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, stats)
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.POST("", h.Create)
		tasks.GET("", h.List)
		tasks.GET("/search", h.Search)
		tasks.GET("/my", h.GetMyTasks)
		tasks.GET("/unclaimed", h.GetUnclaimedTasks)
		tasks.GET("/stats", h.GetTaskStats)
		tasks.GET("/:id", h.Get)
		tasks.PUT("/:id", h.Update)
		tasks.DELETE("/:id", h.Delete)
		tasks.POST("/:id/claim", h.ClaimTask)
		tasks.POST("/:id/assign", h.AssignTask)
		tasks.POST("/:id/complete", h.CompleteTask)
		tasks.POST("/:id/cancel", h.CancelTask)
	}
}
