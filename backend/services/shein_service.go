package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"haodun_manage/backend/utils"
)

const (
	// SheinBaseURL = "https://openapi.sheincorp.com"
	SheinBaseURL    = "https://openapi-test01.sheincorp.cn"
	OrderListPath   = "/open-api/order/order-list"
	OrderDetailPath = "/open-api/order/order-detail"
)

type SheinService struct {
	AppKey    string
	AppSecret string
}

type OrderListRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Status    string `json:"status,omitempty"`
}

type OrderListItem struct {
	OrderNo string `json:"order_no"`
}

type OrderListResponse struct {
	Code int `json:"code"`
	Data struct {
		Total int             `json:"total"`
		List  []OrderListItem `json:"list"`
	} `json:"data"`
	Message string `json:"message"`
}

type OrderDetailResponse struct {
	Code int `json:"code"`
	Data struct {
		OrderNo     string  `json:"order_no"`
		OrderStatus int     `json:"order_status"`
		OrderAmount float64 `json:"order_amount"`
		Currency    string  `json:"currency"`
		CreateTime  string  `json:"create_time"`
		UpdateTime  string  `json:"update_time"`
		Items       []struct {
			Sku         string  `json:"sku"`
			Quantity    int     `json:"quantity"`
			Price       float64 `json:"price"`
			ProductName string  `json:"product_name"`
		} `json:"items"`
		ShippingAddress struct {
			Name    string `json:"name"`
			Phone   string `json:"phone"`
			Address string `json:"address"`
		} `json:"shipping_address"`
	} `json:"data"`
	Message string `json:"message"`
}

// NewSheinService 创建Shein服务实例，优先从环境变量读取配置
func NewSheinService() *SheinService {
	appKey := os.Getenv("SHEIN_APP_KEY")
	appSecret := os.Getenv("SHEIN_APP_SECRET")

	if appKey == "" || appSecret == "" {
		// 测试用的默认值（请替换为实际值）
		appKey = "your_test_app_key"
		appSecret = "your_test_app_secret"
		fmt.Println("警告：使用默认测试配置，请在生产环境中设置环境变量")
	}

	return &SheinService{
		AppKey:    appKey,
		AppSecret: appSecret,
	}
}

// GetAllOrders 获取全部订单（自动分页）
func (s *SheinService) GetAllOrders(startTime, endTime string, status string) ([]string, error) {
	var allOrders []string
	page := 1
	pageSize := 30

	for {
		orders, total, err := s.getOrderListPage(startTime, endTime, status, page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, order := range orders {
			allOrders = append(allOrders, order.OrderNo)
		}

		// 检查是否获取完所有数据
		if len(allOrders) >= total {
			break
		}

		page++
		// 避免请求过快
		time.Sleep(100 * time.Millisecond)
	}

	return allOrders, nil
}

// GetOrderDetails 批量获取订单详情
func (s *SheinService) GetOrderDetails(orderNos []string) ([]OrderDetailResponse, error) {
	var details []OrderDetailResponse

	for _, orderNo := range orderNos {
		detail, err := s.getOrderDetail(orderNo)
		if err != nil {
			return nil, fmt.Errorf("获取订单 %s 详情失败: %v", orderNo, err)
		}
		details = append(details, *detail)

		// 避免请求过快
		time.Sleep(100 * time.Millisecond)
	}

	return details, nil
}

// getOrderListPage 获取单页订单列表
func (s *SheinService) getOrderListPage(startTime, endTime, status string, page, pageSize int) ([]OrderListItem, int, error) {
	params := map[string]string{
		"start_time": startTime,
		"end_time":   endTime,
		"page":       strconv.Itoa(page),
		"page_size":  strconv.Itoa(pageSize),
	}

	if status != "" {
		params["status"] = status
	}

	// 添加公共参数
	params["app_key"] = s.AppKey
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)

	// 生成签名
	signature := utils.SignUtil.SignRequest(params, s.AppSecret)
	params["sign"] = signature

	// 构建请求URL
	u, _ := url.Parse(SheinBaseURL + OrderListPath)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// 发送请求
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var result OrderListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	if result.Code != 0 {
		return nil, 0, fmt.Errorf("API错误: %s", result.Message)
	}

	return result.Data.List, result.Data.Total, nil
}

// getOrderDetail 获取单个订单详情
func (s *SheinService) getOrderDetail(orderNo string) (*OrderDetailResponse, error) {
	params := map[string]string{
		"order_no":  orderNo,
		"app_key":   s.AppKey,
		"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
	}

	// 生成签名
	signature := utils.SignUtil.SignRequest(params, s.AppSecret)
	params["sign"] = signature

	// 构建请求URL
	u, _ := url.Parse(SheinBaseURL + OrderDetailPath)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// 发送请求
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result OrderDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API错误: %s", result.Message)
	}

	return &result, nil
}

// SyncOrders 同步全部订单（列表+详情）
func (s *SheinService) SyncOrders(startTime, endTime string, status string) ([]OrderDetailResponse, error) {
	// 1. 获取全部订单号
	orderNos, err := s.GetAllOrders(startTime, endTime, status)
	if err != nil {
		return nil, fmt.Errorf("获取订单列表失败: %v", err)
	}

	if len(orderNos) == 0 {
		return []OrderDetailResponse{}, nil
	}

	// 2. 获取所有订单详情
	details, err := s.GetOrderDetails(orderNos)
	if err != nil {
		return nil, fmt.Errorf("获取订单详情失败: %v", err)
	}

	return details, nil
}
