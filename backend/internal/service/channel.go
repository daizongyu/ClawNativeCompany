// Package service 提供频道业务逻辑层
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	Type        string `json:"type,omitempty" validate:"omitempty,oneof=public private dm"`
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
	Path        string           `json:"path"`         // 树形路径
	Depth       int              `json:"depth"`        // 层级深度
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

// ChannelTreeNode 频道树节点
type ChannelTreeNode struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	Description string             `json:"description"`
	Path        string             `json:"path"`
	Depth       int                `json:"depth"`
	DocCount    int                `json:"doc_count"`
	ChildCount  int                `json:"child_count"`
	Children    []*ChannelTreeNode `json:"children,omitempty"`
	CreatedBy   string             `json:"created_by"`
	CreatorName string             `json:"creator_name"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// ChannelDetailResponse 频道详情响应（含子频道和文档）
type ChannelDetailResponse struct {
	ChannelResponse
	Children    []*ChannelTreeNode `json:"children"`
	Documents   interface{}        `json:"documents,omitempty"`
	Breadcrumbs []*BreadcrumbItem  `json:"breadcrumbs"`
}

// BreadcrumbItem 面包屑项
type BreadcrumbItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
		Path:        ch.Path,
		Depth:       ch.Depth,
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

	// 查询成员数量（从数据库实时查询）
	memberCount, err := s.repo.GetMemberCount(ctx, ch.ID)
	if err == nil {
		resp.MemberCount = int(memberCount)
	}

	// 转换成员摘要（用于详情展示）
	if len(ch.Members) > 0 {
		resp.Members = make([]*MemberSummary, len(ch.Members))
		for i, m := range ch.Members {
			resp.Members[i] = &MemberSummary{
				ID:   m.ID,
				Name: m.Name,
			}
		}
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
	if req.Type != "" {
		chType := model.ChannelType(req.Type)
		if chType == model.ChannelTypePublic || chType == model.ChannelTypePrivate || chType == model.ChannelTypeDM {
			ch.Type = chType
		}
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

// ==================== 树形结构相关方法（新增） ====================

// CreateChildChannelRequest 创建子频道请求
type CreateChildChannelRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Type        string `json:"type" validate:"required,oneof=public private"`
	Description string `json:"description" validate:"max=500"`
	CreatedBy   string `json:"created_by"`
}

// GetChannelTree 获取频道树
func (s *ChannelService) GetChannelTree(ctx context.Context, rootID string) ([]*ChannelTreeNode, error) {
	// 1. 获取所有频道
	channels, err := s.repo.ListAll(ctx)
	if err != nil {
		logger.Error("获取所有频道失败", "error", err)
		return nil, fmt.Errorf("获取频道列表失败: %w", err)
	}

	// 2. 转换为树节点
	nodeMap := make(map[string]*ChannelTreeNode)
	for _, ch := range channels {
		node := s.toTreeNode(ctx, ch)
		nodeMap[ch.ID] = node
	}

	// 3. 构建树结构
	var roots []*ChannelTreeNode
	for _, ch := range channels {
		node := nodeMap[ch.ID]
		if ch.ParentID != nil && *ch.ParentID != "" {
			if parent, ok := nodeMap[*ch.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				parent.ChildCount++
			}
		} else {
			roots = append(roots, node)
		}
	}

	// 4. 如果指定了根节点，返回该节点的子树
	if rootID != "" {
		if root, ok := nodeMap[rootID]; ok {
			return root.Children, nil
		}
		return []*ChannelTreeNode{}, nil
	}

	return roots, nil
}

// toTreeNode 转换为树节点
func (s *ChannelService) toTreeNode(ctx context.Context, ch *model.Channel) *ChannelTreeNode {
	node := &ChannelTreeNode{
		ID:          ch.ID,
		Name:        ch.Name,
		Type:        string(ch.Type),
		Description: ch.Description,
		Path:        ch.Path,
		Depth:       ch.Depth,
		DocCount:    ch.DocCount,
		ChildCount:  0,
		Children:    []*ChannelTreeNode{},
		CreatedBy:   ch.CreatedBy,
		CreatedAt:   ch.CreatedAt.Format("2006-01-02T15:04:05"),
		UpdatedAt:   ch.UpdatedAt.Format("2006-01-02T15:04:05"),
	}

	// 获取创建者名称
	if ch.CreatedBy != "" {
		creator, err := s.empRepo.GetByID(ctx, ch.CreatedBy)
		if err == nil && creator != nil {
			node.CreatorName = creator.DisplayName
		}
	}

	return node
}

// GetChannelDetail 获取频道详情（含子频道和文档统计）
func (s *ChannelService) GetChannelDetail(ctx context.Context, id string) (*ChannelDetailResponse, error) {
	// 1. 获取频道基本信息
	channel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}

	// 2. 获取子频道
	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		logger.Warn("获取子频道失败", "error", err, "channel_id", id)
	}

	childNodes := make([]*ChannelTreeNode, 0, len(children))
	for _, child := range children {
		childNodes = append(childNodes, s.toTreeNode(ctx, child))
	}

	// 3. 构建面包屑
	breadcrumbs := s.buildBreadcrumbs(ctx, channel)

	// 4. 组合响应
	detail := &ChannelDetailResponse{
		ChannelResponse: *s.toChannelResponse(ctx, channel),
		Children:        childNodes,
		Breadcrumbs:     breadcrumbs,
	}

	return detail, nil
}

// buildBreadcrumbs 构建面包屑路径
func (s *ChannelService) buildBreadcrumbs(ctx context.Context, ch *model.Channel) []*BreadcrumbItem {
	if ch.Path == "" {
		return []*BreadcrumbItem{}
	}

	// 解析路径获取所有父频道ID
	pathParts := strings.Split(ch.Path, "/")
	items := make([]*BreadcrumbItem, 0, len(pathParts))

	for _, part := range pathParts {
		if part == "" {
			continue
		}
		// 路径格式: /parent_id/channel_id
		if part == ch.ID {
			items = append(items, &BreadcrumbItem{
				ID:   ch.ID,
				Name: ch.Name,
			})
		} else {
			// 查询父频道名称
			parent, err := s.repo.GetByID(ctx, part)
			if err == nil && parent != nil {
				items = append(items, &BreadcrumbItem{
					ID:   parent.ID,
					Name: parent.Name,
				})
			}
		}
	}

	return items
}

// CreateChildChannel 创建子频道
func (s *ChannelService) CreateChildChannel(ctx context.Context, parentID string, req *CreateChildChannelRequest) (*ChannelResponse, error) {
	// 1. 检查父频道是否存在
	_, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, repository.ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}

	// 2. 检查权限（需要是父频道管理员或成员）
	// 注意：这里假设调用方已经验证了权限

	// 3. 创建子频道
	chType := model.ChannelType(req.Type)
	if chType != model.ChannelTypePublic && chType != model.ChannelTypePrivate {
		return nil, ErrInvalidChannelType
	}

	channel := &model.Channel{
		Name:        req.Name,
		Type:        chType,
		Description: req.Description,
		ParentID:    &parentID,
		CreatedBy:   req.CreatedBy,
	}

	// BeforeCreate 会自动计算 Path 和 Depth
	if err := s.repo.Create(ctx, channel); err != nil {
		logger.Error("创建子频道失败", "error", err, "parent_id", parentID)
		return nil, fmt.Errorf("创建子频道失败: %w", err)
	}

	// 4. 更新父频道的子频道数量
	if err := s.repo.UpdateChildCount(ctx, parentID, 1); err != nil {
		logger.Warn("更新父频道子频道数量失败", "error", err, "parent_id", parentID)
	}

	// 5. 添加创建者为管理员
	member := &model.ChannelMember{
		ChannelID:  channel.ID,
		EmployeeID: req.CreatedBy,
		Role:       model.ChannelRoleAdmin,
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		logger.Warn("添加频道成员失败", "error", err, "channel_id", channel.ID)
	}

	return s.toChannelResponse(ctx, channel), nil
}
