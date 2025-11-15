package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetLogs(c *gin.Context) {
	var logs []models.Log
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Log{})

	// 搜索条件
	if username := c.Query("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if module := c.Query("module"); module != "" {
		query = query.Where("module LIKE ?", "%"+module+"%")
	}
	if ip := c.Query("ip"); ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var total int64
	query.Count(&total)
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetLogOptions(c *gin.Context) {
	var actionList []string
	database.DB.Model(&models.Log{}).
		Distinct("action").
		Order("action").
		Pluck("action", &actionList)

	// 过滤空字符串
	filteredActions := make([]string, 0, len(actionList))
	for _, action := range actionList {
		if action != "" {
			filteredActions = append(filteredActions, action)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"actions": filteredActions,
		"statuses": []gin.H{
			{"label": "成功", "value": "1"},
			{"label": "失败", "value": "0"},
		},
	})
}

func GetLog(c *gin.Context) {
	id := c.Param("id")
	var log models.Log
	if err := database.DB.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "日志不存在"})
		return
	}
	c.JSON(http.StatusOK, log)
}
