// Package service 提供消息业务逻辑层
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
	"claw/internal/websocket"
)

// 消息服务相关错误
var (
	ErrMessageNotFound    = errors.New("消息不存在")
	ErrNotChannelMember   = errors.New("不是频道成员")
	ErrReadonlyCannotSend = errors.New("只读成员不能发送消息")
	ErrInvalidMessageType = errors.New("无效的消息类型")
	ErrContentEmpty       = errors.New("消息内容不能为空")
	ErrContentTooLong     = errors.New("消息内容过长")
)

// MessageService 消息服务
type MessageService struct {
	msgRepo    repository.MessageRepository
	channelRepo repository.ChannelRepository
}

// NewMessageService 创建消息服务
func NewMessageService() *MessageService {
	return &MessageService{
		msgRepo:     repository.NewMessageRepository(),
		channelRepo: repository.NewChannelRepository(),
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChannelID string   `json:"channel_id" validate:"required"`
	Content   string   `json:"content" validate:"required,max=4000"`
	Type      string   `json:"type,omitempty" validate:"omitempty,oneof=text image file system workflow"`
	ParentID  *string  `json:"parent_id,omitempty"`
}

// UpdateMessageRequest 更新消息请求
type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required,max=4000"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID         string           `json:"id"`
	ChannelID  string           `json:"channel_id"`
	SenderID   string           `json:"sender_id"`
	Sender     *SenderSummary   `json:"sender,omitempty"`
	Type       string           `json:"type"`
	Content    string           `json:"content"`
	Mentions   []string         `json:"mentions,omitempty"`
	Skills     []string         `json:"skills,omitempty"`
	WorkflowID *string          `json:"workflow_id,omitempty"`
	ParentID   *string          `json:"parent_id,omitempty"`
	CreatedAt  string           `json:"created_at"`
	UpdatedAt  string           `json:"updated_at"`
}

// SenderSummary 发送者摘要
type SenderSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListMessageRequest 消息列表请求
type ListMessageRequest struct {
	ChannelID string     `json:"channel_id" validate:"required"`
	Before    *time.Time `json:"before,omitempty"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
}

// ListMessageResponse 消息列表响应
type ListMessageResponse struct {
	List      []*MessageResponse `json:"list"`
	HasMore   bool               `json:"has_more"`
	NextCursor *time.Time        `json:"next_cursor,omitempty"`
}

// SearchMessageRequest 搜索消息请求
type SearchMessageRequest struct {
	ChannelID string `json:"channel_id" validate:"required"`
	Keyword   string `json:"keyword" validate:"required,min=1,max=100"`
	Page      int    `json:"page" validate:"min=1"`
	PageSize  int    `json:"page_size" validate:"min=1,max=50"`
}

// SearchMessageResponse 搜索消息响应
type SearchMessageResponse struct {
	List       []*MessageResponse `json:"list"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPage  int                `json:"total_page"`
}

// MentionInfo 提及信息
type MentionInfo struct {
	EmployeeIDs []string `json:"employee_ids"`
	Skills      []string `json:"skills"`
}

// toMessageResponse 转换模型到响应
func toMessageResponse(msg *model.Message) *MessageResponse {
	resp := &MessageResponse{
		ID:        msg.ID,
		ChannelID: msg.ChannelID,
		SenderID:  msg.SenderID,
		Type:      string(msg.Type),
		Content:   msg.Content,
		Mentions:  []string(msg.Mentions),
		Skills:    []string(msg.Skills),
		CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05"),
		UpdatedAt: msg.UpdatedAt.Format("2006-01-02T15:04:05"),
	}

	if msg.WorkflowID != nil {
		resp.WorkflowID = msg.WorkflowID
	}
	if msg.ParentID != nil {
		resp.ParentID = msg.ParentID
	}

	if msg.Sender.ID != "" {
		resp.Sender = &SenderSummary{
			ID:   msg.Sender.ID,
			Name: msg.Sender.Name,
			Type: string(msg.Sender.Type),
		}
	}

	return resp
}

// Send 发送消息
func (s *MessageService) Send(ctx context.Context, req *SendMessageRequest, senderID string) (*MessageResponse, error) {
	// 验证频道存在
	ch, err := s.channelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("获取频道失败: %w", err)
	}
	_ = ch // 使用变量避免未使用警告

	// 检查是否是频道成员
	member, err := s.channelRepo.GetMember(ctx, req.ChannelID, senderID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return nil, ErrNotChannelMember
		}
		return nil, fmt.Errorf("获取成员信息失败: %w", err)
	}

	// 只读成员不能发送消息
	if member.Role == model.ChannelRoleReadonly {
		return nil, ErrReadonlyCannotSend
	}

	// 验证消息类型
	msgType := model.MessageType(req.Type)
	if req.Type == "" {
		msgType = model.MessageTypeText
	}
	if msgType != model.MessageTypeText &&
		msgType != model.MessageTypeImage &&
		msgType != model.MessageTypeFile &&
		msgType != model.MessageTypeSystem &&
		msgType != model.MessageTypeWorkflow {
		return nil, ErrInvalidMessageType
	}

	// 验证父消息（如果是回复）
	if req.ParentID != nil && *req.ParentID != "" {
		_, err := s.msgRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, repository.ErrMessageNotFound) {
				return nil, ErrMessageNotFound
			}
			return nil, fmt.Errorf("获取父消息失败: %w", err)
		}
	}

	// 解析 @提及
	mentionInfo := parseMentions(req.Content)

	// 创建消息
	msg := &model.Message{
		ChannelID: req.ChannelID,
		SenderID:  senderID,
		Type:      msgType,
		Content:   req.Content,
		Mentions:  model.StringArray(mentionInfo.EmployeeIDs),
		Skills:    model.StringArray(mentionInfo.Skills),
	}

	if req.ParentID != nil && *req.ParentID != "" {
		msg.ParentID = req.ParentID
	}

	if err := s.msgRepo.Create(ctx, msg); err != nil {
		logger.Error("创建消息失败", "error", err)
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}

	// 重新获取消息（包含关联数据）
	msg, err = s.msgRepo.GetByID(ctx, msg.ID)
	if err != nil {
		logger.Error("获取消息失败", "error", err)
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	// 广播消息到频道
	go s.broadcastMessage(ch.ID, msg)

	logger.Info("消息发送成功",
		"id", msg.ID,
		"channel_id", msg.ChannelID,
		"sender_id", msg.SenderID,
	)

	return toMessageResponse(msg), nil
}

// GetByID 根据 ID 获取消息
func (s *MessageService) GetByID(ctx context.Context, id string) (*MessageResponse, error) {
	msg, err := s.msgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			return nil, ErrMessageNotFound
		}
		logger.Error("获取消息失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	return toMessageResponse(msg), nil
}

// List 获取消息列表
func (s *MessageService) List(ctx context.Context, req *ListMessageRequest) (*ListMessageResponse, error) {
	// 设置默认限制
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	// 验证频道存在
	_, err := s.channelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("获取频道失败: %w", err)
	}

	messages, err := s.msgRepo.ListByChannel(ctx, req.ChannelID, req.Before, req.Limit+1)
	if err != nil {
		logger.Error("获取消息列表失败", "error", err)
		return nil, fmt.Errorf("获取消息列表失败: %w", err)
	}

	// 判断是否还有更多
	hasMore := len(messages) > req.Limit
	if hasMore {
		messages = messages[:req.Limit]
	}

	// 转换响应
	list := make([]*MessageResponse, len(messages))
	for i, msg := range messages {
		list[i] = toMessageResponse(msg)
	}

	var nextCursor *time.Time
	if hasMore && len(messages) > 0 {
		t := messages[len(messages)-1].CreatedAt
		nextCursor = &t
	}

	return &ListMessageResponse{
		List:       list,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// Update 更新消息
func (s *MessageService) Update(ctx context.Context, id string, req *UpdateMessageRequest, senderID string) (*MessageResponse, error) {
	msg, err := s.msgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			return nil, ErrMessageNotFound
		}
		logger.Error("获取消息失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	// 只能更新自己的消息
	if msg.SenderID != senderID {
		return nil, ErrPermissionDenied
	}

	// 更新内容
	msg.Content = req.Content

	// 重新解析提及
	mentionInfo := parseMentions(req.Content)
	msg.Mentions = model.StringArray(mentionInfo.EmployeeIDs)
	msg.Skills = model.StringArray(mentionInfo.Skills)

	if err := s.msgRepo.Update(ctx, msg); err != nil {
		logger.Error("更新消息失败", "error", err, "id", id)
		return nil, fmt.Errorf("更新消息失败: %w", err)
	}

	logger.Info("消息更新成功", "id", id)
	return toMessageResponse(msg), nil
}

// Delete 删除消息
func (s *MessageService) Delete(ctx context.Context, id string, senderID string, isAdmin bool) error {
	msg, err := s.msgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			return ErrMessageNotFound
		}
		logger.Error("获取消息失败", "error", err, "id", id)
		return fmt.Errorf("获取消息失败: %w", err)
	}

	// 只能删除自己的消息，或者管理员可以删除任何消息
	if msg.SenderID != senderID && !isAdmin {
		return ErrPermissionDenied
	}

	if err := s.msgRepo.Delete(ctx, id); err != nil {
		logger.Error("删除消息失败", "error", err, "id", id)
		return fmt.Errorf("删除消息失败: %w", err)
	}

	logger.Info("消息删除成功", "id", id)
	return nil
}

// Search 搜索消息
func (s *MessageService) Search(ctx context.Context, req *SearchMessageRequest) (*SearchMessageResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 验证频道存在
	_, err := s.channelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("获取频道失败: %w", err)
	}

	messages, total, err := s.msgRepo.Search(ctx, req.ChannelID, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		logger.Error("搜索消息失败", "error", err)
		return nil, fmt.Errorf("搜索消息失败: %w", err)
	}

	// 转换响应
	list := make([]*MessageResponse, len(messages))
	for i, msg := range messages {
		list[i] = toMessageResponse(msg)
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &SearchMessageResponse{
		List:      list,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// GetThread 获取回复线程
func (s *MessageService) GetThread(ctx context.Context, parentID string) ([]*MessageResponse, error) {
	// 验证父消息存在
	_, err := s.msgRepo.GetByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			return nil, ErrMessageNotFound
		}
		return nil, fmt.Errorf("获取父消息失败: %w", err)
	}

	messages, err := s.msgRepo.GetThread(ctx, parentID)
	if err != nil {
		logger.Error("获取回复线程失败", "error", err)
		return nil, fmt.Errorf("获取回复线程失败: %w", err)
	}

	// 转换响应
	list := make([]*MessageResponse, len(messages))
	for i, msg := range messages {
		list[i] = toMessageResponse(msg)
	}

	return list, nil
}

// parseMentions 解析消息中的 @提及
// 支持 @用户名 和 @#技能名 两种格式
func parseMentions(content string) *MentionInfo {
	info := &MentionInfo{
		EmployeeIDs: []string{},
		Skills:      []string{},
	}

	// 匹配 @用户名（字母数字下划线和中文字符）
	userPattern := regexp.MustCompile(`@([a-zA-Z0-9_\x{4e00}-\x{9fa5}]+)`)
	userMatches := userPattern.FindAllStringSubmatch(content, -1)
	for _, match := range userMatches {
		if len(match) > 1 {
			name := strings.TrimSpace(match[1])
			// 排除技能提及（以 # 开头）
			if !strings.HasPrefix(match[0], "@#") {
				info.EmployeeIDs = append(info.EmployeeIDs, name)
			}
		}
	}

	// 匹配 @#技能名
	skillPattern := regexp.MustCompile(`@#([a-zA-Z0-9_\x{4e00}-\x{9fa5}]+)`)
	skillMatches := skillPattern.FindAllStringSubmatch(content, -1)
	for _, match := range skillMatches {
		if len(match) > 1 {
			skill := strings.TrimSpace(match[1])
			info.Skills = append(info.Skills, skill)
		}
	}

	return info
}

// broadcastMessage 广播消息到频道
func (s *MessageService) broadcastMessage(channelID string, msg *model.Message) {
	resp := toMessageResponse(msg)
	data, err := json.Marshal(map[string]interface{}{
		"type":    "new_message",
		"channel": channelID,
		"data":    resp,
	})
	if err != nil {
		logger.Error("序列化消息失败", "error", err)
		return
	}

	websocket.BroadcastToChannel(channelID, data)
}

// CheckSendPermission 检查发送权限
func (s *MessageService) CheckSendPermission(ctx context.Context, channelID, employeeID string) error {
	// 验证频道存在
	_, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return ErrChannelNotFound
		}
		return fmt.Errorf("获取频道失败: %w", err)
	}

	// 检查是否是频道成员
	member, err := s.channelRepo.GetMember(ctx, channelID, employeeID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrNotChannelMember
		}
		return fmt.Errorf("获取成员信息失败: %w", err)
	}

	// 只读成员不能发送消息
	if member.Role == model.ChannelRoleReadonly {
		return ErrReadonlyCannotSend
	}

	return nil
}
