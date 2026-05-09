// Package model 提供数据模型定义
package model

import (
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// DocumentStatus 文档状态
type DocumentStatus string

const (
	DocumentStatusActive   DocumentStatus = "active"
	DocumentStatusArchived DocumentStatus = "archived"
	DocumentStatusDeleted  DocumentStatus = "deleted"
)

// Document 文档模型
type Document struct {
	Base
	ChannelID   string         `gorm:"size:36;index;not null" json:"channel_id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Summary     string         `gorm:"size:500" json:"summary"`
	AuthorID    string         `gorm:"size:36;index;not null" json:"author_id"`
	EditorID    *string        `gorm:"size:36;index" json:"editor_id"`
	Version     int            `gorm:"default:1" json:"version"`
	FileSize    int64          `gorm:"default:0" json:"file_size"`
	Status      DocumentStatus `gorm:"size:20;default:'active'" json:"status"`

	// 关联
	Channel  *Channel          `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Author   *Employee        `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Editor   *Employee        `gorm:"foreignKey:EditorID" json:"editor,omitempty"`
	Versions []DocumentVersion `gorm:"foreignKey:DocumentID" json:"versions,omitempty"`
}

// TableName 返回表名
func (Document) TableName() string {
	return "documents"
}

// BeforeCreate 创建前处理
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = GenerateID("doc")
	}

	// 计算摘要
	if d.Content != "" {
		d.Summary = ExtractSummary(d.Content, 200)
		d.FileSize = int64(len(d.Content))
	}

	return nil
}

// BeforeUpdate 更新前处理
func (d *Document) BeforeUpdate(tx *gorm.DB) error {
	// 更新时重新计算摘要和大小
	if d.Content != "" {
		d.Summary = ExtractSummary(d.Content, 200)
		d.FileSize = int64(len(d.Content))
	}
	return nil
}

// DocumentVersion 文档历史版本模型
type DocumentVersion struct {
	Base
	DocumentID string `gorm:"size:36;index;not null" json:"document_id"`
	Version    int    `gorm:"not null;index" json:"version"`
	Content    string `gorm:"type:text;not null" json:"content"`
	Summary    string `gorm:"size:500" json:"summary"`
	EditorID   string `gorm:"size:36;not null" json:"editor_id"`
	EditReason string `gorm:"size:200" json:"edit_reason"`

	// 关联
	Document *Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Editor   *Employee `gorm:"foreignKey:EditorID" json:"editor,omitempty"`
}

// TableName 返回表名
func (DocumentVersion) TableName() string {
	return "document_versions"
}

// BeforeCreate 创建前处理
func (v *DocumentVersion) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = GenerateID("ver")
	}

	// 计算摘要
	if v.Content != "" {
		v.Summary = ExtractSummary(v.Content, 200)
	}

	return nil
}

// ExtractSummary 提取摘要
func ExtractSummary(content string, maxLen int) string {
	// 去除 Markdown 语法
	plainText := RemoveMarkdownSyntax(content)
	plainText = strings.TrimSpace(plainText)

	if len(plainText) > maxLen {
		return plainText[:maxLen] + "..."
	}
	return plainText
}

// RemoveMarkdownSyntax 去除 Markdown 语法
func RemoveMarkdownSyntax(content string) string {
	// 移除标题标记
	result := regexp.MustCompile(`^#{1,6}\s*`).ReplaceAllString(content, "")
	// 移除粗体/斜体
	result = regexp.MustCompile(`[*_]{1,2}`).ReplaceAllString(result, "")
	// 移除链接标记，保留文本
	result = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(result, "$1")
	// 移除图片标记
	result = regexp.MustCompile(`!\[[^\]]*\]\([^\)]+\)`).ReplaceAllString(result, "")
	// 移除代码块标记
	result = regexp.MustCompile("`{1,3}").ReplaceAllString(result, "")
	// 移除列表标记
	result = regexp.MustCompile(`^[\s]*[-*+]\s*`).ReplaceAllString(result, "")
	result = regexp.MustCompile(`^[\s]*\d+\.\s*`).ReplaceAllString(result, "")
	// 移除引用标记
	result = regexp.MustCompile(`^>\s*`).ReplaceAllString(result, "")
	// 移除水平线
	result = regexp.MustCompile(`^[-*_]{3,}`).ReplaceAllString(result, "")

	return strings.TrimSpace(result)
}
