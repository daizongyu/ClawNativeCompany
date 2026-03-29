package repository

import (
	"context"
	"errors"

	"claw/internal/database"
	"claw/internal/model"
)

// ExternalMappingRepository 外部映射仓库接口
type ExternalMappingRepository interface {
	Create(ctx context.Context, mapping *model.ExternalMapping) error
	GetBySourceAndExternalID(ctx context.Context, sourceType string, externalID string, mappingType model.ExternalMappingType) (*model.ExternalMapping, error)
	GetByInternalID(ctx context.Context, internalID string, mappingType model.ExternalMappingType) ([]*model.ExternalMapping, error)
	Update(ctx context.Context, mapping *model.ExternalMapping) error
	Delete(ctx context.Context, id string) error
}

// externalMappingRepository 外部映射仓库实现
type externalMappingRepository struct{}

// NewExternalMappingRepository 创建外部映射仓库
func NewExternalMappingRepository() ExternalMappingRepository {
	return &externalMappingRepository{}
}

// Create 创建映射
func (r *externalMappingRepository) Create(ctx context.Context, mapping *model.ExternalMapping) error {
	db := database.GetDB()
	return db.WithContext(ctx).Create(mapping).Error
}

// GetBySourceAndExternalID 根据源类型和外部 ID 获取映射
func (r *externalMappingRepository) GetBySourceAndExternalID(ctx context.Context, sourceType string, externalID string, mappingType model.ExternalMappingType) (*model.ExternalMapping, error) {
	db := database.GetDB()
	var mapping model.ExternalMapping
	err := db.WithContext(ctx).
		Where("source_type = ? AND external_id = ? AND mapping_type = ?", sourceType, externalID, mappingType).
		First(&mapping).Error

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &mapping, nil
}

// GetByInternalID 根据内部 ID 获取映射
func (r *externalMappingRepository) GetByInternalID(ctx context.Context, internalID string, mappingType model.ExternalMappingType) ([]*model.ExternalMapping, error) {
	db := database.GetDB()
	var mappings []*model.ExternalMapping
	err := db.WithContext(ctx).
		Where("internal_id = ? AND mapping_type = ?", internalID, mappingType).
		Find(&mappings).Error

	if err != nil {
		return nil, err
	}

	return mappings, nil
}

// Update 更新映射
func (r *externalMappingRepository) Update(ctx context.Context, mapping *model.ExternalMapping) error {
	db := database.GetDB()
	return db.WithContext(ctx).Save(mapping).Error
}

// Delete 删除映射
func (r *externalMappingRepository) Delete(ctx context.Context, id string) error {
	db := database.GetDB()
	return db.WithContext(ctx).Delete(&model.ExternalMapping{}, "id = ?", id).Error
}
