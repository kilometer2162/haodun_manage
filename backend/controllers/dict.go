package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ========== 字典类型相关接口 ==========

// GetDictTypes 获取字典类型列表
func GetDictTypes(c *gin.Context) {
	var dictTypes []models.DictType
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.DictType{}).Count(&total)
	database.DB.Preload("Items").Offset(offset).Limit(pageSize).Order("id DESC").Find(&dictTypes)

	c.JSON(http.StatusOK, gin.H{
		"data":      dictTypes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDictType 获取单个字典类型
func GetDictType(c *gin.Context) {
	id := c.Param("id")
	var dictType models.DictType
	if err := database.DB.Preload("Items", "deleted_at IS NULL").Order("sort ASC, id ASC").First(&dictType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "字典类型不存在"})
		return
	}
	c.JSON(http.StatusOK, dictType)
}

// CreateDictType 创建字典类型
func CreateDictType(c *gin.Context) {
	var dictType models.DictType
	if err := c.ShouldBindJSON(&dictType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查代码是否已存在
	var existing models.DictType
	if err := database.DB.Where("code = ?", dictType.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "字典类型代码已存在"})
		return
	}

	if err := database.DB.Create(&dictType).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建字典类型失败"})
		return
	}

	database.DB.Preload("Items").First(&dictType, dictType.ID)
	c.JSON(http.StatusOK, dictType)
}

// UpdateDictType 更新字典类型
func UpdateDictType(c *gin.Context) {
	id := c.Param("id")
	var dictType models.DictType
	if err := database.DB.First(&dictType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "字典类型不存在"})
		return
	}

	var req models.DictType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查代码是否与其他记录冲突
	if req.Code != dictType.Code {
		var existing models.DictType
		if err := database.DB.Where("code = ? AND id != ?", req.Code, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "字典类型代码已存在"})
			return
		}
	}

	dictType.Name = req.Name
	dictType.Code = req.Code
	dictType.Description = req.Description
	dictType.Status = req.Status

	database.DB.Save(&dictType)
	database.DB.Preload("Items").First(&dictType, dictType.ID)
	c.JSON(http.StatusOK, dictType)
}

// DeleteDictType 删除字典类型
func DeleteDictType(c *gin.Context) {
	id := c.Param("id")
	
	// 检查是否有字典项
	var count int64
	database.DB.Model(&models.DictItem{}).Where("type_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该字典类型下存在字典项，请先删除字典项"})
		return
	}

	database.DB.Delete(&models.DictType{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ========== 字典项相关接口 ==========

// GetDictItems 获取字典项列表
func GetDictItems(c *gin.Context) {
	typeID := c.Query("type_id")
	var dictItems []models.DictItem
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.DictItem{})
	if typeID != "" {
		query = query.Where("type_id = ?", typeID)
	}

	var total int64
	query.Count(&total)
	query.Preload("Type").Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&dictItems)

	c.JSON(http.StatusOK, gin.H{
		"data":      dictItems,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDictItem 获取单个字典项
func GetDictItem(c *gin.Context) {
	id := c.Param("id")
	var dictItem models.DictItem
	if err := database.DB.Preload("Type").First(&dictItem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "字典项不存在"})
		return
	}
	c.JSON(http.StatusOK, dictItem)
}

// CreateDictItem 创建字典项
func CreateDictItem(c *gin.Context) {
	var dictItem models.DictItem
	if err := c.ShouldBindJSON(&dictItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查字典类型是否存在
	var dictType models.DictType
	if err := database.DB.First(&dictType, dictItem.TypeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "字典类型不存在"})
		return
	}

	// 检查同一类型下值是否重复
	var existing models.DictItem
	if err := database.DB.Where("type_id = ? AND value = ?", dictItem.TypeID, dictItem.Value).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该字典类型下已存在相同的值"})
		return
	}

	if err := database.DB.Create(&dictItem).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建字典项失败"})
		return
	}

	database.DB.Preload("Type").First(&dictItem, dictItem.ID)
	c.JSON(http.StatusOK, dictItem)
}

// UpdateDictItem 更新字典项
func UpdateDictItem(c *gin.Context) {
	id := c.Param("id")
	var dictItem models.DictItem
	if err := database.DB.First(&dictItem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "字典项不存在"})
		return
	}

	var req models.DictItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查值是否与其他记录冲突
	if req.Value != dictItem.Value || req.TypeID != dictItem.TypeID {
		var existing models.DictItem
		if err := database.DB.Where("type_id = ? AND value = ? AND id != ?", req.TypeID, req.Value, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该字典类型下已存在相同的值"})
			return
		}
	}

	dictItem.TypeID = req.TypeID
	dictItem.Label = req.Label
	dictItem.Value = req.Value
	dictItem.Sort = req.Sort
	dictItem.Status = req.Status
	dictItem.Description = req.Description

	database.DB.Save(&dictItem)
	database.DB.Preload("Type").First(&dictItem, dictItem.ID)
	c.JSON(http.StatusOK, dictItem)
}

// DeleteDictItem 删除字典项
func DeleteDictItem(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.DictItem{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetDictByCode 根据字典类型代码获取字典项列表（用于前端下拉选择等）
func GetDictByCode(c *gin.Context) {
	code := c.Param("code")
	var dictType models.DictType
	if err := database.DB.Where("code = ? AND status = 1", code).First(&dictType).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "字典类型不存在"})
		return
	}

	var items []models.DictItem
	database.DB.Where("type_id = ? AND status = 1", dictType.ID).Order("sort ASC, id ASC").Find(&items)

	c.JSON(http.StatusOK, gin.H{
		"code":  dictType.Code,
		"name":  dictType.Name,
		"items": items,
	})
}

