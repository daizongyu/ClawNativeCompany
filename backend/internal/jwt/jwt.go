// Package jwt 提供 JWT Token 生成和验证功能
// 使用 github.com/golang-jwt/jwt/v5
package jwt

import (
	"errors"
	"sync"
	"time"

	"claw/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// 错误定义
var (
	ErrInvalidToken     = errors.New("无效的 token")
	ErrExpiredToken     = errors.New("token 已过期")
	ErrTokenNotYetValid = errors.New("token 尚未生效")
)

// TokenType Token 类型
type TokenType string

const (
	// TokenTypeAccess 访问令牌
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh 刷新令牌
	TokenTypeRefresh TokenType = "refresh"
)

// Claims JWT 声明
type Claims struct {
	EmployeeID string    `json:"emp_id"`
	Name       string    `json:"name"`
	Type       TokenType `json:"type"`
	jwt.RegisteredClaims
}

var (
	expireHours int
	once        sync.Once
)

// initExpireHours 初始化过期时间
func initExpireHours() {
	once.Do(func() {
		expireHours = config.Get().JWT.ExpireHours
		if expireHours < 1 {
			expireHours = 24
		}
	})
}

// GetExpireHours 获取过期时间（小时）
func GetExpireHours() int {
	initExpireHours()
	return expireHours
}

// GenerateToken 生成 JWT Token
// employeeID: 员工 ID
// name: 员工名称
// tokenType: token 类型（access/refresh）
func GenerateToken(employeeID, name string, tokenType TokenType) (string, error) {
	initExpireHours()
	hours := expireHours

	// 刷新令牌有效期更长
	if tokenType == TokenTypeRefresh {
		hours = hours * 7 // 7 天
	}

	claims := Claims{
		EmployeeID: employeeID,
		Name:       name,
		Type:       tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "claw",
			Subject:   employeeID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Get().JWT.Secret))
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func GenerateTokenPair(employeeID, name string) (accessToken, refreshToken string, err error) {
	accessToken, err = GenerateToken(employeeID, name, TokenTypeAccess)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = GenerateToken(employeeID, name, TokenTypeRefresh)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析并验证 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(config.Get().JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotYetValid
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ValidateAccessToken 验证访问令牌
func ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeAccess {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken 验证刷新令牌
func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken 使用刷新令牌生成新的令牌对
func RefreshToken(refreshToken string) (accessToken, newRefreshToken string, err error) {
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	return GenerateTokenPair(claims.EmployeeID, claims.Name)
}

// GetExpireTime 获取 Token 过期时间
func GetExpireTime(tokenString string) (time.Time, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}

// IsExpired 检查 Token 是否已过期
func IsExpired(tokenString string) bool {
	_, err := ParseToken(tokenString)
	return errors.Is(err, ErrExpiredToken)
}
