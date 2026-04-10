// Package repository 提供频道数据访问层
package repository

import (
	"context"
	"errors"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

var (
	// ErrChannelNotFound 频道不存在
	ErrChannelNotFound = errors.New("频道不存在")
	// ErrMemberNotFound 成员不存在
	ErrMemberNotFound = errors.New("频道成员不存在")
	// ErrAlreadyMember 已经是频道成员
	ErrAlreadyMember = errors.New("已经是频道成员")
)

// ListFilter 频道列表筛选条件
type ListFilter struct {
	Type    string // public | private
	Keyword string // 搜索名称或描述
}

// ChannelRepository 频道 Repository 接口
type ChannelRepository interface {
	// 频道 CRUD
	Create(ctx context.Context, ch *model.Channel) error
	GetByID(ctx context.Context, id string) (*model.Channel, error)
	List(ctx context.Context, page, pageSize int) ([]*model.Channel, int64, error)
	ListByMember(ctx context.Context, employeeID string, page, pageSize int) ([]*model.Channel, int64, error)
	ListWithFilter(ctx context.Context, filter ListFilter, page, pageSize int) ([]*model.Channel, int64, error)
	Update(ctx context.Context, ch *model.Channel) error
	Delete(ctx context.Context, id string) error

	// 成员管理
	AddMember(ctx context.Context, member *model.ChannelMember) error
	RemoveMember(ctx context.Context, channelID, employeeID string) error
	UpdateMemberRole(ctx context.Context, channelID, employeeID string, role model.ChannelRole) error
	GetMember(ctx context.Context, channelID, employeeID string) (*model.ChannelMember, error)
	ListMembers(ctx context.Context, channelID string) ([]*model.ChannelMember, error)
	GetMemberCount(ctx context.Context, channelID string) (int64, error)

	// 统计
	Count(ctx context.Context) (int64, error)

	// 权限检查
	CheckPermission(ctx context.Context, channelID, employeeID string, minRole model.ChannelRole) (bool, error)
}

// channelRepo 频道 Repository 实现
type channelRepo struct {
	db *gorm.DB
}

// NewChannelRepository 创建频道 Repository
func NewChannelRepository() ChannelRepository {
	return &channelRepo{db: database.GetDB()}
}

// Create 创建频道
func (r *channelRepo) Create(ctx context.Context, ch *model.Channel) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

// GetByID 根据 ID 获取频道
func (r *channelRepo) GetByID(ctx context.Context, id string) (*model.Channel, error) {
	var ch model.Channel
	err := r.db.WithContext(ctx).
		Preload("Members").
		Where("id = ?", id).
		First(&ch).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	return &ch, nil
}

// List 分页查询频道列表
func (r *channelRepo) List(ctx context.Context, page, pageSize int) ([]*model.Channel, int64, error) {
	var channels []*model.Channel
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Channel{}).Preload("Members")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// ListByMember 获取员工加入的频道列表
func (r *channelRepo) ListByMember(ctx context.Context, employeeID string, page, pageSize int) ([]*model.Channel, int64, error) {
	var channels []*model.Channel
	var total int64

	db := r.db.WithContext(ctx).
		Model(&model.Channel{}).
		Joins("JOIN channel_members ON channel_members.channel_id = channels.id").
		Where("channel_members.employee_id = ?", employeeID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// ListWithFilter 带筛选条件的频道列表
func (r *channelRepo) ListWithFilter(ctx context.Context, filter ListFilter, page, pageSize int) ([]*model.Channel, int64, error) {
	var channels []*model.Channel
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Channel{})

	// 类型筛选
	if filter.Type != "" {
		db = db.Where("type = ?", filter.Type)
	}

	// 关键词搜索（名称或描述）
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.Paginate(page, pageSize)).
		Preload("Creator").
		Preload("Members").
		Order("created_at DESC").
		Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// Update 更新频道
func (r *channelRepo) Update(ctx context.Context, ch *model.Channel) error {
	return r.db.WithContext(ctx).Save(ch).Error
}

// Delete 删除频道
func (r *channelRepo) Delete(ctx context.Context, id string) error {
	// 软删除，同时删除关联的成员
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除频道成员
		if err := tx.Where("channel_id = ?", id).Delete(&model.ChannelMember{}).Error; err != nil {
			return err
		}
		// 删除频道
		return tx.Delete(&model.Channel{}, "id = ?", id).Error
	})
}

// AddMember 添加频道成员
func (r *channelRepo) AddMember(ctx context.Context, member *model.ChannelMember) error {
	// 检查是否已经是成员
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ChannelMember{}).
		Where("channel_id = ? AND employee_id = ?", member.ChannelID, member.EmployeeID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAlreadyMember
	}

	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember 移除频道成员
func (r *channelRepo) RemoveMember(ctx context.Context, channelID, employeeID string) error {
	result := r.db.WithContext(ctx).
		Where("channel_id = ? AND employee_id = ?", channelID, employeeID).
		Delete(&model.ChannelMember{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

// UpdateMemberRole 更新成员角色
func (r *channelRepo) UpdateMemberRole(ctx context.Context, channelID, employeeID string, role model.ChannelRole) error {
	result := r.db.WithContext(ctx).
		Model(&model.ChannelMember{}).
		Where("channel_id = ? AND employee_id = ?", channelID, employeeID).
		Update("role", role)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

// GetMember 获取成员信息
func (r *channelRepo) GetMember(ctx context.Context, channelID, employeeID string) (*model.ChannelMember, error) {
	var member model.ChannelMember
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND employee_id = ?", channelID, employeeID).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

// ListMembers 获取频道成员列表
func (r *channelRepo) ListMembers(ctx context.Context, channelID string) ([]*model.ChannelMember, error) {
	var members []*model.ChannelMember
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Where("channel_id = ?", channelID).
		Find(&members).Error
	return members, err
}

// GetMemberCount 获取频道成员数量
func (r *channelRepo) GetMemberCount(ctx context.Context, channelID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ChannelMember{}).
		Where("channel_id = ?", channelID).
		Count(&count).Error
	return count, err
}

// CheckPermission 检查员工在频道的权限
// minRole: 要求的最低角色权限 (admin > member > readonly)
func (r *channelRepo) CheckPermission(ctx context.Context, channelID, employeeID string, minRole model.ChannelRole) (bool, error) {
	member, err := r.GetMember(ctx, channelID, employeeID)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return false, nil
		}
		return false, err
	}

	// 权限等级: admin=3, member=2, readonly=1
	roleLevels := map[model.ChannelRole]int{
		model.ChannelRoleAdmin:    3,
		model.ChannelRoleMember:   2,
		model.ChannelRoleReadonly: 1,
	}

	userLevel := roleLevels[member.Role]
	requiredLevel := roleLevels[minRole]

	return userLevel >= requiredLevel, nil
}

// Count 获取频道总数
func (r *channelRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Channel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
