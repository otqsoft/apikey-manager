// Package util 通用小工具。
package util

import "strings"

// MaskAPIKey 密钥列表掩码：前 4 后 4，中间星号（契约 0 节要求）。
// 完整明文只允许出现在详情接口与上游转发请求中，禁止落日志。
func MaskAPIKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		// 过短则整体用星号替代，避免泄露
		return strings.Repeat("*", len(key))
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// TruncateRunes 按字符截断标题（契约：首条用户消息前 20 字自动生成标题）。
func TruncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
