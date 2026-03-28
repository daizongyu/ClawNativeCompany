// Package password 提供密码加密和验证功能
// 使用 bcrypt 算法，cost=10
package password

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// 错误定义
var (
	ErrPasswordTooShort = errors.New("密码长度不能少于 8 位")
	ErrPasswordTooLong  = errors.New("密码长度不能超过 72 字节")
	ErrPasswordMismatch = errors.New("密码不匹配")
	ErrInvalidHash      = errors.New("无效的密码哈希")
)

// DefaultCost bcrypt 默认 cost 值
const DefaultCost = 10

// Hash 使用 bcrypt 加密密码
// 返回 base64 编码的哈希字符串
func Hash(password string) (string, error) {
	// 验证密码长度
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	// 生成哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	// 返回 base64 编码
	return base64.StdEncoding.EncodeToString(hash), nil
}

// Verify 验证密码是否匹配
// hashedPassword 是 Hash 函数返回的 base64 编码字符串
func Verify(password, hashedPassword string) error {
	// 解码 base64
	hash, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return ErrInvalidHash
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return fmt.Errorf("密码验证失败: %w", err)
	}

	return nil
}

// IsValid 检查密码哈希是否有效（格式正确）
func IsValid(hashedPassword string) bool {
	hash, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false
	}
	// bcrypt 哈希以特定前缀开头
	return len(hash) > 0 && (hash[0] == '$')
}

// GenerateRandomPassword 生成随机密码
// length: 密码长度（默认16）
func GenerateRandomPassword(length int) string {
	if length < 8 {
		length = 16
	}

	// 字符集
	const (
		lower   = "abcdefghijklmnopqrstuvwxyz"
		upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits  = "0123456789"
		special = "!@#$%^&*"
	)
	all := lower + upper + digits + special

	// 确保包含各类字符
	password := make([]byte, length)
	password[0] = lower[randInt(len(lower))]
	password[1] = upper[randInt(len(upper))]
	password[2] = digits[randInt(len(digits))]
	password[3] = special[randInt(len(special))]

	// 剩余字符随机
	for i := 4; i < length; i++ {
		password[i] = all[randInt(len(all))]
	}

	// Fisher-Yates 洗牌
	for i := len(password) - 1; i > 0; i-- {
		j := randInt(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

// GenerateAPIKey 生成 API Key
// 32字节随机数据，base64编码，长度44字符
func GenerateAPIKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("生成 API Key 失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// randInt 生成 [0, max) 范围的随机整数
func randInt(max int) int {
	if max <= 0 {
		return 0
	}
	b := make([]byte, 4)
	rand.Read(b)
	return int((uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)) % max
}

// CheckStrength 检查密码强度
// 返回强度评分（0-4）和提示信息
type Strength struct {
	Score   int      `json:"score"`   // 0-4
	Level   string   `json:"level"`   // very_weak, weak, fair, strong, very_strong
	Hints   []string `json:"hints"`   // 改进建议
	IsValid bool     `json:"is_valid"` // 是否可用
}

// CheckStrength 检查密码强度
func CheckStrength(password string) Strength {
	s := Strength{Score: 0, Hints: []string{}}

	// 长度检查
	length := len(password)
	if length < 8 {
		s.Hints = append(s.Hints, "密码长度至少 8 位")
	} else if length >= 12 {
		s.Score++
	}

	// 字符类型检查
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case c < ' ' || c > '~':
			hasSpecial = true
		default:
			hasSpecial = true
		}
	}

	charTypes := 0
	if hasLower {
		charTypes++
	} else {
		s.Hints = append(s.Hints, "包含小写字母")
	}
	if hasUpper {
		charTypes++
	} else {
		s.Hints = append(s.Hints, "包含大写字母")
	}
	if hasDigit {
		charTypes++
	} else {
		s.Hints = append(s.Hints, "包含数字")
	}
	if hasSpecial {
		charTypes++
	} else {
		s.Hints = append(s.Hints, "包含特殊字符")
	}

	s.Score += charTypes

	// 确定等级
	switch s.Score {
	case 0, 1:
		s.Level = "very_weak"
		s.IsValid = false
	case 2:
		s.Level = "weak"
		s.IsValid = length >= 8
	case 3:
		s.Level = "fair"
		s.IsValid = true
	case 4:
		s.Level = "strong"
		s.IsValid = true
	case 5:
		s.Level = "very_strong"
		s.IsValid = true
	}

	return s
}
