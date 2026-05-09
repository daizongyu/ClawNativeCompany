// Package repository 提供文档数据访问层
package repository

import (
	"context"
	"errors"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

// DocumentFilter 文档查询筛选条件
type DocumentFilter struct {
	Keyword string
	Sort    string // updated_at | title | file_size
	Order   string // asc | desc
}

// DocumentRepository 文档 Repository 接口
type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	GetByID(ctx context.Context, id string) (*model.Document, error)
	Update(ctx context.Context, doc *model.Document) error
	Delete(ctx context.Context, id string) error
	ListByChannel(ctx context.Context, channelID string, filter *DocumentFilter, page, pageSize int) ([]*model.Document, int64, error)
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Document, int64, error)
	CountByChannel(ctx context.Context, channelID string) (int64, error)
}

// documentRepo 文档 Repository 实现
type documentRepo struct {
	db *gorm.DB
}

// NewDocumentRepository 创建文档 Repository
func NewDocumentRepository() DocumentRepository {
	return &documentRepo{db: database.GetDB()}
}

// Create 创建文档
func (r *documentRepo) Create(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// GetByID 根据ID获取文档
func (r *documentRepo) GetByID(ctx context.Context, id string) (*model.Document, error) {
	var doc model.Document
	err := r.db.WithContext(ctx).
		Preload("Channel").
		Preload("Author").
		Preload("Editor").
		First(&doc, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// Update 更新文档
func (r *documentRepo) Update(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

// Delete 删除文档
func (r *documentRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Document{}, "id = ?", id).Error
}

// ListByChannel 获取频道文档列表
func (r *documentRepo) ListByChannel(ctx context.Context, channelID string, filter *DocumentFilter, page, pageSize int) ([]*model.Document, int64, error) {
	var docs []*model.Document
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Document{}).Where("channel_id = ?", channelID)

	// 关键词筛选
	if filter != nil && filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		db = db.Where("title LIKE ? OR summary LIKE ?", kw, kw)
	}

	// 查询总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	sortField := "updated_at"
	sortOrder := "DESC"
	if filter != nil {
		if filter.Sort == "title" {
			sortField = "title"
		} else if filter.Sort == "file_size" {
			sortField = "file_size"
		}
		if filter.Order == "asc" {
			sortOrder = "ASC"
		}
	}
	db = db.Order(sortField + " " + sortOrder)

	// 分页查询
	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// Search 搜索文档
func (r *documentRepo) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Document, int64, error) {
	var docs []*model.Document
	var total int64

	kw := "%" + keyword + "%"
	db := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("title LIKE ? OR content LIKE ? OR summary LIKE ?", kw, kw, kw)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// CountByChannel 统计频道文档数量
func (r *documentRepo) CountByChannel(ctx context.Context, channelID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Document{}).Where("channel_id = ?", channelID).Count(&count).Error
	return count, err
}
