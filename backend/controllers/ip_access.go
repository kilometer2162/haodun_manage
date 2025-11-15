package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetIPAccesses(c *gin.Context) {
	var ipAccesses []models.IPAccess
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.IPAccess{})

	// 搜索条件
	if ip := c.Query("ip"); ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if country := c.Query("country"); country != "" {
		query = query.Where("country LIKE ?", "%"+country+"%")
	}
	if province := c.Query("province"); province != "" {
		query = query.Where("province LIKE ?", "%"+province+"%")
	}
	if city := c.Query("city"); city != "" {
		query = query.Where("city LIKE ?", "%"+city+"%")
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("date >= ?", t)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("date <= ?", t)
		}
	}

	var total int64
	query.Count(&total)
	query.Order("date DESC, access_count DESC").Offset(offset).Limit(pageSize).Find(&ipAccesses)

	c.JSON(http.StatusOK, gin.H{
		"data":  ipAccesses,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func GetIPStatistics(c *gin.Context) {
	// 按日期统计
	var dateStats []struct {
		Date        string `json:"date"`
		AccessCount int    `json:"access_count"`
		IPCount     int    `json:"ip_count"`
	}

	database.DB.Model(&models.IPAccess{}).
		Select("DATE_FORMAT(date, '%Y-%m-%d') as date, SUM(access_count) as access_count, COUNT(DISTINCT ip) as ip_count").
		Group("date").
		Order("date DESC").
		Limit(30).
		Scan(&dateStats)

	// 按地区统计
	var regionStats []struct {
		Country     string `json:"country"`
		Province    string `json:"province"`
		City        string `json:"city"`
		AccessCount int    `json:"access_count"`
		IPCount     int    `json:"ip_count"`
	}

	database.DB.Model(&models.IPAccess{}).
		Select("country, province, city, SUM(access_count) as access_count, COUNT(DISTINCT ip) as ip_count").
		Group("country, province, city").
		Order("access_count DESC").
		Limit(50).
		Scan(&regionStats)

	c.JSON(http.StatusOK, gin.H{
		"date_stats":   dateStats,
		"region_stats": regionStats,
	})
}

