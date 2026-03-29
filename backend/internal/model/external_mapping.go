package model

import "time"

// ExternalMappingType 映射类型
type ExternalMappingType string

const (
	ExternalMappingTypeChannel  ExternalMappingType = "channel"
	ExternalMappingTypeEmployee ExternalMappingType = "employee"
)

// ExternalMapping 外部系统映射表
// 用于存储外部系统（钉钉、飞书等）ID 与内部 ID 的映射关系
type ExternalMapping struct {
	ID          string              `gorm:"primaryKey;type:varchar(36)" json:"id"`
	SourceType  string              `gorm:"size:50;not null;index:idx_source" json:"source_type"`  // dingtalk, feishu, webhook
	ExternalID  string              `gorm:"size:255;not null;index:idx_external" json:"external_id"` // 外部系统 ID
	MappingType ExternalMappingType `gorm:"size:20;not null" json:"mapping_type"`                    // channel, employee
	InternalID  string              `gorm:"size:36;not null;index:idx_internal" json:"internal_id"`  // 内部 ID
	Name        string              `gorm:"size:200" json:"name"`                                    // 显示名称
	ExtraData   JSONMap             `gorm:"type:text" json:"extra_data,omitempty"`                   // 额外数据
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// TableName 指定表名
func (ExternalMapping) TableName() string {
	return "external_mappings"
}
