package model

import (
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // 待处理
	TaskStatusInProgress TaskStatus = "in_progress" // 进行中
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusCancelled  TaskStatus = "cancelled"  // 已取消
)

// TaskPriority 任务优先级
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// TaskSource 任务来源
type TaskSource string

const (
	TaskSourceManual   TaskSource = "manual"   // 手动创建
	TaskSourceWorkflow TaskSource = "workflow" // 工作流生成
	TaskSourceMention  TaskSource = "mention"  // 提及生成
)

// Task 任务模型
type Task struct {
	Base
	Title       string       `gorm:"size:200;not null" json:"title"`
	Description string       `gorm:"type:text" json:"description"`
	Status      TaskStatus   `gorm:"size:20;default:'pending';index" json:"status"`
	Priority    TaskPriority `gorm:"size:20;default:'medium'" json:"priority"`
	Source      TaskSource   `gorm:"size:20;not null" json:"source"`
	ChannelID   *string      `gorm:"size:36;index" json:"channel_id,omitempty"`
	MessageID   *string      `gorm:"size:36;index" json:"message_id,omitempty"`
	WorkflowID  *string      `gorm:"size:36;index" json:"workflow_id,omitempty"`
	AssigneeID  *string      `gorm:"size:36;index" json:"assignee_id,omitempty"`
	CreatorID   string       `gorm:"size:36;index;not null" json:"creator_id"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Assignee    *Employee    `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Creator     Employee     `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// TableName 返回表名
func (Task) TableName() string {
	return "tasks"
}

// IsOverdue 是否已逾期
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// CanComplete 是否可以完成
func (t *Task) CanComplete() bool {
	return t.Status == TaskStatusPending || t.Status == TaskStatusInProgress
}

// Complete 完成任务
func (t *Task) Complete() {
	if t.CanComplete() {
		t.Status = TaskStatusCompleted
		now := time.Now()
		t.CompletedAt = &now
	}
}

// TaskResponse 任务响应结构
type TaskResponse struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Source      TaskSource   `json:"source"`
	ChannelID   *string      `json:"channel_id,omitempty"`
	MessageID   *string      `json:"message_id,omitempty"`
	WorkflowID  *string      `json:"workflow_id,omitempty"`
	AssigneeID  *string      `json:"assignee_id,omitempty"`
	AssigneeName string      `json:"assignee_name,omitempty"`
	CreatorID   string       `json:"creator_id"`
	CreatorName string       `json:"creator_name"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	IsOverdue   bool         `json:"is_overdue"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	CreatedAt   string       `json:"created_at"`
}

// ToResponse 转换为响应结构
func (t *Task) ToResponse() TaskResponse {
	assigneeName := ""
	if t.Assignee != nil && t.Assignee.ID != "" {
		assigneeName = t.Assignee.Name
	}

	creatorName := ""
	if t.Creator.ID != "" {
		creatorName = t.Creator.Name
	}

	return TaskResponse{
		ID:           t.ID,
		Title:        t.Title,
		Description:  t.Description,
		Status:       t.Status,
		Priority:     t.Priority,
		Source:       t.Source,
		ChannelID:    t.ChannelID,
		MessageID:    t.MessageID,
		WorkflowID:   t.WorkflowID,
		AssigneeID:   t.AssigneeID,
		AssigneeName: assigneeName,
		CreatorID:    t.CreatorID,
		CreatorName:  creatorName,
		DueDate:      t.DueDate,
		IsOverdue:    t.IsOverdue(),
		CompletedAt:  t.CompletedAt,
		CreatedAt:    t.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// TaskCount 任务统计
type TaskCount struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	InProgress int64 `json:"in_progress"`
	Completed  int64 `json:"completed"`
	Overdue    int64 `json:"overdue"`
}
