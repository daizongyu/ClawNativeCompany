// Package validator 提供参数校验功能
package validator

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

// Validate 全局校验器实例
var (
	validate *validator.Validate
	once     sync.Once
)

// Init 初始化校验器
func Init() {
	once.Do(func() {
		validate = validator.New()
		registerCustomValidators()
	})
}

// Get 获取校验器实例
func Get() *validator.Validate {
	if validate == nil {
		Init()
	}
	return validate
}

// registerCustomValidators 注册自定义校验规则
func registerCustomValidators() {
	// 注册自定义校验标签
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("password", validatePassword)
}

// ValidateStruct 校验结构体
func ValidateStruct(s any) error {
	return Get().Struct(s)
}

// ValidateVar 校验单个变量
func ValidateVar(v any, tag string) error {
	return Get().Var(v, tag)
}

// ValidationError 校验错误信息
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors 校验错误列表
type ValidationErrors []ValidationError

// Error 实现 error 接口
func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// FormatValidationError 格式化校验错误
func FormatValidationError(err error) ValidationErrors {
	if err == nil {
		return nil
	}

	var errors ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}
	return errors
}

// getErrorMessage 获取校验错误提示信息
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "该字段为必填项"
	case "email":
		return "邮箱格式不正确"
	case "min":
		return fmt.Sprintf("长度不能少于 %s 个字符", e.Param())
	case "max":
		return fmt.Sprintf("长度不能超过 %s 个字符", e.Param())
	case "len":
		return fmt.Sprintf("长度必须为 %s 个字符", e.Param())
	case "phone":
		return "手机号格式不正确"
	case "password":
		return "密码必须包含字母和数字，长度8-32位"
	case "uuid":
		return "UUID 格式不正确"
	case "oneof":
		return fmt.Sprintf("必须是以下值之一: %s", e.Param())
	case "gt":
		return fmt.Sprintf("必须大于 %s", e.Param())
	case "gte":
		return fmt.Sprintf("必须大于或等于 %s", e.Param())
	case "lt":
		return fmt.Sprintf("必须小于 %s", e.Param())
	case "lte":
		return fmt.Sprintf("必须小于或等于 %s", e.Param())
	default:
		return fmt.Sprintf("校验失败: %s", e.Tag())
	}
}

// 自定义校验函数

// validatePhone 校验手机号
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// 支持中国大陆手机号：1开头，第二位3-9，共11位
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// validatePassword 校验密码强度
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	// 8-32位，必须包含至少一个字母和一个数字
	if len(password) < 8 || len(password) > 32 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasLetter && hasNumber
}

// 便捷校验函数

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty 检查字符串是否非空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// IsEmail 检查是否为有效邮箱
func IsEmail(s string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// IsPhone 检查是否为有效手机号
func IsPhone(s string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// IsUUID 检查是否为有效 UUID
func IsUUID(s string) bool {
	pattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// Length 获取字符串长度（支持中文）
func Length(s string) int {
	return utf8.RuneCountInString(s)
}

// MaxLength 检查字符串最大长度
func MaxLength(s string, max int) bool {
	return Length(s) <= max
}

// MinLength 检查字符串最小长度
func MinLength(s string, min int) bool {
	return Length(s) >= min
}

// RangeLength 检查字符串长度范围
func RangeLength(s string, min, max int) bool {
	length := Length(s)
	return length >= min && length <= max
}
