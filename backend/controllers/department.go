package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
)

// ListDepartments 获取部门列表
func ListDepartments(c *gin.Context) {
	var departments []models.Department

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Department{})

	if statusParam := c.Query("status"); statusParam != "" {
		if status, err := strconv.Atoi(statusParam); err == nil {
			query = query.Where("status = ?", status)
		}
	}

	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", likeKeyword, likeKeyword)
	}

	query = query.Preload("Parent")

	if c.Query("simple") == "1" {
		if err := query.Order("sort ASC").Order("id ASC").Find(&departments).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取部门列表失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": departments})
		return
	}

	offset := (page - 1) * pageSize
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取部门数量失败"})
		return
	}

	if err := query.Order("sort ASC").Order("id ASC").Offset(offset).Limit(pageSize).Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取部门列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      departments,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDepartment 获取部门详情
func GetDepartment(c *gin.Context) {
	id := c.Param("id")
	var department models.Department
	if err := database.DB.First(&department, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "部门不存在"})
		return
	}
	c.JSON(http.StatusOK, department)
}

type departmentRequest struct {
	Name        string `json:"name" binding:"required"`
	ParentID    *uint  `json:"parent_id"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
	Sort        *int   `json:"sort"`
}

// CreateDepartment 创建部门
func CreateDepartment(c *gin.Context) {
	var req departmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ParentID != nil && *req.ParentID != 0 {
		var parent models.Department
		if err := database.DB.First(&parent, *req.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父部门不存在"})
			return
		}
	}

	dept := models.Department{
		Name:        strings.TrimSpace(req.Name),
		ParentID:    req.ParentID,
		Description: req.Description,
		Status:      1,
		Sort:        0,
	}

	if req.Status != nil {
		dept.Status = *req.Status
	}
	if req.Sort != nil {
		dept.Sort = *req.Sort
	}

	if err := database.DB.Create(&dept).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建部门失败"})
		return
	}

	database.DB.Preload("Parent").First(&dept, dept.ID)
	c.JSON(http.StatusOK, dept)
}

// UpdateDepartment 更新部门
func UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	var department models.Department
	if err := database.DB.First(&department, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "部门不存在"})
		return
	}

	var req departmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ParentID != nil {
		if *req.ParentID == department.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父部门不能为自身"})
			return
		}
		if *req.ParentID != 0 {
			var parent models.Department
			if err := database.DB.First(&parent, *req.ParentID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "父部门不存在"})
				return
			}
		}
		department.ParentID = req.ParentID
	}

	if req.Name != "" {
		department.Name = strings.TrimSpace(req.Name)
	}
	department.Description = req.Description

	if req.Status != nil {
		department.Status = *req.Status
	}
	if req.Sort != nil {
		department.Sort = *req.Sort
	}

	if err := database.DB.Save(&department).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "更新部门失败"})
		return
	}

	database.DB.Preload("Parent").First(&department, department.ID)
	c.JSON(http.StatusOK, department)
}

// DeleteDepartment 删除部门
func DeleteDepartment(c *gin.Context) {
	id := c.Param("id")

	var department models.Department
	if err := database.DB.First(&department, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "部门不存在"})
		return
	}

	var childCount int64
	database.DB.Model(&models.Department{}).Where("parent_id = ?", department.ID).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "存在子部门，无法删除"})
		return
	}

	var userCount int64
	database.DB.Model(&models.User{}).Where("department_id = ?", department.ID).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该部门下存在用户，无法删除"})
		return
	}

	if err := database.DB.Delete(&department).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除部门失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
