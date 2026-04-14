// Package repository 提供任务数据访问层
package repository

import (
	"context"
	"errors"
	"time"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

// TaskStats 任务统计
type TaskStats struct {
	Total           int64
	Completed     int64
	Pending       int64
	TodayNew      int64
	TodayCompleted int64
}

// TaskRepository 任务 Repository 接口
// ListTaskFilter 任务列表筛选条件
type ListTaskFilter struct {
	Status      string
	Priority    string
	AssigneeID  string
	Unclaimed   bool
	Keyword     string
}

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id string) (*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]*model.Task, int64, error)
	ListWithFilter(ctx context.Context, filter ListTaskFilter, page, pageSize int) ([]*model.Task, int64, error)
	ListByAssignee(ctx context.Context, assigneeID string, page, pageSize int) ([]*model.Task, int64, error)
	ListByStatus(ctx context.Context, status model.TaskStatus, page, pageSize int) ([]*model.Task, int64, error)
	ListByWorkflow(ctx context.Context, workflowID string, page, pageSize int) ([]*model.Task, int64, error)
	ListUnclaimed(ctx context.Context, page, pageSize int) ([]*model.Task, int64, error)
	ListBySource(ctx context.Context, source model.TaskSource, page, pageSize int) ([]*model.Task, int64, error)
	Search(ctx context.Context, keyword string, status *model.TaskStatus, priority *model.TaskPriority, page, pageSize int) ([]*model.Task, int64, error)
	ClaimTask(ctx context.Context, taskID string, employeeID string) error
	CountByStatus(ctx context.Context, status model.TaskStatus) (int64, error)
	CountByAssignee(ctx context.Context, assigneeID string) (int64, int64, int64, error)
	GetStats(ctx context.Context) (*TaskStats, error)
}

// taskRepo 任务 Repository 实现
type taskRepo struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务 Repository
func NewTaskRepository() TaskRepository {
	return &taskRepo{db: database.GetDB()}
}

// Create 创建任务
func (r *taskRepo) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据 ID 获取任务
func (r *taskRepo) GetByID(ctx context.Context, id string) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).Preload("Creator").Preload("Assignee").First(&task, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &task, nil
}

// Update 更新任务
func (r *taskRepo) Update(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete 删除任务
func (r *taskRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Task{}, "id = ?", id).Error
}

// List 获取任务列表
func (r *taskRepo) List(ctx context.Context, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{})
	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListWithFilter 根据筛选条件获取任务列表
func (r *taskRepo) ListWithFilter(ctx context.Context, filter ListTaskFilter, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{})

	// 状态筛选
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	// 优先级筛选
	if filter.Priority != "" {
		db = db.Where("priority = ?", filter.Priority)
	}

	// 指派给特定员工
	if filter.AssigneeID != "" {
		db = db.Where("assignee_id = ?", filter.AssigneeID)
	}

	// 待认领任务（assignee_id 为空）
	if filter.Unclaimed {
		db = db.Where("assignee_id IS NULL OR assignee_id = ''")
	}

	// 关键词搜索
	if filter.Keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListByAssignee 获取指派给某员工的任务
func (r *taskRepo) ListByAssignee(ctx context.Context, assigneeID string, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{}).Where("assignee_id = ?", assigneeID)
	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListByStatus 根据状态获取任务
func (r *taskRepo) ListByStatus(ctx context.Context, status model.TaskStatus, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{}).Where("status = ?", status)
	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListByWorkflow 获取工作流相关的任务
func (r *taskRepo) ListByWorkflow(ctx context.Context, workflowID string, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{}).Where("workflow_id = ?", workflowID)
	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListUnclaimed 获取未认领的任务池
func (r *taskRepo) ListUnclaimed(ctx context.Context, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{}).
		Where("assignee_id IS NULL").
		Where("status = ?", model.TaskStatusPending)
	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("priority DESC, created_at ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListBySource 根据来源获取任务
func (r *taskRepo) ListBySource(ctx context.Context, source model.TaskSource, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{}).Where("source = ?", source)
	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Search 搜索任务
func (r *taskRepo) Search(ctx context.Context, keyword string, status *model.TaskStatus, priority *model.TaskPriority, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Task{})

	if keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if priority != nil {
		db = db.Where("priority = ?", *priority)
	}

	db.Count(&total)

	err := db.Preload("Creator").Preload("Assignee").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ClaimTask 认领任务（使用事务防止并发问题）
func (r *taskRepo) ClaimTask(ctx context.Context, taskID string, employeeID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询任务
		var task model.Task
		if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		// 检查任务是否已被认领
		if task.AssigneeID != nil && *task.AssigneeID != "" {
			return errors.New("任务已被认领")
		}

		// 检查任务状态
		if task.Status != model.TaskStatusPending {
			return errors.New("任务不在待处理状态")
		}

		// 更新任务
		task.AssigneeID = &employeeID
		task.Status = model.TaskStatusInProgress

		return tx.Save(&task).Error
	})
}

// CountByStatus 统计某状态的任务数
func (r *taskRepo) CountByStatus(ctx context.Context, status model.TaskStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Task{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountByAssignee 统计某员工的任务数
func (r *taskRepo) CountByAssignee(ctx context.Context, assigneeID string) (int64, int64, int64, error) {
	var pending, inProgress, completed int64

	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("assignee_id = ? AND status = ?", assigneeID, model.TaskStatusPending).Count(&pending).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("assignee_id = ? AND status = ?", assigneeID, model.TaskStatusInProgress).Count(&inProgress).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("assignee_id = ? AND status = ?", assigneeID, model.TaskStatusCompleted).Count(&completed).Error; err != nil {
		return 0, 0, 0, err
	}

	return pending, inProgress, completed, nil
}

// GetStats 获取任务统计
func (r *taskRepo) GetStats(ctx context.Context) (*TaskStats, error) {
	stats := &TaskStats{}

	// 总任务数
	if err := r.db.WithContext(ctx).Model(&model.Task{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 已完成任务数
	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("status = ?", model.TaskStatusCompleted).Count(&stats.Completed).Error; err != nil {
		return nil, err
	}

	// 待处理任务数
	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("status = ?", model.TaskStatusPending).Count(&stats.Pending).Error; err != nil {
		return nil, err
	}

	// 今日新建任务数
	today := time.Now().Format("2006-01-01")
	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("DATE(created_at) = ?", today).Count(&stats.TodayNew).Error; err != nil {
		return nil, err
	}

	// 今日完成任务数
	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("status = ? AND DATE(updated_at) = ?", model.TaskStatusCompleted, today).Count(&stats.TodayCompleted).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
