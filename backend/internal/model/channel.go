package model

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
	Members     []Employee  `gorm:"many2many:channel_members;" json:"members,omitempty"`
}

// TableName 返回表名
func (Channel) TableName() string {
	return "channels"
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
