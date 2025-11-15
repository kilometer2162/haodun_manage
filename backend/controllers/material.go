package controllers

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"haodun_manage/backend/utils"
)

type materialFolderRequest struct {
	Name     string  `json:"name" binding:"required"`
	ParentID *uint64 `json:"parent_id"`
}

type materialFolderUpdateRequest struct {
	Name     *string `json:"name"`
	ParentID *uint64 `json:"parent_id"`
}

var allowedImageFormats = map[string]struct{}{
	"jpg":  {},
	"jpeg": {},
	"png":  {},
	"gif":  {},
	"bmp":  {},
	"webp": {},
}

func ListMaterialFolders(c *gin.Context) {
	var folders []models.MaterialFolder
	if err := database.DB.Order("path ASC").Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件夹列表失败"})
		return
	}

	if c.Query("flat") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": folders})
		return
	}

	tree := buildMaterialFolderTree(folders)
	c.JSON(http.StatusOK, gin.H{"data": tree})
}

func CreateMaterialFolder(c *gin.Context) {
	var req materialFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹名称不能为空"})
		return
	}

	var parent *models.MaterialFolder
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			req.ParentID = nil
		} else {
			parent = &models.MaterialFolder{}
			if err := database.DB.First(parent, *req.ParentID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "父文件夹不存在"})
				return
			}
		}
	}

	path := name
	if parent != nil && parent.Path != "" {
		path = fmt.Sprintf("%s/%s", strings.TrimSuffix(parent.Path, "/"), name)
	}

	folder := models.MaterialFolder{
		Name:     name,
		ParentID: req.ParentID,
		Path:     path,
	}

	if err := database.DB.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件夹失败"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func UpdateMaterialFolder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹ID格式错误"})
		return
	}

	var folder models.MaterialFolder
	if err := database.DB.First(&folder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件夹不存在"})
		return
	}

	var req materialFolderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newName := folder.Name
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹名称不能为空"})
			return
		}
		newName = trimmed
	}

	var newParentID *uint64 = folder.ParentID
	var newParent *models.MaterialFolder
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			newParentID = nil
		} else {
			if *req.ParentID == folder.ID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "父文件夹不能为自身"})
				return
			}
			parent := models.MaterialFolder{}
			if err := database.DB.First(&parent, *req.ParentID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "父文件夹不存在"})
				return
			}
			if strings.HasPrefix(parent.Path+"/", folder.Path+"/") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "不能将文件夹移动到自己的子级"})
				return
			}
			newParent = &parent
			newParentID = req.ParentID
		}
	} else if folder.ParentID != nil {
		// keep current parent loaded for path building
		parent := models.MaterialFolder{}
		if err := database.DB.First(&parent, *folder.ParentID).Error; err == nil {
			newParent = &parent
		}
	}

	newPath := newName
	if newParent != nil {
		newPath = fmt.Sprintf("%s/%s", strings.TrimSuffix(newParent.Path, "/"), newName)
	}

	oldPath := folder.Path

	folder.Name = newName
	folder.ParentID = newParentID
	folder.Path = newPath

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法开启事务"})
		return
	}

	if err := tx.Model(&models.MaterialFolder{}).Where("id = ?", folder.ID).Updates(map[string]interface{}{
		"name":      folder.Name,
		"parent_id": folder.ParentID,
		"path":      folder.Path,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文件夹失败"})
		return
	}

	if oldPath != newPath {
		if err := updateMaterialFolderChildPaths(tx, oldPath, newPath); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新子文件夹路径失败"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func DeleteMaterialFolder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹ID格式错误"})
		return
	}

	var folder models.MaterialFolder
	if err := database.DB.First(&folder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件夹不存在"})
		return
	}

	var childCount int64
	database.DB.Model(&models.MaterialFolder{}).Where("parent_id = ?", folder.ID).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "存在子文件夹，无法删除"})
		return
	}

	var materialCount int64
	database.DB.Model(&models.MaterialAsset{}).Where("folder_id = ?", folder.ID).Count(&materialCount)
	if materialCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹下存在素材，无法删除"})
		return
	}

	if err := database.DB.Delete(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件夹失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

type materialFilter struct {
	Keyword  string
	FolderID *uint64
	Shape    string
	Format   string
}

type materialRequest struct {
	Code       string  `json:"code"`
	FileName   string  `json:"file_name"`
	Title      string  `json:"title"`
	Width      *int    `json:"width"`
	Height     *int    `json:"height"`
	Format     string  `json:"format"`
	FileSize   *int64  `json:"file_size"`
	OrderCount *int    `json:"order_count"`
	FolderID   *uint64 `json:"folder_id"`
	Shape      string  `json:"shape"`
	Storage    string  `json:"storage"`
	FilePath   string  `json:"file_path"`
}

func ListMaterials(c *gin.Context) {
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "10"), 10)
	if pageSize > 200 {
		pageSize = 200
	}

	filters := materialFilter{
		Keyword: strings.TrimSpace(c.Query("keyword")),
		Shape:   strings.TrimSpace(c.Query("shape")),
		Format:  strings.TrimSpace(c.Query("format")),
	}

	if folderIDParam := c.Query("folder_id"); folderIDParam != "" {
		if v, err := strconv.ParseUint(folderIDParam, 10, 64); err == nil {
			filters.FolderID = &v
		}
	}

	query := database.DB.Model(&models.MaterialAsset{})

	if filters.Keyword != "" {
		like := "%" + filters.Keyword + "%"
		query = query.Where("code LIKE ? OR file_name LIKE ? OR title LIKE ?", like, like, like)
	}
	if filters.FolderID != nil {
		query = query.Where("folder_id = ?", *filters.FolderID)
	}
	if filters.Shape != "" {
		query = query.Where("shape = ?", filters.Shape)
	}
	if filters.Format != "" {
		query = query.Where("format = ?", filters.Format)
	}
	if !isAdminUser(c) {
		query = query.Where("created_by = ?", currentUserID(c))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取素材数量失败"})
		return
	}

	if total == 0 {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{
			"materials": []models.MaterialAsset{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		}})
		return
	}

	maxPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	if page > maxPage {
		page = maxPage
	}

	offset := (page - 1) * pageSize
	var materials []models.MaterialAsset
	dataQuery := query.Session(&gorm.Session{})
	if err := dataQuery.Preload("Folder").Order("id DESC").Offset(offset).Limit(pageSize).Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取素材列表失败"})
		return
	}

	responses := enrichMaterialAssets(c, materials)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"materials": responses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}})
}

func GetMaterial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "素材ID格式错误"})
		return
	}

	var material models.MaterialAsset
	if err := database.DB.Preload("Folder").First(&material, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "素材不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取素材失败"})
		return
	}

	if !canAccessMaterial(c, material) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该素材"})
		return
	}

	enriched := enrichMaterialAssets(c, []models.MaterialAsset{material})
	if len(enriched) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理素材数据失败"})
		return
	}

	c.JSON(http.StatusOK, enriched[0])
}

func CreateMaterial(c *gin.Context) {
	var req materialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件名不能为空"})
		return
	}

	var folderPath string
	if req.FolderID != nil {
		var folder models.MaterialFolder
		if err := database.DB.First(&folder, *req.FolderID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "归属文件夹不存在"})
			return
		}
		folderPath = folder.Path
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = utils.GenerateMaterialCode()
	} else {
		var count int64
		database.DB.Model(&models.MaterialAsset{}).Where("code = ?", code).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "素材编号已存在"})
			return
		}
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		baseName := strings.TrimSuffix(req.FileName, filepath.Ext(req.FileName))
		if baseName != "" {
			title = baseName
		} else {
			title = req.FileName
		}
	}

	width, height := extractDimension(req.Width), extractDimension(req.Height)
	shape := strings.TrimSpace(req.Shape)
	if shape == "" {
		shape = utils.DetermineMaterialShape(width, height)
	}

	dimensions := utils.FormatMaterialDimensions(width, height)

	storage := utils.NormalizeStorageDriver(req.Storage)
	if storage == "" {
		storage = utils.GetStorageDriver()
	}

	material := models.MaterialAsset{
		Code:       code,
		FileName:   req.FileName,
		Title:      title,
		Width:      width,
		Height:     height,
		Dimensions: dimensions,
		Format:     strings.TrimPrefix(strings.ToLower(req.Format), "."),
		FileSize:   extractFileSize(req.FileSize),
		Storage:    storage,
		FilePath:   strings.TrimSpace(req.FilePath),
		CreatedBy:  uint64(c.GetUint("user_id")),
		UpdatedBy:  uint64(c.GetUint("user_id")),
		OrderCount: extractOrderCount(req.OrderCount),
		FolderID:   req.FolderID,
		Shape:      shape,
	}

	if err := database.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建素材失败"})
		return
	}

	if folderPath != "" {
		material.Folder = &models.MaterialFolder{ID: *req.FolderID, Path: folderPath}
	}

	enriched := enrichMaterialAssets(c, []models.MaterialAsset{material})
	c.JSON(http.StatusOK, enriched[0])
}

func UpdateMaterial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "素材ID格式错误"})
		return
	}

	var material models.MaterialAsset
	if err := database.DB.Preload("Folder").First(&material, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "素材不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询素材失败"})
		return
	}

	if !canAccessMaterial(c, material) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该素材"})
		return
	}

	var req materialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if req.FileName != "" {
		updates["file_name"] = req.FileName
	}
	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Format != "" {
		updates["format"] = strings.TrimPrefix(strings.ToLower(req.Format), ".")
	}
	if req.FilePath != "" {
		updates["file_path"] = strings.TrimSpace(req.FilePath)
	}
	if req.Storage != "" {
		updates["storage"] = utils.NormalizeStorageDriver(req.Storage)
	}
	if req.OrderCount != nil {
		updates["order_count"] = extractOrderCount(req.OrderCount)
	}
	if req.FileSize != nil {
		updates["file_size"] = extractFileSize(req.FileSize)
	}

	if req.Width != nil || req.Height != nil {
		width := extractDimension(req.Width)
		height := extractDimension(req.Height)
		updates["width"] = width
		updates["height"] = height
		updates["dimensions"] = utils.FormatMaterialDimensions(width, height)
		shape := strings.TrimSpace(req.Shape)
		if shape == "" {
			shape = utils.DetermineMaterialShape(width, height)
		}
		updates["shape"] = shape
	} else if req.Shape != "" {
		updates["shape"] = strings.TrimSpace(req.Shape)
	}

	if req.FolderID != nil {
		if *req.FolderID == 0 {
			updates["folder_id"] = nil
		} else {
			var folder models.MaterialFolder
			if err := database.DB.First(&folder, *req.FolderID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "归属文件夹不存在"})
				return
			}
			updates["folder_id"] = req.FolderID
		}
	}

	updates["updated_by"] = uint64(c.GetUint("user_id"))
	updates["updated_at"] = time.Now()

	if len(updates) == 0 {
		c.JSON(http.StatusOK, material)
		return
	}

	if err := database.DB.Model(&models.MaterialAsset{}).Where("id = ?", material.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新素材失败"})
		return
	}

	database.DB.Preload("Folder").First(&material, material.ID)
	enriched := enrichMaterialAssets(c, []models.MaterialAsset{material})
	c.JSON(http.StatusOK, enriched[0])
}

func DeleteMaterial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "素材ID格式错误"})
		return
	}

	var material models.MaterialAsset
	if err := database.DB.Preload("Folder").First(&material, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "素材不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询素材失败"})
		return
	}

	if !canAccessMaterial(c, material) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该素材"})
		return
	}

	if material.FilePath != "" {
		if err := utils.DeleteAttachment(c.Request.Context(), material.Storage, material.FilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除素材文件失败"})
			return
		}
	}

	if err := database.DB.Delete(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除素材记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func UploadMaterial(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		baseName := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
		if baseName != "" {
			title = baseName
		} else {
			title = fileHeader.Filename
		}
	}
	folderIDStr := c.PostForm("folder_id")
	var folderID *uint64
	if folderIDStr != "" {
		if v, err := strconv.ParseUint(folderIDStr, 10, 64); err == nil {
			folderID = &v
			var folder models.MaterialFolder
			if err := database.DB.First(&folder, v).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "归属文件夹不存在"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹ID格式错误"})
			return
		}
	}

	orderCount := 0
	if v := c.PostForm("order_count"); v != "" {
		if num, err := strconv.Atoi(v); err == nil && num >= 0 {
			orderCount = num
		}
	}

	code := strings.TrimSpace(c.PostForm("code"))
	if code == "" {
		code = utils.GenerateMaterialCode()
	} else {
		var count int64
		database.DB.Model(&models.MaterialAsset{}).Where("code = ?", code).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "素材编号已存在"})
			return
		}
	}

	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")
	if _, ok := allowedImageFormats[format]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "素材仅支持上传图片文件 (jpg/jpeg/png/gif/bmp/webp)"})
		return
	}
	width, height := detectImageDimensions(fileHeader)
	if width == 0 || height == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法识别图片尺寸，确认文件是否为有效图片"})
		return
	}
	src, err := fileHeader.Open()

	dimensions := utils.FormatMaterialDimensions(width, height)
	shape := utils.DetermineMaterialShape(width, height)

	storage := utils.GetStorageDriver()
	objectKey := utils.BuildMaterialObjectKey(code, fileHeader.Filename)

	header := make([]byte, 512)
	n, _ := io.ReadFull(src, header)
	if n > 0 {
		header = header[:n]
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" && len(header) > 0 {
		contentType = http.DetectContentType(header)
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "素材仅支持上传图片文件 (jpg/jpeg/png/gif/bmp/webp)"})
		return
	}

	reader := utils.CreateReusableReader(src, header)
	url, err := utils.UploadAttachment(c.Request.Context(), storage, objectKey, reader, fileHeader.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("上传文件失败: %v", err)})
		return
	}

	material := models.MaterialAsset{
		Code:       code,
		FileName:   fileHeader.Filename,
		Title:      title,
		Width:      width,
		Height:     height,
		Dimensions: dimensions,
		Format:     format,
		FileSize:   fileHeader.Size,
		Storage:    storage,
		FilePath:   objectKey,
		CreatedBy:  uint64(c.GetUint("user_id")),
		UpdatedBy:  uint64(c.GetUint("user_id")),
		OrderCount: orderCount,
		FolderID:   folderID,
		Shape:      shape,
	}

	if err := database.DB.Create(&material).Error; err != nil {
		_ = utils.DeleteAttachment(c.Request.Context(), storage, objectKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存素材信息失败"})
		return
	}

	var previewURL string
	if utils.NormalizeStorageDriver(storage) == utils.StorageDriverLocal {
		if url == "" {
			url = fmt.Sprintf("/api/materials/%d/download?disposition=attachment", material.ID)
		}
		previewURL = fmt.Sprintf("/api/materials/%d/download?disposition=inline", material.ID)
	}

	enriched := enrichMaterialAssets(c, []models.MaterialAsset{material})
	if len(enriched) > 0 {
		if enriched[0].DownloadURL == "" {
			enriched[0].DownloadURL = url
		}
		if enriched[0].PreviewURL == "" {
			if previewURL != "" {
				enriched[0].PreviewURL = previewURL
			} else {
				enriched[0].PreviewURL = enriched[0].DownloadURL
			}
		}
		c.JSON(http.StatusOK, enriched[0])
		return
	}

	material.DownloadURL = url
	if previewURL != "" {
		material.PreviewURL = previewURL
	} else {
		material.PreviewURL = url
	}
	c.JSON(http.StatusOK, material)
}

func DownloadMaterial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "素材ID格式错误"})
		return
	}

	var material models.MaterialAsset
	if err := database.DB.Preload("Folder").First(&material, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "素材不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询素材失败"})
		return
	}

	if !canAccessMaterial(c, material) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该素材"})
		return
	}

	if material.FilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "素材未关联文件"})
		return
	}

	disposition := strings.ToLower(strings.TrimSpace(c.DefaultQuery("disposition", "attachment")))
	if disposition != "inline" {
		disposition = "attachment"
	}

	storage := utils.NormalizeStorageDriver(material.Storage)
	switch storage {
	case utils.StorageDriverLocal:
		fullPath, err := utils.GetLocalFilePath(material.FilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "定位文件失败"})
			return
		}
		if disposition == "inline" {
			c.File(fullPath)
		} else {
			c.FileAttachment(fullPath, material.FileName)
		}
	case utils.StorageDriverCOS:
		url, err := utils.GenerateAttachmentDownloadURL(c.Request.Context(), material.Storage, material.FilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成下载链接失败"})
			return
		}
		if url == "" {
			if u, err2 := utils.BuildAttachmentURL(material.Storage, material.FilePath); err2 == nil {
				url = u
			}
		}
		if url == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取文件链接"})
			return
		}
		c.Redirect(http.StatusFound, url)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未知的存储方式"})
	}
}

func enrichMaterialAssets(c *gin.Context, materials []models.MaterialAsset) []models.MaterialAsset {
	if len(materials) == 0 {
		return []models.MaterialAsset{}
	}

	userIDs := make(map[uint64]struct{})
	for _, asset := range materials {
		if asset.CreatedBy != 0 {
			userIDs[asset.CreatedBy] = struct{}{}
		}
		if asset.UpdatedBy != 0 {
			userIDs[asset.UpdatedBy] = struct{}{}
		}
	}

	userNameMap := make(map[uint64]string)
	if len(userIDs) > 0 {
		idList := make([]uint, 0, len(userIDs))
		for id := range userIDs {
			idList = append(idList, uint(id))
		}

		var users []models.User
		if err := database.DB.Model(&models.User{}).Select("id, username").Where("id IN ?", idList).Find(&users).Error; err == nil {
			for _, user := range users {
				userNameMap[uint64(user.ID)] = user.Username
			}
		}
	}

	// 统计每个素材关联的订单数
	materialIDs := make([]uint64, len(materials))
	for i, asset := range materials {
		materialIDs[i] = asset.ID
	}

	orderCountMap := make(map[uint64]int)
	if len(materialIDs) > 0 {
		type OrderCountResult struct {
			MaterialID uint64
			Count      int64
		}
		var orderCounts []OrderCountResult
		if err := database.DB.Model(&models.OrderAttachment{}).
			Select("material_id, COUNT(DISTINCT order_id) as count").
			Where("material_id IN ? AND material_id IS NOT NULL AND file_type = ?", materialIDs, "material_image").
			Group("material_id").
			Find(&orderCounts).Error; err == nil {
			for _, oc := range orderCounts {
				orderCountMap[oc.MaterialID] = int(oc.Count)
			}
		}
	}

	results := make([]models.MaterialAsset, len(materials))
	for i := range materials {
		asset := materials[i]
		if asset.Dimensions == "" {
			asset.Dimensions = utils.FormatMaterialDimensions(asset.Width, asset.Height)
		}

		if asset.Shape == "" {
			asset.Shape = utils.DetermineMaterialShape(asset.Width, asset.Height)
		}

		// 更新订单数为实际关联的订单数
		if count, ok := orderCountMap[asset.ID]; ok {
			asset.OrderCount = count
		} else {
			asset.OrderCount = 0
		}

		if asset.FilePath != "" {
			storage := utils.NormalizeStorageDriver(asset.Storage)
			switch storage {
			case utils.StorageDriverLocal:
				asset.PreviewURL = fmt.Sprintf("/api/materials/%d/download?disposition=inline", asset.ID)
				asset.DownloadURL = fmt.Sprintf("/api/materials/%d/download?disposition=attachment", asset.ID)
			case utils.StorageDriverCOS:
				if url, err := utils.GenerateAttachmentDownloadURL(c.Request.Context(), asset.Storage, asset.FilePath); err == nil && url != "" {
					asset.DownloadURL = url
				} else if url, err := utils.BuildAttachmentURL(asset.Storage, asset.FilePath); err == nil {
					asset.DownloadURL = url
				}
				asset.PreviewURL = asset.DownloadURL
			default:
				if asset.DownloadURL == "" {
					asset.DownloadURL = fmt.Sprintf("/api/materials/%d/download", asset.ID)
				}
			}
		}

		if asset.PreviewURL == "" && asset.DownloadURL != "" {
			asset.PreviewURL = asset.DownloadURL
		}

		if name, ok := userNameMap[asset.CreatedBy]; ok {
			asset.CreatedByName = name
		}
		if name, ok := userNameMap[asset.UpdatedBy]; ok {
			asset.UpdatedByName = name
		}

		results[i] = asset
	}

	return results
}

func detectImageDimensions(fileHeader *multipart.FileHeader) (int, int) {
	file, err := fileHeader.Open()
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0
	}
	return config.Width, config.Height
}

func extractDimension(value *int) int {
	if value == nil || *value <= 0 {
		return 0
	}
	return *value
}

func extractFileSize(value *int64) int64 {
	if value == nil || *value < 0 {
		return 0
	}
	return *value
}

func extractOrderCount(value *int) int {
	if value == nil || *value < 0 {
		return 0
	}
	return *value
}

func updateMaterialFolderChildPaths(tx *gorm.DB, oldPath, newPath string) error {
	if oldPath == newPath {
		return nil
	}
	var children []models.MaterialFolder
	like := oldPath + "/%"
	if err := tx.Where("path LIKE ?", like).Find(&children).Error; err != nil {
		return err
	}
	for _, child := range children {
		newChildPath := strings.Replace(child.Path, oldPath, newPath, 1)
		if err := tx.Model(&models.MaterialFolder{}).Where("id = ?", child.ID).Update("path", newChildPath).Error; err != nil {
			return err
		}
	}
	return nil
}

func buildMaterialFolderTree(folders []models.MaterialFolder) []models.MaterialFolder {
	roots := make([]models.MaterialFolder, 0)
	childrenMap := make(map[uint64][]models.MaterialFolder)

	for _, folder := range folders {
		folder.Children = nil
		if folder.ParentID == nil {
			roots = append(roots, folder)
		} else {
			parentID := *folder.ParentID
			childrenMap[parentID] = append(childrenMap[parentID], folder)
		}
	}

	var attachChildren func(node *models.MaterialFolder)
	attachChildren = func(node *models.MaterialFolder) {
		kids := childrenMap[node.ID]
		node.Children = kids
		for i := range node.Children {
			attachChildren(&node.Children[i])
		}
	}

	for i := range roots {
		attachChildren(&roots[i])
	}

	return roots
}

func canAccessMaterial(c *gin.Context, material models.MaterialAsset) bool {
	if isAdminUser(c) {
		return true
	}
	return material.CreatedBy == uint64(c.GetUint("user_id"))
}
