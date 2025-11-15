package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetRoles(c *gin.Context) {
	var roles []models.Role
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.Role{}).Count(&total)
	database.DB.Preload("Permissions").Preload("Permissions.Resource").Offset(offset).Limit(pageSize).Find(&roles)

	c.JSON(http.StatusOK, gin.H{
		"data":  roles,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func GetRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := database.DB.Preload("Permissions").Preload("Permissions.Resource").First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func CreateRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建角色失败"})
		return
	}

	c.JSON(http.StatusOK, role)
}

func UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}

	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Save(&role)
	c.JSON(http.StatusOK, role)
}

func DeleteRole(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.Role{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func AssignPermissions(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var permissions []models.Permission
	database.DB.Where("id IN ?", req.PermissionIDs).Find(&permissions)
	database.DB.Model(&role).Association("Permissions").Replace(permissions)

	c.JSON(http.StatusOK, gin.H{"message": "权限分配成功"})
}

