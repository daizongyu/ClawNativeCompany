// Package handler 提供 HTTP 请求处理
// Agent Handler 处理 Agent 相关请求
package handler

import (
	"net/http"

	"claw/internal/service"
	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AgentHandler Agent Handler
type AgentHandler struct {
	agentService *service.AgentService
}

// NewAgentHandler 创建 Agent Handler
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		agentService: service.NewAgentService(),
	}
}

// InboundTask 接收 Agent 提交的任务结果
// POST /api/v1/agent/tasks/:id/complete
func (h *AgentHandler) InboundTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	var req struct {
		Success bool                   `json:"success" binding:"required"`
		Result  map[string]interface{} `json:"result"`
		Error   string                 `json:"error"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 从上下文获取 Agent ID
	agentID := c.GetString("agent_id")
	if agentID == "" {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	err := h.agentService.CompleteTask(c.Request.Context(), taskID, agentID, req.Success, req.Result, req.Error)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			utils.Error(c, http.StatusNotFound, "任务不存在")
		case service.ErrNotTaskAssignee:
			utils.Error(c, http.StatusForbidden, "不是任务负责人")
		default:
			utils.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "任务结果已提交"})
}

// GetMyTasks Agent 获取分配给自己的任务
// GET /api/v1/agent/tasks
func (h *AgentHandler) GetMyTasks(c *gin.Context) {
	agentID := c.GetString("agent_id")
	if agentID == "" {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	tasks, err := h.agentService.GetAgentTasks(c.Request.Context(), agentID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, tasks)
}

// GetTaskDetail Agent 获取任务详情
// GET /api/v1/agent/tasks/:id
func (h *AgentHandler) GetTaskDetail(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.ValidationError(c, "任务ID不能为空")
		return
	}

	agentID := c.GetString("agent_id")
	if agentID == "" {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	task, err := h.agentService.GetTaskDetail(c.Request.Context(), taskID, agentID)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			utils.Error(c, http.StatusNotFound, "任务不存在")
		case service.ErrNotTaskAssignee:
			utils.Error(c, http.StatusForbidden, "无权访问此任务")
		default:
			utils.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.SuccessWithData(c, task)
}

// UpdateStatus Agent 更新在线状态
// POST /api/v1/agent/heartbeat
func (h *AgentHandler) UpdateStatus(c *gin.Context) {
	agentID := c.GetString("agent_id")
	if agentID == "" {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=online busy offline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.agentService.UpdateAgentStatus(c.Request.Context(), agentID, req.Status); err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "状态已更新"})
}

// GetAgentInfo 获取 Agent 信息
// GET /api/v1/agent/info
func (h *AgentHandler) GetAgentInfo(c *gin.Context) {
	agentID := c.GetString("agent_id")
	if agentID == "" {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	info, err := h.agentService.GetAgentInfo(c.Request.Context(), agentID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, info)
}

// SendMessage Agent 发送消息到频道
// POST /api/v1/agent/messages
func (h *AgentHandler) SendMessage(c *gin.Context) {
	agentID := c.GetString("agent_id")
	if agentID == "" {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req struct {
		ChannelID string `json:"channel_id" binding:"required"`
		Content   string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	msg, err := h.agentService.SendMessage(c.Request.Context(), agentID, req.ChannelID, req.Content)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, msg)
}

// RegisterRoutes 注册路由
func (h *AgentHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Agent API 使用 API Key 认证
	agent := r.Group("/agent")
	{
		agent.GET("/tasks", h.GetMyTasks)
		agent.GET("/tasks/:id", h.GetTaskDetail)
		agent.POST("/tasks/:id/complete", h.InboundTask)
		agent.POST("/heartbeat", h.UpdateStatus)
		agent.GET("/info", h.GetAgentInfo)
		agent.POST("/messages", h.SendMessage)
	}
}
