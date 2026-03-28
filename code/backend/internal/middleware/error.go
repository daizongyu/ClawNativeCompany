// Package middleware 提供 Gin 中间件
package middleware

import (
	"net/http"

	"claw/internal/logger"
	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 统一错误处理中间件
// 捕获 panic 并返回统一格式的错误响应
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 错误
				logger.Error("服务器内部错误",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				// 返回统一错误响应
				utils.Error(c, http.StatusInternalServerError, "服务器内部错误")
				c.Abort()
			}
		}()

		c.Next()

		// 处理业务错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("请求处理错误",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)

			// 根据错误类型返回不同状态码
			status := http.StatusBadRequest
			if err.Type == gin.ErrorTypePrivate {
				status = http.StatusInternalServerError
			}

			utils.Error(c, status, err.Error())
		}
	}
}

// Recovery 错误恢复中间件（别名，与 ErrorHandler 功能相同）
func Recovery() gin.HandlerFunc {
	return ErrorHandler()
}

// NotFound 处理 404 路由
func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.Error(c, http.StatusNotFound, "接口不存在")
	}
}

// MethodNotAllowed 处理 405 方法不允许
func MethodNotAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.Error(c, http.StatusMethodNotAllowed, "请求方法不允许")
	}
}
