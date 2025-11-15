package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetConfigs 获取系统参数列表
func GetConfigs(c *gin.Context) {
	var configs []models.Config
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	group := c.Query("group")
	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Config{})
	if group != "" {
		query = query.Where("`group` = ?", group)
	}

	var total int64
	query.Count(&total)

	if err := query.Order("`group` ASC, sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询系统参数失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      configs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetConfig 获取单个系统参数
func GetConfig(c *gin.Context) {
	id := c.Param("id")
	var config models.Config
	if err := database.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "系统参数不存在"})
		return
	}
	c.JSON(http.StatusOK, config)
}

// CreateConfig 创建系统参数
func CreateConfig(c *gin.Context) {
	var config models.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查键是否已存在
	var existing models.Config
	if err := database.DB.Where("`key` = ?", config.Key).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数键已存在"})
		return
	}

	if err := database.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建系统参数失败"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig 更新系统参数
func UpdateConfig(c *gin.Context) {
	id := c.Param("id")
	var config models.Config
	if err := database.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "系统参数不存在"})
		return
	}

	var req models.Config
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查键是否与其他记录冲突
	if req.Key != config.Key {
		var existing models.Config
		if err := database.DB.Where("`key` = ? AND id != ?", req.Key, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数键已存在"})
			return
		}
	}

	config.Key = req.Key
	config.Value = req.Value
	config.Label = req.Label
	config.Type = req.Type
	config.Group = req.Group
	config.Description = req.Description
	config.Sort = req.Sort
	config.Status = req.Status

	database.DB.Save(&config)
	c.JSON(http.StatusOK, config)
}

// DeleteConfig 删除系统参数
func DeleteConfig(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.Config{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetConfigByKey 根据键获取参数值（用于前端获取配置）
func GetConfigByKey(c *gin.Context) {
	key := c.Param("key")
	var config models.Config
	if err := database.DB.Where("`key` = ? AND status = 1", key).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "系统参数不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   config.Key,
		"value": config.Value,
		"label": config.Label,
		"type":  config.Type,
	})
}

// GetConfigGroups 获取参数分组列表
func GetConfigGroups(c *gin.Context) {
	var groups []string
	database.DB.Model(&models.Config{}).Distinct("group").Pluck("group", &groups)
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}
