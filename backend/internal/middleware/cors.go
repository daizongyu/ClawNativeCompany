package middleware

import (
	"net/http"
	"time"

	"claw/internal/logger"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig 默认 CORS 配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// CORS 跨域中间件
func CORS(config ...CORSConfig) gin.HandlerFunc {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查允许的源
		allowOrigin := ""
		if len(cfg.AllowOrigins) == 0 {
			allowOrigin = origin
		} else {
			for _, o := range cfg.AllowOrigins {
				if o == "*" || o == origin {
					allowOrigin = origin
					if o == "*" {
						allowOrigin = "*"
					}
					break
				}
			}
		}

		// 设置 CORS 响应头
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}

		if len(cfg.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods, ", "))
		}

		if len(cfg.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders, ", "))
		}

		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(cfg.ExposeHeaders, ", "))
		}

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(rune(cfg.MaxAge.Seconds())))
		}

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Logger 请求日志中间件
// 记录每个 HTTP 请求的详细信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 构建日志条目
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		// 获取员工 ID（如果已认证）
		employeeID := GetEmployeeID(c)

		entry := logger.LogEntry{
			Method:     method,
			Path:       path,
			Status:     statusCode,
			Duration:   duration,
			ClientIP:   clientIP,
			UserAgent:  userAgent,
			EmployeeID: employeeID,
		}

		// 如果有错误，记录错误信息
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		logger.LogRequest(entry)
	}
}

// joinStrings 将字符串切片用分隔符连接
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
