package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText     MessageType = "text"     // 文本消息
	MessageTypeImage    MessageType = "image"    // 图片
	MessageTypeFile     MessageType = "file"     // 文件
	MessageTypeSystem   MessageType = "system"   // 系统消息
	MessageTypeWorkflow MessageType = "workflow" // 工作流触发
)

// Message 消息模型
type Message struct {
	Base
	ChannelID   string      `gorm:"size:36;index;not null" json:"channel_id"`
	SenderID    string      `gorm:"size:36;index;not null" json:"sender_id"`
	Type        MessageType `gorm:"size:20;not null" json:"type"`
	Content     string      `gorm:"type:text" json:"content"`
	Mentions    StringArray `gorm:"type:text" json:"mentions"`    // JSON 数组存储提及的员工ID
	Skills      StringArray `gorm:"type:text" json:"skills"`      // JSON 数组存储触发的技能
	WorkflowID  *string     `gorm:"size:36;index" json:"workflow_id,omitempty"`
	ParentID    *string     `gorm:"size:36;index" json:"parent_id,omitempty"` // 回复的消息ID
	IsDeleted   bool        `gorm:"default:false" json:"is_deleted"`
	Sender      Employee    `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

// TableName 返回表名
func (Message) TableName() string {
	return "messages"
}

// IsWorkflowTrigger 是否为工作流触发消息
func (m *Message) IsWorkflowTrigger() bool {
	return m.Type == MessageTypeWorkflow || len(m.Skills) > 0
}

// StringArray 字符串数组类型（存储为 JSON）
type StringArray []string

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("无法扫描 StringArray 类型")
	}

	return json.Unmarshal(bytes, a)
}

// MessageResponse 消息响应结构
type MessageResponse struct {
	ID         string      `json:"id"`
	ChannelID  string      `json:"channel_id"`
	SenderID   string      `json:"sender_id"`
	SenderName string      `json:"sender_name"`
	Type       MessageType `json:"type"`
	Content    string      `json:"content"`
	Mentions   StringArray `json:"mentions"`
	Skills     StringArray `json:"skills"`
	WorkflowID *string     `json:"workflow_id,omitempty"`
	ParentID   *string     `json:"parent_id,omitempty"`
	CreatedAt  string      `json:"created_at"`
}

// ToResponse 转换为响应结构
func (m *Message) ToResponse() MessageResponse {
	senderName := ""
	if m.Sender.ID != "" {
		senderName = m.Sender.Name
	}

	return MessageResponse{
		ID:         m.ID,
		ChannelID:  m.ChannelID,
		SenderID:   m.SenderID,
		SenderName: senderName,
		Type:       m.Type,
		Content:    m.Content,
		Mentions:   m.Mentions,
		Skills:     m.Skills,
		WorkflowID: m.WorkflowID,
		ParentID:   m.ParentID,
		CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
