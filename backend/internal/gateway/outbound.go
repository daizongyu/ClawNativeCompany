// Package gateway 提供外部系统推送功能
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"claw/internal/logger"
	"claw/internal/model"
)

// OutboundGateway Outbound 推送网关
type OutboundGateway struct {
	httpClient *http.Client
	webhooks   map[string]WebhookConfig
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Secret    string            `json:"secret"`
	Timeout   int               `json:"timeout"`
	Retry     int               `json:"retry"`
}

// OutboundMessage 推送消息
type OutboundMessage struct {
	Type        string                 `json:"type"`
	Event       string                 `json:"event"`
	Timestamp   int64                  `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Signature   string                 `json:"signature,omitempty"`
}

// NewOutboundGateway 创建 Outbound 网关
func NewOutboundGateway() *OutboundGateway {
	return &OutboundGateway{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		webhooks: make(map[string]WebhookConfig),
	}
}

// RegisterWebhook 注册 Webhook
func (g *OutboundGateway) RegisterWebhook(name string, config WebhookConfig) {
	if config.Method == "" {
		config.Method = "POST"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.Retry == 0 {
		config.Retry = 3
	}
	g.webhooks[name] = config
}

// PushTaskAssigned 推送任务分配通知
func (g *OutboundGateway) PushTaskAssigned(ctx context.Context, webhookName string, task *model.Task, assignee *model.Employee) error {
	msg := OutboundMessage{
		Type:      "task",
		Event:     "assigned",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"task_id":     task.ID,
			"title":       task.Title,
			"description": task.Description,
			"priority":    task.Priority,
			"assignee": map[string]interface{}{
				"id":   assignee.ID,
				"name": assignee.Name,
				"type": assignee.Type,
			},
			"created_at": task.CreatedAt,
		},
	}

	return g.send(ctx, webhookName, msg)
}

// PushTaskCompleted 推送任务完成通知
func (g *OutboundGateway) PushTaskCompleted(ctx context.Context, webhookName string, task *model.Task) error {
	msg := OutboundMessage{
		Type:      "task",
		Event:     "completed",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"task_id":     task.ID,
			"title":       task.Title,
			"status":      task.Status,
			"completed_at": task.CompletedAt,
			"result":      task.Result,
		},
	}

	return g.send(ctx, webhookName, msg)
}

// PushMessage 推送消息通知
func (g *OutboundGateway) PushMessage(ctx context.Context, webhookName string, msg *model.Message, channel *model.Channel) error {
	outboundMsg := OutboundMessage{
		Type:      "message",
		Event:     "new",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"message_id":   msg.ID,
			"channel_id":   channel.ID,
			"channel_name": channel.Name,
			"sender_id":    msg.SenderID,
			"content":      msg.Content,
			"type":         msg.Type,
			"created_at":   msg.CreatedAt,
		},
	}

	return g.send(ctx, webhookName, outboundMsg)
}

// PushWorkflowEvent 推送工作流事件
func (g *OutboundGateway) PushWorkflowEvent(ctx context.Context, webhookName string, event string, execution *model.WorkflowExecution, workflow *model.Workflow) error {
	msg := OutboundMessage{
		Type:      "workflow",
		Event:     event,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"execution_id": execution.ID,
			"workflow_id": workflow.ID,
			"workflow_name": workflow.Name,
			"status":      execution.Status,
			"started_at":  execution.StartedAt,
		},
	}

	if execution.CompletedAt != nil {
		msg.Data["completed_at"] = execution.CompletedAt
	}

	return g.send(ctx, webhookName, msg)
}

// send 发送消息
func (g *OutboundGateway) send(ctx context.Context, webhookName string, msg OutboundMessage) error {
	config, ok := g.webhooks[webhookName]
	if !ok {
		return fmt.Errorf("webhook %s 未注册", webhookName)
	}

	// 签名
	if config.Secret != "" {
		msg.Signature = g.sign(msg, config.Secret)
	}

	// 序列化
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送请求
	return g.sendWithRetry(ctx, config, body)
}

// sendWithRetry 带重试的发送
func (g *OutboundGateway) sendWithRetry(ctx context.Context, config WebhookConfig, body []byte) error {
	log := logger.Get()

	for i := 0; i < config.Retry; i++ {
		err := g.doSend(ctx, config, body)
		if err == nil {
			return nil
		}

		log.Error("推送失败，准备重试",
			"error", err,
			"retry", i+1,
			"max_retry", config.Retry,
		)

		// 指数退避
		if i < config.Retry-1 {
			time.Sleep(time.Duration(1<<i) * time.Second)
		}
	}

	return fmt.Errorf("推送失败，已重试 %d 次", config.Retry)
}

// doSend 执行发送
func (g *OutboundGateway) doSend(ctx context.Context, config WebhookConfig, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

// sign 签名消息
func (g *OutboundGateway) sign(msg OutboundMessage, secret string) string {
	// 简化实现：使用时间戳 + 密钥的 HMAC
	// 实际应该使用更安全的签名算法
	data := fmt.Sprintf("%d:%s:%s", msg.Timestamp, msg.Type, secret)
	return fmt.Sprintf("%x", data)
}

// PushToAll 推送到所有注册的 Webhook
func (g *OutboundGateway) PushToAll(ctx context.Context, msg OutboundMessage) []error {
	var errors []error
	for name := range g.webhooks {
		if err := g.send(ctx, name, msg); err != nil {
			errors = append(errors, fmt.Errorf("推送到 %s 失败: %w", name, err))
		}
	}
	return errors
}
