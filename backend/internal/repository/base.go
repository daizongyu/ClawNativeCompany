// Package repository 提供数据访问层
// 封装所有数据库操作
package repository

import (
	"context"
	"errors"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

// 通用错误定义
var (
	ErrNotFound = errors.New("记录不存在")
	ErrExists   = errors.New("记录已存在")
)

// BaseRepository 基础 Repository 接口
type BaseRepository[T model.Model] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id string) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]*T, int64, error)
}

// baseRepo 基础 Repository 实现
type baseRepo[T model.Model] struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础 Repository
func NewBaseRepository[T model.Model]() BaseRepository[T] {
	return &baseRepo[T]{db: database.GetDB()}
}

// Create 创建记录
func (r *baseRepo[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByID 根据 ID 获取记录
func (r *baseRepo[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// Update 更新记录
func (r *baseRepo[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete 软删除记录
func (r *baseRepo[T]) Delete(ctx context.Context, id string) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, "id = ?", id).Error
}

// List 分页查询
func (r *baseRepo[T]) List(ctx context.Context, page, pageSize int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	db := r.db.WithContext(ctx).Model(new(T))

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// WithTx 在事务中执行
func (r *baseRepo[T]) WithTx(fn func(repo BaseRepository[T]) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &baseRepo[T]{db: tx}
		return fn(txRepo)
	})
}

// DB 获取 GORM DB 实例（用于自定义查询）
func (r *baseRepo[T]) DB() *gorm.DB {
	return r.db
}
