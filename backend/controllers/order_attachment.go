package controllers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"haodun_manage/backend/utils"
)

func ListOrderAttachments(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID格式错误"})
		return
	}

	var attachments []models.OrderAttachment
	if err := database.DB.Where("order_id = ?", orderID).Order("created_at DESC").Find(&attachments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取附件失败"})
		return
	}

	resp := make([]gin.H, 0, len(attachments))
	for _, att := range attachments {
		url, err := utils.BuildAttachmentURL(att.Storage, att.FilePath)
		if err != nil {
			url = ""
		}
		if utils.NormalizeStorageDriver(att.Storage) == utils.StorageDriverLocal {
			url = fmt.Sprintf("/api/orders/%d/attachments/%d/download?inline=1", att.OrderID, att.ID)
		} else if url == "" {
			url = fmt.Sprintf("/api/orders/%d/attachments/%d/download?inline=1", att.OrderID, att.ID)
		}
		resp = append(resp, gin.H{
			"id":          att.ID,
			"order_id":    att.OrderID,
			"file_type":   att.FileType,
			"file_name":   att.FileName,
			"file_path":   att.FilePath,
			"file_ext":    att.FileExt,
			"file_size":   att.FileSize,
			"checksum":    att.Checksum,
			"storage":     att.Storage,
			"uploader_id": att.UploaderID,
			"material_id": att.MaterialID,
			"created_at":  att.CreatedAt,
			"url":         url,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func UploadOrderAttachment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID格式错误"})
		return
	}

	fileType := strings.TrimSpace(c.PostForm("file_type"))
	if fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "附件类型不能为空"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择上传文件"})
		return
	}
	if fileHeader.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件不能为空"})
		return
	}

	var order models.OrderInfo
	if err := database.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	fileName := fileHeader.Filename
	fileExt := strings.ToLower(filepath.Ext(fileName))
	fileBase := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	switch fileType {
	case "material_image":
		expected := strings.TrimSpace(order.ItemNo)
		if expected == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "订单未配置货号，无法上传素材图"})
			return
		}
		if fileBase != expected && fileName != expected {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("素材图文件名需与订单货号完全一致（%s）", expected),
			})
			return
		}
	case "shipping_label":
		expected := strings.TrimSpace(order.GSPOrderNo)
		if expected == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "订单未配置订单号，无法上传面单"})
			return
		}
		if fileExt != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "面单文件仅支持PDF格式"})
			return
		}
		if fileBase != expected && fileName != expected {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("面单文件名需与订单号完全一致（%s）", expected),
			})
			return
		}
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}
	defer src.Close()

	header := make([]byte, 512)
	n, err := io.ReadFull(src, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件头失败"})
		return
	}
	if n < len(header) {
		header = header[:n]
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" && len(header) > 0 {
		contentType = http.DetectContentType(header)
	}

	reader := utils.CreateReusableReader(src, header)
	hasher := sha256.New()
	stream := io.TeeReader(reader, hasher)

	storageDriver := utils.GetStorageDriver()
	objectKey := utils.BuildAttachmentKey(order.ID, fileHeader.Filename)
	url, err := utils.UploadAttachment(c.Request.Context(), storageDriver, objectKey, stream, fileHeader.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传附件失败"})
		return
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	uploaderID := c.GetUint("user_id")

	attachment := models.OrderAttachment{
		OrderID:    order.ID,
		FileType:   fileType,
		FileName:   fileHeader.Filename,
		FilePath:   objectKey,
		FileExt:    fileExt,
		FileSize:   fileHeader.Size,
		Checksum:   checksum,
		Storage:    storageDriver,
		UploaderID: uploaderID,
	}

	if fileType == "material_image" {
		if assetID, err := ensureMaterialAssetForOrder(c.Request.Context(), &order, fileHeader, header, contentType, storageDriver, uploaderID); err != nil {
			log.Printf("failed to sync material to library: %v", err)
		} else if assetID != nil {
			attachment.MaterialID = assetID
		}
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var existing models.OrderAttachment
		if err := tx.Where("order_id = ? AND file_type = ?", order.ID, fileType).First(&existing).Error; err == nil {
			if existing.MaterialID == nil {
				if delErr := utils.DeleteAttachment(c.Request.Context(), existing.Storage, existing.FilePath); delErr != nil {
					log.Printf("failed to delete old attachment: %v", delErr)
				}
			}
			if err := tx.Unscoped().Delete(&existing).Error; err != nil {
				return err
			}
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return tx.Create(&attachment).Error
	}); err != nil {
		_ = utils.DeleteAttachment(c.Request.Context(), storageDriver, objectKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存附件信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":          attachment.ID,
			"order_id":    attachment.OrderID,
			"file_type":   attachment.FileType,
			"file_name":   attachment.FileName,
			"file_path":   attachment.FilePath,
			"file_ext":    attachment.FileExt,
			"file_size":   attachment.FileSize,
			"checksum":    attachment.Checksum,
			"storage":     attachment.Storage,
			"uploader_id": attachment.UploaderID,
			"material_id": attachment.MaterialID,
			"url":         url,
		},
	})
}

func ensureMaterialAssetForOrder(ctx context.Context, order *models.OrderInfo, fileHeader *multipart.FileHeader, header []byte, contentType string, storage string, uploaderID uint) (*uint64, error) {
	if order == nil || fileHeader == nil {
		return nil, nil
	}

	code := strings.TrimSpace(order.ItemNo)
	if code == "" {
		base := strings.TrimSpace(strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename)))
		if base != "" {
			code = base
		} else {
			code = utils.GenerateMaterialCode()
		}
	}

	var existing models.MaterialAsset
	if err := database.DB.Where("code = ?", code).Or("file_name = ?", fileHeader.Filename).First(&existing).Error; err == nil {
		return &existing.ID, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	readerForUpload := bytes.NewReader(data)

	if contentType == "" {
		contentType = fileHeader.Header.Get("Content-Type")
	}
	if contentType == "" && len(header) > 0 {
		contentType = http.DetectContentType(header)
	}

	objectKey := utils.BuildMaterialObjectKey(code, fileHeader.Filename)
	if _, err := utils.UploadAttachment(ctx, storage, objectKey, readerForUpload, int64(len(data)), contentType); err != nil {
		return nil, err
	}

	width, height := 0, 0
	if config, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		width = config.Width
		height = config.Height
	}

	shape := utils.DetermineMaterialShape(width, height)
	dimensions := utils.FormatMaterialDimensions(width, height)
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")

	folderID, err := getDefaultMaterialFolderID()
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(order.ProductName)
	if title == "" {
		title = strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
	}

	orderCount := order.ItemCount
	if orderCount <= 0 {
		orderCount = 1
	}

	asset := models.MaterialAsset{
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
		CreatedBy:  uint64(uploaderID),
		UpdatedBy:  uint64(uploaderID),
		OrderCount: orderCount,
		Shape:      shape,
	}
	if folderID != nil {
		asset.FolderID = folderID
	}

	if err := database.DB.Create(&asset).Error; err != nil {
		_ = utils.DeleteAttachment(ctx, storage, objectKey)
		return nil, err
	}

	return &asset.ID, nil
}

func getDefaultMaterialFolderID() (*uint64, error) {
	var folder models.MaterialFolder
	if err := database.DB.Where("name = ?", "默认文件夹").First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			folder = models.MaterialFolder{Name: "默认文件夹", Path: "默认文件夹"}
			if err := database.DB.Create(&folder).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &folder.ID, nil
}

func AttachOrderMaterial(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID格式错误"})
		return
	}

	var req struct {
		MaterialID uint64 `json:"material_id" binding:"required"`
		FileType   string `json:"file_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileType := strings.TrimSpace(req.FileType)
	if fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "附件类型不能为空"})
		return
	}

	var order models.OrderInfo
	if err := database.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	var material models.MaterialAsset
	if err := database.DB.First(&material, req.MaterialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "素材不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询素材失败"})
		return
	}

	// 权限检查：非管理员只能引用自己创建的素材
	if roleID := c.GetUint("role_id"); roleID != 1 {
		userID := uint64(c.GetUint("user_id"))
		if material.CreatedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权引用该素材"})
			return
		}
	}

	baseName := strings.TrimSuffix(material.FileName, filepath.Ext(material.FileName))
	switch fileType {
	case "material_image":
		expected := strings.TrimSpace(order.ItemNo)
		if expected != "" && !strings.EqualFold(baseName, expected) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("素材文件名需与订单货号一致（%s）", expected)})
			return
		}
	case "shipping_label":
		expected := strings.TrimSpace(order.GSPOrderNo)
		if expected == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "订单未配置订单号，无法关联面单"})
			return
		}
		if !strings.EqualFold(filepath.Ext(material.FileName), ".pdf") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "面单素材需为 PDF 文件"})
			return
		}
		if !strings.EqualFold(baseName, expected) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("面单文件名需与订单号一致（%s）", expected)})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的附件类型"})
		return
	}

	attachment, err := upsertOrderMaterialAttachment(c.Request.Context(), nil, orderID, fileType, &material, c.GetUint("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var url string
	if utils.NormalizeStorageDriver(attachment.Storage) == utils.StorageDriverCOS {
		url, _ = utils.GenerateAttachmentDownloadURL(c.Request.Context(), attachment.Storage, attachment.FilePath)
	}
	if url == "" {
		url, _ = utils.BuildAttachmentURL(attachment.Storage, attachment.FilePath)
	}
	if url == "" {
		url = fmt.Sprintf("/api/orders/%d/attachments/%d/download?inline=1", orderID, attachment.ID)
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"id":          attachment.ID,
		"order_id":    attachment.OrderID,
		"file_type":   attachment.FileType,
		"file_name":   attachment.FileName,
		"file_path":   attachment.FilePath,
		"file_ext":    attachment.FileExt,
		"file_size":   attachment.FileSize,
		"storage":     attachment.Storage,
		"material_id": attachment.MaterialID,
		"uploader_id": attachment.UploaderID,
		"url":         url,
	}})
}

func upsertOrderMaterialAttachment(ctx context.Context, tx *gorm.DB, orderID uint64, fileType string, material *models.MaterialAsset, uploaderID uint) (*models.OrderAttachment, error) {
	if material == nil {
		return nil, fmt.Errorf("素材不存在")
	}
	storage := utils.NormalizeStorageDriver(material.Storage)
	objectKey := strings.TrimSpace(material.FilePath)
	if objectKey == "" {
		return nil, fmt.Errorf("素材未关联文件")
	}
	if tx == nil {
		tx = database.DB
	}

	var attachment models.OrderAttachment
	err := tx.Where("order_id = ? AND file_type = ?", orderID, fileType).First(&attachment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			attachment = models.OrderAttachment{OrderID: orderID, FileType: fileType}
		} else {
			return nil, fmt.Errorf("查询附件失败: %v", err)
		}
	}

	if attachment.ID != 0 && attachment.MaterialID == nil && strings.TrimSpace(attachment.FilePath) != "" {
		if delErr := utils.DeleteAttachment(ctx, attachment.Storage, attachment.FilePath); delErr != nil {
			log.Printf("failed to delete legacy attachment file: %v", delErr)
		}
	}

	if attachment.MaterialID != nil && *attachment.MaterialID == material.ID {
		return &attachment, nil
	}

	materialID := material.ID
	attachment.FileName = material.FileName
	attachment.FilePath = objectKey
	attachment.FileExt = strings.ToLower(filepath.Ext(material.FileName))
	attachment.FileSize = material.FileSize
	attachment.Checksum = ""
	attachment.Storage = storage
	attachment.UploaderID = uploaderID
	attachment.MaterialID = &materialID

	if attachment.ID == 0 {
		if err := tx.Create(&attachment).Error; err != nil {
			return nil, fmt.Errorf("保存附件失败: %v", err)
		}
	} else {
		if err := tx.Save(&attachment).Error; err != nil {
			return nil, fmt.Errorf("更新附件失败: %v", err)
		}
	}

	return &attachment, nil
}

func DownloadOrderAttachment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID格式错误"})
		return
	}
	attachmentID, err := strconv.ParseUint(c.Param("attachmentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "附件ID格式错误"})
		return
	}

	var attachment models.OrderAttachment
	if err := database.DB.Where("order_id = ? AND id = ?", orderID, attachmentID).First(&attachment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "附件不存在"})
		return
	}

	url, err := utils.GenerateAttachmentDownloadURL(c.Request.Context(), attachment.Storage, attachment.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取下载链接失败"})
		return
	}

	if url != "" {
		c.JSON(http.StatusOK, gin.H{"url": url})
		return
	}

	if utils.NormalizeStorageDriver(attachment.Storage) == utils.StorageDriverLocal {
		filePath, err := utils.GetLocalFilePath(attachment.FilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "定位本地文件失败"})
			return
		}
		if c.Query("inline") == "1" {
			c.File(filePath)
		} else {
			c.FileAttachment(filePath, attachment.FileName)
		}
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取附件下载地址"})
}

func DeleteOrderAttachment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID格式错误"})
		return
	}
	attachmentID, err := strconv.ParseUint(c.Param("attachmentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "附件ID格式错误"})
		return
	}

	var attachment models.OrderAttachment
	if err := database.DB.Where("order_id = ? AND id = ?", orderID, attachmentID).First(&attachment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "附件不存在"})
		return
	}

	if attachment.MaterialID == nil {
		if err := utils.DeleteAttachment(c.Request.Context(), attachment.Storage, attachment.FilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除附件文件失败"})
			return
		}
	}

	if err := database.DB.Unscoped().Delete(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除附件记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
