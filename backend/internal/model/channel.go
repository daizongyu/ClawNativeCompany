package model

import "gorm.io/gorm"

// ChannelType 频道类型
type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"  // 公开频道
	ChannelTypePrivate ChannelType = "private" // 私有频道
	ChannelTypeDM      ChannelType = "dm"      // 私聊
)

// ChannelRole 频道角色
type ChannelRole string

const (
	ChannelRoleAdmin    ChannelRole = "admin"    // 管理员
	ChannelRoleMember   ChannelRole = "member"   // 成员
	ChannelRoleReadonly ChannelRole = "readonly" // 只读
)

// Channel 频道模型
type Channel struct {
	Base
	Name        string      `gorm:"size:100;not null" json:"name"`
	Type        ChannelType `gorm:"size:20;not null" json:"type"`
	Description string      `gorm:"size:500" json:"description"`
	CreatedBy   string      `gorm:"size:36;index;not null" json:"created_by"`
	Creator     *Employee   `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
	Members     []Employee  `gorm:"many2many:channel_members;" json:"members,omitempty"`

	// 树形结构字段（新增）
	ParentID   *string    `gorm:"size:36;index" json:"parent_id"`       // 父频道ID，null为根频道
	Path       string     `gorm:"size:500;index" json:"path"`           // 路径：/产品需求/子需求分析
	Depth      int        `gorm:"default:0" json:"depth"`               // 层级深度：0=根
	DocCount   int        `gorm:"default:0" json:"doc_count"`           // 文档数量（缓存）
	ChildCount int        `gorm:"default:0" json:"child_count"`         // 子频道数量（缓存）

	// 关联
	Parent   *Channel  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Channel `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Documents []Document `gorm:"foreignKey:ChannelID" json:"documents,omitempty"`
}

// TableName 返回表名
func (Channel) TableName() string {
	return "channels"
}

// BeforeCreate 创建前处理 - 自动计算 Path 和 Depth
func (c *Channel) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = GenerateID("ch")
	}

	// 自动计算 Path 和 Depth（使用 ID 而非 Name）
	if c.ParentID != nil && *c.ParentID != "" {
		var parent Channel
		if err := tx.First(&parent, "id = ?", *c.ParentID).Error; err == nil {
			c.Path = parent.Path + "/" + c.ID
			c.Depth = parent.Depth + 1
		}
	} else {
		c.Path = "/" + c.ID
		c.Depth = 0
	}

	return nil
}

// BeforeUpdate 更新前处理 - 如果 ParentID 变更，需要更新 Path
func (c *Channel) BeforeUpdate(tx *gorm.DB) error {
	// 如果 ParentID 发生变化，重新计算 Path 和 Depth（使用 ID 而非 Name）
	if c.ParentID != nil && *c.ParentID != "" {
		var parent Channel
		if err := tx.First(&parent, "id = ?", *c.ParentID).Error; err == nil {
			c.Path = parent.Path + "/" + c.ID
			c.Depth = parent.Depth + 1
		}
	} else {
		c.Path = "/" + c.ID
		c.Depth = 0
	}
	return nil
}

// IsRoot 是否为根频道
func (c *Channel) IsRoot() bool {
	return c.ParentID == nil
}

// GetFullPath 获取完整路径
func (c *Channel) GetFullPath() string {
	return c.Path
}

// IsPublic 是否为公开频道
func (c *Channel) IsPublic() bool {
	return c.Type == ChannelTypePublic
}

// IsPrivate 是否为私有频道
func (c *Channel) IsPrivate() bool {
	return c.Type == ChannelTypePrivate
}

// IsDM 是否为私聊
func (c *Channel) IsDM() bool {
	return c.Type == ChannelTypeDM
}

// ChannelMember 频道成员关联表模型
type ChannelMember struct {
	Base
	ChannelID  string      `gorm:"size:36;uniqueIndex:idx_channel_member;not null" json:"channel_id"`
	EmployeeID string      `gorm:"size:36;uniqueIndex:idx_channel_member;not null;index" json:"employee_id"`
	Role       ChannelRole `gorm:"size:20;default:'member'" json:"role"` // admin | member | readonly
	Employee   *Employee   `gorm:"foreignKey:EmployeeID;references:ID" json:"employee,omitempty"`
}

// TableName 返回表名
func (ChannelMember) TableName() string {
	return "channel_members"
}

// IsAdmin 是否为频道管理员
func (cm *ChannelMember) IsAdmin() bool {
	return cm.Role == ChannelRoleAdmin
}

// ChannelResponse 频道响应结构
type ChannelResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        ChannelType `json:"type"`
	Description string      `json:"description"`
	CreatedBy   string      `json:"created_by"`
	MemberCount int         `json:"member_count"`
	CreatedAt   string      `json:"created_at"`
}

// ToResponse 转换为响应结构
func (c *Channel) ToResponse(memberCount int) ChannelResponse {
	return ChannelResponse{
		ID:          c.ID,
		Name:        c.Name,
		Type:        c.Type,
		Description: c.Description,
		CreatedBy:   c.CreatedBy,
		MemberCount: memberCount,
		CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
