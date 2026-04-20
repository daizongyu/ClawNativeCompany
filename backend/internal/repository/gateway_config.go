package repository

import (
	"claw/internal/model"

	"gorm.io/gorm"
)

// GatewayConfigRepository Gateway 配置仓储
type GatewayConfigRepository struct {
	db *gorm.DB
}

// NewGatewayConfigRepository 创建 Gateway 配置仓储
func NewGatewayConfigRepository(db *gorm.DB) *GatewayConfigRepository {
	return &GatewayConfigRepository{db: db}
}

// Create 创建 Gateway 配置
func (r *GatewayConfigRepository) Create(config *model.GatewayConfig) error {
	return r.db.Create(config).Error
}

// GetByID 根据 ID 获取 Gateway 配置
func (r *GatewayConfigRepository) GetByID(id string) (*model.GatewayConfig, error) {
	var config model.GatewayConfig
	if err := r.db.First(&config, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByIDWithEmployee 根据 ID 获取 Gateway 配置（包含员工信息）
func (r *GatewayConfigRepository) GetByIDWithEmployee(id string) (*model.GatewayConfig, error) {
	var config model.GatewayConfig
	if err := r.db.Preload("Employee").First(&config, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByEmployeeID 获取员工的所有 Gateway 配置
func (r *GatewayConfigRepository) GetByEmployeeID(employeeID string) ([]model.GatewayConfig, error) {
	var configs []model.GatewayConfig
	if err := r.db.Where("employee_id = ?", employeeID).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// GetByEmployeeIDAndType 获取员工指定类型的 Gateway 配置
func (r *GatewayConfigRepository) GetByEmployeeIDAndType(employeeID string, gatewayType model.GatewayType) ([]model.GatewayConfig, error) {
	var configs []model.GatewayConfig
	if err := r.db.Where("employee_id = ? AND type = ?", employeeID, gatewayType).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// GetDefaultByEmployeeID 获取员工的默认 Gateway 配置
func (r *GatewayConfigRepository) GetDefaultByEmployeeID(employeeID string) (*model.GatewayConfig, error) {
	var config model.GatewayConfig
	if err := r.db.Where("employee_id = ? AND is_default = ?", employeeID, true).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// Update 更新 Gateway 配置
func (r *GatewayConfigRepository) Update(config *model.GatewayConfig) error {
	return r.db.Save(config).Error
}

// UpdateStatus 更新 Gateway 配置状态
func (r *GatewayConfigRepository) UpdateStatus(id string, status model.GatewayStatus) error {
	return r.db.Model(&model.GatewayConfig{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateVerifyStatus 更新验证状态
func (r *GatewayConfigRepository) UpdateVerifyStatus(id string, verified bool, errorMsg string) error {
	updates := map[string]interface{}{
		"verify_error": errorMsg,
	}
	if verified {
		updates["status"] = model.GatewayStatusActive
		updates["last_verified_at"] = gorm.Expr("NOW()")
	} else {
		updates["status"] = model.GatewayStatusInvalid
	}
	return r.db.Model(&model.GatewayConfig{}).Where("id = ?", id).Updates(updates).Error
}

// SetDefault 设置默认 Gateway 配置
func (r *GatewayConfigRepository) SetDefault(id string, employeeID string) error {
	// 先将该员工的所有配置设为非默认
	if err := r.db.Model(&model.GatewayConfig{}).Where("employee_id = ?", employeeID).Update("is_default", false).Error; err != nil {
		return err
	}
	// 再将指定配置设为默认
	return r.db.Model(&model.GatewayConfig{}).Where("id = ?", id).Update("is_default", true).Error
}

// Delete 删除 Gateway 配置
func (r *GatewayConfigRepository) Delete(id string) error {
	return r.db.Delete(&model.GatewayConfig{}, "id = ?", id).Error
}

// List 列出 Gateway 配置（支持分页和筛选）
func (r *GatewayConfigRepository) List(employeeID string, gatewayType model.GatewayType, status model.GatewayStatus, page, pageSize int) ([]model.GatewayConfig, int64, error) {
	var configs []model.GatewayConfig
	var total int64

	query := r.db.Model(&model.GatewayConfig{})

	if employeeID != "" {
		query = query.Where("employee_id = ?", employeeID)
	}
	if gatewayType != "" {
		query = query.Where("type = ?", gatewayType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

// CountByEmployeeID 统计员工的 Gateway 配置数量
func (r *GatewayConfigRepository) CountByEmployeeID(employeeID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.GatewayConfig{}).Where("employee_id = ?", employeeID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ExistsByName 检查名称是否已存在（同一员工）
func (r *GatewayConfigRepository) ExistsByName(employeeID string, name string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.GatewayConfig{}).Where("employee_id = ? AND name = ?", employeeID, name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
