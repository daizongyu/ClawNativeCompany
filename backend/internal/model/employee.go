package model

import (
	"encoding/json"
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
	// 基础信息
	Username    string         `gorm:"size:50;uniqueIndex" json:"username"` // 用户名，唯一标识
	DisplayName string         `gorm:"size:100" json:"display_name"`        // 显示名
	Name        string         `gorm:"size:100;not null" json:"name"`                // 保留兼容
	Type        EmployeeType   `gorm:"size:20;not null" json:"type"`                 // human | agent
	Email       string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password    string         `gorm:"size:255" json:"-"`                            // 人类必填，Agent 可选

	// 职能信息
	Role   string `gorm:"size:100" json:"role"`     // 职能：开发、产品、设计等
	Skills string `gorm:"type:text" json:"skills"`  // JSON 数组: ["Go", "React"]

	// 扩展资料
	Avatar     string `gorm:"size:500" json:"avatar"`     // 头像URL
	Department string `gorm:"size:100" json:"department"` // 部门
	Position   string `gorm:"size:100" json:"position"`   // 职位
	Phone      string `gorm:"size:20" json:"phone"`        // 电话

	// 通知偏好（JSON 存储）
	NotificationPrefs string `gorm:"type:text" json:"notification_prefs"` // JSON: {"email":true,"webhook":true,"internal":true}

	Status     EmployeeStatus `gorm:"size:20;default:'active'" json:"status"`
	LastSeenAt *time.Time     `json:"last_seen_at,omitempty"`
	APIKey     string         `gorm:"size:255" json:"-"`                      // Agent 认证用

	// 关联
	GatewayConfigs []GatewayConfig `gorm:"foreignKey:EmployeeID" json:"gateway_configs,omitempty"`
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

// GetNotificationPrefs 获取通知偏好
func (e *Employee) GetNotificationPrefs() NotificationPreferences {
	if e.NotificationPrefs == "" {
		return DefaultNotificationPreferences()
	}

	var prefs NotificationPreferences
	if err := json.Unmarshal([]byte(e.NotificationPrefs), &prefs); err != nil {
		return DefaultNotificationPreferences()
	}
	return prefs
}

// SetNotificationPrefs 设置通知偏好
func (e *Employee) SetNotificationPrefs(prefs NotificationPreferences) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	e.NotificationPrefs = string(data)
	return nil
}

// NotificationPreferences 通知偏好结构
type NotificationPreferences struct {
	Channels NotificationChannels `json:"channels"` // 通知渠道开关
	Events   EventNotifications   `json:"events"`   // 事件通知开关
}

// NotificationChannels 通知渠道
type NotificationChannels struct {
	Email    bool `json:"email"`    // 邮件通知
	Webhook  bool `json:"webhook"`  // Webhook 通知
	Internal bool `json:"internal"` // 站内信通知
}

// EventNotifications 事件通知
type EventNotifications struct {
	TaskAssigned     bool `json:"task_assigned"`     // 任务分配
	TaskCompleted    bool `json:"task_completed"`    // 任务完成
	Mentioned        bool `json:"mentioned"`         // 被提及
	WorkflowExecuted bool `json:"workflow_executed"` // 工作流执行
}

// DefaultNotificationPreferences 默认通知偏好
func DefaultNotificationPreferences() NotificationPreferences {
	return NotificationPreferences{
		Channels: NotificationChannels{
			Email:    true,
			Webhook:  true,
			Internal: true,
		},
		Events: EventNotifications{
			TaskAssigned:     true,
			TaskCompleted:    true,
			Mentioned:        true,
			WorkflowExecuted: false,
		},
	}
}

// EmployeeResponse 员工响应结构（脱敏）
type EmployeeResponse struct {
	ID                  string         `json:"id"`
	Username            string         `json:"username"`
	DisplayName         string         `json:"display_name"`
	Name                string         `json:"name"` // 兼容旧字段
	Type                EmployeeType   `json:"type"`
	Email               string         `json:"email"`
	Role                string         `json:"role"`
	Skills              string         `json:"skills"`
	Avatar              string         `json:"avatar"`
	Department          string         `json:"department"`
	Position            string         `json:"position"`
	Phone               string         `json:"phone"`
	NotificationPrefs   string         `json:"notification_prefs"`
	Status              EmployeeStatus `json:"status"`
	LastSeenAt          *time.Time     `json:"last_seen_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	GatewayConfigs      []GatewayConfigResponse `json:"gateway_configs,omitempty"`
}

// ToResponse 转换为响应结构
func (e *Employee) ToResponse() EmployeeResponse {
	resp := EmployeeResponse{
		ID:                e.ID,
		Username:          e.Username,
		DisplayName:       e.DisplayName,
		Name:              e.Name,
		Type:              e.Type,
		Email:             e.Email,
		Role:              e.Role,
		Skills:            e.Skills,
		Avatar:            e.Avatar,
		Department:        e.Department,
		Position:          e.Position,
		Phone:             e.Phone,
		NotificationPrefs: e.NotificationPrefs,
		Status:            e.Status,
		LastSeenAt:        e.LastSeenAt,
		CreatedAt:         e.CreatedAt,
	}

	// 转换 GatewayConfigs
	if len(e.GatewayConfigs) > 0 {
		configs := make([]GatewayConfigResponse, len(e.GatewayConfigs))
		for i, cfg := range e.GatewayConfigs {
			configs[i] = cfg.ToResponse()
		}
		resp.GatewayConfigs = configs
	}

	return resp
}

// ToResponseWithBasic 转换为基本响应结构（不含敏感信息）
func (e *Employee) ToResponseWithBasic() EmployeeResponse {
	return EmployeeResponse{
		ID:          e.ID,
		Username:    e.Username,
		DisplayName: e.DisplayName,
		Name:        e.Name,
		Type:        e.Type,
		Avatar:      e.Avatar,
		Department:  e.Department,
		Position:    e.Position,
		Status:      e.Status,
	}
}
