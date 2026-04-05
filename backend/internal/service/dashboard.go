// Package service 提供业务逻辑层
// Dashboard Service 处理仪表盘统计业务
package service

import (
	"context"

	"claw/internal/repository"
)

// DashboardService 仪表盘服务接口
type DashboardService interface {
	GetStats(ctx context.Context) (*DashboardStats, error)
}

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	EmployeeCount        int64 `json:"employee_count"`
	ChannelCount       int64 `json:"channel_count"`
	PendingTaskCount int64 `json:"pending_task_count"`
	ActiveWorkflowCount int64 `json:"active_workflow_count"`
	TotalTasks       int64 `json:"total_tasks"`
	CompletedTasks   int64 `json:"completed_tasks"`
	TaskCompletionRate float64 `json:"task_completion_rate"`
	TodayNewTasks    int64 `json:"today_new_tasks"`
	TodayCompletedTasks int64 `json:"today_completed_tasks"`
	RunningWorkflows  int64 `json:"running_workflows"`
	PausedWorkflows   int64 `json:"paused_workflows"`
	FailedWorkflows   int64 `json:"failed_workflows"`
	TotalExecutions   int64 `json:"total_executions"`
}

// dashboardService Dashboard 服务实现
type dashboardService struct {
	employeeRepo repository.EmployeeRepository
	taskRepo     repository.TaskRepository
	channelRepo  repository.ChannelRepository
	workflowRepo repository.WorkflowRepository
}

// NewDashboardService 创建 Dashboard 服务
func NewDashboardService(
	employeeRepo repository.EmployeeRepository,
	taskRepo repository.TaskRepository,
	channelRepo repository.ChannelRepository,
	workflowRepo repository.WorkflowRepository,
) DashboardService {
	return &dashboardService{
		employeeRepo: employeeRepo,
		taskRepo:     taskRepo,
		channelRepo:  channelRepo,
		workflowRepo: workflowRepo,
	}
}

// GetStats 获取仪表盘统计数据
func (s *dashboardService) GetStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 员工统计
	employeeCount, err := s.employeeRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.EmployeeCount = employeeCount

	// 频道统计
	channelCount, err := s.channelRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ChannelCount = channelCount

	// 任务统计
	taskStats, err := s.taskRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalTasks = taskStats.Total
	stats.CompletedTasks = taskStats.Completed
	stats.PendingTaskCount = taskStats.Pending
	stats.TodayNewTasks = taskStats.TodayNew
	stats.TodayCompletedTasks = taskStats.TodayCompleted
	if taskStats.Total > 0 {
		stats.TaskCompletionRate = float64(taskStats.Completed) / float64(taskStats.Total) * 100
	}

	// 工作流统计
	workflowStats, err := s.workflowRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	stats.ActiveWorkflowCount = workflowStats.Active
	stats.RunningWorkflows = workflowStats.Running
	stats.PausedWorkflows = workflowStats.Paused
	stats.FailedWorkflows = workflowStats.Failed
	stats.TotalExecutions = workflowStats.TotalExecutions

	return stats, nil
}
