// Package handler 提供 HTTP 请求处理
// 消息 Handler 处理消息相关请求
package handler

import (
	"net/http"
	"time"

	"claw/internal/service"
	"claw/pkg/utils"
	"claw/pkg/validator"

	"github.com/gin-gonic/gin"
)

// MessageHandler 消息 Handler
type MessageHandler struct {
	messageService *service.MessageService
}

// NewMessageHandler 创建消息 Handler
func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		messageService: service.NewMessageService(),
	}
}

// Send 发送消息
// POST /api/v1/messages
func (h *MessageHandler) Send(c *gin.Context) {
	var req service.SendMessageRequest
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
	senderID := utils.GetEmployeeID(c)
	if senderID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := h.messageService.Send(c.Request.Context(), &req, senderID)
	if err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		case service.ErrNotChannelMember:
			utils.Error(c, http.StatusForbidden, "不是频道成员")
		case service.ErrReadonlyCannotSend:
			utils.Error(c, http.StatusForbidden, "只读成员不能发送消息")
		case service.ErrInvalidMessageType:
			utils.ValidationError(c, "无效的消息类型")
		case service.ErrMessageNotFound:
			utils.Error(c, http.StatusNotFound, "回复的消息不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "发送消息失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// ListAll 获取所有消息列表
// GET /api/v1/messages
func (h *MessageHandler) ListAll(c *gin.Context) {
	channelID := c.Query("channel_id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	limit := utils.GetIntQuery(c, "limit", 50)

	req := &service.ListMessageRequest{
		ChannelID: channelID,
		Limit:     limit,
	}

	resp, err := h.messageService.List(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取消息列表失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Get 获取消息详情
// GET /api/v1/messages/:id
func (h *MessageHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "消息 ID 不能为空")
		return
	}

	resp, err := h.messageService.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			utils.Error(c, http.StatusNotFound, "消息不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取消息失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// List 获取消息列表
// GET /api/v1/channels/:id/messages
func (h *MessageHandler) List(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	var req service.ListMessageRequest
	req.ChannelID = channelID

	// 解析 limit
	req.Limit = utils.GetIntQuery(c, "limit", 50)

	// 解析 before 时间戳
	beforeStr := c.Query("before")
	if beforeStr != "" {
		before, err := time.Parse(time.RFC3339, beforeStr)
		if err == nil {
			req.Before = &before
		}
	}

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	resp, err := h.messageService.List(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取消息列表失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Update 更新消息
// PUT /api/v1/messages/:id
func (h *MessageHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "消息 ID 不能为空")
		return
	}

	var req service.UpdateMessageRequest
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
	senderID := utils.GetEmployeeID(c)
	if senderID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := h.messageService.Update(c.Request.Context(), id, &req, senderID)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			utils.Error(c, http.StatusNotFound, "消息不存在")
		case service.ErrPermissionDenied:
			utils.Error(c, http.StatusForbidden, "只能更新自己的消息")
		default:
			utils.Error(c, http.StatusInternalServerError, "更新消息失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Delete 删除消息
// DELETE /api/v1/messages/:id
func (h *MessageHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "消息 ID 不能为空")
		return
	}

	// 获取当前用户 ID
	senderID := utils.GetEmployeeID(c)
	if senderID == "" {
		utils.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 检查是否是管理员（简化处理，实际应该从上下文获取角色）
	isAdmin := false
	if role, exists := c.Get("role"); exists {
		if r, ok := role.(string); ok && r == "admin" {
			isAdmin = true
		}
	}

	err := h.messageService.Delete(c.Request.Context(), id, senderID, isAdmin)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			utils.Error(c, http.StatusNotFound, "消息不存在")
		case service.ErrPermissionDenied:
			utils.Error(c, http.StatusForbidden, "只能删除自己的消息")
		default:
			utils.Error(c, http.StatusInternalServerError, "删除消息失败")
		}
		return
	}

	utils.Success(c)
}

// Search 搜索消息
// GET /api/v1/channels/:id/messages/search
func (h *MessageHandler) Search(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		utils.ValidationError(c, "频道 ID 不能为空")
		return
	}

	var req service.SearchMessageRequest
	req.ChannelID = channelID
	req.Keyword = c.Query("keyword")
	req.Page = utils.GetIntQuery(c, "page", 1)
	req.PageSize = utils.GetIntQuery(c, "page_size", 20)

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	resp, err := h.messageService.Search(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "搜索消息失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// GetThread 获取回复线程
// GET /api/v1/messages/:id/thread
func (h *MessageHandler) GetThread(c *gin.Context) {
	parentID := c.Param("id")
	if parentID == "" {
		utils.ValidationError(c, "消息 ID 不能为空")
		return
	}

	resp, err := h.messageService.GetThread(c.Request.Context(), parentID)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			utils.Error(c, http.StatusNotFound, "消息不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取回复线程失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// RegisterRoutes 注册路由
func (h *MessageHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 消息相关路由
	messages := r.Group("/messages")
	{
		messages.GET("", h.ListAll)
		messages.POST("", h.Send)
		messages.GET("/:id", h.Get)
		messages.PUT("/:id", h.Update)
		messages.DELETE("/:id", h.Delete)
		messages.GET("/:id/thread", h.GetThread)
	}

	// 频道消息路由
	// 频道消息路由（避免使用 Group，确保路由顺序正确）
	r.GET("/channels/:id/messages", h.List)
	r.GET("/channels/:id/messages/search", h.Search)
}
