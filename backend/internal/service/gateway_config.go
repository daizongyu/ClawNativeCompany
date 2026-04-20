package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
)

// GatewayConfigService Gateway 配置服务
type GatewayConfigService struct {
	repo       *repository.GatewayConfigRepository
	encryptKey []byte
}

// NewGatewayConfigService 创建 Gateway 配置服务
func NewGatewayConfigService(repo *repository.GatewayConfigRepository) *GatewayConfigService {
	// 从环境变量获取加密密钥，或使用默认密钥（生产环境必须修改）
	key := []byte("claw-gateway-config-key-32bytes!")
	return &GatewayConfigService{
		repo:       repo,
		encryptKey: key,
	}
}

// CreateGatewayConfigRequest 创建 Gateway 配置请求
type CreateGatewayConfigRequest struct {
	Type       model.GatewayType `json:"type" validate:"required,oneof=dingtalk slack custom"`
	Name       string            `json:"name" validate:"required,min=1,max=100"`
	IsDefault  bool              `json:"is_default"`

	// 钉钉配置
	AppKey    string `json:"app_key,omitempty"`
	AppSecret string `json:"app_secret,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	CorpID    string `json:"corp_id,omitempty"`

	// Slack 配置
	BotToken  string `json:"bot_token,omitempty"`
	AppToken  string `json:"app_token,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`

	// 自定义配置
	WebhookURL string `json:"webhook_url,omitempty"`
	AuthType   string `json:"auth_type,omitempty"`
	AuthToken  string `json:"auth_token,omitempty"`
}

// UpdateGatewayConfigRequest 更新 Gateway 配置请求
type UpdateGatewayConfigRequest struct {
	Name      string            `json:"name" validate:"omitempty,min=1,max=100"`
	Status    model.GatewayStatus `json:"status" validate:"omitempty,oneof=active inactive"`
	IsDefault bool              `json:"is_default"`

	// 钉钉配置
	AppKey    string `json:"app_key,omitempty"`
	AppSecret string `json:"app_secret,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	CorpID    string `json:"corp_id,omitempty"`

	// Slack 配置
	BotToken  string `json:"bot_token,omitempty"`
	AppToken  string `json:"app_token,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`

	// 自定义配置
	WebhookURL string `json:"webhook_url,omitempty"`
	AuthType   string `json:"auth_type,omitempty"`
	AuthToken  string `json:"auth_token,omitempty"`
}

// List 列出 Gateway 配置（带筛选）
func (s *GatewayConfigService) List(employeeID string, gatewayType model.GatewayType, status model.GatewayStatus, page, pageSize int) ([]model.GatewayConfig, int64, error) {
	return s.repo.List(employeeID, gatewayType, status, page, pageSize)
}

// Create 创建 Gateway 配置
func (s *GatewayConfigService) Create(ctx context.Context, employeeID string, req CreateGatewayConfigRequest) (*model.GatewayConfig, error) {
	// 检查名称是否已存在
	exists, err := s.repo.ExistsByName(employeeID, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("配置名称已存在")
	}

	config := &model.GatewayConfig{
		EmployeeID: employeeID,
		Type:       req.Type,
		Name:       req.Name,
		Status:     model.GatewayStatusActive,
		IsDefault:  req.IsDefault,
	}

	// 根据类型设置配置
	switch req.Type {
	case model.GatewayTypeDingTalk:
		if req.AppKey == "" || req.AppSecret == "" {
			return nil, errors.New("钉钉配置需要 AppKey 和 AppSecret")
		}
		config.AppKey = req.AppKey
		config.AppSecret = s.encrypt(req.AppSecret)
		config.AgentID = req.AgentID
		config.CorpID = req.CorpID

	case model.GatewayTypeSlack:
		if req.BotToken == "" {
			return nil, errors.New("Slack 配置需要 BotToken")
		}
		config.BotToken = s.encrypt(req.BotToken)
		config.AppToken = s.encrypt(req.AppToken)
		config.ChannelID = req.ChannelID

	case model.GatewayTypeCustom:
		if req.WebhookURL == "" {
			return nil, errors.New("自定义配置需要 WebhookURL")
		}
		config.WebhookURL = req.WebhookURL
		config.AuthType = req.AuthType
		if req.AuthToken != "" {
			config.AuthToken = s.encrypt(req.AuthToken)
		}
	}

	// 如果是第一个配置，设为默认
	count, err := s.repo.CountByEmployeeID(employeeID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		config.IsDefault = true
	}

	// 如果设为默认，取消其他默认配置
	if config.IsDefault {
		if err := s.setDefaultConfig(config.ID, employeeID); err != nil {
			logger.Warn("设置默认配置失败", "error", err)
		}
	}

	if err := s.repo.Create(config); err != nil {
		return nil, err
	}

	return config, nil
}

// Update 更新 Gateway 配置
func (s *GatewayConfigService) Update(ctx context.Context, id string, employeeID string, req UpdateGatewayConfigRequest) (*model.GatewayConfig, error) {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("配置不存在")
	}

	if config.EmployeeID != employeeID {
		return nil, errors.New("无权修改此配置")
	}

	// 更新名称
	if req.Name != "" && req.Name != config.Name {
		exists, err := s.repo.ExistsByName(employeeID, req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("配置名称已存在")
		}
		config.Name = req.Name
	}

	// 更新状态
	if req.Status != "" {
		config.Status = req.Status
	}

	// 更新配置内容
	switch config.Type {
	case model.GatewayTypeDingTalk:
		if req.AppKey != "" {
			config.AppKey = req.AppKey
		}
		if req.AppSecret != "" {
			config.AppSecret = s.encrypt(req.AppSecret)
		}
		if req.AgentID != "" {
			config.AgentID = req.AgentID
		}
		if req.CorpID != "" {
			config.CorpID = req.CorpID
		}

	case model.GatewayTypeSlack:
		if req.BotToken != "" {
			config.BotToken = s.encrypt(req.BotToken)
		}
		if req.AppToken != "" {
			config.AppToken = s.encrypt(req.AppToken)
		}
		if req.ChannelID != "" {
			config.ChannelID = req.ChannelID
		}

	case model.GatewayTypeCustom:
		if req.WebhookURL != "" {
			config.WebhookURL = req.WebhookURL
		}
		if req.AuthType != "" {
			config.AuthType = req.AuthType
		}
		if req.AuthToken != "" {
			config.AuthToken = s.encrypt(req.AuthToken)
		}
	}

	// 更新默认设置
	if req.IsDefault && !config.IsDefault {
		config.IsDefault = true
		if err := s.setDefaultConfig(config.ID, employeeID); err != nil {
			logger.Warn("设置默认配置失败", "error", err)
		}
	}

	if err := s.repo.Update(config); err != nil {
		return nil, err
	}

	return config, nil
}

// Delete 删除 Gateway 配置
func (s *GatewayConfigService) Delete(ctx context.Context, id string, employeeID string) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("配置不存在")
	}

	if config.EmployeeID != employeeID {
		return errors.New("无权删除此配置")
	}

	return s.repo.Delete(id)
}

// GetByID 根据 ID 获取配置
func (s *GatewayConfigService) GetByID(ctx context.Context, id string) (*model.GatewayConfig, error) {
	return s.repo.GetByID(id)
}

// GetByEmployeeID 获取员工的所有配置
func (s *GatewayConfigService) GetByEmployeeID(ctx context.Context, employeeID string) ([]model.GatewayConfig, error) {
	return s.repo.GetByEmployeeID(employeeID)
}

// GetDefaultByEmployeeID 获取员工的默认配置
func (s *GatewayConfigService) GetDefaultByEmployeeID(ctx context.Context, employeeID string) (*model.GatewayConfig, error) {
	return s.repo.GetDefaultByEmployeeID(employeeID)
}

// SetDefault 设置默认配置
func (s *GatewayConfigService) SetDefault(ctx context.Context, id string, employeeID string) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("配置不存在")
	}

	if config.EmployeeID != employeeID {
		return errors.New("无权修改此配置")
	}

	return s.repo.SetDefault(id, employeeID)
}

// setDefaultConfig 设置默认配置（内部方法）
func (s *GatewayConfigService) setDefaultConfig(id string, employeeID string) error {
	return s.repo.SetDefault(id, employeeID)
}

// Verify 验证 Gateway 配置
func (s *GatewayConfigService) Verify(ctx context.Context, id string, employeeID string) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("配置不存在")
	}

	if config.EmployeeID != employeeID {
		return errors.New("无权验证此配置")
	}

	// 根据类型进行验证
	var verifyErr error
	switch config.Type {
	case model.GatewayTypeDingTalk:
		verifyErr = s.verifyDingTalk(ctx, config)
	case model.GatewayTypeSlack:
		verifyErr = s.verifySlack(ctx, config)
	case model.GatewayTypeCustom:
		verifyErr = s.verifyCustom(ctx, config)
	}

	// 更新验证状态
	if verifyErr != nil {
		if err := s.repo.UpdateVerifyStatus(id, false, verifyErr.Error()); err != nil {
			logger.Error("更新验证状态失败", "error", err)
		}
		return verifyErr
	}

	if err := s.repo.UpdateVerifyStatus(id, true, ""); err != nil {
		logger.Error("更新验证状态失败", "error", err)
	}

	return nil
}

// verifyDingTalk 验证钉钉配置
func (s *GatewayConfigService) verifyDingTalk(ctx context.Context, config *model.GatewayConfig) error {
	// 解密 AppSecret
	appSecret := s.decrypt(config.AppSecret)

	// 调用钉钉 API 获取 access_token
	url := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s",
		config.AppKey, appSecret)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求钉钉 API 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("钉钉 API 返回错误状态码: %d", resp.StatusCode)
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		Token   string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析钉钉响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("钉钉 API 错误: %s", result.ErrMsg)
	}

	return nil
}

// verifySlack 验证 Slack 配置
func (s *GatewayConfigService) verifySlack(ctx context.Context, config *model.GatewayConfig) error {
	// 解密 BotToken
	botToken := s.decrypt(config.BotToken)

	// 调用 Slack API 验证
	req, err := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+botToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 Slack API 失败: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析 Slack 响应失败: %v", err)
	}

	if !result.OK {
		return fmt.Errorf("Slack API 错误: %s", result.Error)
	}

	return nil
}

// verifyCustom 验证自定义 Webhook
func (s *GatewayConfigService) verifyCustom(ctx context.Context, config *model.GatewayConfig) error {
	// 发送测试请求
	testPayload := map[string]interface{}{
		"event": "test",
		"timestamp": time.Now().Unix(),
		"message": "Gateway 配置测试",
	}

	payload, err := json.Marshal(testPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加认证
	if config.AuthType == "bearer" && config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.decrypt(config.AuthToken))
	} else if config.AuthType == "basic" && config.AuthToken != "" {
		req.Header.Set("Authorization", "Basic "+s.decrypt(config.AuthToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 Webhook 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// SendTestMessage 发送测试消息
func (s *GatewayConfigService) SendTestMessage(ctx context.Context, id string, employeeID string) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("配置不存在")
	}

	if config.EmployeeID != employeeID {
		return errors.New("无权测试此配置")
	}

	// 构建测试消息
	testMsg := map[string]interface{}{
		"event": "test",
		"timestamp": time.Now().Unix(),
		"title": "Gateway 配置测试",
		"content": "这是一条测试消息，来自 Claw Native 平台。",
	}

	// 根据类型发送
	switch config.Type {
	case model.GatewayTypeDingTalk:
		return s.sendDingTalkMessage(ctx, config, testMsg)
	case model.GatewayTypeSlack:
		return s.sendSlackMessage(ctx, config, testMsg)
	case model.GatewayTypeCustom:
		return s.sendCustomMessage(ctx, config, testMsg)
	}

	return errors.New("不支持的 Gateway 类型")
}

// sendDingTalkMessage 发送钉钉消息
func (s *GatewayConfigService) sendDingTalkMessage(ctx context.Context, config *model.GatewayConfig, msg map[string]interface{}) error {
	// 获取 access_token
	appSecret := s.decrypt(config.AppSecret)
	tokenURL := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s",
		config.AppKey, appSecret)

	resp, err := http.Get(tokenURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResult struct {
		ErrCode int    `json:"errcode"`
		Token   string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&tokenResult)

	if tokenResult.ErrCode != 0 {
		return errors.New("获取钉钉 access_token 失败")
	}

	// 发送消息
	msgURL := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", tokenResult.Token)

	payload := map[string]interface{}{
		"msg": map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": msg["content"].(string),
			},
		},
	}

	payloadJSON, _ := json.Marshal(payload)
	_, err = http.Post(msgURL, "application/json", strings.NewReader(string(payloadJSON)))
	return err
}

// sendSlackMessage 发送 Slack 消息
func (s *GatewayConfigService) sendSlackMessage(ctx context.Context, config *model.GatewayConfig, msg map[string]interface{}) error {
	botToken := s.decrypt(config.BotToken)

	payload := map[string]interface{}{
		"channel": config.ChannelID,
		"text":    msg["content"].(string),
	}

	payloadJSON, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", strings.NewReader(string(payloadJSON)))
	req.Header.Set("Authorization", "Bearer "+botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.OK {
		return fmt.Errorf("Slack API 错误: %s", result.Error)
	}

	return nil
}

// sendCustomMessage 发送自定义 Webhook 消息
func (s *GatewayConfigService) sendCustomMessage(ctx context.Context, config *model.GatewayConfig, msg map[string]interface{}) error {
	payload, _ := json.Marshal(msg)

	req, _ := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")

	if config.AuthType == "bearer" && config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.decrypt(config.AuthToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// encrypt 加密数据
func (s *GatewayConfigService) encrypt(plaintext string) string {
	if plaintext == "" {
		return ""
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		logger.Error("创建加密 cipher 失败", "error", err)
		return plaintext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("创建 GCM 失败", "error", err)
		return plaintext
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Error("生成 nonce 失败", "error", err)
		return plaintext
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// decrypt 解密数据
func (s *GatewayConfigService) decrypt(ciphertext string) string {
	if ciphertext == "" {
		return ""
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		// 可能是未加密的数据，直接返回
		return ciphertext
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		logger.Error("创建解密 cipher 失败", "error", err)
		return ciphertext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("创建 GCM 失败", "error", err)
		return ciphertext
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return ciphertext
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		// 解密失败，返回原文
		return ciphertext
	}

	return string(plaintext)
}
