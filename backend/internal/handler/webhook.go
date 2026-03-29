// Package handler 提供 HTTP 请求处理
// Webhook Handler 处理外部系统推送
package handler

import (
	"net/http"

	"claw/internal/service"
	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

// WebhookHandler Webhook Handler
type WebhookHandler struct {
	webhookService *service.WebhookService
}

// NewWebhookHandler 创建 Webhook Handler
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		webhookService: service.NewWebhookService(),
	}
}

// DingTalk 接收钉钉消息推送
// POST /webhooks/dingtalk
func (h *WebhookHandler) DingTalk(c *gin.Context) {
	var req struct {
		MsgType    string `json:"msgtype"`
		Text       struct {
			Content string `json:"content"`
		} `json:"text"`
		SenderStaffID string `json:"senderStaffId"`
		SenderNick    string `json:"senderNick"`
		ConversationID string `json:"conversationId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 处理钉钉消息
	err := h.webhookService.ProcessDingTalkMessage(c.Request.Context(), service.DingTalkMessage{
		MsgType:        req.MsgType,
		Content:        req.Text.Content,
		SenderStaffID:  req.SenderStaffID,
		SenderNick:     req.SenderNick,
		ConversationID: req.ConversationID,
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "消息已接收"})
}

// Feishu 接收飞书消息推送
// POST /webhooks/feishu
func (h *WebhookHandler) Feishu(c *gin.Context) {
	var req struct {
		Header struct {
			EventType string `json:"event_type"`
		} `json:"header"`
		Event struct {
			Message struct {
				MessageType string `json:"message_type"`
				Content     string `json:"content"`
			} `json:"message"`
			Sender struct {
				SenderID struct {
					UnionID string `json:"union_id"`
				} `json:"sender_id"`
				SenderType string `json:"sender_type"`
				Nickname   string `json:"nickname"`
			} `json:"sender"`
		} `json:"event"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 处理飞书消息
	err := h.webhookService.ProcessFeishuMessage(c.Request.Context(), service.FeishuMessage{
		EventType:   req.Header.EventType,
		MessageType: req.Event.Message.MessageType,
		Content:     req.Event.Message.Content,
		SenderID:    req.Event.Sender.SenderID.UnionID,
		SenderType:  req.Event.Sender.SenderType,
		Nickname:    req.Event.Sender.Nickname,
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "消息已接收"})
}

// Generic 接收通用 Webhook
// POST /webhooks/generic
func (h *WebhookHandler) Generic(c *gin.Context) {
	// 获取签名验证
	signature := c.GetHeader("X-Webhook-Signature")
	if signature == "" {
		utils.Error(c, http.StatusUnauthorized, "缺少签名")
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 处理通用 Webhook
	err := h.webhookService.ProcessGenericWebhook(c.Request.Context(), signature, payload)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessWithData(c, gin.H{"message": "Webhook 已接收"})
}

// RegisterRoutes 注册路由
func (h *WebhookHandler) RegisterRoutes(r *gin.RouterGroup) {
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/dingtalk", h.DingTalk)
		webhooks.POST("/feishu", h.Feishu)
		webhooks.POST("/generic", h.Generic)
	}
}
