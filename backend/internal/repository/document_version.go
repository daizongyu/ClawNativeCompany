// Package repository 提供文档版本数据访问层
package repository

import (
	"context"
	"errors"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

// DocumentVersionRepository 文档版本 Repository 接口
type DocumentVersionRepository interface {
	Create(ctx context.Context, version *model.DocumentVersion) error
	GetByID(ctx context.Context, id string) (*model.DocumentVersion, error)
	GetByVersion(ctx context.Context, documentID string, version int) (*model.DocumentVersion, error)
	ListByDocument(ctx context.Context, documentID string, page, pageSize int) ([]*model.DocumentVersion, int64, error)
	Delete(ctx context.Context, id string) error
	DeleteByDocument(ctx context.Context, documentID string) error
	CountByDocument(ctx context.Context, documentID string) (int64, error)
	GetOldestVersions(ctx context.Context, documentID string, limit int) ([]*model.DocumentVersion, error)
}

// documentVersionRepo 文档版本 Repository 实现
type documentVersionRepo struct {
	db *gorm.DB
}

// NewDocumentVersionRepository 创建文档版本 Repository
func NewDocumentVersionRepository() DocumentVersionRepository {
	return &documentVersionRepo{db: database.GetDB()}
}

// Create 创建版本记录
func (r *documentVersionRepo) Create(ctx context.Context, version *model.DocumentVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetByID 根据ID获取版本
func (r *documentVersionRepo) GetByID(ctx context.Context, id string) (*model.DocumentVersion, error) {
	var ver model.DocumentVersion
	err := r.db.WithContext(ctx).
		Preload("Document").
		Preload("Editor").
		First(&ver, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ver, nil
}

// GetByVersion 根据文档ID和版本号获取版本
func (r *documentVersionRepo) GetByVersion(ctx context.Context, documentID string, version int) (*model.DocumentVersion, error) {
	var ver model.DocumentVersion
	err := r.db.WithContext(ctx).
		Preload("Editor").
		Where("document_id = ? AND version = ?", documentID, version).
		First(&ver).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ver, nil
}

// ListByDocument 获取文档版本列表（按版本号倒序）
func (r *documentVersionRepo) ListByDocument(ctx context.Context, documentID string, page, pageSize int) ([]*model.DocumentVersion, int64, error) {
	var versions []*model.DocumentVersion
	var total int64

	db := r.db.WithContext(ctx).Model(&model.DocumentVersion{}).Where("document_id = ?", documentID)

	// 查询总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 按版本号倒序（最新版本在前）
	if err := db.Order("version DESC").
		Scopes(database.Paginate(page, pageSize)).
		Preload("Editor").
		Find(&versions).Error; err != nil {
		return nil, 0, err
	}

	return versions, total, nil
}

// Delete 删除版本记录
func (r *documentVersionRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.DocumentVersion{}, "id = ?", id).Error
}

// DeleteByDocument 删除文档的所有版本记录
func (r *documentVersionRepo) DeleteByDocument(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&model.DocumentVersion{}).Error
}

// CountByDocument 统计文档版本数量
func (r *documentVersionRepo) CountByDocument(ctx context.Context, documentID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.DocumentVersion{}).Where("document_id = ?", documentID).Count(&count).Error
	return count, err
}

// GetOldestVersions 获取最旧的N个版本（用于清理）
func (r *documentVersionRepo) GetOldestVersions(ctx context.Context, documentID string, limit int) ([]*model.DocumentVersion, error) {
	var versions []*model.DocumentVersion
	err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("version ASC").
		Limit(limit).
		Find(&versions).Error
	return versions, err
}
