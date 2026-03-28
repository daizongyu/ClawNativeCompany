package model

import (
	"time"
)

// EmployeeType 员工类型
type EmployeeType string

const (
	EmployeeTypeHuman EmployeeType = "human" // 人类员工
	EmployeeTypeAgent EmployeeType = "agent" // AI Agent
)

// EmployeeStatus 员工状态
type EmployeeStatus string

const (
	EmployeeStatusActive   EmployeeStatus = "active"
	EmployeeStatusInactive EmployeeStatus = "inactive"
)

// Employee 员工模型
type Employee struct {
	Base
	Name       string         `gorm:"size:100;not null" json:"name"`
	Type       EmployeeType   `gorm:"size:20;not null" json:"type"`           // human | agent
	Email      string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password   string         `gorm:"size:255" json:"-"`                      // 人类必填，Agent 可选
	Skills     string         `gorm:"type:text" json:"skills"`                // JSON 数组: ["Go", "React"]
	Status     EmployeeStatus `gorm:"size:20;default:'active'" json:"status"`
	LastSeenAt *time.Time     `json:"last_seen_at,omitempty"`
	APIKey     string         `gorm:"size:255" json:"-"`                      // Agent 认证用
}

// TableName 返回表名
func (Employee) TableName() string {
	return "employees"
}

// IsHuman 是否为人类员工
func (e *Employee) IsHuman() bool {
	return e.Type == EmployeeTypeHuman
}

// IsAgent 是否为 AI Agent
func (e *Employee) IsAgent() bool {
	return e.Type == EmployeeTypeAgent
}

// IsActive 检查是否激活
func (e *Employee) IsActive() bool {
	return e.Status == EmployeeStatusActive
}

// UpdateLastSeen 更新最后在线时间
func (e *Employee) UpdateLastSeen() {
	now := time.Now()
	e.LastSeenAt = &now
}

// EmployeeResponse 员工响应结构（脱敏）
type EmployeeResponse struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       EmployeeType   `json:"type"`
	Email      string         `json:"email"`
	Skills     string         `json:"skills"`
	Status     EmployeeStatus `json:"status"`
	LastSeenAt *time.Time     `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ToResponse 转换为响应结构
func (e *Employee) ToResponse() EmployeeResponse {
	return EmployeeResponse{
		ID:         e.ID,
		Name:       e.Name,
		Type:       e.Type,
		Email:      e.Email,
		Skills:     e.Skills,
		Status:     e.Status,
		LastSeenAt: e.LastSeenAt,
		CreatedAt:  e.CreatedAt,
	}
}
