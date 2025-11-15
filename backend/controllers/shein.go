package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"haodun_manage/backend/services"
)

type SheinController struct {
	sheinService *services.SheinService
}

func NewSheinController() *SheinController {
	return &SheinController{
		sheinService: services.NewSheinService(),
	}
}

// SyncOrders 同步Shein订单接口
// @Summary 同步Shein订单
// @Description 获取指定时间范围内的全部Shein订单（含详情）
// @Tags Shein
// @Accept json
// @Produce json
// @Param start_time query string true "开始时间 (格式: 2006-01-02 15:04:05)"
// @Param end_time query string true "结束时间 (格式: 2006-01-02 15:04:05)"
// @Param status query string false "订单状态"
// @Success 200 {object} gin.H{"data": []services.OrderDetailResponse}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /api/shein/sync-orders [get]
func (c *SheinController) SyncOrders(ctx *gin.Context) {
	startTime := ctx.Query("start_time")
	endTime := ctx.Query("end_time")
	status := ctx.Query("status")

	if startTime == "" || endTime == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_time 和 end_time 参数必填"})
		return
	}

	// 验证时间格式
	_, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_time 格式错误，应为 2006-01-02 15:04:05"})
		return
	}

	_, err = time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "end_time 格式错误，应为 2006-01-02 15:04:05"})
		return
	}

	// 同步订单
	orders, err := c.sheinService.SyncOrders(startTime, endTime, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    orders,
		"count":   len(orders),
		"message": "同步成功",
	})
}

// GetOrderList 获取订单列表（仅列表，不含详情）
// @Summary 获取Shein订单列表
// @Description 获取指定时间范围内的订单号列表
// @Tags Shein
// @Accept json
// @Produce json
// @Param start_time query string true "开始时间"
// @Param end_time query string true "结束时间"
// @Param status query string false "订单状态"
// @Success 200 {object} gin.H{"data": []string}
// @Router /api/shein/order-list [get]
func (c *SheinController) GetOrderList(ctx *gin.Context) {
	startTime := ctx.Query("start_time")
	endTime := ctx.Query("end_time")
	status := ctx.Query("status")

	if startTime == "" || endTime == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_time 和 end_time 参数必填"})
		return
	}

	orderNos, err := c.sheinService.GetAllOrders(startTime, endTime, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  orderNos,
		"count": len(orderNos),
	})
}

// GetOrderDetail 获取单个订单详情
// @Summary 获取Shein订单详情
// @Description 根据订单号获取订单详情
// @Tags Shein
// @Accept json
// @Produce json
// @Param order_no query string true "订单号"
// @Success 200 {object} services.OrderDetailResponse
// @Router /api/shein/order-detail [get]
func (c *SheinController) GetOrderDetail(ctx *gin.Context) {
	orderNo := ctx.Query("order_no")
	if orderNo == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "order_no 参数必填"})
		return
	}

	detail, err := c.sheinService.GetOrderDetails([]string{orderNo})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(detail) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	ctx.JSON(http.StatusOK, detail[0])
}
