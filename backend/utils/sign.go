package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// SignUtil 签名工具
var SignUtil = &signUtil{}

type signUtil struct{}

// BuildQueryString 构建待签名字符串
func (s *signUtil) BuildQueryString(params map[string]string) string {
	// 过滤空值并排序
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接 key=value&key=value
	pairs := make([]string, len(keys))
	for i, k := range keys {
		pairs[i] = k + "=" + params[k]
	}
	return strings.Join(pairs, "&")
}

// Md5 MD5 签名
func (s *signUtil) Md5(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HmacSHA256 HMAC-SHA256 签名
func (s *signUtil) HmacSHA256(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignRequest 生成请求签名
func (s *signUtil) SignRequest(params map[string]string, secret string) string {
	queryString := s.BuildQueryString(params)
	return s.HmacSHA256(queryString, secret)
}

// VerifySignature 验证签名
func (s *signUtil) VerifySignature(params map[string]string, secret, signature string) bool {
	expected := s.SignRequest(params, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
