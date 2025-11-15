package controllers

import (
	"haodun_manage/backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSystemMetrics 返回服务器系统资源使用情况
func GetSystemMetrics(c *gin.Context) {
	metrics, err := utils.GetSystemMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统监控数据失败", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"memory": gin.H{
			"total":   metrics.Memory.Total,
			"used":    metrics.Memory.Used,
			"free":    metrics.Memory.Free,
			"percent": metrics.Memory.Percent,
		},
		"cpu": gin.H{
			"cores":       metrics.CPU.Cores,
			"usage":       metrics.CPU.Usage,
			"user":        metrics.CPU.User,
			"system":      metrics.CPU.System,
			"idle":        metrics.CPU.Idle,
			"description": metrics.CPU.Description,
		},
		"disk": gin.H{
			"path":    metrics.Disk.Path,
			"total":   metrics.Disk.Total,
			"used":    metrics.Disk.Used,
			"free":    metrics.Disk.Free,
			"percent": metrics.Disk.Percent,
		},
		"updated_at": metrics.UpdateTime.Format("2006-01-02 15:04:05"),
	})
}
