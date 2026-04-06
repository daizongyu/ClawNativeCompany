// Package repository 提供消息数据访问层
package repository

import (
	"context"
	"errors"
	"time"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

var (
	// ErrMessageNotFound 消息不存在
	ErrMessageNotFound = errors.New("消息不存在")
)

// MessageRepository 消息 Repository 接口
type MessageRepository interface {
	Create(ctx context.Context, msg *model.Message) error
	GetByID(ctx context.Context, id string) (*model.Message, error)
	ListByChannel(ctx context.Context, channelID string, before *time.Time, limit int) ([]*model.Message, error)
	Update(ctx context.Context, msg *model.Message) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, channelID string, keyword string, page, pageSize int) ([]*model.Message, int64, error)
	GetThread(ctx context.Context, parentID string) ([]*model.Message, error)
}

// messageRepo 消息 Repository 实现
type messageRepo struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息 Repository
func NewMessageRepository() MessageRepository {
	return &messageRepo{db: database.GetDB()}
}

// Create 创建消息
func (r *messageRepo) Create(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// GetByID 根据 ID 获取消息
func (r *messageRepo) GetByID(ctx context.Context, id string) (*model.Message, error) {
	var msg model.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Where("id = ?", id).
		First(&msg).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return &msg, nil
}

// ListByChannel 获取频道消息列表（支持游标分页）
// before: 获取此时间之前的消息，nil 表示获取最新消息
func (r *messageRepo) ListByChannel(ctx context.Context, channelID string, before *time.Time, limit int) ([]*model.Message, error) {
	var messages []*model.Message

	db := r.db.WithContext(ctx).
		Preload("Sender").
		Where("channel_id = ?", channelID).
		Where("parent_id IS NULL"). // 只获取顶级消息
		Order("created_at DESC")

	if before != nil {
		db = db.Where("created_at < ?", before)
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	err := db.Limit(limit).Find(&messages).Error
	return messages, err
}

// Update 更新消息
func (r *messageRepo) Update(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Save(msg).Error
}

// Delete 删除消息（软删除）
func (r *messageRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Message{}, "id = ?", id).Error
}

// Search 搜索消息
func (r *messageRepo) Search(ctx context.Context, channelID string, keyword string, page, pageSize int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	db := r.db.WithContext(ctx).
		Preload("Sender").
		Where("channel_id = ?", channelID).
		Where("content LIKE ?", "%"+keyword+"%")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetThread 获取回复线程
func (r *messageRepo) GetThread(ctx context.Context, parentID string) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Where("parent_id = ?", parentID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}
