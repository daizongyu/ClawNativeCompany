// Package service 提供任务业务逻辑层
package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
	"claw/internal/websocket"
)

// 任务服务错误
var (
	ErrTaskNotFound        = errors.New("任务不存在")
	ErrTaskAlreadyClaimed  = errors.New("任务已被认领")
	ErrTaskNotClaimable    = errors.New("任务不可认领")
	ErrInvalidTaskStatus   = errors.New("无效的任务状态")
	ErrInvalidPriority     = errors.New("无效的优先级")
	ErrNotTaskAssignee     = errors.New("不是任务负责人")
	ErrTaskAlreadyComplete = errors.New("任务已完成")
)

// TaskService 任务服务
type TaskService struct {
	taskRepo   repository.TaskRepository
	empRepo    repository.EmployeeRepository
	msgService *MessageService
	wsManager  *websocket.Manager
}

// NewTaskService 创建任务服务
func NewTaskService() *TaskService {
	return &TaskService{
		taskRepo:   repository.NewTaskRepository(),
		empRepo:    repository.NewEmployeeRepository(),
		msgService: NewMessageService(),
		wsManager:  websocket.GetManager(),
	}
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Title       string    `json:"title" validate:"required,min=2,max=200"`
	Description string    `json:"description" validate:"max=2000"`
	Priority  string    `json:"priority" validate:"required,oneof=low medium high urgent"`
	AssigneeID *string  `json:"assignee_id,omitempty"`
	DueDate    *string  `json:"due_date,omitempty"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Title       string   `json:"title" validate:"omitempty,min=2,max=200"`
	Description string   `json:"description" validate:"max:2000"`
	Priority  string   `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate   *string  `json:"due_date,omitempty"`
}

// TaskResponse 任务响应（使用 model.TaskResponse）

// Create 创建任务
func (s *TaskService) Create(ctx context.Context, req CreateTaskRequest, creatorID string) (*model.TaskResponse, error) {
	// 验证指派人
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		_, err := s.empRepo.GetByID(ctx, *req.AssigneeID)
		if err != nil {
			if err == repository.ErrNotFound {
				return nil, errors.New("指派的员工不存在")
			}
			return nil, err
		}
	}

	// 解析优先级
	priority := model.TaskPriority(req.Priority)

	// 解析截止日期
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, errors.New("无效的截止日期格式")
		}
		dueDate = &parsed
	}

	// 创建任务
	task := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatusPending,
		Priority:    priority,
		Source:      model.TaskSourceManual,
		AssigneeID:  req.AssigneeID,
		CreatorID:   creatorID,
		DueDate:     dueDate,
	}

	// 如果指定了指派人，直接设为进行中
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		task.Status = model.TaskStatusInProgress
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// 发送通知
	s.notifyTaskCreated(ctx, task)

	return s.toTaskResponse(ctx, task), nil
}

// GetByID 根据 ID 获取任务
func (s *TaskService) GetByID(ctx context.Context, id string) (*model.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return s.toTaskResponse(ctx, task), nil
}

// Update 更新任务
func (s *TaskService) Update(ctx context.Context, id string, req UpdateTaskRequest) (*model.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// 更新字段
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Priority != "" {
		task.Priority = model.TaskPriority(req.Priority)
	}
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, errors.New("无效的截止日期格式")
		}
		task.DueDate = &parsed
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return s.toTaskResponse(ctx, task), nil
}

// Delete 删除任务
func (s *TaskService) Delete(ctx context.Context, id string) error {
	_, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrTaskNotFound
		}
		return err
	}

	return s.taskRepo.Delete(ctx, id)
}

// List 获取任务列表
func (s *TaskService) List(ctx context.Context, page, pageSize int) ([]*model.TaskResponse, int64, error) {
	tasks, total, err := s.taskRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*model.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = s.toTaskResponse(ctx, t)
	}

	return responses, total, nil
}

// ListByAssignee 获取指派给某员工的任务
func (s *TaskService) ListByAssignee(ctx context.Context, assigneeID string, page, pageSize int) ([]*model.TaskResponse, int64, error) {
	tasks, total, err := s.taskRepo.ListByAssignee(ctx, assigneeID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*model.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = s.toTaskResponse(ctx, t)
	}

	return responses, total, nil
}

// ListByStatus 根据状态获取任务
func (s *TaskService) ListByStatus(ctx context.Context, status string, page, pageSize int) ([]*model.TaskResponse, int64, error) {
	taskStatus := model.TaskStatus(status)
	tasks, total, err := s.taskRepo.ListByStatus(ctx, taskStatus, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*model.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = s.toTaskResponse(ctx, t)
	}

	return responses, total, nil
}

// ListUnclaimed 获取未认领任务池
func (s *TaskService) ListUnclaimed(ctx context.Context, page, pageSize int) ([]*model.TaskResponse, int64, error) {
	tasks, total, err := s.taskRepo.ListUnclaimed(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*model.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = s.toTaskResponse(ctx, t)
	}

	return responses, total, nil
}

// Search 搜索任务
func (s *TaskService) Search(ctx context.Context, keyword string, status, priority *string, page, pageSize int) ([]*model.TaskResponse, int64, error) {
	var taskStatus *model.TaskStatus
	var taskPriority *model.TaskPriority

	if status != nil {
		ts := model.TaskStatus(*status)
		taskStatus = &ts
	}

	if priority != nil {
		tp := model.TaskPriority(*priority)
		taskPriority = &tp
	}

	tasks, total, err := s.taskRepo.Search(ctx, keyword, taskStatus, taskPriority, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*model.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = s.toTaskResponse(ctx, t)
	}

	return responses, total, nil
}

// ClaimTask 认领任务
func (s *TaskService) ClaimTask(ctx context.Context, taskID string, employeeID string) (*model.TaskResponse, error) {
	// 使用事务认领任务
	if err := s.taskRepo.ClaimTask(ctx, taskID, employeeID); err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		if err.Error() == "任务已被认领" {
			return nil, ErrTaskAlreadyClaimed
		}
		if err.Error() == "任务不在待处理状态" {
			return nil, ErrTaskNotClaimable
		}
		return nil, err
	}

	// 获取更新后的任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 发送通知
	s.notifyTaskClaimed(ctx, task, employeeID)

	return s.toTaskResponse(ctx, task), nil
}

// AssignTask 指派任务
func (s *TaskService) AssignTask(ctx context.Context, taskID string, assigneeID string) (*model.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// 验证指派人
	_, err = s.empRepo.GetByID(ctx, assigneeID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, errors.New("指派的员工不存在")
		}
		return nil, err
	}

	// 更新任务
	task.AssigneeID = &assigneeID
	if task.Status == model.TaskStatusPending {
		task.Status = model.TaskStatusInProgress
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// 发送通知
	s.notifyTaskAssigned(ctx, task, assigneeID)

	return s.toTaskResponse(ctx, task), nil
}

// CompleteTask 完成任务
func (s *TaskService) CompleteTask(ctx context.Context, taskID string, employeeID string, result map[string]interface{}) (*model.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// 检查权限
	if task.AssigneeID == nil || *task.AssigneeID != employeeID {
		return nil, ErrNotTaskAssignee
	}

	// 检查状态
	if task.Status == model.TaskStatusCompleted {
		return nil, ErrTaskAlreadyComplete
	}

	// 完成任务
	task.Complete()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// 发送通知
	s.notifyTaskCompleted(ctx, task)

	return s.toTaskResponse(ctx, task), nil
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(ctx context.Context, taskID string) (*model.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if task.Status == model.TaskStatusCompleted {
		return nil, ErrTaskAlreadyComplete
	}

	task.Status = model.TaskStatusCancelled
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return s.toTaskResponse(ctx, task), nil
}

// GetMyTasks 获取我的任务
func (s *TaskService) GetMyTasks(ctx context.Context, employeeID string, page, pageSize int) ([]*model.TaskResponse, int64, error) {
	return s.ListByAssignee(ctx, employeeID, page, pageSize)
}

// GetTaskStats 获取任务统计
func (s *TaskService) GetTaskStats(ctx context.Context, employeeID string) (map[string]interface{}, error) {
	pending, inProgress, completed, err := s.taskRepo.CountByAssignee(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"pending":     pending,
		"in_progress": inProgress,
		"completed":   completed,
		"total":       pending + inProgress + completed,
	}, nil
}

// CreateFromWorkflow 从工作流创建任务（由工作流引擎调用）
func (s *TaskService) CreateFromWorkflow(ctx context.Context, workflowID string, config map[string]interface{}) (*model.TaskResponse, error) {
	// 解析配置
	title, _ := config["title"].(string)
	description, _ := config["description"].(string)
	priorityStr, _ := config["priority"].(string)
	assigneeID, _ := config["assignee_id"].(string)

	if title == "" {
		title = "工作流任务"
	}

	priority := model.TaskPriorityMedium
	if priorityStr != "" {
		priority = model.TaskPriority(priorityStr)
	}

	var assigneeIDPtr *string
	if assigneeID != "" {
		assigneeIDPtr = &assigneeID
	}

	task := &model.Task{
		Title:       title,
		Description: description,
		Status:      model.TaskStatusPending,
		Priority:    priority,
		Source:      model.TaskSourceWorkflow,
		AssigneeID:  assigneeIDPtr,
		CreatorID:   "system",
		WorkflowID:  &workflowID,
	}

	if assigneeIDPtr != nil {
		task.Status = model.TaskStatusInProgress
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// 发送通知
	s.notifyTaskCreated(ctx, task)

	return s.toTaskResponse(ctx, task), nil
}

// toTaskResponse 转换为任务响应
func (s *TaskService) toTaskResponse(ctx context.Context, t *model.Task) *model.TaskResponse {
	resp := t.ToResponse()
	return &resp
}

// notifyTaskCreated 发送任务创建通知
func (s *TaskService) notifyTaskCreated(ctx context.Context, task *model.Task) {
	log := logger.Get()

	// 如果指定了指派人，通知被指派人
	if task.AssigneeID != nil && s.wsManager != nil {
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "task_assigned",
			"data": map[string]interface{}{
				"task_id":  task.ID,
				"title":    task.Title,
				"priority": task.Priority,
			},
		})
		s.wsManager.BroadcastToUser(*task.AssigneeID, msg)
	}

	log.Info("任务创建通知已发送", "task_id", task.ID)
}

// notifyTaskClaimed 发送任务认领通知
func (s *TaskService) notifyTaskClaimed(ctx context.Context, task *model.Task, employeeID string) {
	log := logger.Get()

	if s.wsManager != nil {
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "task_claimed",
			"data": map[string]interface{}{
				"task_id": task.ID,
				"title":   task.Title,
			},
		})
		s.wsManager.BroadcastToUser(employeeID, msg)
	}

	log.Info("任务认领通知已发送", "task_id", task.ID, "employee_id", employeeID)
}

// notifyTaskAssigned 发送任务指派通知
func (s *TaskService) notifyTaskAssigned(ctx context.Context, task *model.Task, assigneeID string) {
	log := logger.Get()

	if s.wsManager != nil {
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "task_assigned",
			"data": map[string]interface{}{
				"task_id":  task.ID,
				"title":    task.Title,
				"priority": task.Priority,
			},
		})
		s.wsManager.BroadcastToUser(assigneeID, msg)
	}

	log.Info("任务指派通知已发送", "task_id", task.ID, "assignee_id", assigneeID)
}

// notifyTaskCompleted 发送任务完成通知
func (s *TaskService) notifyTaskCompleted(ctx context.Context, task *model.Task) {
	log := logger.Get()

	// 通知创建人
	if s.wsManager != nil && task.CreatorID != "" && task.CreatorID != "system" {
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "task_completed",
			"data": map[string]interface{}{
				"task_id":  task.ID,
				"title":    task.Title,
				"assignee": task.AssigneeID,
			},
		})
		s.wsManager.BroadcastToUser(task.CreatorID, msg)
	}

	log.Info("任务完成通知已发送", "task_id", task.ID)
}
