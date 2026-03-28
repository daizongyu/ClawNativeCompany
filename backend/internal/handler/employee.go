// Package handler 提供 HTTP 请求处理
// 员工 Handler 处理员工管理相关请求
package handler

import (
	"net/http"

	"claw/internal/middleware"
	"claw/internal/service"
	"claw/pkg/utils"
	"claw/pkg/validator"

	"github.com/gin-gonic/gin"
)

// EmployeeHandler 员工 Handler
type EmployeeHandler struct {
	employeeService *service.EmployeeService
}

// NewEmployeeHandler 创建员工 Handler
func NewEmployeeHandler() *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: service.NewEmployeeService(),
	}
}

// Create 创建员工
// POST /api/v1/employees
func (h *EmployeeHandler) Create(c *gin.Context) {
	var req service.CreateEmployeeRequest
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

	// 执行创建
	resp, err := h.employeeService.Create(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrEmployeeExists:
			utils.Error(c, http.StatusConflict, "邮箱已被注册")
		case service.ErrInvalidEmployeeType:
			utils.ValidationError(c, "无效的员工类型")
		default:
			utils.Error(c, http.StatusInternalServerError, "创建员工失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Get 获取员工详情
// GET /api/v1/employees/:id
func (h *EmployeeHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "员工 ID 不能为空")
		return
	}

	resp, err := h.employeeService.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrEmployeeNotFound:
			utils.Error(c, http.StatusNotFound, "员工不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取员工失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// List 获取员工列表
// GET /api/v1/employees
func (h *EmployeeHandler) List(c *gin.Context) {
	var req service.ListEmployeeRequest
	
	// 绑定查询参数
	req.Page = utils.GetIntQuery(c, "page", 1)
	req.PageSize = utils.GetIntQuery(c, "page_size", 20)
	req.Type = c.Query("type")
	req.Status = c.Query("status")

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	resp, err := h.employeeService.List(c.Request.Context(), &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取员工列表失败")
		return
	}

	utils.SuccessWithData(c, resp)
}

// Update 更新员工
// PUT /api/v1/employees/:id
func (h *EmployeeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "员工 ID 不能为空")
		return
	}

	var req service.UpdateEmployeeRequest
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

	resp, err := h.employeeService.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrEmployeeNotFound:
			utils.Error(c, http.StatusNotFound, "员工不存在")
		case service.ErrEmployeeExists:
			utils.Error(c, http.StatusConflict, "邮箱已被使用")
		default:
			utils.Error(c, http.StatusInternalServerError, "更新员工失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Delete 删除员工
// DELETE /api/v1/employees/:id
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "员工 ID 不能为空")
		return
	}

	// 获取当前用户 ID
	currentUserID, _ := c.Get(string(middleware.ContextKeyEmployeeID))
	currentUserIDStr, _ := currentUserID.(string)

	err := h.employeeService.Delete(c.Request.Context(), id, currentUserIDStr)
	if err != nil {
		switch err {
		case service.ErrEmployeeNotFound:
			utils.Error(c, http.StatusNotFound, "员工不存在")
		case service.ErrCannotDeleteSelf:
			utils.Error(c, http.StatusForbidden, "不能删除自己")
		default:
			utils.Error(c, http.StatusInternalServerError, "删除员工失败")
		}
		return
	}

	utils.Success(c)
}

// SearchBySkills 按技能搜索员工
// GET /api/v1/employees/search
func (h *EmployeeHandler) SearchBySkills(c *gin.Context) {
	var req service.SearchEmployeeRequest
	
	// 绑定查询参数
	req.Page = utils.GetIntQuery(c, "page", 1)
	req.PageSize = utils.GetIntQuery(c, "page_size", 20)
	
	// 解析 skills 数组
	skills := c.QueryArray("skills")
	if len(skills) == 0 {
		// 尝试从单个参数解析
		skillStr := c.Query("skills")
		if skillStr != "" {
			skills = []string{skillStr}
		}
	}
	req.Skills = skills

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	resp, err := h.employeeService.SearchBySkills(c.Request.Context(), &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "搜索员工失败")
		return
	}

	utils.SuccessWithData(c, resp)
}

// GenerateAPIKey 生成 API Key
// POST /api/v1/employees/:id/apikey
func (h *EmployeeHandler) GenerateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "员工 ID 不能为空")
		return
	}

	resp, err := h.employeeService.GenerateAPIKey(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrEmployeeNotFound:
			utils.Error(c, http.StatusNotFound, "员工不存在")
		default:
			if err.Error() == "只有 Agent 类型员工可以生成 API Key" {
				utils.Error(c, http.StatusBadRequest, err.Error())
			} else {
				utils.Error(c, http.StatusInternalServerError, "生成 API Key 失败")
			}
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// ResetAPIKey 重置 API Key
// PUT /api/v1/employees/:id/apikey
func (h *EmployeeHandler) ResetAPIKey(c *gin.Context) {
	// 复用生成逻辑
	h.GenerateAPIKey(c)
}

// RegisterRoutes 注册路由
func (h *EmployeeHandler) RegisterRoutes(r *gin.RouterGroup) {
	employees := r.Group("/employees")
	{
		employees.POST("", h.Create)
		employees.GET("", h.List)
		employees.GET("/search", h.SearchBySkills)
		employees.GET("/:id", h.Get)
		employees.PUT("/:id", h.Update)
		employees.DELETE("/:id", h.Delete)
		employees.POST("/:id/apikey", h.GenerateAPIKey)
		employees.PUT("/:id/apikey", h.ResetAPIKey)
	}
}
