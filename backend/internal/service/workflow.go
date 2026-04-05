// Package service 提供工作流业务逻辑层
package service

import (
	"context"
	"errors"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
	"claw/internal/workflow"
	"claw/internal/workflow/expression"
	"claw/pkg/utils"
)

// 工作流服务错误
var (
	ErrWorkflowNotFound   = errors.New("工作流不存在")
	ErrWorkflowInactive   = errors.New("工作流未激活")
	ErrInvalidStepConfig  = errors.New("无效的步骤配置")
	ErrExecutionNotFound  = errors.New("工作流执行不存在")
	ErrInvalidTriggerType = errors.New("无效的触发器类型")
	ErrInvalidWFStatus    = errors.New("无效的工作流状态")
)

// WorkflowService 工作流服务
type WorkflowService struct {
	workflowRepo  repository.WorkflowRepository
	executionRepo repository.WorkflowExecutionRepository
	engine        *workflow.Engine
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService() *WorkflowService {
	engine := workflow.NewEngine()

	// 注册内置处理器
	engine.RegisterHandler("start", &StartStepHandler{})
	engine.RegisterHandler("end", &EndStepHandler{})
	engine.RegisterHandler("condition", NewConditionStepHandler())

	return &WorkflowService{
		workflowRepo:  repository.NewWorkflowRepository(),
		executionRepo: repository.NewWorkflowExecutionRepository(),
		engine:        engine,
	}
}

// SetTaskCallback 设置任务回调（由 TaskService 调用）
func (s *WorkflowService) SetTaskCallback(callback func(ctx context.Context, executionID string, stepID string, stepConfig map[string]string) error) {
	s.engine.SetTaskCallback(callback)
}

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	Name          string           `json:"name" validate:"required,min=2,max=100"`
	Description   string           `json:"description" validate:"max=500"`
	TriggerType   string           `json:"trigger_type" validate:"required,oneof=keyword manual schedule"`
	TriggerConfig model.JSONMap    `json:"trigger_config" validate:"required"`
	Steps         model.WorkflowSteps `json:"steps" validate:"required,min=1"`
}

// UpdateWorkflowRequest 更新工作流请求
type UpdateWorkflowRequest struct {
	Name          string           `json:"name" validate:"omitempty,min=2,max=100"`
	Description   string           `json:"description" validate:"max=500"`
	TriggerType   string           `json:"trigger_type" validate:"omitempty,oneof=keyword manual schedule"`
	TriggerConfig model.JSONMap    `json:"trigger_config"`
	Steps         model.WorkflowSteps `json:"steps" validate:"omitempty,min=1"`
}

// ListWorkflowRequest 工作流列表请求
type ListWorkflowRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Status   string `json:"status,omitempty"`
}

// ListWorkflowResponse 工作流列表响应
type ListWorkflowResponse struct {
	List       []*WorkflowResponse `json:"list"`
	Pagination utils.Pagination    `json:"pagination"`
}

// WorkflowResponse 工作流响应
type WorkflowResponse struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Status        string          `json:"status"`
	TriggerType   string          `json:"trigger_type"`
	TriggerConfig model.JSONMap   `json:"trigger_config"`
	Steps         model.WorkflowSteps `json:"steps"`
	CreatedBy     string          `json:"created_by"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

// Create 创建工作流
func (s *WorkflowService) Create(ctx context.Context, req CreateWorkflowRequest, createdBy string) (*WorkflowResponse, error) {
	wf := &model.Workflow{
		Name:          req.Name,
		Description:   req.Description,
		Status:        model.WorkflowStatusActive,
		CreatedBy:     createdBy,
		TriggerType:     model.WorkflowTriggerType(req.TriggerType),
		TriggerConfig:   req.TriggerConfig,
		Steps:         req.Steps,
	}

	if err := s.engine.CreateWorkflow(ctx, wf); err != nil {
		return nil, err
	}

	return s.toWorkflowResponse(wf), nil
}

// List 获取工作流列表
func (s *WorkflowService) List(ctx context.Context, req ListWorkflowRequest) (*ListWorkflowResponse, error) {
	var workflows []*model.Workflow
	var total int64
	var err error

	if req.Status != "" {
		status := model.WorkflowStatus(req.Status)
		workflows, total, err = s.workflowRepo.ListByStatus(ctx, status, req.Page, req.PageSize)
	} else {
		workflows, total, err = s.workflowRepo.List(ctx, req.Page, req.PageSize)
	}

	if err != nil {
		return nil, err
	}

	responses := make([]*WorkflowResponse, len(workflows))
	for i, wf := range workflows {
		responses[i] = s.toWorkflowResponse(wf)
	}

	totalPage := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	return &ListWorkflowResponse{
		List: responses,
		Pagination: utils.Pagination{
			Page:      req.Page,
			PageSize:  req.PageSize,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

// GetByID 根据 ID 获取工作流
func (s *WorkflowService) GetByID(ctx context.Context, id string) (*WorkflowResponse, error) {
	wf, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}
	return s.toWorkflowResponse(wf), nil
}

// Update 更新工作流
func (s *WorkflowService) Update(ctx context.Context, id string, req UpdateWorkflowRequest) (*WorkflowResponse, error) {
	wf, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}

	if req.Name != "" {
		wf.Name = req.Name
	}
	if req.Description != "" {
		wf.Description = req.Description
	}
	if req.TriggerType != "" {
		wf.TriggerType = model.WorkflowTriggerType(req.TriggerType)
	}
	if req.TriggerConfig != nil {
		wf.TriggerConfig = req.TriggerConfig
	}
	if len(req.Steps) > 0 {
		wf.Steps = req.Steps
	}

	if err := s.workflowRepo.Update(ctx, wf); err != nil {
		return nil, err
	}

	return s.toWorkflowResponse(wf), nil
}

// Delete 删除工作流
func (s *WorkflowService) Delete(ctx context.Context, id string) error {
	_, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrWorkflowNotFound
		}
		return err
	}
	return s.workflowRepo.Delete(ctx, id)
}

// Search 搜索工作流
func (s *WorkflowService) Search(ctx context.Context, keyword string, page, pageSize int) ([]*WorkflowResponse, int64, error) {
	workflows, total, err := s.workflowRepo.SearchByName(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*WorkflowResponse, len(workflows))
	for i, w := range workflows {
		responses[i] = s.toWorkflowResponse(w)
	}
	return responses, total, nil
}

// UpdateStatus 更新工作流状态
func (s *WorkflowService) UpdateStatus(ctx context.Context, id string, status string) (*WorkflowResponse, error) {
	wf, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}

	wfStatus := model.WorkflowStatus(status)
	if wfStatus != model.WorkflowStatusActive && wfStatus != model.WorkflowStatusInactive {
		return nil, ErrInvalidWFStatus
	}

	wf.Status = wfStatus
	if err := s.workflowRepo.Update(ctx, wf); err != nil {
		return nil, err
	}
	return s.toWorkflowResponse(wf), nil
}

// TriggerWorkflow 触发工作流
func (s *WorkflowService) TriggerWorkflow(ctx context.Context, workflowID string, triggeredBy string, data map[string]interface{}) (*WorkflowExecutionResponse, error) {
	execution, err := s.engine.TriggerWorkflow(ctx, workflowID, triggeredBy, data)
	if err != nil {
		return nil, err
	}
	return s.toExecutionResponse(execution), nil
}

// GetExecution 获取工作流执行
func (s *WorkflowService) GetExecution(ctx context.Context, executionID string) (*WorkflowExecutionResponse, error) {
	execution, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrExecutionNotFound
		}
		return nil, err
	}
	return s.toExecutionResponse(execution), nil
}

// ListExecutions 获取工作流执行列表
func (s *WorkflowService) ListExecutions(ctx context.Context, workflowID string, page, pageSize int) ([]*WorkflowExecutionResponse, int64, error) {
	executions, total, err := s.executionRepo.ListByWorkflow(ctx, workflowID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*WorkflowExecutionResponse, len(executions))
	for i, e := range executions {
		responses[i] = s.toExecutionResponse(e)
	}
	return responses, total, nil
}

// ListActiveExecutions 获取活跃执行列表
func (s *WorkflowService) ListActiveExecutions(ctx context.Context, page, pageSize int) ([]*WorkflowExecutionResponse, int64, error) {
	executions, total, err := s.executionRepo.ListActive(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*WorkflowExecutionResponse, len(executions))
	for i, e := range executions {
		responses[i] = s.toExecutionResponse(e)
	}
	return responses, total, nil
}

// CancelExecution 取消执行
func (s *WorkflowService) CancelExecution(ctx context.Context, executionID string) error {
	return s.engine.CancelExecution(ctx, executionID)
}

// CompleteStep 完成步骤（由 TaskService 调用）
func (s *WorkflowService) CompleteStep(ctx context.Context, executionID string, success bool, output map[string]interface{}, nextStep *string, errMsg string) error {
	result := &workflow.StepResult{
		Success:   success,
		Output:    output,
		NextStep:   nextStep,
		Error:     errMsg,
	}
	return s.engine.CompleteStep(ctx, executionID, result)
}

// toWorkflowResponse 转换为工作流响应
func (s *WorkflowService) toWorkflowResponse(w *model.Workflow) *WorkflowResponse {
	return &WorkflowResponse{
		ID:            w.ID,
		Name:          w.Name,
		Description:   w.Description,
		Status:        string(w.Status),
		TriggerType:   string(w.TriggerType),
		TriggerConfig: w.TriggerConfig,
		Steps:         w.Steps,
		CreatedBy:     w.CreatedBy,
		CreatedAt:     w.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     w.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}


// WorkflowExecutionResponse 工作流执行响应
type WorkflowExecutionResponse struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	TriggeredBy   string                 `json:"triggered_by"`
	TriggerType   string                 `json:"trigger_type"`
	Input         map[string]interface{}   `json:"input"`
	Output        map[string]interface{}   `json:"output"`
	Status        string                 `json:"status"`
	StartedAt     int64                  `json:"started_at"`
	CompletedAt   *int64               `json:"completed_at,omitempty"`
	ErrorMessage  string               `json:"error_message,omitempty"`
	CreatedAt     string                 `json:"created_at"`
}

// toExecutionResponse 转换为执行响应
func (s *WorkflowService) toExecutionResponse(e *model.WorkflowExecution) *WorkflowExecutionResponse {
	resp := &WorkflowExecutionResponse{
		ID:          e.ID,
		WorkflowID:    e.WorkflowID,
		TriggeredBy:   e.TriggeredBy,
		TriggerType:   string(e.TriggerType),
		Input:       e.Input,
		Output:      e.Output,
		Status:      string(e.Status),
		StartedAt:   e.StartedAt,
		ErrorMessage: e.ErrorMessage,
		CreatedAt:   e.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if e.CompletedAt != nil {
		resp.CompletedAt = e.CompletedAt
	}

	return resp
}

// CheckKeywordTrigger 检查关键词触发
func (s *WorkflowService) CheckKeywordTrigger(ctx context.Context, content string) (*model.Workflow, map[string]interface{}, error) {
	// 获取所有活跃的关键词触发工作流
	workflows, _, err := s.workflowRepo.ListByStatus(ctx, model.WorkflowStatusActive, 1, 1000)
	if err != nil {
		return nil, nil, err
	}

	for _, w := range workflows {
		if w.TriggerType != model.WorkflowTriggerKeyword {
			continue
		}

		// 检查关键词
		keywords, ok := w.TriggerConfig["keywords"].([]interface{})
		if !ok {
			continue
		}

		for _, kw := range keywords {
			keyword, ok := kw.(string)
			if !ok {
				continue
			}

			if contains(content, keyword) {
				return w, map[string]interface{}{
					"trigger_keyword": keyword,
					"content":         content,
				}, nil
			}
		}
	}

	return nil, nil, nil
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==================== 处理器实现 ====================

// StartStepHandler 开始处理器
type StartStepHandler struct{}


// Execute 执行开始
func (h *StartStepHandler) Execute(ctx context.Context, step *model.WorkflowStep, input map[string]interface{}) (*workflow.StepResult, error) {
	log := logger.Get()
	log.Info("执行开始", "step_id", step.ID)

	nextStep := ""
	if step.NextStepID != nil {
		nextStep = *step.NextStepID
	}

	return &workflow.StepResult{
		Success:   true,
		Output:    input,
		NextStep:  &nextStep,
	}, nil
}

// EndStepHandler 结束处理器
type EndStepHandler struct{}

// Execute 执行结束
func (h *EndStepHandler) Execute(ctx context.Context, step *model.WorkflowStep, input map[string]interface{}) (*workflow.StepResult, error) {
	log := logger.Get()
	log.Info("执行结束", "step_id", step.ID)

	return &workflow.StepResult{
		Success: true,
		Output:  input,
		NextStep: nil,
	}, nil
}

// ConditionStepHandler 条件处理器
type ConditionStepHandler struct {
	engine *expression.Engine
}

// NewConditionStepHandler 创建条件处理器
func NewConditionStepHandler() *ConditionStepHandler {
	return &ConditionStepHandler{
		engine: expression.NewEngine(),
	}
}

// Execute 执行条件
func (h *ConditionStepHandler) Execute(ctx context.Context, step *model.WorkflowStep, input map[string]interface{}) (*workflow.StepResult, error) {
	log := logger.Get()
	log.Info("执行条件", "step_id", step.ID)

	// 获取条件配置
	condition, ok := step.Config["condition"]
	if !ok {
		return &workflow.StepResult{
			Success: false,
			Error:   "条件缺少 condition 配置",
		}, nil
	}

	// 使用表达式引擎评估条件
	success, err := h.engine.Evaluate(condition, input)
	if err != nil {
		log.Error("条件表达式求值失败", "error", err, "condition", condition)
		return &workflow.StepResult{
			Success: false,
			Error:   "条件表达式求值失败: " + err.Error(),
		}, nil
	}

	nextStep := ""
	if success && step.NextStepID != nil {
		nextStep = *step.NextStepID
	}

	return &workflow.StepResult{
		Success:  success,
		Output:   input,
		NextStep: &nextStep,
	}, nil
}
