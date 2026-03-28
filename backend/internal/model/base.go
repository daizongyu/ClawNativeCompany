// Package model 提供数据模型定义
// 所有模型继承 Base 基础模型
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base 基础模型，所有模型嵌入此结构
type Base struct {
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate 创建前自动生成 UUID
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// IsDeleted 检查记录是否已软删除
func (b *Base) IsDeleted() bool {
	return b.DeletedAt.Valid
}

// Model 接口定义
type Model interface {
	GetID() string
}

// GetID 获取模型 ID
func (b Base) GetID() string {
	return b.ID
}
