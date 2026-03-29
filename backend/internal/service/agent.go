// Package service 提供 Agent 业务逻辑层
package service

import (
	"context"
	"errors"
	"time"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
)

// AgentService Agent 服务
type AgentService struct {
	taskRepo   repository.TaskRepository
	empRepo    repository.EmployeeRepository
	msgService *MessageService
}

// NewAgentService 创建 Agent 服务
func NewAgentService() *AgentService {
	return &AgentService{
		taskRepo:   repository.NewTaskRepository(),
		empRepo:   repository.NewEmployeeRepository(),
		msgService: NewMessageService(),
	}
}

// CompleteTask Agent 完成任务
func (s *AgentService) CompleteTask(ctx context.Context, taskID string, agentID string, success bool, result map[string]interface{}, errMsg string) error {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrTaskNotFound
		}
		return err
	}

	// 验证 Agent 是任务的负责人
	if task.AssigneeID == nil || *task.AssigneeID != agentID {
		return ErrNotTaskAssignee
	}

	// 完成任务
	task.Complete()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// 如果任务来自工作流，通知工作流引擎
	if task.WorkflowID != nil {
		log := logger.Get()
		log.Info("任务完成，通知工作流",
			"task_id", taskID,
			"workflow_id", *task.WorkflowID,
		)
	}

	return nil
}

// GetAgentTasks 获取 Agent 的任务列表
func (s *AgentService) GetAgentTasks(ctx context.Context, agentID string) ([]*model.TaskResponse, error) {
	// 获取待处理和进行中的任务
	tasks, _, err := s.taskRepo.ListByAssignee(ctx, agentID, 1, 100)
	if err != nil {
		return nil, err
	}

	// 过滤出待处理和进行中的任务
	var activeTasks []*model.Task
	for _, t := range tasks {
		if t.Status == model.TaskStatusPending || t.Status == model.TaskStatusInProgress {
			activeTasks = append(activeTasks, t)
		}
	}

	// 转换为响应
	responses := make([]*model.TaskResponse, len(activeTasks))
	for i, t := range activeTasks {
		resp := t.ToResponse()
		responses[i] = &resp
	}

	return responses, nil
}

// GetTaskDetail 获取任务详情
func (s *AgentService) GetTaskDetail(ctx context.Context, taskID string, agentID string) (*model.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// 验证 Agent 可以访问此任务
	if task.AssigneeID == nil || *task.AssigneeID != agentID {
		return nil, ErrNotTaskAssignee
	}

	resp := task.ToResponse()
	return &resp, nil
}

// UpdateAgentStatus 更新 Agent 状态
func (s *AgentService) UpdateAgentStatus(ctx context.Context, agentID string, status string) error {
	emp, err := s.empRepo.GetByID(ctx, agentID)
	if err != nil {
		if err == repository.ErrNotFound {
			return errors.New("Agent 不存在")
		}
		return err
	}

	// 更新最后在线时间
	now := time.Now()
	emp.LastSeenAt = &now

	return s.empRepo.Update(ctx, emp)
}

// GetAgentInfo 获取 Agent 信息
func (s *AgentService) GetAgentInfo(ctx context.Context, agentID string) (map[string]interface{}, error) {
	emp, err := s.empRepo.GetByID(ctx, agentID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, errors.New("Agent 不存在")
		}
		return nil, err
	}

	// 获取任务统计
	pending, inProgress, completed, err := s.taskRepo.CountByAssignee(ctx, agentID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":        emp.ID,
		"name":      emp.Name,
		"email":     emp.Email,
		"type":      emp.Type,
		"skills":    emp.Skills,
		"status":    emp.Status,
		"last_seen": emp.LastSeenAt,
		"tasks": map[string]int64{
			"pending":     pending,
			"in_progress": inProgress,
			"completed":   completed,
		},
	}, nil
}

// SendMessage Agent 发送消息
func (s *AgentService) SendMessage(ctx context.Context, agentID string, channelID string, content string) (map[string]interface{}, error) {
	// 使用消息服务发送消息
	req := &SendMessageRequest{
		ChannelID: channelID,
		Content:   content,
		Type:      "text",
	}

	msg, err := s.msgService.Send(ctx, req, agentID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":         msg.ID,
		"channel_id": msg.ChannelID,
		"content":    msg.Content,
		"created_at": msg.CreatedAt,
	}, nil
}
