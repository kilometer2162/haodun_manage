package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 构建树形结构
func buildResourceTree(resources []models.Resource, parentID *uint) []models.Resource {
	var tree []models.Resource
	for i := range resources {
		var pid *uint
		if resources[i].ParentID != nil {
			pid = resources[i].ParentID
		}
		if (parentID == nil && pid == nil) || (parentID != nil && pid != nil && *parentID == *pid) {
			children := buildResourceTree(resources, &resources[i].ID)
			resources[i].Children = children
			tree = append(tree, resources[i])
		}
	}
	return tree
}

func GetResources(c *gin.Context) {
	treeMode := c.Query("tree") == "true"
	
	if treeMode {
		// 树形结构模式
		var resources []models.Resource
		database.DB.Order("sort ASC, id ASC").Find(&resources)
		
		tree := buildResourceTree(resources, nil)
		c.JSON(http.StatusOK, gin.H{
			"data": tree,
		})
		return
	}
	
	// 分页模式
	var resources []models.Resource
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.Resource{}).Count(&total)
	database.DB.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&resources)

	c.JSON(http.StatusOK, gin.H{
		"data":  resources,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func GetResource(c *gin.Context) {
	id := c.Param("id")
	var resource models.Resource
	if err := database.DB.First(&resource, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resource)
}

func CreateResource(c *gin.Context) {
	var resource models.Resource
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证父资源是否存在（如果指定了父资源）
	if resource.ParentID != nil {
		var parent models.Resource
		if err := database.DB.First(&parent, *resource.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父资源不存在"})
			return
		}
		// 确保父资源是菜单类型
		if parent.Type != "menu" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父资源必须是菜单类型"})
			return
		}
	}

	if err := database.DB.Create(&resource).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建资源失败"})
		return
	}

	// 重新加载关联数据
	database.DB.Preload("Parent").First(&resource, resource.ID)
	c.JSON(http.StatusOK, resource)
}

func UpdateResource(c *gin.Context) {
	id := c.Param("id")
	var resource models.Resource
	if err := database.DB.First(&resource, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
		return
	}

	var updateData models.Resource
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证父资源（如果指定了父资源）
	if updateData.ParentID != nil {
		// 不能将自己设为父资源
		if *updateData.ParentID == resource.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能将自己设为父资源"})
			return
		}
		// 验证父资源是否存在
		var parent models.Resource
		if err := database.DB.First(&parent, *updateData.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父资源不存在"})
			return
		}
		// 确保父资源是菜单类型
		if parent.Type != "menu" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父资源必须是菜单类型"})
			return
		}
		// 检查是否会形成循环引用（简单检查：父资源不能是自己的子资源）
		var checkChild func(uint) bool
		checkChild = func(parentID uint) bool {
			var children []models.Resource
			database.DB.Where("parent_id = ?", parentID).Find(&children)
			for _, child := range children {
				if child.ID == resource.ID {
					return true
				}
				if checkChild(child.ID) {
					return true
				}
			}
			return false
		}
		if checkChild(resource.ID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能将子资源设为父资源"})
			return
		}
	}

	// 更新资源
	resource.Name = updateData.Name
	resource.Path = updateData.Path
	resource.Method = updateData.Method
	resource.Description = updateData.Description
	resource.Type = updateData.Type
	resource.ParentID = updateData.ParentID
	resource.Sort = updateData.Sort
	resource.Icon = updateData.Icon

	database.DB.Save(&resource)
	
	// 重新加载关联数据
	database.DB.Preload("Parent").First(&resource, resource.ID)
	c.JSON(http.StatusOK, resource)
}

func DeleteResource(c *gin.Context) {
	id := c.Param("id")
	var resource models.Resource
	if err := database.DB.First(&resource, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
		return
	}

	// 检查是否有子资源
	var childrenCount int64
	database.DB.Model(&models.Resource{}).Where("parent_id = ?", id).Count(&childrenCount)
	if childrenCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该资源下存在子资源，请先删除子资源"})
		return
	}

	database.DB.Delete(&resource)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

