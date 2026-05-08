// Package service 提供员工业务逻辑层
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
	"claw/pkg/password"
)

// 员工服务相关错误
var (
	ErrEmployeeNotFound    = errors.New("员工不存在")
	ErrEmployeeExists      = errors.New("邮箱已被注册")
	ErrInvalidEmployeeType = errors.New("无效的员工类型")
	ErrInvalidStatus       = errors.New("无效的状态")
	ErrCannotDeleteSelf    = errors.New("不能删除自己")
	ErrInvalidPassword     = errors.New("密码不能为空")
)

// EmployeeService 员工服务
type EmployeeService struct {
	repo repository.EmployeeRepository
}

// NewEmployeeService 创建员工服务
func NewEmployeeService() *EmployeeService {
	return &EmployeeService{
		repo: repository.NewEmployeeRepository(),
	}
}

// CreateEmployeeRequest 创建员工请求
type CreateEmployeeRequest struct {
	Username    string   `json:"username" validate:"required,min=3,max=50,username"`
	DisplayName string   `json:"display_name" validate:"required,min=1,max=100"`
	Name        string   `json:"name"` // 自动从 display_name 复制，无需前端传入
	Type        string   `json:"type" validate:"required,oneof=human agent"`
	Email       string   `json:"email" validate:"required,email"`
	Password    string   `json:"password" validate:"min=6"`
	Role        string   `json:"role"`
	Skills      []string `json:"skills"`

	// 扩展资料
	Avatar     string `json:"avatar,omitempty"`
	Department string `json:"department,omitempty"`
	Position   string `json:"position,omitempty"`
	Phone      string `json:"phone,omitempty"`

	// 通知偏好
	NotificationPrefs model.NotificationPreferences `json:"notification_prefs,omitempty"`
}

// UpdateEmployeeRequest 更新员工请求
type UpdateEmployeeRequest struct {
	Username    string   `json:"username,omitempty" validate:"omitempty,min=3,max=50,username"`
	DisplayName string   `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`
	Name        string   `json:"name,omitempty"` // 自动从 display_name 复制
	Type        string   `json:"type,omitempty" validate:"omitempty,oneof=human agent"`
	Email       string   `json:"email,omitempty" validate:"omitempty,email"`
	Role        string   `json:"role,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	Status      string   `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`

	// 扩展资料
	Avatar     string `json:"avatar,omitempty"`
	Department string `json:"department,omitempty"`
	Position   string `json:"position,omitempty"`
	Phone      string `json:"phone,omitempty"`

	// 通知偏好
	NotificationPrefs model.NotificationPreferences `json:"notification_prefs,omitempty"`
}

// UpdateNotificationPrefsRequest 更新通知偏好请求
type UpdateNotificationPrefsRequest struct {
	Channels model.NotificationChannels `json:"channels"`
	Events   model.EventNotifications   `json:"events"`
}

// EmployeeResponse 员工响应
type EmployeeResponse struct {
	ID                string                        `json:"id"`
	Username          string                        `json:"username"`
	DisplayName       string                        `json:"display_name"`
	Name              string                        `json:"name"` // 兼容旧字段
	Type              string                        `json:"type"`
	Email             string                        `json:"email"`
	Role              string                        `json:"role,omitempty"`
	Skills            []string                      `json:"skills"`
	Avatar            string                        `json:"avatar,omitempty"`
	Department        string                        `json:"department,omitempty"`
	Position          string                        `json:"position,omitempty"`
	Phone             string                        `json:"phone,omitempty"`
	NotificationPrefs model.NotificationPreferences `json:"notification_prefs,omitempty"`
	Status            string                        `json:"status"`
	LastSeenAt        *string                       `json:"last_seen_at,omitempty"`
	CreatedAt         string                        `json:"created_at"`
}

// ListEmployeeRequest 员工列表请求
type ListEmployeeRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Type     string `json:"type,omitempty" validate:"omitempty,oneof=human agent"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Role     string `json:"role,omitempty"`
	Keyword  string `json:"keyword,omitempty"`
}

// ListEmployeeResponse 员工列表响应
type ListEmployeeResponse struct {
	List       []*EmployeeResponse `json:"list"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPage  int                 `json:"total_page"`
}

// SearchEmployeeRequest 搜索员工请求
type SearchEmployeeRequest struct {
	Page     int      `json:"page" validate:"min=1"`
	PageSize int      `json:"page_size" validate:"min=1,max=100"`
	Skills   []string `json:"skills" validate:"required,min=1"`
}

// APIKeyResponse API Key 响应
type APIKeyResponse struct {
	APIKey    string `json:"api_key"`
	UpdatedAt string `json:"updated_at"`
}

// toEmployeeResponse 转换模型到响应
func toEmployeeResponse(emp *model.Employee) *EmployeeResponse {
	// 解析 skills JSON
	var skills []string
	if emp.Skills != "" {
		json.Unmarshal([]byte(emp.Skills), &skills)
	}

	// 获取通知偏好
	notificationPrefs := emp.GetNotificationPrefs()

	resp := &EmployeeResponse{
		ID:                emp.ID,
		Username:          emp.Username,
		DisplayName:       emp.DisplayName,
		Name:              emp.Name,
		Type:              string(emp.Type),
		Email:             emp.Email,
		Role:              emp.Role,
		Skills:            skills,
		Avatar:            emp.Avatar,
		Department:        emp.Department,
		Position:          emp.Position,
		Phone:             emp.Phone,
		NotificationPrefs: notificationPrefs,
		Status:            string(emp.Status),
		CreatedAt:         emp.CreatedAt.Format("2006-01-02T15:04:05"),
	}

	if emp.LastSeenAt != nil {
		lastSeen := emp.LastSeenAt.Format("2006-01-02T15:04:05")
		resp.LastSeenAt = &lastSeen
	}

	return resp
}

// Create 创建员工
func (s *EmployeeService) Create(ctx context.Context, req *CreateEmployeeRequest) (*EmployeeResponse, error) {
	// 检查邮箱是否已存在
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, ErrEmployeeExists
	}

	// 验证员工类型
	empType := model.EmployeeType(req.Type)
	if empType != model.EmployeeTypeHuman && empType != model.EmployeeTypeAgent {
		return nil, ErrInvalidEmployeeType
	}

	// human 类型必须有密码
	if empType == model.EmployeeTypeHuman && req.Password == "" {
		return nil, ErrInvalidPassword
	}

	// 构建技能 JSON
	skillsJSON, _ := json.Marshal(req.Skills)

	// 如果 name 为空，则使用 display_name
	name := req.Name
	if name == "" {
		name = req.DisplayName
	}

	// 创建员工模型
	emp := &model.Employee{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Name:        name,
		Type:        empType,
		Email:       req.Email,
		Role:        req.Role,
		Skills:      string(skillsJSON),
		Status:      model.EmployeeStatusActive,
	}

	// 设置扩展资料
	if req.Avatar != "" {
		emp.Avatar = req.Avatar
	}
	if req.Department != "" {
		emp.Department = req.Department
	}
	if req.Position != "" {
		emp.Position = req.Position
	}
	if req.Phone != "" {
		emp.Phone = req.Phone
	}

	// 设置通知偏好
	if req.NotificationPrefs.Channels.Email || req.NotificationPrefs.Channels.Webhook || req.NotificationPrefs.Channels.Internal {
		// 使用提供的偏好
		if err := emp.SetNotificationPrefs(req.NotificationPrefs); err != nil {
			logger.Warn("设置通知偏好失败", "error", err)
		}
	} else {
		// 使用默认偏好
		if err := emp.SetNotificationPrefs(model.DefaultNotificationPreferences()); err != nil {
			logger.Warn("设置默认通知偏好失败", "error", err)
		}
	}

	// 人类员工需要密码，Agent 可选
	if req.Password != "" {
		hashedPwd, err := password.Hash(req.Password)
		if err != nil {
			logger.Error("密码哈希失败", "error", err)
			return nil, fmt.Errorf("密码处理失败: %w", err)
		}
		emp.Password = hashedPwd
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, emp); err != nil {
		logger.Error("创建员工失败", "error", err, "email", req.Email)
		return nil, fmt.Errorf("创建员工失败: %w", err)
	}

	logger.Info("员工创建成功",
		"id", emp.ID,
		"name", emp.Name,
		"type", emp.Type,
	)

	return toEmployeeResponse(emp), nil
}

// GetByID 根据 ID 获取员工
func (s *EmployeeService) GetByID(ctx context.Context, id string) (*EmployeeResponse, error) {
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		logger.Error("获取员工失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取员工失败: %w", err)
	}

	return toEmployeeResponse(emp), nil
}

// List 获取员工列表
func (s *EmployeeService) List(ctx context.Context, req *ListEmployeeRequest) (*ListEmployeeResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 构建筛选条件
	filter := &repository.EmployeeFilter{
		Type:    req.Type,
		Status:  req.Status,
		Role:    req.Role,
		Keyword: req.Keyword,
	}

	// 获取列表
	emps, total, err := s.repo.ListWithFilter(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		logger.Error("获取员工列表失败", "error", err)
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}

	// 转换响应
	list := make([]*EmployeeResponse, len(emps))
	for i, emp := range emps {
		list[i] = toEmployeeResponse(emp)
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &ListEmployeeResponse{
		List:      list,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Update 更新员工
func (s *EmployeeService) Update(ctx context.Context, id string, req *UpdateEmployeeRequest) (*EmployeeResponse, error) {
	// 获取现有员工
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		logger.Error("获取员工失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取员工失败: %w", err)
	}

	// 更新字段
	if req.Username != "" {
		emp.Username = req.Username
	}
	if req.DisplayName != "" {
		emp.DisplayName = req.DisplayName
	}
	if req.Name != "" {
		emp.Name = req.Name
	}
	if req.Type != "" {
		empType := model.EmployeeType(req.Type)
		if empType == model.EmployeeTypeHuman || empType == model.EmployeeTypeAgent {
			emp.Type = empType
		}
	}
	if req.Email != "" {
		// 检查新邮箱是否已被使用
		existing, err := s.repo.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrEmployeeExists
		}
		emp.Email = req.Email
	}
	if req.Role != "" {
		emp.Role = req.Role
	}
	if req.Skills != nil {
		skillsJSON, _ := json.Marshal(req.Skills)
		emp.Skills = string(skillsJSON)
	}
	if req.Avatar != "" {
		emp.Avatar = req.Avatar
	}
	if req.Department != "" {
		emp.Department = req.Department
	}
	if req.Position != "" {
		emp.Position = req.Position
	}
	if req.Phone != "" {
		emp.Phone = req.Phone
	}
	if req.NotificationPrefs.Channels.Email || req.NotificationPrefs.Channels.Webhook || req.NotificationPrefs.Channels.Internal {
		if err := emp.SetNotificationPrefs(req.NotificationPrefs); err != nil {
			logger.Warn("更新通知偏好失败", "error", err)
		}
	}
	if req.Status != "" {
		emp.Status = model.EmployeeStatus(req.Status)
	}

	// 保存更新
	if err := s.repo.Update(ctx, emp); err != nil {
		logger.Error("更新员工失败", "error", err, "id", id)
		return nil, fmt.Errorf("更新员工失败: %w", err)
	}

	logger.Info("员工更新成功", "id", id)
	return toEmployeeResponse(emp), nil
}

// Delete 删除员工
func (s *EmployeeService) Delete(ctx context.Context, id string, currentUserID string) error {
	// 不能删除自己
	if id == currentUserID {
		return ErrCannotDeleteSelf
	}

	// 检查员工是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrEmployeeNotFound
		}
		logger.Error("获取员工失败", "error", err, "id", id)
		return fmt.Errorf("获取员工失败: %w", err)
	}

	// 执行硬删除
	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("删除员工失败", "error", err, "id", id)
		return fmt.Errorf("删除员工失败: %w", err)
	}

	logger.Info("员工删除成功", "id", id)
	return nil
}

// SearchBySkills 按技能搜索员工
func (s *EmployeeService) SearchBySkills(ctx context.Context, req *SearchEmployeeRequest) (*ListEmployeeResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 标准化技能名称
	skills := make([]string, len(req.Skills))
	for i, skill := range req.Skills {
		skills[i] = strings.ToLower(strings.TrimSpace(skill))
	}

	// 搜索
	emps, total, err := s.repo.SearchBySkills(ctx, skills, req.Page, req.PageSize)
	if err != nil {
		logger.Error("搜索员工失败", "error", err)
		return nil, fmt.Errorf("搜索员工失败: %w", err)
	}

	// 转换响应
	list := make([]*EmployeeResponse, len(emps))
	for i, emp := range emps {
		list[i] = toEmployeeResponse(emp)
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &ListEmployeeResponse{
		List:      list,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// UpdateNotificationPrefs 更新通知偏好
func (s *EmployeeService) UpdateNotificationPrefs(ctx context.Context, id string, req UpdateNotificationPrefsRequest) (*EmployeeResponse, error) {
	// 获取员工
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		logger.Error("获取员工失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取员工失败: %w", err)
	}

	// 更新通知偏好
	prefs := model.NotificationPreferences{
		Channels: req.Channels,
		Events:   req.Events,
	}
	if err := emp.SetNotificationPrefs(prefs); err != nil {
		logger.Error("设置通知偏好失败", "error", err, "id", id)
		return nil, fmt.Errorf("设置通知偏好失败: %w", err)
	}

	// 保存到数据库
	if err := s.repo.Update(ctx, emp); err != nil {
		logger.Error("更新员工失败", "error", err, "id", id)
		return nil, fmt.Errorf("更新员工失败: %w", err)
	}

	logger.Info("通知偏好更新成功", "id", id)
	return toEmployeeResponse(emp), nil
}

// GenerateAPIKey 生成 API Key
func (s *EmployeeService) GenerateAPIKey(ctx context.Context, id string) (*APIKeyResponse, error) {
	// 获取员工
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		logger.Error("获取员工失败", "error", err, "id", id)
		return nil, fmt.Errorf("获取员工失败: %w", err)
	}

	// 去掉类型限制，允许所有员工生成 API Key
	// 人类员工也可通过智能体代理使用 API Key

	// 生成新的 API Key
	apiKey := generateAPIKey()

	// 更新 API Key
	if err := s.repo.UpdateAPIKey(ctx, id, apiKey); err != nil {
		logger.Error("更新 API Key 失败", "error", err, "id", id)
		return nil, fmt.Errorf("更新 API Key 失败: %w", err)
	}

	logger.Info("API Key 生成成功", "id", id)

	return &APIKeyResponse{
		APIKey:    apiKey,
		UpdatedAt: emp.UpdatedAt.Format("2006-01-02T15:04:05"),
	}, nil
}

// ResetAPIKey 重置 API Key
func (s *EmployeeService) ResetAPIKey(ctx context.Context, id string) (*APIKeyResponse, error) {
	return s.GenerateAPIKey(ctx, id)
}

// GetDistinctRoles 获取所有已有的职能值
func (s *EmployeeService) GetDistinctRoles(ctx context.Context) ([]string, error) {
	return s.repo.GetDistinctRoles(ctx)
}

// generateAPIKey 生成随机 API Key
func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "claw_" + hex.EncodeToString(bytes)
}
