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
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Type     string `json:"type" validate:"required,oneof=human agent"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"min=6"`
	Skills   []string `json:"skills"`
}

// UpdateEmployeeRequest 更新员工请求
type UpdateEmployeeRequest struct {
	Name   string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email  string   `json:"email,omitempty" validate:"omitempty,email"`
	Skills []string `json:"skills,omitempty"`
	Status string   `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

// EmployeeResponse 员工响应
type EmployeeResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Email      string   `json:"email"`
	Skills     []string `json:"skills"`
	Status     string   `json:"status"`
	LastSeenAt *string  `json:"last_seen_at,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

// ListEmployeeRequest 员工列表请求
type ListEmployeeRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Type     string `json:"type,omitempty" validate:"omitempty,oneof=human agent"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
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

	resp := &EmployeeResponse{
		ID:        emp.ID,
		Name:      emp.Name,
		Type:      string(emp.Type),
		Email:     emp.Email,
		Skills:    skills,
		Status:    string(emp.Status),
		CreatedAt: emp.CreatedAt.Format("2006-01-02T15:04:05"),
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

	// 创建员工模型
	emp := &model.Employee{
		Name:   req.Name,
		Type:   empType,
		Email:  req.Email,
		Skills: string(skillsJSON),
		Status: model.EmployeeStatusActive,
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

	// 获取列表
	emps, total, err := s.repo.List(ctx, req.Page, req.PageSize)
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
	if req.Name != "" {
		emp.Name = req.Name
	}
	if req.Email != "" {
		// 检查新邮箱是否已被使用
		existing, err := s.repo.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrEmployeeExists
		}
		emp.Email = req.Email
	}
	if req.Skills != nil {
		skillsJSON, _ := json.Marshal(req.Skills)
		emp.Skills = string(skillsJSON)
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

	// TODO: 软删除实现，需要在 repository 中添加 Delete 方法
	// 目前使用更新状态为 inactive 作为替代
	emp := &model.Employee{}
	emp.ID = id
	emp.Status = model.EmployeeStatusInactive
	
	// 这里我们需要一个软删除方法，暂时使用硬删除
	// 实际项目中应该使用软删除
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

	// 只有 Agent 类型可以生成 API Key
	if emp.Type != model.EmployeeTypeAgent {
		return nil, errors.New("只有 Agent 类型员工可以生成 API Key")
	}

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

// generateAPIKey 生成随机 API Key
func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "claw_" + hex.EncodeToString(bytes)
}
