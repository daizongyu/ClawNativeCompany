// Package utils 提供通用工具函数
package utils

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ResponseCode 响应码类型
type ResponseCode int

const (
	// CodeSuccess 成功
	CodeSuccess ResponseCode = 0
	// CodeError 通用错误
	CodeError ResponseCode = 1
	// CodeUnauthorized 未授权
	CodeUnauthorized ResponseCode = 401
	// CodeForbidden 禁止访问
	CodeForbidden ResponseCode = 403
	// CodeNotFound 资源不存在
	CodeNotFound ResponseCode = 404
	// CodeValidation 参数验证错误
	CodeValidation ResponseCode = 400
	// CodeInternal 服务器内部错误
	CodeInternal ResponseCode = 500
)

// Response 统一响应结构
type Response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    any          `json:"data,omitempty"`
}

// Pagination 分页信息
type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

// ListResponse 列表响应
type ListResponse struct {
	List       any         `json:"list"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Success 成功响应（无数据）
func Success(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
	})
}

// SuccessWithData 成功响应（带数据）
func SuccessWithData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithList 成功响应（带列表和分页）
func SuccessWithList(c *gin.Context, list any, pagination *Pagination) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: ListResponse{
			List:       list,
			Pagination: pagination,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Code:    ResponseCode(statusCode),
		Message: message,
	})
}

// ErrorWithCode 带自定义错误码的错误响应
func ErrorWithCode(c *gin.Context, statusCode int, code ResponseCode, message string) {
	c.JSON(statusCode, Response{
		Code:    code,
		Message: message,
	})
}

// ValidationError 参数验证错误
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 未授权错误
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权，请先登录"
	}
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 禁止访问错误
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "禁止访问"
	}
	Error(c, http.StatusForbidden, message)
}

// NotFound 资源不存在错误
func NotFound(c *gin.Context, resource string) {
	message := "资源不存在"
	if resource != "" {
		message = resource + "不存在"
	}
	Error(c, http.StatusNotFound, message)
}

// InternalError 服务器内部错误
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	Error(c, http.StatusInternalServerError, message)
}

// GetEmployeeID 从上下文获取员工 ID（用于 service 层）
func GetEmployeeID(c *gin.Context) string {
	val, exists := c.Get("employee_id")
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
	val, exists := c.Get("employee_name")
	if !exists {
		return ""
	}
	if name, ok := val.(string); ok {
		return name
	}
	return ""
}

// GetIntQuery 从查询参数获取整数，带默认值
func GetIntQuery(c *gin.Context, key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// GetInt64Query 从查询参数获取 int64，带默认值
func GetInt64Query(c *gin.Context, key string, defaultValue int64) int64 {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intVal
}
