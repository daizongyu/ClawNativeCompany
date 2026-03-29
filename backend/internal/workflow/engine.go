// Package workflow 提供工作流引擎核心功能
package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
)

// 工作流引擎错误
var (
	ErrWorkflowNotFound      = errors.New("工作流不存在")
	ErrWorkflowInactive    = errors.New("工作流未激活")
	ErrInvalidStepConfig   = errors.New("无效的步骤配置")
	ErrStepNotFound      = errors.New("步骤不存在")
	ErrExecutionNotFound    = errors.New("工作流执行不存在")
	ErrExecutionNotRunning   = errors.New("工作流执行未运行")
	ErrNoStartStep      = errors.New("工作流缺少开始步骤")
	ErrNoHandler       = errors.New("步骤处理器未找到")
	ErrInvalidTrigger  = errors.New("无效的触发器配置")
)

// StepHandler 步骤处理器接口
type StepHandler interface {
	Execute(ctx context.Context, step *model.WorkflowStep, input map[string]interface{}) (*StepResult, error)
}

// StepResult 步骤执行结果
type StepResult struct {
	Success   bool                   `json:"success"`
	Output    map[string]interface{} `json:"output"`
	NextStep  *string                `json:"next_step,omitempty"`
	Error     string                 `json:"error,omitempty"`
}


// Engine 工作流引擎
type Engine struct {
	workflowRepo    repository.WorkflowRepository
	executionRepo repository.WorkflowExecutionRepository
	handlers     map[string]StepHandler
	taskCallback  func(ctx context.Context, executionID string, stepID string, stepConfig map[string]string) error
}

// NewEngine 创建工作流引擎
func NewEngine() *Engine {
	return &Engine{
		workflowRepo:    repository.NewWorkflowRepository(),
		executionRepo: repository.NewWorkflowExecutionRepository(),
		handlers:      make(map[string]StepHandler),
	}
}

// RegisterHandler 注册步骤处理器
func (e *Engine) RegisterHandler(stepType string, handler StepHandler) {
	e.handlers[stepType] = handler
}

// SetTaskCallback 设置任务回调函数
func (e *Engine) SetTaskCallback(callback func(ctx context.Context, executionID string, stepID string, stepConfig map[string]string) error) {
	e.taskCallback = callback
}

// CreateWorkflow 创建工作流
func (e *Engine) CreateWorkflow(ctx context.Context, workflow *model.Workflow) error {
	// 验证步骤定义
	if err := e.validateSteps(workflow.Steps); err != nil {
		return err
	}

	// 验证触发器配置
	if err := e.validateTrigger(workflow); err != nil {
		return err
	}

	return e.workflowRepo.Create(ctx, workflow)
}

// TriggerWorkflow 触发工作流
func (e *Engine) TriggerWorkflow(ctx context.Context, workflowID string, triggeredBy string, input map[string]interface{}) (*model.WorkflowExecution, error) {
	// 获取工作流
	workflow, err := e.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}

	// 检查工作流状态
	if workflow.Status != model.WorkflowStatusActive {
		return nil, ErrWorkflowInactive
	}

	// 创建执行记录
	execution := &model.WorkflowExecution{
		WorkflowID:  workflowID,
		TriggeredBy: triggeredBy,
		TriggerType: workflow.TriggerType,
		Input:       input,
		Status:      model.ExecutionStatusRunning,
		StartedAt:    time.Now().Unix(),
	Output:      make(model.JSONMap),
	}

	if err := e.executionRepo.Create(ctx, execution); err != nil {
		return nil, err
	}

	// 找到开始步骤并执行
	startStep := e.findStartStep(workflow.Steps)
	if startStep == nil && len(workflow.Steps) > 0 {
		// 如果没有明确的开始步骤，使用第一个步骤
		startStep = &workflow.Steps[0]
	}

	if startStep != nil {
		go e.executeStep(context.Background(), execution.ID, workflow, startStep, input)
	} else {
		// 没有步骤，直接完成
		execution.Status = model.ExecutionStatusSuccess
		execution.CompletedAt = timePtr(time.Now().Unix())
		e.executionRepo.Update(ctx, execution)
	}

	return execution, nil
}

// CompleteStep 完成步骤执行
func (e *Engine) CompleteStep(ctx context.Context, executionID string, result *StepResult) error {
	execution, err := e.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return err
	}

	if !result.Success {
		execution.Status = model.ExecutionStatusFailed
		execution.ErrorMessage = result.Error
		execution.CompletedAt = timePtr(time.Now().Unix())
		return e.executionRepo.Update(ctx, execution)
	}

	// 更新输出
	if execution.Output == nil {
		execution.Output = make(model.JSONMap)
	}
	for k, v := range result.Output {
		execution.Output[k] = v
	}

	// 获取下一
	if result.NextStep != nil && *result.NextStep != "" {
		// 继续执行下一
		workflow, err := e.workflowRepo.GetByID(ctx, execution.WorkflowID)
		if err != nil {
			return err
		}
		nextStep := e.findStepByID(workflow.Steps, *result.NextStep)
		if nextStep != nil {
			go e.executeStep(context.Background(), executionID, workflow, nextStep, result.Output)
		} else {
			// 没有下一，完成执行
			execution.Status = model.ExecutionStatusSuccess
			execution.CompletedAt = timePtr(time.Now().Unix())
		}
	} else {
		// 没有下一，完成执行
		execution.Status = model.ExecutionStatusSuccess
		execution.CompletedAt = timePtr(time.Now().Unix())
	}

	return e.executionRepo.Update(ctx, execution)
}

// CancelExecution 取消工作流执行
func (e *Engine) CancelExecution(ctx context.Context, executionID string) error {
	execution, err := e.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return err
	}

	if execution.IsFinished() {
		return errors.New("执行已结束")
	}


	execution.Status = model.ExecutionStatusCancelled
	execution.CompletedAt = timePtr(time.Now().Unix())
	return e.executionRepo.Update(ctx, execution)
}

// executeStep 执行
func (e *Engine) executeStep(ctx context.Context, executionID string, workflow *model.Workflow, step *model.WorkflowStep, input map[string]interface{}) {
	log := logger.Get()
	log.Info("执行工作流",
		"execution_id", executionID,
		"step_id", step.ID,
		"step_type", step.Type,
	)

	// 获取处理器
	handler, ok := e.handlers[step.Type]
	if !ok {
		// 对于任务，使用默认的任务回调
		if step.Type == "task" && e.taskCallback != nil {
			err := e.taskCallback(ctx, executionID, step.ID, step.Config)
			if err != nil {
				log.Error("任务回调失败", "error", err)
				e.CompleteStep(ctx, executionID, &StepResult{
					Success: false,
					Error:   err.Error(),
				})
			}
			return
		}

		log.Error("处理器未找到", "step_type", step.Type)
		e.CompleteStep(ctx, executionID, &StepResult{
			Success: false,
			Error:   fmt.Sprintf("类型 %s 未实现", step.Type),
		})
		return
	}

	// 执行
	result, err := handler.Execute(ctx, step, input)
	if err != nil {
		log.Error("执行失败", "error", err)
		e.CompleteStep(ctx, executionID, &StepResult{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// 确定下一
	if result.Success && result.NextStep == nil && step.NextStepID != nil {
		result.NextStep = step.NextStepID
	}

	// 完成
	if err := e.CompleteStep(ctx, executionID, result); err != nil {
		log.Error("完成失败", "error", err)
	}
}

// validateSteps 验证定义
func (e *Engine) validateSteps(steps model.WorkflowSteps) error {
	if len(steps) == 0 {
		return errors.New("工作流至少需要包含一个")
	}

	// 检查ID唯一性
	stepIDs := make(map[string]bool)
	for _, step := range steps {
		if step.ID == "" {
			return errors.New("ID不能为空")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("重复的ID: %s", step.ID)
		}
		stepIDs[step.ID] = true
	}

	return nil
}

// validateTrigger 验证触发器配置
func (e *Engine) validateTrigger(workflow *model.Workflow) error {
	switch workflow.TriggerType {
	case model.WorkflowTriggerKeyword:
		// 验证关键词配置
		keywords, ok := workflow.TriggerConfig["keywords"].([]interface{})
		if !ok || len(keywords) == 0 {
			return errors.New("关键词触发器需要配置关键词列表")
		}
	case model.WorkflowTriggerManual:
		// 手动触发无需额外配置
	case model.WorkflowTriggerSchedule:
		// 验证定时配置
		cron, ok := workflow.TriggerConfig["cron"].(string)
		if !ok || cron == "" {
			return errors.New("定时触发器需要配置 cron 表达式")
		}
	default:
		return fmt.Errorf("无效的触发器类型: %s", workflow.TriggerType)
	}
	return nil
}

// findStartStep 查找开始
func (e *Engine) findStartStep(steps model.WorkflowSteps) *model.WorkflowStep {
	for i := range steps {
		if steps[i].Type == "start" {
			return &steps[i]
		}
	}
	return nil
}

// findStepByID 根据ID查找
func (e *Engine) findStepByID(steps model.WorkflowSteps, id string) *model.WorkflowStep {
	for i := range steps {
		if steps[i].ID == id {
			return &steps[i]
		}
	}
	return nil
}

// 辅助函数
func timePtr(t int64) *int64 {
	return &t
}
