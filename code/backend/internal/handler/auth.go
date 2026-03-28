// Package handler 提供 HTTP 请求处理
// 认证 Handler 处理登录、登出、刷新等请求
package handler

import (
	"net/http"

	"claw/internal/middleware"
	"claw/internal/service"
	"claw/pkg/utils"
	"claw/pkg/validator"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证 Handler
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	// 执行登录
	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			utils.Error(c, http.StatusUnauthorized, "邮箱或密码错误")
		case service.ErrEmployeeNotActive:
			utils.Error(c, http.StatusForbidden, "账号未激活")
		default:
			utils.InternalError(c, "登录失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Refresh 刷新令牌
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	if req.RefreshToken == "" {
		utils.ValidationError(c, "刷新令牌不能为空")
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrTokenExpired:
			utils.Error(c, http.StatusUnauthorized, "刷新令牌已过期，请重新登录")
		case service.ErrTokenInvalid:
			utils.Error(c, http.StatusUnauthorized, "无效的刷新令牌")
		case service.ErrEmployeeNotActive:
			utils.Error(c, http.StatusForbidden, "账号已禁用")
		default:
			utils.InternalError(c, "刷新令牌失败")
		}
		return
	}

	utils.SuccessWithData(c, resp)
}

// Logout 用户登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	employeeID := middleware.GetEmployeeID(c)
	if employeeID == "" {
		utils.Unauthorized(c, "未登录")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), employeeID); err != nil {
		utils.InternalError(c, "登出失败")
		return
	}

	utils.Success(c)
}

// GetMe 获取当前登录用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	employeeID := middleware.GetEmployeeID(c)
	if employeeID == "" {
		utils.Unauthorized(c, "未登录")
		return
	}

	employee, err := h.authService.GetEmployeeByID(c.Request.Context(), employeeID)
	if err != nil {
		utils.InternalError(c, "获取用户信息失败")
		return
	}

	utils.SuccessWithData(c, employee.ToResponse())
}

// GenerateAPIKey 生成 API Key
// POST /api/v1/auth/api-key
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	employeeID := middleware.GetEmployeeID(c)
	if employeeID == "" {
		utils.Unauthorized(c, "未登录")
		return
	}

	apiKey, err := h.authService.GenerateAPIKey(c.Request.Context(), employeeID)
	if err != nil {
		utils.InternalError(c, "生成 API Key 失败")
		return
	}

	utils.SuccessWithData(c, gin.H{
		"api_key": apiKey,
		"note":    "请妥善保存此 API Key，它只会显示一次",
	})
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(r *gin.RouterGroup) {
	handler := NewAuthHandler()

	// 公开路由
	r.POST("/login", handler.Login)
	r.POST("/refresh", handler.Refresh)

	// 需要认证的路由
	auth := r.Group("")
	auth.Use(middleware.DualAuth())
	{
		auth.POST("/logout", handler.Logout)
		auth.GET("/me", handler.GetMe)
		auth.POST("/api-key", handler.GenerateAPIKey)
	}
}

// RegisterRoutes 注册路由（方法形式，用于统一接口）
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 公开路由
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)

	// 需要认证的路由
	auth := r.Group("")
	auth.Use(middleware.DualAuth())
	{
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.GetMe)
		auth.POST("/api-key", h.GenerateAPIKey)
	}
}
