// Package service 提供业务逻辑层
// 认证服务处理登录、登出、Token 刷新等
package service

import (
	"context"
	"errors"
	"fmt"

	"claw/internal/jwt"
	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
	"claw/pkg/password"
)

// 认证相关错误
var (
	ErrInvalidCredentials = errors.New("邮箱或密码错误")
	ErrEmployeeNotActive  = errors.New("账号未激活")
	ErrTokenInvalid       = errors.New("无效的令牌")
	ErrTokenExpired       = errors.New("令牌已过期")
	ErrAPIKeyInvalid      = errors.New("无效的 API Key")
)

// AuthService 认证服务
type AuthService struct {
	employeeRepo repository.EmployeeRepository
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{
		employeeRepo: repository.NewEmployeeRepository(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Employee     model.EmployeeResponse `json:"employee"`
	AccessToken  string                 `json:"access_token"`
	RefreshToken string                 `json:"refresh_token"`
	ExpiresIn    int                    `json:"expires_in"`
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// 1. 查找员工
	employee, err := s.employeeRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		logger.Error("登录查询失败", "error", err, "email", req.Email)
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	// 2. 检查是否为人类员工（Agent 不能通过密码登录）
	if !employee.IsHuman() {
		logger.Warn("Agent 尝试密码登录", "email", req.Email)
		return nil, ErrInvalidCredentials
	}

	// 3. 检查账号状态
	if !employee.IsActive() {
		return nil, ErrEmployeeNotActive
	}

	// 4. 验证密码
	if err := password.Verify(req.Password, employee.Password); err != nil {
		logger.Warn("密码验证失败", "email", req.Email)
		return nil, ErrInvalidCredentials
	}

	// 4. 生成 Token
	accessToken, refreshToken, err := jwt.GenerateTokenPair(employee.ID, employee.Name)
	if err != nil {
		logger.Error("生成 Token 失败", "error", err)
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	// 5. 更新最后在线时间
	if err := s.employeeRepo.UpdateLastSeen(ctx, employee.ID); err != nil {
		logger.Warn("更新最后在线时间失败", "error", err, "employee_id", employee.ID)
	}

	logger.Info("用户登录成功", "employee_id", employee.ID, "email", employee.Email)

	return &LoginResponse{
		Employee:     employee.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    jwt.GetExpireHours() * 3600,
	}, nil
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse 刷新令牌响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// RefreshToken 刷新访问令牌
func (s *AuthService) RefreshToken(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	// 1. 验证刷新令牌
	claims, err := jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	// 2. 检查员工是否存在且有效
	employee, err := s.employeeRepo.GetByID(ctx, claims.EmployeeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("刷新令牌失败: %w", err)
	}

	if !employee.IsActive() {
		return nil, ErrEmployeeNotActive
	}

	// 3. 生成新的令牌对
	accessToken, refreshToken, err := jwt.GenerateTokenPair(employee.ID, employee.Name)
	if err != nil {
		logger.Error("生成新 Token 失败", "error", err)
		return nil, fmt.Errorf("刷新令牌失败: %w", err)
	}

	logger.Info("令牌刷新成功", "employee_id", employee.ID)

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    jwt.GetExpireHours() * 3600,
	}, nil
}

// Logout 用户登出
// 当前实现为无状态，客户端只需删除本地 Token
// 未来可实现 Token 黑名单
func (s *AuthService) Logout(ctx context.Context, employeeID string) error {
	logger.Info("用户登出", "employee_id", employeeID)
	return nil
}

// ValidateToken 验证访问令牌
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claims, err := jwt.ValidateAccessToken(token)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	// 验证员工是否存在且有效
	employee, err := s.employeeRepo.GetByID(ctx, claims.EmployeeID)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	if !employee.IsActive() {
		return nil, ErrEmployeeNotActive
	}

	return claims, nil
}

// ValidateAPIKey 验证 API Key
func (s *AuthService) ValidateAPIKey(ctx context.Context, apiKey string) (*model.Employee, error) {
	if apiKey == "" {
		return nil, ErrAPIKeyInvalid
	}

	// 查询数据库验证 API Key
	employee, err := s.employeeRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAPIKeyInvalid
		}
		return nil, fmt.Errorf("验证 API Key 失败: %w", err)
	}

	if !employee.IsActive() {
		return nil, ErrEmployeeNotActive
	}

	return employee, nil
}

// GenerateAPIKey 为员工生成新的 API Key
func (s *AuthService) GenerateAPIKey(ctx context.Context, employeeID string) (string, error) {
	// 生成新的 API Key
	apiKey, err := password.GenerateAPIKey()
	if err != nil {
		return "", fmt.Errorf("生成 API Key 失败: %w", err)
	}

	// 更新到数据库
	if err := s.employeeRepo.UpdateAPIKey(ctx, employeeID, apiKey); err != nil {
		return "", fmt.Errorf("保存 API Key 失败: %w", err)
	}

	logger.Info("生成 API Key", "employee_id", employeeID)
	return apiKey, nil
}

// GetEmployeeByID 根据 ID 获取员工
func (s *AuthService) GetEmployeeByID(ctx context.Context, id string) (*model.Employee, error) {
	return s.employeeRepo.GetByID(ctx, id)
}
