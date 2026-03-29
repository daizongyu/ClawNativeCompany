// Package service 提供 Webhook 业务逻辑层
package service

import (
	"context"
	"errors"
	"regexp"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
	"github.com/google/uuid"
)

// DingTalkMessage 钉钉消息
type DingTalkMessage struct {
	MsgType        string `json:"msgtype"`
	Content        string `json:"content"`
	SenderStaffID  string `json:"senderStaffId"`
	SenderNick     string `json:"senderNick"`
	ConversationID string `json:"conversationId"`
}

// FeishuMessage 飞书消息
type FeishuMessage struct {
	EventType   string `json:"event_type"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
	SenderID    string `json:"sender_id"`
	SenderType  string `json:"sender_type"`
	Nickname    string `json:"nickname"`
}

// WebhookService Webhook 服务
type WebhookService struct {
	msgRepo         repository.MessageRepository
	channelRepo     repository.ChannelRepository
	empRepo         repository.EmployeeRepository
	mappingRepo     repository.ExternalMappingRepository
	msgService      *MessageService
}

// NewWebhookService 创建 Webhook 服务
func NewWebhookService() *WebhookService {
	return &WebhookService{
		msgRepo:         repository.NewMessageRepository(),
		channelRepo:     repository.NewChannelRepository(),
		empRepo:         repository.NewEmployeeRepository(),
		mappingRepo:     repository.NewExternalMappingRepository(),
		msgService:      NewMessageService(),
	}
}

// ProcessDingTalkMessage 处理钉钉消息
func (s *WebhookService) ProcessDingTalkMessage(ctx context.Context, msg DingTalkMessage) error {
	log := logger.Get()
	log.Info("处理钉钉消息",
		"sender", msg.SenderNick,
		"content_preview", truncate(msg.Content, 50),
	)

	// 查找或创建对应的频道（使用映射表）
	channel, err := s.findOrCreateChannelWithMapping(ctx, "dingtalk", msg.ConversationID, msg.SenderNick)
	if err != nil {
		return err
	}

	// 查找或创建发送者（使用映射表）
	sender, err := s.findOrCreateEmployeeWithMapping(ctx, "dingtalk", msg.SenderStaffID, msg.SenderNick)
	if err != nil {
		return err
	}

	// 检查是否触发工作流
	workflowService := NewWorkflowService()
	workflow, triggerData, err := workflowService.CheckKeywordTrigger(ctx, msg.Content)
	if err != nil {
		log.Error("检查关键词触发失败", "error", err)
	}

	if workflow != nil {
		// 触发工作流
		_, err := workflowService.TriggerWorkflow(ctx, workflow.ID, sender.ID, map[string]interface{}{
			"content":      msg.Content,
			"sender_id":    sender.ID,
			"sender_name":  msg.SenderNick,
			"channel_id":   channel.ID,
			"trigger_data": triggerData,
		})
		if err != nil {
			log.Error("触发工作流失败", "error", err)
		}
	}

	// 发送消息到频道
	_, err = s.msgService.Send(ctx, &SendMessageRequest{
		ChannelID: channel.ID,
		Content:   msg.Content,
		Type:      "text",
	}, sender.ID)

	return err
}

// ProcessFeishuMessage 处理飞书消息
func (s *WebhookService) ProcessFeishuMessage(ctx context.Context, msg FeishuMessage) error {
	log := logger.Get()
	log.Info("处理飞书消息",
		"sender", msg.Nickname,
		"content_preview", truncate(msg.Content, 50),
	)

	// 查找或创建对应的频道（使用映射表）
	channel, err := s.findOrCreateChannelWithMapping(ctx, "feishu", "default", msg.Nickname)
	if err != nil {
		return err
	}

	// 查找或创建发送者（使用映射表）
	sender, err := s.findOrCreateEmployeeWithMapping(ctx, "feishu", msg.SenderID, msg.Nickname)
	if err != nil {
		return err
	}

	// 检查是否触发工作流
	workflowService := NewWorkflowService()
	workflow, triggerData, err := workflowService.CheckKeywordTrigger(ctx, msg.Content)
	if err != nil {
		log.Error("检查关键词触发失败", "error", err)
	}

	if workflow != nil {
		// 触发工作流
		_, err := workflowService.TriggerWorkflow(ctx, workflow.ID, sender.ID, map[string]interface{}{
			"content":      msg.Content,
			"sender_id":    sender.ID,
			"sender_name":  msg.Nickname,
			"channel_id":   channel.ID,
			"trigger_data": triggerData,
		})
		if err != nil {
			log.Error("触发工作流失败", "error", err)
		}
	}

	// 发送消息到频道
	_, err = s.msgService.Send(ctx, &SendMessageRequest{
		ChannelID: channel.ID,
		Content:   msg.Content,
		Type:      "text",
	}, sender.ID)

	return err
}

// ProcessGenericWebhook 处理通用 Webhook
func (s *WebhookService) ProcessGenericWebhook(ctx context.Context, signature string, payload map[string]interface{}) error {
	log := logger.Get()
	log.Info("处理通用 Webhook", "signature", signature)

	// 验证签名（简化实现，实际应该使用密钥验证）
	if signature == "" {
		return errors.New("无效的签名")
	}

	// 提取消息内容
	content, _ := payload["content"].(string)
	if content == "" {
		return errors.New("消息内容不能为空")
	}

	// 查找或创建频道（使用映射表）
	channel, err := s.findOrCreateChannelWithMapping(ctx, "webhook", "generic", "Webhook")
	if err != nil {
		return err
	}

	// 查找或创建发送者（使用映射表）
	sender, err := s.findOrCreateEmployeeWithMapping(ctx, "webhook", "generic", "Webhook")
	if err != nil {
		return err
	}

	// 发送消息到频道
	_, err = s.msgService.Send(ctx, &SendMessageRequest{
		ChannelID: channel.ID,
		Content:   content,
		Type:      "text",
	}, sender.ID)

	return err
}

// findOrCreateChannelWithMapping 使用映射表查找或创建频道
func (s *WebhookService) findOrCreateChannelWithMapping(ctx context.Context, sourceType string, externalID string, name string) (*model.Channel, error) {
	// 1. 先查映射表
	mapping, err := s.mappingRepo.GetBySourceAndExternalID(ctx, sourceType, externalID, model.ExternalMappingTypeChannel)
	if err == nil {
		// 找到映射，获取频道
		channel, err := s.channelRepo.GetByID(ctx, mapping.InternalID)
		if err == nil {
			return channel, nil
		}
		// 映射存在但频道不存在，删除无效映射
		s.mappingRepo.Delete(ctx, mapping.ID)
	}

	// 2. 未找到映射，创建新频道
	channel := &model.Channel{
		Name:        sourceType + "-" + name,
		Type:        model.ChannelTypePublic,
		Description: "外部系统同步频道: " + sourceType,
		CreatedBy:   "system",
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	// 3. 创建映射关系
	mapping = &model.ExternalMapping{
		ID:          uuid.New().String(),
		SourceType:  sourceType,
		ExternalID:  externalID,
		MappingType: model.ExternalMappingTypeChannel,
		InternalID:  channel.ID,
		Name:        name,
	}

	if err := s.mappingRepo.Create(ctx, mapping); err != nil {
		// 映射创建失败，但频道已创建，记录错误但不返回
		logger.Get().Error("创建频道映射失败", "error", err, "channel_id", channel.ID)
	}

	return channel, nil
}

// findOrCreateEmployeeWithMapping 使用映射表查找或创建员工
func (s *WebhookService) findOrCreateEmployeeWithMapping(ctx context.Context, sourceType string, externalID string, name string) (*model.Employee, error) {
	// 1. 先查映射表
	mapping, err := s.mappingRepo.GetBySourceAndExternalID(ctx, sourceType, externalID, model.ExternalMappingTypeEmployee)
	if err == nil {
		// 找到映射，获取员工
		emp, err := s.empRepo.GetByID(ctx, mapping.InternalID)
		if err == nil {
			return emp, nil
		}
		// 映射存在但员工不存在，删除无效映射
		s.mappingRepo.Delete(ctx, mapping.ID)
	}

	// 2. 未找到映射，创建新员工
	// 生成唯一邮箱
	email := sourceType + "_" + externalID + "@external.local"

	emp := &model.Employee{
		Name:   name,
		Type:   model.EmployeeTypeAgent,
		Email:  email,
		Status: model.EmployeeStatusActive,
	}

	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}

	// 3. 创建映射关系
	mapping = &model.ExternalMapping{
		ID:          uuid.New().String(),
		SourceType:  sourceType,
		ExternalID:  externalID,
		MappingType: model.ExternalMappingTypeEmployee,
		InternalID:  emp.ID,
		Name:        name,
	}

	if err := s.mappingRepo.Create(ctx, mapping); err != nil {
		// 映射创建失败，但员工已创建，记录错误但不返回
		logger.Get().Error("创建员工映射失败", "error", err, "employee_id", emp.ID)
	}

	return emp, nil
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// extractMentions 提取 @提及
func extractMentions(content string) []string {
	// 使用正则表达式提取 @用户名
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	var mentions []string
	for _, match := range matches {
		if len(match) > 1 {
			mentions = append(mentions, match[1])
		}
	}

	return mentions
}

// extractSkills 提取技能标签
func extractSkills(content string) []string {
	// 使用正则表达式提取 #技能
	re := regexp.MustCompile(`#(\w+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	var skills []string
	for _, match := range matches {
		if len(match) > 1 {
			skills = append(skills, match[1])
		}
	}

	return skills
}
