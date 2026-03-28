// Package repository 提供员工数据访问层
package repository

import (
	"context"
	"errors"

	"claw/internal/database"
	"claw/internal/model"

	"gorm.io/gorm"
)

// EmployeeRepository 员工 Repository 接口
type EmployeeRepository interface {
	Create(ctx context.Context, emp *model.Employee) error
	GetByID(ctx context.Context, id string) (*model.Employee, error)
	GetByEmail(ctx context.Context, email string) (*model.Employee, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.Employee, error)
	Update(ctx context.Context, emp *model.Employee) error
	UpdateAPIKey(ctx context.Context, id, apiKey string) error
	UpdateLastSeen(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]*model.Employee, int64, error)
	SearchBySkills(ctx context.Context, skills []string, page, pageSize int) ([]*model.Employee, int64, error)
}

// employeeRepo 员工 Repository 实现
type employeeRepo struct {
	db *gorm.DB
}

// NewEmployeeRepository 创建员工 Repository
func NewEmployeeRepository() EmployeeRepository {
	return &employeeRepo{db: database.GetDB()}
}

// Create 创建员工
func (r *employeeRepo) Create(ctx context.Context, emp *model.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

// GetByID 根据 ID 获取员工
func (r *employeeRepo) GetByID(ctx context.Context, id string) (*model.Employee, error) {
	var emp model.Employee
	if err := r.db.WithContext(ctx).First(&emp, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}

// GetByEmail 根据邮箱获取员工
func (r *employeeRepo) GetByEmail(ctx context.Context, email string) (*model.Employee, error) {
	var emp model.Employee
	if err := r.db.WithContext(ctx).First(&emp, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}

// GetByAPIKey 根据 API Key 获取员工
func (r *employeeRepo) GetByAPIKey(ctx context.Context, apiKey string) (*model.Employee, error) {
	var emp model.Employee
	if err := r.db.WithContext(ctx).First(&emp, "api_key = ?", apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}

// Update 更新员工
func (r *employeeRepo) Update(ctx context.Context, emp *model.Employee) error {
	return r.db.WithContext(ctx).Save(emp).Error
}

// UpdateAPIKey 更新 API Key
func (r *employeeRepo) UpdateAPIKey(ctx context.Context, id, apiKey string) error {
	return r.db.WithContext(ctx).Model(&model.Employee{}).Where("id = ?", id).Update("api_key", apiKey).Error
}

// UpdateLastSeen 更新最后在线时间
func (r *employeeRepo) UpdateLastSeen(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.Employee{}).Where("id = ?", id).UpdateColumn("last_seen_at", gorm.Expr("datetime('now')")).Error
}

// List 分页查询员工列表
func (r *employeeRepo) List(ctx context.Context, page, pageSize int) ([]*model.Employee, int64, error) {
	var emps []*model.Employee
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Employee{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&emps).Error; err != nil {
		return nil, 0, err
	}

	return emps, total, nil
}

// SearchBySkills 根据技能搜索员工
// 使用 JSON 数组匹配，支持模糊搜索
func (r *employeeRepo) SearchBySkills(ctx context.Context, skills []string, page, pageSize int) ([]*model.Employee, int64, error) {
	var emps []*model.Employee
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Employee{})

	// 构建技能匹配条件
	for _, skill := range skills {
		db = db.Where("skills LIKE ?", "%"+skill+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.Paginate(page, pageSize)).Find(&emps).Error; err != nil {
		return nil, 0, err
	}

	return emps, total, nil
}
