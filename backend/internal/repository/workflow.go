// Package repository 提供工作流数据访问层
package repository

import (
	"context"
	"errors"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

// WorkflowStats 工作流统计
type WorkflowStats struct {
	Active         int64
	Running      int64
	Paused       int64
	Failed       int64
	TotalExecutions int64
}


// WorkflowRepository 工作流 Repository 接口
type WorkflowRepository interface {
	Create(ctx context.Context, workflow *model.Workflow) error
	GetByID(ctx context.Context, id string) (*model.Workflow, error)
	Update(ctx context.Context, workflow *model.Workflow) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]*model.Workflow, int64, error)
	ListByStatus(ctx context.Context, status model.WorkflowStatus, page, pageSize int) ([]*model.Workflow, int64, error)
	SearchByName(ctx context.Context, keyword string, page, pageSize int) ([]*model.Workflow, int64, error)
	GetStats(ctx context.Context) (*WorkflowStats, error)
}

// WorkflowExecutionRepository 工作流执行 Repository 接口
type WorkflowExecutionRepository interface {
	Create(ctx context.Context, execution *model.WorkflowExecution) error
	GetByID(ctx context.Context, id string) (*model.WorkflowExecution, error)
	Update(ctx context.Context, execution *model.WorkflowExecution) error
	ListByWorkflow(ctx context.Context, workflowID string, page, pageSize int) ([]*model.WorkflowExecution, int64, error)
	ListByStatus(ctx context.Context, status model.ExecutionStatus, page, pageSize int) ([]*model.WorkflowExecution, int64, error)
	ListActive(ctx context.Context, page, pageSize int) ([]*model.WorkflowExecution, int64, error)
}

// workflowRepo 工作流 Repository 实现
type workflowRepo struct {
	db *gorm.DB
}

// workflowExecutionRepo 工作流执行 Repository 实现
type workflowExecutionRepo struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流 Repository
func NewWorkflowRepository() WorkflowRepository {
	return &workflowRepo{db: database.GetDB()}
}

// NewWorkflowExecutionRepository 创建工作流执行 Repository
func NewWorkflowExecutionRepository() WorkflowExecutionRepository {
	return &workflowExecutionRepo{db: database.GetDB()}
}

// Create 创建工作流
func (r *workflowRepo) Create(ctx context.Context, workflow *model.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

// GetByID 根据 ID 获取工作流
func (r *workflowRepo) GetByID(ctx context.Context, id string) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.WithContext(ctx).First(&workflow, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &workflow, nil
}

// Update 更新工作流
func (r *workflowRepo) Update(ctx context.Context, workflow *model.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

// Delete 删除工作流
func (r *workflowRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Workflow{}, "id = ?", id).Error
}

// List 获取工作流列表
func (r *workflowRepo) List(ctx context.Context, page, pageSize int) ([]*model.Workflow, int64, error) {
	var workflows []*model.Workflow
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Workflow{})
	db.Count(&total)

	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&workflows).Error
	if err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

// ListByStatus 根据状态获取工作流列表
func (r *workflowRepo) ListByStatus(ctx context.Context, status model.WorkflowStatus, page, pageSize int) ([]*model.Workflow, int64, error) {
	var workflows []*model.Workflow
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Workflow{}).Where("status = ?", status)
	db.Count(&total)

	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&workflows).Error
	if err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

// SearchByName 根据名称搜索工作流
func (r *workflowRepo) SearchByName(ctx context.Context, keyword string, page, pageSize int) ([]*model.Workflow, int64, error) {
	var workflows []*model.Workflow
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Workflow{}).Where("name LIKE ?", "%"+keyword+"%")
	db.Count(&total)

	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&workflows).Error
	if err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

// Create 创建工作流执行
func (r *workflowExecutionRepo) Create(ctx context.Context, execution *model.WorkflowExecution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

// GetByID 根据 ID 获取工作流执行
func (r *workflowExecutionRepo) GetByID(ctx context.Context, id string) (*model.WorkflowExecution, error) {
	var execution model.WorkflowExecution
	err := r.db.WithContext(ctx).First(&execution, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &execution, nil
}

// Update 更新工作流执行
func (r *workflowExecutionRepo) Update(ctx context.Context, execution *model.WorkflowExecution) error {
	return r.db.WithContext(ctx).Save(execution).Error
}

// ListByWorkflow 获取工作流执行列表
func (r *workflowExecutionRepo) ListByWorkflow(ctx context.Context, workflowID string, page, pageSize int) ([]*model.WorkflowExecution, int64, error) {
	var executions []*model.WorkflowExecution
	var total int64

	db := r.db.WithContext(ctx).Model(&model.WorkflowExecution{}).Where("workflow_id = ?", workflowID)
	db.Count(&total)

	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&executions).Error
	if err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

// ListByStatus 根据状态获取工作流执行列表
func (r *workflowExecutionRepo) ListByStatus(ctx context.Context, status model.ExecutionStatus, page, pageSize int) ([]*model.WorkflowExecution, int64, error) {
	var executions []*model.WorkflowExecution
	var total int64

	db := r.db.WithContext(ctx).Model(&model.WorkflowExecution{}).Where("status = ?", status)
	db.Count(&total)

	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&executions).Error
	if err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

// ListActive 获取活跃的工作流执行
func (r *workflowExecutionRepo) ListActive(ctx context.Context, page, pageSize int) ([]*model.WorkflowExecution, int64, error) {
	var executions []*model.WorkflowExecution
	var total int64

	db := r.db.WithContext(ctx).Model(&model.WorkflowExecution{}).Where("status = ?", model.ExecutionStatusRunning)
	db.Count(&total)

	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&executions).Error
	if err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

// GetStats 获取工作流统计
func (r *workflowRepo) GetStats(ctx context.Context) (*WorkflowStats, error) {
	stats := &WorkflowStats{}

	// 活跃工作流数 (status = active)
	if err := r.db.WithContext(ctx).Model(&model.Workflow{}).Where("status = ?", model.WorkflowStatusActive).Count(&stats.Active).Error; err != nil {
		return nil, err
	}

	// 运行中的执行数 (从 executions 表统计)
	if err := r.db.WithContext(ctx).Model(&model.WorkflowExecution{}).Where("status = ?", model.ExecutionStatusRunning).Count(&stats.Running).Error; err != nil {
		return nil, err
	}

	// 失败的执行数 (从 executions 表统计)
	if err := r.db.WithContext(ctx).Model(&model.WorkflowExecution{}).Where("status = ?", model.ExecutionStatusFailed).Count(&stats.Failed).Error; err != nil {
		return nil, err
	}

	// 总执行次数
	if err := r.db.WithContext(ctx).Model(&model.WorkflowExecution{}).Count(&stats.TotalExecutions).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
