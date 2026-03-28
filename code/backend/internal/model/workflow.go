package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
)

// WorkflowTriggerType 触发器类型
type WorkflowTriggerType string

const (
	WorkflowTriggerKeyword WorkflowTriggerType = "keyword" // 关键词触发
	WorkflowTriggerManual  WorkflowTriggerType = "manual"  // 手动触发
	WorkflowTriggerSchedule WorkflowTriggerType = "schedule" // 定时触发
)

// Workflow 工作流模型
type Workflow struct {
	Base
	Name        string              `gorm:"size:100;not null" json:"name"`
	Description string              `gorm:"size:500" json:"description"`
	Status      WorkflowStatus      `gorm:"size:20;default:'active'" json:"status"`
	CreatedBy   string              `gorm:"size:36;index;not null" json:"created_by"`
	TriggerType WorkflowTriggerType `gorm:"size:20;not null" json:"trigger_type"`
	TriggerConfig JSONMap           `gorm:"type:text" json:"trigger_config"` // 触发器配置（JSON）
	Steps       WorkflowSteps       `gorm:"type:text" json:"steps"`          // 步骤列表（JSON）
}

// TableName 返回表名
func (Workflow) TableName() string {
	return "workflows"
}

// IsActive 是否激活
func (w *Workflow) IsActive() bool {
	return w.Status == WorkflowStatusActive
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`        // http, script, condition, delay
	Config      map[string]string `json:"config"`      // 步骤配置
	NextStepID  *string           `json:"next_step_id,omitempty"`
	OnErrorStep *string           `json:"on_error_step,omitempty"`
}

// WorkflowSteps 工作流步骤列表
type WorkflowSteps []WorkflowStep

// Value 实现 driver.Valuer 接口
func (s WorkflowSteps) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *WorkflowSteps) Scan(value interface{}) error {
	if value == nil {
		*s = WorkflowSteps{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("无法扫描 WorkflowSteps 类型")
	}

	return json.Unmarshal(bytes, s)
}

// JSONMap 通用 JSON Map 类型
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("无法扫描 JSONMap 类型")
	}

	return json.Unmarshal(bytes, m)
}

// WorkflowExecution 工作流执行记录
type WorkflowExecution struct {
	Base
	WorkflowID   string              `gorm:"size:36;index;not null" json:"workflow_id"`
	TriggeredBy  string              `gorm:"size:36;index;not null" json:"triggered_by"` // 员工ID或system
	TriggerType  WorkflowTriggerType `gorm:"size:20" json:"trigger_type"`
	Input        JSONMap             `gorm:"type:text" json:"input"`
	Status       ExecutionStatus     `gorm:"size:20;index;not null" json:"status"`
	Output       JSONMap             `gorm:"type:text" json:"output"`
	ErrorMessage string              `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    int64               `json:"started_at"`
	CompletedAt  *int64              `json:"completed_at,omitempty"`
}

// TableName 返回表名
func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// IsFinished 是否已完成
func (e *WorkflowExecution) IsFinished() bool {
	return e.Status == ExecutionStatusSuccess ||
		e.Status == ExecutionStatusFailed ||
		e.Status == ExecutionStatusCancelled
}

// WorkflowResponse 工作流响应结构
type WorkflowResponse struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Status        WorkflowStatus      `json:"status"`
	CreatedBy     string              `json:"created_by"`
	TriggerType   WorkflowTriggerType `json:"trigger_type"`
	TriggerConfig JSONMap             `json:"trigger_config"`
	Steps         WorkflowSteps       `json:"steps"`
	CreatedAt     string              `json:"created_at"`
}

// ToResponse 转换为响应结构
func (w *Workflow) ToResponse() WorkflowResponse {
	return WorkflowResponse{
		ID:            w.ID,
		Name:          w.Name,
		Description:   w.Description,
		Status:        w.Status,
		CreatedBy:     w.CreatedBy,
		TriggerType:   w.TriggerType,
		TriggerConfig: w.TriggerConfig,
		Steps:         w.Steps,
		CreatedAt:     w.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ExecutionResponse 执行记录响应结构
type ExecutionResponse struct {
	ID           string          `json:"id"`
	WorkflowID   string          `json:"workflow_id"`
	WorkflowName string          `json:"workflow_name"`
	TriggeredBy  string          `json:"triggered_by"`
	TriggerType  string          `json:"trigger_type"`
	Status       ExecutionStatus `json:"status"`
	ErrorMessage string          `json:"error_message,omitempty"`
	StartedAt    int64           `json:"started_at"`
	CompletedAt  *int64          `json:"completed_at,omitempty"`
}
