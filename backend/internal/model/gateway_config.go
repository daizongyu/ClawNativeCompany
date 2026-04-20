package model

import (
	"time"
)

// GatewayType Gateway 类型
type GatewayType string

const (
	GatewayTypeDingTalk GatewayType = "dingtalk" // 钉钉
	GatewayTypeSlack    GatewayType = "slack"    // Slack
	GatewayTypeCustom   GatewayType = "custom"   // 自定义 Webhook
)

// GatewayStatus Gateway 状态
type GatewayStatus string

const (
	GatewayStatusActive   GatewayStatus = "active"    // 已激活
	GatewayStatusInactive GatewayStatus = "inactive"  // 已禁用
	GatewayStatusInvalid  GatewayStatus = "invalid"   // 验证失败
)

// GatewayConfig Gateway 配置模型
type GatewayConfig struct {
	Base
	EmployeeID  string        `gorm:"size:36;index;not null" json:"employee_id"`      // 所属员工ID
	Type        GatewayType   `gorm:"size:20;not null" json:"type"`                   // 类型
	Name        string        `gorm:"size:100;not null" json:"name"`                  // 配置名称
	Status      GatewayStatus `gorm:"size:20;default:'active'" json:"status"`         // 状态
	IsDefault   bool          `gorm:"default:false" json:"is_default"`                // 是否默认

	// 钉钉配置
	AppKey       string `gorm:"size:255" json:"-"`                              // AppKey
	AppSecret    string `gorm:"size:255" json:"-"`                              // AppSecret（加密存储）
	AgentID      string `gorm:"size:50" json:"agent_id,omitempty"`              // AgentID
	CorpID       string `gorm:"size:100" json:"corp_id,omitempty"`              // CorpID

	// Slack 配置
	BotToken     string `gorm:"size:255" json:"-"`                              // Bot Token（加密存储）
	AppToken     string `gorm:"size:255" json:"-"`                              // App Token（加密存储）
	ChannelID    string `gorm:"size:50" json:"channel_id,omitempty"`            // 默认频道ID

	// 自定义 Webhook 配置
	WebhookURL   string `gorm:"size:500" json:"webhook_url,omitempty"`          // Webhook URL
	AuthType     string `gorm:"size:20" json:"auth_type,omitempty"`             // 认证类型: none/basic/bearer
	AuthToken    string `gorm:"size:255" json:"-"`                              // 认证令牌（加密存储）

	// 验证状态
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`                // 最后验证时间
	VerifyError    string     `gorm:"type:text" json:"verify_error,omitempty"`    // 验证错误信息

	// 关联
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// TableName 返回表名
func (GatewayConfig) TableName() string {
	return "gateway_configs"
}

// IsActive 检查是否激活
func (g *GatewayConfig) IsActive() bool {
	return g.Status == GatewayStatusActive
}

// IsDingTalk 是否为钉钉配置
func (g *GatewayConfig) IsDingTalk() bool {
	return g.Type == GatewayTypeDingTalk
}

// IsSlack 是否为 Slack 配置
func (g *GatewayConfig) IsSlack() bool {
	return g.Type == GatewayTypeSlack
}

// IsCustom 是否为自定义配置
func (g *GatewayConfig) IsCustom() bool {
	return g.Type == GatewayTypeCustom
}

// GetMaskedSecret 获取脱敏后的密钥
func (g *GatewayConfig) GetMaskedSecret() string {
	if g.AppSecret != "" {
		return maskSecret(g.AppSecret)
	}
	if g.BotToken != "" {
		return maskSecret(g.BotToken)
	}
	if g.AuthToken != "" {
		return maskSecret(g.AuthToken)
	}
	return ""
}

// maskSecret 脱敏处理
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}

// GatewayConfigResponse Gateway 配置响应结构（脱敏）
type GatewayConfigResponse struct {
	ID             string        `json:"id"`
	EmployeeID     string        `json:"employee_id"`
	Type           GatewayType   `json:"type"`
	Name           string        `json:"name"`
	Status         GatewayStatus `json:"status"`
	IsDefault      bool          `json:"is_default"`

	// 钉钉配置（脱敏）
	AppKey       string `json:"app_key,omitempty"`
	AgentID      string `json:"agent_id,omitempty"`
	CorpID       string `json:"corp_id,omitempty"`
	MaskedSecret string `json:"masked_secret,omitempty"`

	// Slack 配置（脱敏）
	BotToken     string `json:"bot_token,omitempty"`  // 脱敏后
	ChannelID    string `json:"channel_id,omitempty"`

	// 自定义配置
	WebhookURL   string `json:"webhook_url,omitempty"`
	AuthType     string `json:"auth_type,omitempty"`

	// 验证状态
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`
	VerifyError    string     `json:"verify_error,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse 转换为响应结构
func (g *GatewayConfig) ToResponse() GatewayConfigResponse {
	resp := GatewayConfigResponse{
		ID:             g.ID,
		EmployeeID:     g.EmployeeID,
		Type:           g.Type,
		Name:           g.Name,
		Status:         g.Status,
		IsDefault:      g.IsDefault,
		LastVerifiedAt: g.LastVerifiedAt,
		VerifyError:    g.VerifyError,
		CreatedAt:      g.CreatedAt,
		UpdatedAt:      g.UpdatedAt,
	}

	// 根据类型返回对应配置
	switch g.Type {
	case GatewayTypeDingTalk:
		resp.AppKey = g.AppKey
		resp.AgentID = g.AgentID
		resp.CorpID = g.CorpID
		resp.MaskedSecret = g.GetMaskedSecret()
	case GatewayTypeSlack:
		resp.BotToken = g.GetMaskedSecret()
		resp.ChannelID = g.ChannelID
	case GatewayTypeCustom:
		resp.WebhookURL = g.WebhookURL
		resp.AuthType = g.AuthType
	}

	return resp
}
