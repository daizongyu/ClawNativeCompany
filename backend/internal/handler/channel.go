// Package handler 提供 HTTP 请求处理
// 频道 Handler 处理频道管理相关请求
package handler

import (
	"net/http"

	"claw/internal/model"
	"claw/internal/service"
	"claw/pkg/utils"
	"claw/pkg/validator"

	"github.com/gin-gonic/gin"
)

// ChannelHandler 频道 Handler
type ChannelHandler struct {
	channelService *service.ChannelService
}

// NewChannelHandler 创建频道 Handler
func NewChannelHandler() *ChannelHandler {
	return &ChannelHandler{
		channelService: service.NewChannelService(),
	}
}

// Create 创建频道
// POST /api/v1/channels
func (h *ChannelHandler) Create(c *gin.Context) {
	var req service.CreateChannelRequest
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

	resp, err := h.channelService.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		switch err {
		case service.ErrInvalidChannelType:
			utils.ValidationError(c, "无效的频道类型")
		default:
			utils.Error(c, http.StatusInternalServerError, "创建频道失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Get 获取频道详情
// GET /api/v1/channels/:id
func (h *ChannelHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	resp, err := h.channelService.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取频道失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// List 获取频道列表
// GET /api/v1/channels
func (h *ChannelHandler) List(c *gin.Context) {
	var req service.ListChannelRequest

	// 绑定查询参数
	req.Page = utils.GetIntQuery(c, "page", 1)
	req.PageSize = utils.GetIntQuery(c, "page_size", 20)
	req.Type = c.Query("type")
	req.Keyword = c.Query("keyword")

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	resp, err := h.channelService.List(c.Request.Context(), &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取频道列表失败")
		return
	}

	utils.SuccessWithData(c, resp)
}

// ListMyChannels 获取我加入的频道
// GET /api/v1/channels/my
func (h *ChannelHandler) ListMyChannels(c *gin.Context) {
	// 获取当前用户 ID
	employeeID := utils.GetEmployeeID(c)
	if employeeID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	page := utils.GetIntQuery(c, "page", 1)
	pageSize := utils.GetIntQuery(c, "page_size", 20)

	resp, err := h.channelService.ListByMember(c.Request.Context(), employeeID, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取频道列表失败")
		return
	}

	utils.SuccessWithData(c, resp)
}

// Update 更新频道
// PUT /api/v1/channels/:id
func (h *ChannelHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	var req service.UpdateChannelRequest
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

	// 检查权限（需要 admin 角色）
	employeeID := utils.GetEmployeeID(c)
	hasPermission, err := h.channelService.CheckPermission(c.Request.Context(), id, employeeID, model.ChannelRoleAdmin)
	if err != nil || !hasPermission {
		utils.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	resp, err := h.channelService.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "更新频道失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Delete 删除频道
// DELETE /api/v1/channels/:id
func (h *ChannelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	// 检查权限（需要 admin 角色）
	employeeID := utils.GetEmployeeID(c)
	hasPermission, err := h.channelService.CheckPermission(c.Request.Context(), id, employeeID, model.ChannelRoleAdmin)
	if err != nil || !hasPermission {
		utils.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	if err := h.channelService.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "删除频道失败")
		}
		return
	}

	utils.Success(c)
}

// AddMember 添加成员
// POST /api/v1/channels/:id/members
func (h *ChannelHandler) AddMember(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	var req service.AddMemberRequest
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

	// 检查权限（需要 admin 角色）
	employeeID := utils.GetEmployeeID(c)
	hasPermission, err := h.channelService.CheckPermission(c.Request.Context(), channelID, employeeID, model.ChannelRoleAdmin)
	if err != nil || !hasPermission {
		utils.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	resp, err := h.channelService.AddMember(c.Request.Context(), channelID, &req)
	if err != nil {
		switch err {
		case service.ErrAlreadyMember:
			utils.Error(c, http.StatusConflict, "已经是频道成员")
		case service.ErrInvalidRole:
			utils.ValidationError(c, "无效的角色")
		default:
			utils.Error(c, http.StatusInternalServerError, "添加成员失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// RemoveMember 移除成员
// DELETE /api/v1/channels/:id/members/:employee_id
func (h *ChannelHandler) RemoveMember(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	memberID := c.Param("employee_id")
	if memberID == "" {
		utils.ValidationError(c, "成员 ID 不能为空")
		return
	}

	// 检查权限（需要 admin 角色）
	employeeID := utils.GetEmployeeID(c)
	hasPermission, err := h.channelService.CheckPermission(c.Request.Context(), channelID, employeeID, model.ChannelRoleAdmin)
	if err != nil || !hasPermission {
		utils.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	err = h.channelService.RemoveMember(c.Request.Context(), channelID, memberID, employeeID)
	if err != nil {
		switch err {
		case service.ErrMemberNotFound:
			utils.Error(c, http.StatusNotFound, "成员不存在")
		case service.ErrCannotRemoveSelf:
			utils.Error(c, http.StatusForbidden, "不能移除自己")
		case service.ErrLastAdmin:
			utils.Error(c, http.StatusForbidden, "不能移除最后一个管理员")
		default:
			utils.Error(c, http.StatusInternalServerError, "移除成员失败")
		}
		return
	}

	utils.Success(c)
}

// UpdateMemberRole 更新成员角色
// PUT /api/v1/channels/:id/members/:employee_id/role
func (h *ChannelHandler) UpdateMemberRole(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	memberID := c.Param("employee_id")
	if memberID == "" {
		utils.ValidationError(c, "成员 ID 不能为空")
		return
	}

	var req service.UpdateMemberRoleRequest
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

	// 检查权限（需要 admin 角色）
	employeeID := utils.GetEmployeeID(c)
	hasPermission, err := h.channelService.CheckPermission(c.Request.Context(), channelID, employeeID, model.ChannelRoleAdmin)
	if err != nil || !hasPermission {
		utils.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	resp, err := h.channelService.UpdateMemberRole(c.Request.Context(), channelID, memberID, &req)
	if err != nil {
		switch err {
		case service.ErrMemberNotFound:
			utils.Error(c, http.StatusNotFound, "成员不存在")
		case service.ErrInvalidRole:
			utils.ValidationError(c, "无效的角色")
		default:
			utils.Error(c, http.StatusInternalServerError, "更新成员角色失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// ListMembers 获取成员列表
// GET /api/v1/channels/:id/members
func (h *ChannelHandler) ListMembers(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	resp, err := h.channelService.ListMembers(c.Request.Context(), channelID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取成员列表失败")
		return
	}

	utils.SuccessWithData(c, resp)
}

// GetMyRole 获取我在频道的角色
// GET /api/v1/channels/:id/my-role
func (h *ChannelHandler) GetMyRole(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	employeeID := utils.GetEmployeeID(c)
	if employeeID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	role, err := h.channelService.GetMemberRole(c.Request.Context(), channelID, employeeID)
	if err != nil {
		switch err {
		case service.ErrMemberNotFound:
			utils.Error(c, http.StatusNotFound, "不是频道成员")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取角色失败")
		}
		return
	}

	utils.SuccessWithData(c, gin.H{"role": role})
}

// RegisterRoutes 注册路由
func (h *ChannelHandler) RegisterRoutes(r *gin.RouterGroup) {
	channels := r.Group("/channels")
	{
		channels.POST("", h.Create)
		channels.GET("", h.List)
		channels.GET("/my", h.ListMyChannels)
		channels.GET("/:id", h.Get)
		channels.PUT("/:id", h.Update)
		channels.DELETE("/:id", h.Delete)
		channels.GET("/:id/members", h.ListMembers)
		channels.POST("/:id/members", h.AddMember)
		channels.DELETE("/:id/members/:employee_id", h.RemoveMember)
		channels.PUT("/:id/members/:employee_id/role", h.UpdateMemberRole)
		channels.GET("/:id/my-role", h.GetMyRole)
	}
}
