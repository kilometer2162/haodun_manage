package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IPGeoInfo IP地理位置信息
type IPGeoInfo struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	RegionName  string  `json:"regionName"`  // 省份
	City        string  `json:"city"`
	ISP         string  `json:"isp"`
	Query       string  `json:"query"`
	Message     string  `json:"message,omitempty"`
}

// GetIPGeoInfo 获取IP地理位置信息
// 使用 ip-api.com 免费API（无需API key）
// 限制：每分钟45次请求
func GetIPGeoInfo(ip string) (*IPGeoInfo, error) {
	// 跳过本地IP
	if ip == "127.0.0.1" || ip == "localhost" || ip == "::1" {
		return &IPGeoInfo{
			Country:    "本地",
			RegionName: "本地",
			City:       "本地",
			ISP:        "本地",
		}, nil
	}

	// 构建API URL，使用中文语言
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip)

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 发送请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求IP地理位置API失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IP地理位置API返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON
	var geoInfo IPGeoInfo
	if err := json.Unmarshal(body, &geoInfo); err != nil {
		return nil, fmt.Errorf("解析响应JSON失败: %v", err)
	}

	// 检查API返回的状态
	if geoInfo.Status == "fail" {
		return nil, fmt.Errorf("IP地理位置查询失败: %s", geoInfo.Message)
	}

	// 处理空值，设置为"未知"
	if geoInfo.Country == "" {
		geoInfo.Country = "未知"
	}
	if geoInfo.RegionName == "" {
		geoInfo.RegionName = "未知"
	}
	if geoInfo.City == "" {
		geoInfo.City = "未知"
	}
	if geoInfo.ISP == "" {
		geoInfo.ISP = "未知"
	}

	return &geoInfo, nil
}

// GetIPGeoInfoSafe 安全获取IP地理位置信息，失败时返回默认值
func GetIPGeoInfoSafe(ip string) (country, province, city, isp string) {
	geoInfo, err := GetIPGeoInfo(ip)
	if err != nil {
		// API调用失败，返回默认值
		return "未知", "未知", "未知", "未知"
	}

	return geoInfo.Country, geoInfo.RegionName, geoInfo.City, geoInfo.ISP
}

