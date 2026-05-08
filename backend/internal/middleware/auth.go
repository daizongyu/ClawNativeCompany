// Package middleware 提供 Gin 中间件
package middleware

import (
	"net/http"
	"strings"

	"claw/internal/jwt"
	"claw/internal/logger"
	"claw/internal/repository"
	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ContextKey 上下文键类型定义（避免冲突）
type ContextKey string

const (
	// ContextKeyEmployeeID 员工 ID 上下文键
	ContextKeyEmployeeID ContextKey = "employee_id"
	// ContextKeyEmployeeName 员工名称上下文键
	ContextKeyEmployeeName ContextKey = "employee_name"
	// ContextKeyAuthType 认证类型上下文键
	ContextKeyAuthType ContextKey = "auth_type"
)

// AuthType 认证类型
type AuthType string

const (
	// AuthTypeJWT JWT 认证（人类用户）
	AuthTypeJWT AuthType = "jwt"
	// AuthTypeAPIKey API Key 认证（Agent）
	AuthTypeAPIKey AuthType = "api_key"
)

// JWTAuth JWT 认证中间件
// 从 Authorization header 提取并验证 JWT token
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "缺少认证信息")
			c.Abort()
			return
		}

		// 提取 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(c, http.StatusUnauthorized, "认证格式错误，请使用 Bearer token")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwt.ValidateAccessToken(token)
		if err != nil {
			logger.Warn("JWT 验证失败",
				"error", err.Error(),
				"path", c.Request.URL.Path,
			)
			utils.Error(c, http.StatusUnauthorized, "认证已过期或无效")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(string(ContextKeyEmployeeID), claims.EmployeeID)
		c.Set(string(ContextKeyEmployeeName), claims.Name)
		c.Set(string(ContextKeyAuthType), AuthTypeJWT)

		c.Next()
	}
}

// APIKeyAuth API Key 认证中间件
// 从 X-API-Key header 验证 API Key
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			utils.Error(c, http.StatusUnauthorized, "缺少 API Key")
			c.Abort()
			return
		}

		// 验证 API Key 格式（支持44字符纯base64 或 69字符带claw_前缀）
		if len(apiKey) != 44 && len(apiKey) != 69 {
			utils.Error(c, http.StatusUnauthorized, "API Key 格式错误")
			c.Abort()
			return
		}

		// 查询数据库验证 API Key
		employeeRepo := repository.NewEmployeeRepository()
		employee, err := employeeRepo.GetByAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			logger.Warn("API Key 验证失败",
				"error", err.Error(),
				"path", c.Request.URL.Path,
			)
			utils.Error(c, http.StatusUnauthorized, "无效的 API Key")
			c.Abort()
			return
		}

		// 检查账号状态
		if !employee.IsActive() {
			utils.Error(c, http.StatusForbidden, "账号已禁用")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(string(ContextKeyEmployeeID), employee.ID)
		c.Set(string(ContextKeyEmployeeName), employee.Name)
		c.Set(string(ContextKeyAuthType), AuthTypeAPIKey)

		c.Next()
	}
}

// DualAuth 双认证模式中间件
// 支持 JWT（人类）或 API Key（Agent）任一方式认证
func DualAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		apiKey := c.GetHeader("X-API-Key")

		// 优先尝试 JWT 认证
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token := parts[1]
				claims, err := jwt.ValidateAccessToken(token)
				if err == nil {
					// 验证员工状态
					employeeRepo := repository.NewEmployeeRepository()
					employee, err := employeeRepo.GetByID(c.Request.Context(), claims.EmployeeID)
					if err == nil && employee.IsActive() {
						c.Set(string(ContextKeyEmployeeID), claims.EmployeeID)
						c.Set(string(ContextKeyEmployeeName), claims.Name)
						c.Set(string(ContextKeyAuthType), AuthTypeJWT)
						c.Next()
						return
					}
				}
			}
		}

		// 尝试 API Key 认证（支持44或69字符）
		if apiKey != "" && (len(apiKey) == 44 || len(apiKey) == 69) {
			logger.Info("API Key auth attempt", "key_len", len(apiKey), "key_prefix", apiKey[:10])
			employeeRepo := repository.NewEmployeeRepository()
			employee, err := employeeRepo.GetByAPIKey(c.Request.Context(), apiKey)
			logger.Info("API Key result", "employee", employee, "err", err)
			if err == nil && employee.IsActive() {
				c.Set(string(ContextKeyEmployeeID), employee.ID)
				c.Set(string(ContextKeyEmployeeName), employee.Name)
				c.Set(string(ContextKeyAuthType), AuthTypeAPIKey)
				c.Next()
				return
			}
		}

		// 都失败了
		utils.Error(c, http.StatusUnauthorized, "认证失败，请提供有效的 JWT 或 API Key")
		c.Abort()
	}
}

// OptionalAuth 可选认证中间件
// 如果提供了认证信息则解析，否则匿名访问
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token := parts[1]
				claims, err := jwt.ValidateAccessToken(token)
				if err == nil {
					c.Set(string(ContextKeyEmployeeID), claims.EmployeeID)
					c.Set(string(ContextKeyEmployeeName), claims.Name)
					c.Set(string(ContextKeyAuthType), AuthTypeJWT)
				}
			}
		}
		c.Next()
	}
}

// AdminOnly 仅管理员访问
// 检查员工是否为频道管理员（需要在频道上下文中使用）
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeID := GetEmployeeID(c)
		if employeeID == "" {
			utils.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// TODO: 检查员工在特定频道的角色
		// 当前简化实现：允许所有已认证用户
		// 实际应该查询 channel_members 表检查 role 字段

		c.Next()
	}
}

// GetEmployeeID 从上下文获取员工 ID
func GetEmployeeID(c *gin.Context) string {
	val, exists := c.Get(string(ContextKeyEmployeeID))
	if !exists {
		return ""
	}
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}

// GetEmployeeName 从上下文获取员工名称
func GetEmployeeName(c *gin.Context) string {
	val, exists := c.Get(string(ContextKeyEmployeeName))
	if !exists {
		return ""
	}
	if name, ok := val.(string); ok {
		return name
	}
	return ""
}

// GetAuthType 从上下文获取认证类型
func GetAuthType(c *gin.Context) AuthType {
	val, exists := c.Get(string(ContextKeyAuthType))
	if !exists {
		return ""
	}
	if authType, ok := val.(AuthType); ok {
		return authType
	}
	return ""
}

// IsAuthenticated 检查是否已认证
func IsAuthenticated(c *gin.Context) bool {
	return GetEmployeeID(c) != ""
}

// IsAPIKeyAuth 是否为 API Key 认证
func IsAPIKeyAuth(c *gin.Context) bool {
	return GetAuthType(c) == AuthTypeAPIKey
}

// IsJWTAuth 是否为 JWT 认证
func IsJWTAuth(c *gin.Context) bool {
	return GetAuthType(c) == AuthTypeJWT
}
