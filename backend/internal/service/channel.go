// Package service 提供频道业务逻辑层
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
)

// 频道服务相关错误
var (
	ErrChannelNotFound   = errors.New("频道不存在")
	ErrChannelExists     = errors.New("频道名称已存在")
	ErrMemberNotFound    = errors.New("成员不存在")
	ErrAlreadyMember     = errors.New("已经是频道成员")
	ErrPermissionDenied  = errors.New("权限不足")
	ErrCannotRemoveSelf  = errors.New("不能移除自己")
	ErrLastAdmin         = errors.New("不能移除最后一个管理员")
	ErrInvalidRole       = errors.New("无效的角色")
	ErrInvalidChannelType = errors.New("无效的频道类型")
)

// ChannelService 频道服务
type ChannelService struct {
	repo     repository.ChannelRepository
	empRepo  repository.EmployeeRepository
}

// NewChannelService 创建频道服务
func NewChannelService() *ChannelService {
	return &ChannelService{
		repo:    repository.NewChannelRepository(),
		empRepo: repository.NewEmployeeRepository(),
	}
}

// CreateChannelRequest 创建频道请求
type CreateChannelRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Type        string `json:"type" validate:"required,oneof=public private dm"`
	Description string `json:"description" validate:"max=500"`
}

// UpdateChannelRequest 更新频道请求
type UpdateChannelRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	EmployeeID string `json:"employee_id" validate:"required"`
	Role       string `json:"role" validate:"required,oneof=admin member readonly"`
}

// UpdateMemberRoleRequest 更新成员角色请求
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member readonly"`
}

// ChannelResponse 频道响应
type ChannelResponse struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Description string           `json:"description"`
	CreatedBy   string           `json:"created_by"`
	CreatorName string           `json:"creator_name"` // 创建者名称
	MemberCount int              `json:"member_count"` // 成员数量
	Members     []*MemberSummary `json:"members,omitempty"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

// MemberSummary 成员摘要
type MemberSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// ChannelMemberResponse 频道成员响应
type ChannelMemberResponse struct {
	ChannelID  string           `json:"channel_id"`
	EmployeeID string           `json:"employee_id"`
	Role       string           `json:"role"`
	Employee   *EmployeeSummary `json:"employee,omitempty"`
	JoinedAt   string           `json:"joined_at"`
}

// EmployeeSummary 员工摘要
type EmployeeSummary struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Email  string   `json:"email"`
	Skills []string `json:"skills"`
}

// ListChannelRequest 频道列表请求
type ListChannelRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Type     string `json:"type,omitempty" validate:"omitempty,oneof=public private dm"`
	Keyword  string `json:"keyword,omitempty"`
}

// ListChannelResponse 频道列表响应
type ListChannelResponse struct {
	List      []*ChannelResponse `json:"list"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
	TotalPage int                `json:"total_page"`
}

// toChannelResponse 转换模型到响应
func (s *ChannelService) toChannelResponse(ctx context.Context, ch *model.Channel) *ChannelResponse {
	resp := &ChannelResponse{
		ID:          ch.ID,
		Name:        ch.Name,
		Type:        string(ch.Type),
		Description: ch.Description,
		CreatedBy:   ch.CreatedBy,
		CreatedAt:   ch.CreatedAt.Format("2006-01-02T15:04:05"),
		UpdatedAt:   ch.UpdatedAt.Format("2006-01-02T15:04:05"),
	}

	// 查询创建者名称
	if ch.CreatedBy != "" {
		creator, err := s.empRepo.GetByID(ctx, ch.CreatedBy)
		if err == nil && creator != nil {
			resp.CreatorName = creator.Name
		}
	}

	// 转换成员摘要
	if len(ch.Members) > 0 {
		resp.Members = make([]*MemberSummary, len(ch.Members))
		for i, m := range ch.Members {
			resp.Members[i] = &MemberSummary{
				ID:   m.ID,
				Name: m.Name,
			}
		}
		resp.MemberCount = len(ch.Members)
	}

	return resp
}

// toChannelMemberResponse 转换成员模型到响应
func toChannelMemberResponse(m *model.ChannelMember) *ChannelMemberResponse {
	resp := &ChannelMemberResponse{
		ChannelID:  m.ChannelID,
		EmployeeID: m.EmployeeID,
		Role:       string(m.Role),
		JoinedAt:   m.CreatedAt.Format("2006-01-02T15:04:05"),
	}

	// 添加员工信息
	if m.Employee != nil {
		skills := []string{}
		if m.Employee.Skills != "" {
			_ = json.Unmarshal([]byte(m.Employee.Skills), &skills)
		}
		resp.Employee = &EmployeeSummary{
			ID:     m.Employee.ID,
			Name:   m.Employee.Name,
			Type:   string(m.Employee.Type),
			Email:  m.Employee.Email,
			Skills: skills,
		}
	}

	return resp
}

// Create 创建频道
func (s *ChannelService) Create(ctx context.Context, req *CreateChannelRequest, createdBy string) (*ChannelResponse, error) {
	// 验证频道类型
	chType := model.ChannelType(req.Type)
	if chType != model.ChannelTypePublic && chType != model.ChannelTypePrivate && chType != model.ChannelTypeDM {
		return nil, ErrInvalidChannelType
	}

	// 创建频道
	ch := &model.Channel{
		Name:        req.Name,
		Type:        chType,
		Description: req.Description,
		CreatedBy:   createdBy,
	}

	if err := s.repo.Create(ctx, ch); err != nil {
		logger.Error("创建频道失败", "error", err, "name", req.Name)
		return nil, fmt.Errorf("创建频道失败: %w", err)
	}

	// 创建者自动成为管理员
	member := &model.ChannelMember{
		ChannelID:  ch.ID,
		EmployeeID: createdBy,
		Role:       model.ChannelRoleAdmin,
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		logger.Error("添加频道创建者失败", "error", err, "channel_id", ch.ID)
		// 继续，不中断流程
	}

	logger.Info("频道创建成功",
		"id", ch.ID,
		"name", ch.Name,
		"type", ch.Type,
	)

	return s.toChannelResponse(ctx, ch), nil
}

// GetByID 根据 ID 获取频道
func (s *ChannelService) GetByID(ctx context.Context, id string) (*ChannelResponse, error) {
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}
		logger.Error("获取频道失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取频道失败: %w", err)
	}

	return s.toChannelResponse(ctx, ch), nil
}

// List 获取频道列表
func (s *ChannelService) List(ctx context.Context, req *ListChannelRequest) (*ListChannelResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 构建筛选条件
	filter := repository.ListFilter{
		Type:    req.Type,
		Keyword: req.Keyword,
	}

	channels, total, err := s.repo.ListWithFilter(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		logger.Error("获取频道列表失败", "error", err)
		return nil, fmt.Errorf("获取频道列表失败: %w", err)
	}

	// 转换响应
	list := make([]*ChannelResponse, len(channels))
	for i, ch := range channels {
		list[i] = s.toChannelResponse(ctx, ch)
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &ListChannelResponse{
		List:      list,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// ListByMember 获取员工加入的频道
func (s *ChannelService) ListByMember(ctx context.Context, employeeID string, page, pageSize int) (*ListChannelResponse, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	channels, total, err := s.repo.ListByMember(ctx, employeeID, page, pageSize)
	if err != nil {
		logger.Error("获取员工频道列表失败", "error", err)
		return nil, fmt.Errorf("获取频道列表失败: %w", err)
	}

	// 转换响应
	list := make([]*ChannelResponse, len(channels))
	for i, ch := range channels {
		list[i] = s.toChannelResponse(ctx, ch)
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &ListChannelResponse{
		List:      list,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}, nil
}

// Update 更新频道
func (s *ChannelService) Update(ctx context.Context, id string, req *UpdateChannelRequest) (*ChannelResponse, error) {
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}
		logger.Error("获取频道失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取频道失败: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		ch.Name = req.Name
	}
	if req.Description != "" {
		ch.Description = req.Description
	}

	if err := s.repo.Update(ctx, ch); err != nil {
		logger.Error("更新频道失败", "error", err, "id", id)
		return nil, fmt.Errorf("更新频道失败: %w", err)
	}

	logger.Info("频道更新成功", "id", id)
	return s.toChannelResponse(ctx, ch), nil
}

// Delete 删除频道
func (s *ChannelService) Delete(ctx context.Context, id string) error {
	// 检查频道是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return ErrChannelNotFound
		}
		logger.Error("获取频道失败", "error", err, "id", id)
		return fmt.Errorf("获取频道失败: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("删除频道失败", "error", err, "id", id)
		return fmt.Errorf("删除频道失败: %w", err)
	}

	logger.Info("频道删除成功", "id", id)
	return nil
}

// AddMember 添加成员
func (s *ChannelService) AddMember(ctx context.Context, channelID string, req *AddMemberRequest) (*ChannelMemberResponse, error) {
	// 验证角色
	role := model.ChannelRole(req.Role)
	if role != model.ChannelRoleAdmin && role != model.ChannelRoleMember && role != model.ChannelRoleReadonly {
		return nil, ErrInvalidRole
	}

	member := &model.ChannelMember{
		ChannelID:  channelID,
		EmployeeID: req.EmployeeID,
		Role:       role,
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		if errors.Is(err, repository.ErrAlreadyMember) {
			return nil, ErrAlreadyMember
		}
		logger.Error("添加成员失败", "error", err, "channel_id", channelID, "employee_id", req.EmployeeID)
		return nil, fmt.Errorf("添加成员失败: %w", err)
	}

	logger.Info("成员添加成功",
		"channel_id", channelID,
		"employee_id", req.EmployeeID,
		"role", role,
	)

	return toChannelMemberResponse(member), nil
}

// RemoveMember 移除成员
func (s *ChannelService) RemoveMember(ctx context.Context, channelID, employeeID string, requestedBy string) error {
	// 不能移除自己
	if employeeID == requestedBy {
		return ErrCannotRemoveSelf
	}

	// 检查是否是最后一个管理员
	members, err := s.repo.ListMembers(ctx, channelID)
	if err != nil {
		logger.Error("获取成员列表失败", "error", err, "channel_id", channelID)
		return fmt.Errorf("获取成员列表失败: %w", err)
	}

	adminCount := 0
	for _, m := range members {
		if m.Role == model.ChannelRoleAdmin {
			adminCount++
		}
	}

	// 如果要移除的是管理员，且只有一个管理员
	for _, m := range members {
		if m.EmployeeID == employeeID && m.Role == model.ChannelRoleAdmin && adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	if err := s.repo.RemoveMember(ctx, channelID, employeeID); err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrMemberNotFound
		}
		logger.Error("移除成员失败", "error", err, "channel_id", channelID, "employee_id", employeeID)
		return fmt.Errorf("移除成员失败: %w", err)
	}

	logger.Info("成员移除成功",
		"channel_id", channelID,
		"employee_id", employeeID,
	)

	return nil
}

// UpdateMemberRole 更新成员角色
func (s *ChannelService) UpdateMemberRole(ctx context.Context, channelID, employeeID string, req *UpdateMemberRoleRequest) (*ChannelMemberResponse, error) {
	// 验证角色
	role := model.ChannelRole(req.Role)
	if role != model.ChannelRoleAdmin && role != model.ChannelRoleMember && role != model.ChannelRoleReadonly {
		return nil, ErrInvalidRole
	}

	if err := s.repo.UpdateMemberRole(ctx, channelID, employeeID, role); err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return nil, ErrMemberNotFound
		}
		logger.Error("更新成员角色失败", "error", err, "channel_id", channelID, "employee_id", employeeID)
		return nil, fmt.Errorf("更新成员角色失败: %w", err)
	}

	// 获取更新后的成员信息
	member, err := s.repo.GetMember(ctx, channelID, employeeID)
	if err != nil {
		return nil, err
	}

	logger.Info("成员角色更新成功",
		"channel_id", channelID,
		"employee_id", employeeID,
		"role", role,
	)

	return toChannelMemberResponse(member), nil
}

// ListMembers 获取成员列表
func (s *ChannelService) ListMembers(ctx context.Context, channelID string) ([]*ChannelMemberResponse, error) {
	members, err := s.repo.ListMembers(ctx, channelID)
	if err != nil {
		logger.Error("获取成员列表失败", "error", err, "channel_id", channelID)
		return nil, fmt.Errorf("获取成员列表失败: %w", err)
	}

	// 转换响应
	list := make([]*ChannelMemberResponse, len(members))
	for i, m := range members {
		list[i] = toChannelMemberResponse(m)
	}

	return list, nil
}

// CheckPermission 检查权限
func (s *ChannelService) CheckPermission(ctx context.Context, channelID, employeeID string, minRole model.ChannelRole) (bool, error) {
	return s.repo.CheckPermission(ctx, channelID, employeeID, minRole)
}

// GetMemberRole 获取成员角色
func (s *ChannelService) GetMemberRole(ctx context.Context, channelID, employeeID string) (model.ChannelRole, error) {
	member, err := s.repo.GetMember(ctx, channelID, employeeID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return "", ErrMemberNotFound
		}
		return "", err
	}
	return member.Role, nil
}
