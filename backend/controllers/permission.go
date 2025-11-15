package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPermissions(c *gin.Context) {
	var permissions []models.Permission
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.Permission{}).Count(&total)
	database.DB.Preload("Resource").Offset(offset).Limit(pageSize).Find(&permissions)

	c.JSON(http.StatusOK, gin.H{
		"data":  permissions,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func GetPermission(c *gin.Context) {
	id := c.Param("id")
	var permission models.Permission
	if err := database.DB.Preload("Resource").First(&permission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		return
	}
	c.JSON(http.StatusOK, permission)
}

func CreatePermission(c *gin.Context) {
	var permission models.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&permission).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建权限失败"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

func UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	var permission models.Permission
	if err := database.DB.First(&permission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		return
	}

	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Save(&permission)
	c.JSON(http.StatusOK, permission)
}

func DeletePermission(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.Permission{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

