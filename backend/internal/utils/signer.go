// Package utils 提供签名工具
package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Signer HMAC-SHA256 签名器
type Signer struct {
	secret []byte
}

// NewSigner 创建签名器
func NewSigner(secret string) *Signer {
	return &Signer{
		secret: []byte(secret),
	}
}

// Sign 生成签名
// timestamp: Unix 时间戳（秒）
// method: HTTP 方法（GET, POST, PUT, DELETE 等）
// path: API 路径（如 /api/v1/webhook/dingtalk）
// body: 请求体（map 会被序列化为 JSON）
func (s *Signer) Sign(timestamp int64, method, path string, body map[string]interface{}) string {
	// 构建签名字符串
	// 格式: timestamp + "\n" + method + "\n" + path + "\n" + sorted_body_json
	signString := fmt.Sprintf("%d\n%s\n%s\n%s",
		timestamp,
		strings.ToUpper(method),
		path,
		s.serializeBody(body),
	)

	// 使用 HMAC-SHA256 计算签名
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(signString))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// Verify 验证签名
// timestamp: 签名时的时间戳
// method: HTTP 方法
// path: API 路径
// body: 请求体
// signature: 待验证的签名
// maxAge: 签名最大有效期（秒），0 表示不检查
func (s *Signer) Verify(timestamp int64, method, path string, body map[string]interface{}, signature string, maxAge int) bool {
	// 检查时间戳（防止重放攻击）
	if maxAge > 0 {
		now := time.Now().Unix()
		if timestamp < now-int64(maxAge) || timestamp > now+60 {
			return false
		}
	}

	// 计算期望的签名
	expected := s.Sign(timestamp, method, path, body)

	// 使用 constant-time 比较防止时序攻击
	return hmac.Equal([]byte(signature), []byte(expected))
}

// SignWithHeader 生成带时间戳的签名头
// 返回格式: "v1=" + hex(signature)
func (s *Signer) SignWithHeader(timestamp int64, method, path string, body map[string]interface{}) string {
	signature := s.Sign(timestamp, method, path, body)
	return "v1=" + signature
}

// VerifyHeader 验证签名头
func (s *Signer) VerifyHeader(header string, timestamp int64, method, path string, body map[string]interface{}, maxAge int) bool {
	// 解析版本和签名
	parts := strings.SplitN(header, "=", 2)
	if len(parts) != 2 {
		return false
	}

	version := parts[0]
	signature := parts[1]

	// 目前只支持 v1
	if version != "v1" {
		return false
	}

	return s.Verify(timestamp, method, path, body, signature, maxAge)
}

// serializeBody 序列化请求体
// 对 map 按键排序后序列化为 JSON
func (s *Signer) serializeBody(body map[string]interface{}) string {
	if body == nil || len(body) == 0 {
		return ""
	}

	// 按键排序
	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建有序 map
	sorted := make(map[string]interface{})
	for _, k := range keys {
		sorted[k] = body[k]
	}

	// 序列化为 JSON
	jsonBytes, err := json.Marshal(sorted)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

// GenerateNonce 生成随机 nonce
func GenerateNonce() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// SignQuery 对查询参数签名
// params: 查询参数（会被排序）
func (s *Signer) SignQuery(params map[string]string) string {
	// 按键排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建查询字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	queryString := strings.Join(parts, "&")

	// 计算签名
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}
