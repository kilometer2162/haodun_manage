package controllers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"haodun_manage/backend/config"
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"haodun_manage/backend/utils"

	"github.com/gin-gonic/gin"
	excelize "github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var (
	specPattern    = regexp.MustCompile(`(?i)^\d+(\.\d+)?\s*(?:\*|x)\s*\d+(\.\d+)?$`)
	chinesePattern = regexp.MustCompile(`^[\p{Han}]+$`)
)

const adminRoleID uint = 1

func isAdminUser(c *gin.Context) bool {
	return c.GetUint("role_id") == adminRoleID
}

func currentUserID(c *gin.Context) uint {
	return c.GetUint("user_id")
}

func isCompletedStatus(value int8) bool {
	return value == 1
}

// saveOrderParams 包含创建/更新订单时的所有字段
type saveOrderParams struct {
	GSPOrderNo             string   `json:"gsp_order_no"`
	OrderType              string   `json:"order_type"`
	OrderCreatedAt         string   `json:"order_created_at"`
	Status                 *int8    `json:"status"`
	PaymentTime            *string  `json:"payment_time"`
	CompletedAt            *string  `json:"completed_at"`
	ShippingWarehouseCode  string   `json:"shipping_warehouse_code"`
	RequiredSignAt         *string  `json:"required_sign_at"`
	ShopCode               string   `json:"shop_code"`
	ProductID              string   `json:"product_id"`
	OwnerName              string   `json:"owner_name"`
	ProductName            string   `json:"product_name"`
	Spec                   string   `json:"spec"`
	ItemNo                 string   `json:"item_no"`
	SellerSKU              string   `json:"seller_sku"`
	PlatformSKU            string   `json:"platform_sku"`
	PlatformSKC            string   `json:"platform_skc"`
	PlatformSPU            string   `json:"platform_spu"`
	ProductPrice           *float64 `json:"product_price"`
	ExpectedRevenue        *float64 `json:"expected_revenue"`
	SpecialProductNote     *string  `json:"special_product_note"`
	CurrencyCode           string   `json:"currency_code"`
	ExpectedFulfillmentQty *int     `json:"expected_fulfillment_qty"`
	ItemCount              *int     `json:"item_count"`
	PostalCode             string   `json:"postal_code"`
	Country                string   `json:"country"`
	Province               string   `json:"province"`
	City                   string   `json:"city"`
	District               string   `json:"district"`
	AddressLine1           string   `json:"address_line1"`
	AddressLine2           string   `json:"address_line2"`
	CustomerFullName       string   `json:"customer_full_name"`
	CustomerLastName       string   `json:"customer_last_name"`
	CustomerFirstName      string   `json:"customer_first_name"`
	PhoneNumber            string   `json:"phone_number"`
	Email                  string   `json:"email"`
	TaxNumber              *string  `json:"tax_number"`
}

func CreateOrder(c *gin.Context) {
	var params saveOrderParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateOrderParams(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.Status != nil && params.CompletedAt == nil && isCompletedStatus(*params.Status) {
		now := time.Now().Format("2006-01-02 15:04:05")
		params.CompletedAt = &now
	}

	order := models.OrderInfo{}
	assignOrderFields(&order, &params)
	userID := uint64(c.GetUint("user_id"))
	order.CreatedBy = userID
	order.UpdatedBy = userID

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存订单失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func UpdateOrder(c *gin.Context) {
	id := c.Param("id")

	var order models.OrderInfo
	if err := database.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	var params saveOrderParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateOrderParams(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.Status != nil {
		log.Printf("[DEBUG] UpdateOrder - Status: %d, isCompletedStatus: %v, CompletedAt before: %v",
			*params.Status, isCompletedStatus(*params.Status), params.CompletedAt)
		if isCompletedStatus(*params.Status) && params.CompletedAt == nil {
			now := time.Now().Format("2006-01-02 15:04:05")
			params.CompletedAt = &now
			log.Printf("[DEBUG] UpdateOrder - Set CompletedAt to current time: %s", now)
		} else if !isCompletedStatus(*params.Status) {
			empty := ""
			params.CompletedAt = &empty
			log.Printf("[DEBUG] UpdateOrder - Set CompletedAt to empty string (will be cleared)")
		}
		log.Printf("[DEBUG] UpdateOrder - CompletedAt after: %v", params.CompletedAt)
	}
	assignOrderFields(&order, &params)
	order.UpdatedBy = uint64(c.GetUint("user_id"))

	if err := database.DB.Model(&models.OrderInfo{}).
		Where("id = ?", order.ID).
		Select("*").
		Updates(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新订单失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func DeleteOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID格式错误"})
		return
	}

	var order models.OrderInfo
	if err := database.DB.Where("id = ?", id).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法启动数据库事务"})
		return
	}

	var attachments []models.OrderAttachment
	if err := tx.Where("order_id = ?", order.ID).Find(&attachments).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单附件失败"})
		return
	}

	for _, att := range attachments {
		if err := utils.DeleteAttachment(c.Request.Context(), att.Storage, att.FilePath); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除附件文件失败"})
			return
		}
	}

	if len(attachments) > 0 {
		if err := tx.Unscoped().Where("order_id = ?", order.ID).Delete(&models.OrderAttachment{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除附件记录失败"})
			return
		}
	}

	if err := tx.Unscoped().Delete(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除订单失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func BatchUploadOrderAttachments(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "上传数据格式错误"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}

	successes := make([]gin.H, 0, len(files))
	failures := make([]gin.H, 0)
	storageDriver := utils.GetStorageDriver()
	uploaderID := c.GetUint("user_id")

	for _, fileHeader := range files {
		result, err := processBatchAttachment(c, fileHeader, storageDriver, uploaderID)
		if err != nil {
			failures = append(failures, gin.H{
				"file_name": fileHeader.Filename,
				"message":   err.Error(),
			})
			continue
		}
		successes = append(successes, result)
	}

	message := fmt.Sprintf("成功上传%d个，失败%d个", len(successes), len(failures))
	data := gin.H{
		"success": successes,
		"failed":  failures,
	}

	if len(successes) == 0 && len(failures) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": message,
			"data":  data,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"data":    data,
	})
}

func processBatchAttachment(c *gin.Context, fileHeader *multipart.FileHeader, storage string, uploaderID uint) (gin.H, error) {
	fileName := strings.TrimSpace(fileHeader.Filename)
	if fileName == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}

	originalExt := filepath.Ext(fileName)
	ext := strings.ToLower(originalExt)
	baseName := strings.TrimSpace(strings.TrimSuffix(fileName, originalExt))
	if baseName == "" {
		return nil, fmt.Errorf("文件名需与订单编号一致")
	}

	var order models.OrderInfo
	var lookupErr error
	fileType := "material_image"
	if ext == ".pdf" {
		fileType = "shipping_label"
		lookupErr = database.DB.Where("gsp_order_no = ?", baseName).First(&order).Error
	} else {
		lookupErr = database.DB.Where("item_no = ?", baseName).First(&order).Error
	}
	if lookupErr != nil {
		if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			lookupField := "货号"
			if fileType == "shipping_label" {
				lookupField = "GSP订单号"
			}
			return nil, fmt.Errorf("文件 %s: 未找到匹配的订单（%s: %s）", fileName, lookupField, baseName)
		}
		return nil, fmt.Errorf("文件 %s: 查询订单失败: %v", fileName, lookupErr)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败")
	}
	defer src.Close()

	header := make([]byte, 512)
	n, err := io.ReadFull(src, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("读取文件头失败")
	}
	if n < len(header) {
		header = header[:n]
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" && len(header) > 0 {
		contentType = http.DetectContentType(header)
	}

	if fileType == "shipping_label" {
		if ext != ".pdf" {
			return nil, fmt.Errorf("面单文件仅支持 PDF 格式")
		}
	} else {
		if contentType != "" && !strings.HasPrefix(contentType, "image/") {
			return nil, fmt.Errorf("素材图仅支持图片文件")
		}
	}

	reader := utils.CreateReusableReader(src, header)
	hasher := sha256.New()
	stream := io.TeeReader(reader, hasher)

	objectKey := utils.BuildAttachmentKey(order.ID, fileName)
	url, err := utils.UploadAttachment(c.Request.Context(), storage, objectKey, stream, fileHeader.Size, contentType)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %v", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))

	attachment := models.OrderAttachment{
		OrderID:    order.ID,
		FileType:   fileType,
		FileName:   fileHeader.Filename,
		FilePath:   objectKey,
		FileExt:    fileExt,
		FileSize:   fileHeader.Size,
		Checksum:   checksum,
		Storage:    storage,
		UploaderID: uploaderID,
	}

	if fileType == "material_image" {
		if assetID, err := ensureMaterialAssetForOrder(c.Request.Context(), &order, fileHeader, header, contentType, storage, uploaderID); err != nil {
			log.Printf("failed to sync material to library (batch): %v", err)
		} else if assetID != nil {
			attachment.MaterialID = assetID
		}
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
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
	})
	if err != nil {
		_ = utils.DeleteAttachment(c.Request.Context(), storage, objectKey)
		return nil, fmt.Errorf("保存附件信息失败: %v", err)
	}

	return gin.H{
		"file_name": fileHeader.Filename,
		"order_id":  order.ID,
		"order_no":  order.GSPOrderNo,
		"file_type": fileType,
		"url":       url,
		"message":   fmt.Sprintf("订单 %s %s 上传成功", order.GSPOrderNo, map[string]string{"material_image": "素材图", "shipping_label": "面单"}[fileType]),
	}, nil
}

func ListOrders(c *gin.Context) {
	tab := c.DefaultQuery("tab", "platform")
	var orders []models.OrderInfo

	query := database.DB.Model(&models.OrderInfo{})
	if tab == "factory" {
		query = query.Where("order_type = ?", "factory").Order("order_created_at DESC")
	} else {
		query = query.Where("order_type = ?", "platform").Order("id DESC")
	}

	query = applyOrderFilters(c, query)

	if !isAdminUser(c) {
		userID := currentUserID(c)
		query = query.Where("created_by = ?", userID)
	}

	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "10"), 10)
	if pageSize > 200 {
		pageSize = 200
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单列表失败"})
		return
	}

	if total == 0 {
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"orders":    []orderResponse{},
				"total":     0,
				"page":      1,
				"page_size": pageSize,
			},
		})
		return
	}

	maxPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	if page > maxPage {
		page = maxPage
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单列表失败"})
		return
	}

	responses := enrichOrdersWithAttachments(c.Request.Context(), orders)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"orders":    responses,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func enrichOrdersWithAttachments(ctx context.Context, orders []models.OrderInfo) []orderResponse {
	if len(orders) == 0 {
		return []orderResponse{}
	}

	ids := make([]uint64, len(orders))
	orderIndex := make(map[uint64]int, len(orders))
	for i, order := range orders {
		ids[i] = order.ID
		orderIndex[order.ID] = i
	}

	type attachmentSummary struct {
		MaterialURL      string
		MaterialID       uint64
		MaterialStorage  string
		MaterialFileName string
		MaterialAssetID  *uint64
		LabelURL         string
		LabelID          uint64
		LabelStorage     string
		LabelFileName    string
		LabelAssetID     *uint64
	}

	summaries := make(map[uint64]attachmentSummary)

	var attachments []models.OrderAttachment
	if err := database.DB.Where("order_id IN ?", ids).Find(&attachments).Error; err == nil {
		for _, att := range attachments {
			summary := summaries[att.OrderID]
			var url string
			storage := utils.NormalizeStorageDriver(att.Storage)
			if storage == utils.StorageDriverCOS {
				url, _ = utils.GenerateAttachmentDownloadURL(ctx, att.Storage, att.FilePath)
			}
			if url == "" {
				url, _ = utils.BuildAttachmentURL(att.Storage, att.FilePath)
			}
			switch att.FileType {
			case "material_image":
				summary.MaterialURL = url
				summary.MaterialID = att.ID
				summary.MaterialStorage = att.Storage
				summary.MaterialFileName = att.FileName
				summary.MaterialAssetID = att.MaterialID
			case "shipping_label":
				summary.LabelURL = url
				summary.LabelID = att.ID
				summary.LabelStorage = att.Storage
				summary.LabelFileName = att.FileName
				summary.LabelAssetID = att.MaterialID
			}
			summaries[att.OrderID] = summary
		}
	}

	results := make([]orderResponse, len(orders))
	for i, order := range orders {
		summary := summaries[order.ID]
		if summary.MaterialURL == "" && summary.MaterialID != 0 {
			summary.MaterialURL = fmt.Sprintf("/api/orders/%d/attachments/%d/download?inline=1", order.ID, summary.MaterialID)
		}
		if summary.MaterialURL != "" && utils.NormalizeStorageDriver(summary.MaterialStorage) == utils.StorageDriverLocal {
			summary.MaterialURL = fmt.Sprintf("/api/orders/%d/attachments/%d/download?inline=1", order.ID, summary.MaterialID)
		}
		if summary.LabelURL == "" && summary.LabelID != 0 {
			summary.LabelURL = fmt.Sprintf("/api/orders/%d/attachments/%d/download?inline=1", order.ID, summary.LabelID)
		}
		if summary.LabelURL != "" && utils.NormalizeStorageDriver(summary.LabelStorage) == utils.StorageDriverLocal {
			summary.LabelURL = fmt.Sprintf("/api/orders/%d/attachments/%d/download?inline=1", order.ID, summary.LabelID)
		}
		results[i] = orderResponse{
			OrderInfo:                 order,
			MaterialImageURL:          summary.MaterialURL,
			MaterialAttachmentID:      summary.MaterialID,
			MaterialStorage:           summary.MaterialStorage,
			MaterialFileName:          summary.MaterialFileName,
			MaterialAssetID:           summary.MaterialAssetID,
			ShippingLabelURL:          summary.LabelURL,
			ShippingLabelAttachmentID: summary.LabelID,
			ShippingLabelStorage:      summary.LabelStorage,
			ShippingLabelFileName:     summary.LabelFileName,
			ShippingLabelAssetID:      summary.LabelAssetID,
		}
	}

	return results
}

func applyOrderFilters(c *gin.Context, query *gorm.DB) *gorm.DB {
	timeField := c.Query("time_field")
	allowedTimeFields := map[string]bool{
		"order_created_at": true,
		"completed_at":     true,
		"payment_time":     true,
		"required_sign_at": true,
	}
	if allowedTimeFields[timeField] {
		startStr := c.Query("time_start")
		endStr := c.Query("time_end")
		if startStr != "" && endStr != "" {
			startTime, err1 := parseQueryTime(startStr)
			endTime, err2 := parseQueryTime(endStr)
			if err1 == nil && err2 == nil {
				if endTime.Before(startTime) {
					startTime, endTime = endTime, startTime
				}
				query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", timeField), startTime, endTime)
			}
		}
	}

	exactField := c.Query("exact_field")
	exactValue := c.Query("exact_value")
	if exactField != "" && exactValue != "" {
		switch exactField {
		case "product_price", "expected_revenue":
			if num, err := strconv.ParseFloat(exactValue, 64); err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", exactField), num)
			}
		case "expected_fulfillment_qty":
			if num, err := strconv.Atoi(exactValue); err == nil {
				query = query.Where("expected_fulfillment_qty = ?", num)
			}
		}
	}

	fuzzyField := c.Query("fuzzy_field")
	fuzzyKeyword := c.Query("fuzzy_keyword")
	allowedFuzzyFields := map[string]bool{
		"gsp_order_no":            true,
		"shipping_warehouse_code": true,
		"shop_code":               true,
		"owner_name":              true,
		"product_name":            true,
		"spec":                    true,
		"item_no":                 true,
		"seller_sku":              true,
		"platform_sku":            true,
		"platform_skc":            true,
		"platform_spu":            true,
		"special_product_note":    true,
		"currency_code":           true,
		"postal_code":             true,
		"country":                 true,
		"province":                true,
		"city":                    true,
		"district":                true,
		"address_line1":           true,
		"address_line2":           true,
		"customer_full_name":      true,
		"customer_last_name":      true,
		"customer_first_name":     true,
		"phone_number":            true,
		"email":                   true,
		"tax_number":              true,
	}
	if allowedFuzzyFields[fuzzyField] && fuzzyKeyword != "" {
		likeValue := fmt.Sprintf("%%%s%%", fuzzyKeyword)
		query = query.Where(fmt.Sprintf("%s LIKE ?", fuzzyField), likeValue)
	}

	return query
}

func parseQueryTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s", value)
}

func parsePositiveInt(value string, defaultValue int) int {
	if v, err := strconv.Atoi(value); err == nil && v > 0 {
		return v
	}
	return defaultValue
}

func ImportOrders(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择上传文件"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	sheets, err := parseWorkbook(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法解析Excel文件: %v", err)})
		return
	}

	defaultAddress := getConfigValue("default_address")
	if defaultAddress == "" {
		defaultAddress = config.AppConfig.LocalBaseURL
	}

	shippingSet := loadShippingWarehouseSet()

	platformOrders, platformErrs := parsePlatformSheet(sheets, defaultAddress, shippingSet)
	factoryOrders, factoryErrs := parseFactorySheet(sheets, defaultAddress, shippingSet)

	validationErrors := append(platformErrs, factoryErrs...)

	orders := append(platformOrders, factoryOrders...)
	if len(orders) == 0 && len(validationErrors) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Excel 中未找到可导入的数据"})
		return
	}

	if formatted := formatValidationErrors(validationErrors); len(formatted) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(formatted, "; ")})
		return
	}

	tx := database.DB.Begin()
	operatorID := c.GetUint("user_id")
	userID := uint64(operatorID)
	for i := range orders {
		orders[i].UpdatedBy = userID
		if orders[i].CreatedBy == 0 {
			orders[i].CreatedBy = userID
		}
		if err := upsertOrder(tx, &orders[i]); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("导入失败: %v", err)})
			return
		}
		if err := autoLinkMaterialForOrder(c.Request.Context(), tx, &orders[i], operatorID); err != nil {
			log.Printf("auto link material failed for order %s: %v", orders[i].GSPOrderNo, err)
		} else {
			log.Printf("auto link material success for order %s", orders[i].GSPOrderNo)
		}
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("导入失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功导入%d条订单", len(orders))})
}

func ExportOrders(c *gin.Context) {
	templatePath := filepath.Join("template", "运营提交表格模板.xlsx")
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		log.Printf("failed to open template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败，模板不可用"})
		return
	}
	defer func() { _ = f.Close() }()

	type sheetConfig struct {
		SheetName string
		OrderType string
		Writer    func(*excelize.File, string, int, models.OrderInfo)
	}
	sheets := []sheetConfig{
		{SheetName: "平台面单", OrderType: "platform", Writer: writePlatformRow},
		{SheetName: "工厂物流", OrderType: "factory", Writer: writeFactoryRow},
	}

	for _, sheet := range sheets {
		if err := clearSheetData(f, sheet.SheetName); err != nil {
			log.Printf("failed to clear sheet %s: %v", sheet.SheetName, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
			return
		}

		orders, err := queryOrdersForExport(c, sheet.OrderType)
		if err != nil {
			log.Printf("failed to query %s orders: %v", sheet.OrderType, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
			return
		}

		for i, order := range orders {
			rowNum := i + 2 // 数据从第二行开始
			sheet.Writer(f, sheet.SheetName, rowNum, order)
		}
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		log.Printf("failed to write workbook to buffer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}

	filename := fmt.Sprintf("orders_all_%d.xlsx", time.Now().Unix())
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := c.Writer.Write(buffer.Bytes()); err != nil {
		log.Printf("failed to send workbook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}
}

func loadShippingWarehouseSet() map[string]struct{} {
	var dictType models.DictType
	if err := database.DB.Where("code = ?", "shipping_warehouse").First(&dictType).Error; err != nil {
		return nil
	}

	var items []models.DictItem
	if err := database.DB.Where("type_id = ? AND status = 1", dictType.ID).Find(&items).Error; err != nil {
		return nil
	}

	result := make(map[string]struct{})
	for _, item := range items {
		if value := strings.TrimSpace(item.Label); value != "" {
			result[strings.ToUpper(value)] = struct{}{}
		}
		if value := strings.TrimSpace(item.Value); value != "" {
			result[strings.ToUpper(value)] = struct{}{}
		}
	}
	return result
}

type columnIndex struct {
	GSPOrderNo        int
	OrderType         string
	OrderCreatedAt    int
	RequiredSignAt    int
	ShippingWarehouse int
	ShopCode          int
	OwnerName         int
	ProductID         int
	ProductName       int
	Spec              int
	ItemNo            int
	SellerSKU         int
	PlatformSKU       int
	PlatformSKC       int
	PlatformSPU       int
	ProductPrice      int
	ExpectedRevenue   int
	SpecialNote       int
	ExpectedQty       int
	CurrencyCode      int
	PostalCode        int
	Country           int
	Province          int
	City              int
	District          int
	AddressLine1      int
	AddressLine2      int
	CustomerFullName  int
	CustomerLastName  int
	CustomerFirstName int
	PhoneNumber       int
	Email             int
	TaxNumber         int
	PaymentTime       int
	CompletedAt       int
}

func newColumnIndex() columnIndex {
	return columnIndex{
		GSPOrderNo:        -1,
		OrderCreatedAt:    -1,
		RequiredSignAt:    -1,
		ShippingWarehouse: -1,
		ShopCode:          -1,
		OwnerName:         -1,
		ProductID:         -1,
		ProductName:       -1,
		Spec:              -1,
		ItemNo:            -1,
		SellerSKU:         -1,
		PlatformSKU:       -1,
		PlatformSKC:       -1,
		PlatformSPU:       -1,
		ProductPrice:      -1,
		ExpectedRevenue:   -1,
		SpecialNote:       -1,
		ExpectedQty:       -1,
		CurrencyCode:      -1,
		PostalCode:        -1,
		Country:           -1,
		Province:          -1,
		City:              -1,
		District:          -1,
		AddressLine1:      -1,
		AddressLine2:      -1,
		CustomerFullName:  -1,
		CustomerLastName:  -1,
		CustomerFirstName: -1,
		PhoneNumber:       -1,
		Email:             -1,
		TaxNumber:         -1,
		PaymentTime:       -1,
		CompletedAt:       -1,
	}
}

func detectColumns(headers []string) columnIndex {
	idx := newColumnIndex()
	for i, header := range headers {
		title := strings.TrimSpace(header)
		if title == "" || strings.HasPrefix(title, "#") {
			continue
		}
		replacements := []struct{ old, new string }{
			{"（", "("},
			{"）", ")"},
			{"[", "("},
			{"]", ")"},
			{"，", ""},
			{",", ""},
			{"。", ""},
			{"/", ""},
			{"\\", ""},
			{"-", ""},
			{"_", ""},
			{":", ""},
			{"：", ""},
			{"\n", ""},
			{" ", ""},
		}
		for _, rep := range replacements {
			title = strings.ReplaceAll(title, rep.old, rep.new)
		}
		lower := strings.ToLower(title)

		set := func(current *int, value int) {
			if *current < 0 {
				*current = value
			}
		}

		if strings.Contains(title, "GSP订单号") || strings.Contains(lower, "gsporderid") {
			set(&idx.GSPOrderNo, i)
		}
		if strings.Contains(title, "订单创建时间") || strings.Contains(lower, "ordercreated") {
			set(&idx.OrderCreatedAt, i)
		}
		if strings.Contains(title, "要求签收时间") || strings.Contains(lower, "requiredsign") {
			set(&idx.RequiredSignAt, i)
		}
		if strings.Contains(title, "发货仓库") || strings.Contains(lower, "warehouse") || strings.Contains(lower, "warehousecode") || strings.Contains(lower, "shipfrom") {
			set(&idx.ShippingWarehouse, i)
		}
		if strings.Contains(title, "店铺编号") || strings.Contains(lower, "shop") || strings.Contains(lower, "store") {
			set(&idx.ShopCode, i)
		}
		if strings.Contains(title, "负责人") || strings.Contains(lower, "owner") {
			set(&idx.OwnerName, i)
		}
		if strings.Contains(title, "商品ID") || strings.Contains(lower, "productid") {
			set(&idx.ProductID, i)
		}
		if strings.Contains(title, "商品名称") || strings.Contains(lower, "productname") || strings.Contains(lower, "itemname") {
			set(&idx.ProductName, i)
		}
		if strings.Contains(title, "规格") || strings.Contains(lower, "spec") {
			set(&idx.Spec, i)
		}
		if strings.Contains(title, "货号") || strings.Contains(lower, "itemno") {
			set(&idx.ItemNo, i)
		}
		if strings.Contains(title, "卖家SKU") || strings.Contains(lower, "sellersku") {
			set(&idx.SellerSKU, i)
		}
		if (strings.Contains(title, "平台SKU") && !strings.Contains(title, "SKC")) || strings.Contains(lower, "platformsku") {
			set(&idx.PlatformSKU, i)
		}
		if strings.Contains(title, "平台SKC") || strings.Contains(lower, "platformskc") {
			set(&idx.PlatformSKC, i)
		}
		if strings.Contains(title, "平台SPU") || strings.Contains(lower, "platformspu") {
			set(&idx.PlatformSPU, i)
		}
		if strings.Contains(title, "商品价格") || strings.Contains(lower, "price") {
			set(&idx.ProductPrice, i)
		}
		if strings.Contains(title, "商品预计收入") || strings.Contains(lower, "expectedrevenue") {
			set(&idx.ExpectedRevenue, i)
		}
		if strings.Contains(title, "特殊产品备注") || strings.Contains(lower, "special") {
			set(&idx.SpecialNote, i)
		}
		if strings.Contains(title, "履约件数") || strings.Contains(lower, "expectedfulfillment") || strings.Contains(title, "件数") {
			set(&idx.ExpectedQty, i)
		}
		if strings.Contains(title, "币种") || strings.Contains(lower, "currency") {
			set(&idx.CurrencyCode, i)
		}
		if strings.Contains(title, "邮编") || strings.Contains(lower, "postal") || strings.Contains(lower, "zipcode") {
			set(&idx.PostalCode, i)
		}
		if strings.Contains(title, "国家") || strings.Contains(lower, "country") {
			set(&idx.Country, i)
		}
		if strings.Contains(title, "省份") || strings.Contains(lower, "province") {
			set(&idx.Province, i)
		}
		if strings.Contains(title, "城市") || strings.Contains(lower, "city") {
			set(&idx.City, i)
		}
		if strings.Contains(title, "区") || strings.Contains(lower, "district") {
			set(&idx.District, i)
		}
		if strings.Contains(title, "用户地址1") || strings.Contains(lower, "address1") {
			set(&idx.AddressLine1, i)
		}
		if strings.Contains(title, "用户地址2") || strings.Contains(lower, "address2") {
			set(&idx.AddressLine2, i)
		}
		if strings.Contains(title, "用户全称") || strings.Contains(lower, "fullname") {
			set(&idx.CustomerFullName, i)
		}
		if strings.Contains(title, "用户姓氏") || strings.Contains(lower, "lastname") {
			set(&idx.CustomerLastName, i)
		}
		if strings.Contains(title, "用户名字") || strings.Contains(lower, "firstname") {
			set(&idx.CustomerFirstName, i)
		}
		if strings.Contains(title, "手机号") || strings.Contains(lower, "phone") || strings.Contains(lower, "mobile") {
			set(&idx.PhoneNumber, i)
		}
		if strings.Contains(title, "邮箱") || strings.Contains(lower, "email") {
			set(&idx.Email, i)
		}
		if strings.Contains(title, "税号") || strings.Contains(lower, "tax") {
			set(&idx.TaxNumber, i)
		}
		if strings.Contains(title, "支付时间") || strings.Contains(lower, "payment") {
			set(&idx.PaymentTime, i)
		}
		if strings.Contains(title, "完成时间") || strings.Contains(lower, "complete") || strings.Contains(lower, "finished") {
			set(&idx.CompletedAt, i)
		}
	}
	return idx
}

func missingRequiredColumns(idx columnIndex, orderType string) []string {
	requirements := []struct {
		column int
		name   string
	}{
		{idx.GSPOrderNo, "GSP订单号"},
		{idx.ShopCode, "店铺编号"},
		{idx.OwnerName, "负责人"},
		{idx.ProductName, "商品名称"},
		{idx.Spec, "规格"},
		{idx.ItemNo, "货号"},
		{idx.SellerSKU, "卖家SKU"},
		{idx.PlatformSKU, "平台SKU"},
		{idx.PlatformSKC, "平台SKC"},
		{idx.PlatformSPU, "平台SPU"},
		{idx.ProductPrice, "商品价格"},
		{idx.PostalCode, "邮编"},
		{idx.Country, "国家"},
		{idx.Province, "省份"},
		{idx.City, "城市"},
		{idx.CustomerFullName, "用户全称"},
		{idx.CustomerLastName, "用户姓氏"},
		{idx.CustomerFirstName, "用户名字"},
		{idx.PhoneNumber, "手机号"},
		{idx.Email, "用户邮箱"},
	}
	if orderType == "platform" {
		requirements = append(requirements,
			struct {
				column int
				name   string
			}{idx.ShippingWarehouse, "发货仓库"},
			struct {
				column int
				name   string
			}{idx.ExpectedQty, "应履约件数"},
		)
	}

	var missing []string
	for _, req := range requirements {
		if req.column < 0 {
			missing = append(missing, req.name)
		}
	}
	return missing
}

func applyPlatformColumnDefaults(idx *columnIndex) {
	if idx.ShippingWarehouse < 0 {
		idx.ShippingWarehouse = 1 // B列
	}
	if idx.ShopCode < 0 {
		idx.ShopCode = 2 // C列
	}
	if idx.OwnerName < 0 {
		idx.OwnerName = 3 // D列
	}
	if idx.Spec < 0 {
		idx.Spec = 5 // F列
	}
	if idx.ItemNo < 0 {
		idx.ItemNo = 6 // G列
	}
	if idx.SpecialNote < 0 {
		idx.SpecialNote = 12 // M列
	}
	if idx.ExpectedQty < 0 {
		idx.ExpectedQty = 13 // N列
	}
}

func applyFactoryColumnDefaults(idx *columnIndex) {
	if idx.ShopCode < 0 {
		idx.ShopCode = 2 // C列
	}
	if idx.OwnerName < 0 {
		idx.OwnerName = 3 // D列
	}
	if idx.Spec < 0 {
		idx.Spec = 5 // F列
	}
	if idx.ItemNo < 0 {
		idx.ItemNo = 6 // G列
	}
	if idx.SpecialNote < 0 {
		idx.SpecialNote = 12 // M列
	}
	if idx.ExpectedQty < 0 {
		idx.ExpectedQty = 13 // N列
	}
}

type workbookSheets map[string][][]string

func parseWorkbook(data []byte) (result workbookSheets, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Excel解析异常: %v", r)
			result = nil
		}
	}()

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	result = make(workbookSheets)
	for _, rawName := range f.GetSheetList() {
		name := strings.TrimSpace(rawName)
		rows, err := f.GetRows(rawName)
		if err != nil {
			return nil, err
		}
		cleanedRows := make([][]string, len(rows))
		for i, row := range rows {
			cleanedRow := make([]string, len(row))
			for j, value := range row {
				cleanedRow[j] = strings.TrimSpace(value)
			}
			cleanedRows[i] = cleanedRow
		}
		result[name] = cleanedRows
	}
	return result, nil
}

var errSheetNotFound = errors.New("sheet not found")

func getSheetRows(sheets workbookSheets, keywords []string) ([][]string, string, error) {
	if len(sheets) == 0 {
		return nil, "", fmt.Errorf("%w: workbook has no sheets", errSheetNotFound)
	}

	normalizedKeywords := make([]string, 0, len(keywords))
	for _, k := range keywords {
		normalizedKeywords = append(normalizedKeywords, strings.ToLower(strings.TrimSpace(k)))
	}

	for name, rows := range sheets {
		trimmed := strings.TrimSpace(name)
		lowerName := strings.ToLower(trimmed)
		for _, keyword := range normalizedKeywords {
			if lowerName == keyword || strings.Contains(lowerName, keyword) {
				return rows, trimmed, nil
			}
		}
	}

	return nil, "", fmt.Errorf("%w: 未找到匹配的工作表, 期望关键词: %s", errSheetNotFound, strings.Join(keywords, ","))
}

func findHeaderRow(rows [][]string, orderType string) (int, columnIndex, []string) {
	for i, row := range rows {
		idx := detectColumns(row)
		if !looksLikeHeader(idx) {
			continue
		}
		missing := missingRequiredColumns(idx, orderType)
		if len(missing) > 0 {
			return -1, columnIndex{}, missing
		}
		return i, idx, nil
	}
	return -1, columnIndex{}, []string{"未找到有效的表头，请确认模板是否正确"}
}

func looksLikeHeader(idx columnIndex) bool {
	positiveCount := 0
	fields := []int{
		idx.GSPOrderNo,
		idx.ShopCode,
		idx.OwnerName,
		idx.ProductName,
		idx.Spec,
	}
	for _, val := range fields {
		if val >= 0 {
			positiveCount++
		}
	}
	return positiveCount >= 3
}

func parsePlatformSheet(sheets workbookSheets, defaultAddress string, shippingSet map[string]struct{}) ([]models.OrderInfo, []validationError) {
	rows, sheetName, err := getSheetRows(sheets, []string{"平台面单", "platform"})
	if err != nil {
		if errors.Is(err, errSheetNotFound) {
			return nil, nil
		}
		return nil, []validationError{{Sheet: "平台面单", Row: 1, Field: "文件", Code: "custom", Extra: err.Error()}}
	}
	if len(rows) == 0 {
		return []models.OrderInfo{}, nil
	}

	headerRow, idx, missing := findHeaderRow(rows, "platform")
	if headerRow < 0 {
		applyPlatformColumnDefaults(&idx)
		missing = missingRequiredColumns(idx, "platform")
		if len(missing) > 0 {
			errs := make([]validationError, len(missing))
			for idxMissing, name := range missing {
				errs[idxMissing] = validationError{Sheet: sheetName, Row: headerRow + 1, Field: name, Code: "custom", Extra: "缺少列"}
			}
			return nil, errs
		}
		headerRow = 0
	} else {
		applyPlatformColumnDefaults(&idx)
	}

	var orders []models.OrderInfo
	var validationErrors []validationError
	var duplicateRows []int

	for i := headerRow + 1; i < len(rows); i++ {
		row := rows[i]
		if rowIsEmpty(row) {
			continue
		}
		excelRowNum := i + 1
		order, rowErrors := buildOrderFromRow(row, idx, defaultAddress, shippingSet, excelRowNum, sheetName, "platform")
		if len(rowErrors) > 0 {
			validationErrors = append(validationErrors, rowErrors...)
			continue
		}
		exists, err := platformOrderExists(order)
		if err != nil {
			validationErrors = append(validationErrors, validationError{
				Sheet: sheetName,
				Row:   excelRowNum,
				Field: "记录",
				Code:  "duplicate_check",
				Extra: fmt.Sprintf("检查重复时出错: %v", err),
			})
			continue
		}
		if exists {
			duplicateRows = append(duplicateRows, excelRowNum)
		}
		orders = append(orders, order)
	}
	if len(duplicateRows) > 0 {
		addDuplicateRowsWarning(&validationErrors, sheetName, duplicateRows)
	}
	return orders, validationErrors
}

func parseFactorySheet(sheets workbookSheets, defaultAddress string, shippingSet map[string]struct{}) ([]models.OrderInfo, []validationError) {
	rows, sheetName, err := getSheetRows(sheets, []string{"工厂物流", "factory"})
	if err != nil {
		if errors.Is(err, errSheetNotFound) {
			return nil, nil
		}
		return nil, []validationError{{Sheet: "工厂物流", Row: 1, Field: "文件", Code: "custom", Extra: err.Error()}}
	}
	if len(rows) == 0 {
		return []models.OrderInfo{}, nil
	}

	headerRow, idx, missing := findHeaderRow(rows, "factory")
	if headerRow < 0 {
		applyFactoryColumnDefaults(&idx)
		missing = missingRequiredColumns(idx, "factory")
		if len(missing) > 0 {
			errs := make([]validationError, len(missing))
			for idxMissing, name := range missing {
				errs[idxMissing] = validationError{Sheet: sheetName, Row: headerRow + 1, Field: name, Code: "custom", Extra: "缺少列"}
			}
			return nil, errs
		}
		headerRow = 0
	} else {
		applyFactoryColumnDefaults(&idx)
	}

	var orders []models.OrderInfo
	var validationErrors []validationError
	var duplicateRows []int

	for i := headerRow + 1; i < len(rows); i++ {
		row := rows[i]
		if rowIsEmpty(row) {
			continue
		}
		excelRowNum := i + 1
		order, rowErrors := buildOrderFromRow(row, idx, defaultAddress, shippingSet, excelRowNum, sheetName, "factory")
		if len(rowErrors) > 0 {
			validationErrors = append(validationErrors, rowErrors...)
			continue
		}
		exists, err := factoryOrderExists(order)
		if err != nil {
			validationErrors = append(validationErrors, validationError{
				Sheet: sheetName,
				Row:   excelRowNum,
				Field: "记录",
				Code:  "duplicate_check",
				Extra: fmt.Sprintf("检查重复时出错: %v", err),
			})
			continue
		}
		if exists {
			duplicateRows = append(duplicateRows, excelRowNum)
		}
		orders = append(orders, order)
	}
	if len(duplicateRows) > 0 {
		addDuplicateRowsWarning(&validationErrors, sheetName, duplicateRows)
	}
	return orders, validationErrors
}

func platformOrderExists(order models.OrderInfo) (bool, error) {
	var count int64
	err := database.DB.Model(&models.OrderInfo{}).
		Where("order_type = ?", "platform").
		Where("gsp_order_no = ? AND shipping_warehouse_code = ? AND shop_code = ? AND owner_name = ? AND spec = ? AND item_no = ? AND seller_sku = ?",
			order.GSPOrderNo,
			order.ShippingWarehouseCode,
			order.ShopCode,
			order.OwnerName,
			order.Spec,
			order.ItemNo,
			order.SellerSKU,
		).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func factoryOrderExists(order models.OrderInfo) (bool, error) {
	var count int64
	err := database.DB.Model(&models.OrderInfo{}).
		Where("order_type = ?", "factory").
		Where("gsp_order_no = ? AND shop_code = ? AND owner_name = ? AND spec = ? AND item_no = ? AND seller_sku = ?",
			order.GSPOrderNo,
			order.ShopCode,
			order.OwnerName,
			order.Spec,
			order.ItemNo,
			order.SellerSKU,
		).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func addDuplicateRowsWarning(errors *[]validationError, sheetName string, rows []int) {
	if len(rows) == 0 {
		return
	}
	rowTexts := make([]string, len(rows))
	for i, row := range rows {
		rowTexts[i] = strconv.Itoa(row)
	}
	message := fmt.Sprintf("%s[第%s行]已有类似记录，请检查!", sheetName, strings.Join(rowTexts, ","))
	*errors = append(*errors, validationError{
		Sheet: sheetName,
		Row:   rows[0],
		Field: "记录",
		Code:  "duplicate",
		Extra: message,
	})
}

func buildOrderFromRow(row []string, idx columnIndex, defaultAddress string, shippingSet map[string]struct{}, rowNum int, sheetName, orderType string) (models.OrderInfo, []validationError) {
	var errors []validationError
	order := models.OrderInfo{OrderType: orderType}
	isPlatform := orderType == "platform"

	order.GSPOrderNo = getValue(row, idx.GSPOrderNo)
	if order.GSPOrderNo == "" {
		appendValidationError(&errors, sheetName, rowNum, "GSP订单号", "required", "")
	}

	order.ShippingWarehouseCode = getValue(row, idx.ShippingWarehouse)
	if isPlatform {
		if order.ShippingWarehouseCode == "" {
			appendValidationError(&errors, sheetName, rowNum, "发货仓库", "required", "")
		} else if len(shippingSet) > 0 {
			key := strings.ToUpper(strings.TrimSpace(order.ShippingWarehouseCode))
			if _, ok := shippingSet[key]; !ok {
				appendValidationError(&errors, sheetName, rowNum, "发货仓库", "dict", order.ShippingWarehouseCode)
			}
		}
	} else if order.ShippingWarehouseCode != "" && len(shippingSet) > 0 {
		key := strings.ToUpper(strings.TrimSpace(order.ShippingWarehouseCode))
		if _, ok := shippingSet[key]; !ok {
			appendValidationError(&errors, sheetName, rowNum, "发货仓库", "dict", order.ShippingWarehouseCode)
		}
	}

	order.ShopCode = getValue(row, idx.ShopCode)
	if order.ShopCode == "" {
		appendValidationError(&errors, sheetName, rowNum, "店铺编号", "required", "")
	}

	order.OwnerName = getValue(row, idx.OwnerName)
	if order.OwnerName == "" {
		appendValidationError(&errors, sheetName, rowNum, "负责人", "required", "")
	}

	order.ProductName = getValue(row, idx.ProductName)
	// 平台、工厂均允许商品名称为空

	order.Spec = strings.TrimSpace(getValue(row, idx.Spec))
	if order.Spec != "" {
		normalizedSpec := strings.ReplaceAll(strings.ToLower(order.Spec), " ", "")
		normalizedSpec = strings.ReplaceAll(normalizedSpec, "x", "*")
		order.Spec = normalizedSpec
	}
	if order.Spec == "" {
		appendValidationError(&errors, sheetName, rowNum, "规格", "required", "")
	} else if !specPattern.MatchString(order.Spec) {
		appendValidationError(&errors, sheetName, rowNum, "规格", "spec_format", "")
	}

	order.ItemNo = getValue(row, idx.ItemNo)
	if order.ItemNo == "" {
		appendValidationError(&errors, sheetName, rowNum, "货号", "required", "")
	}

	order.SellerSKU = getValue(row, idx.SellerSKU)
	// 平台、工厂均允许卖家SKU为空

	order.PlatformSKU = getValue(row, idx.PlatformSKU)
	// 平台、工厂均允许平台SKU为空

	order.PlatformSKC = getValue(row, idx.PlatformSKC)
	// 平台、工厂均允许平台SKC为空

	order.PlatformSPU = getValue(row, idx.PlatformSPU)
	// 平台、工厂均允许平台SPU为空

	if idx.ProductPrice >= 0 {
		priceStr := getValue(row, idx.ProductPrice)
		if priceStr != "" {
			if price, err := parseFloat(priceStr); err != nil {
				appendValidationError(&errors, sheetName, rowNum, "商品价格", "numeric", "")
			} else {
				order.ProductPrice = price
			}
		}
	}

	if idx.ExpectedQty >= 0 {
		qtyStr := getValue(row, idx.ExpectedQty)
		if isPlatform {
			if qtyStr == "" {
				appendValidationError(&errors, sheetName, rowNum, "应履约件数", "required", "")
			} else if qty, err := parseInt(qtyStr); err != nil {
				appendValidationError(&errors, sheetName, rowNum, "应履约件数", "numeric", "")
			} else {
				order.ExpectedFulfillmentQty = qty
			}
		} else {
			// 工厂物流无需校验应履约件数，若有值则尝试解析（解析失败直接忽略）
			if qtyStr != "" {
				if qty, err := parseInt(qtyStr); err == nil {
					order.ExpectedFulfillmentQty = qty
				}
			}
		}
	} else if isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "应履约件数", "required", "")
	}

	if idx.SpecialNote >= 0 {
		note := getValue(row, idx.SpecialNote)
		if note != "" && !chinesePattern.MatchString(note) {
			appendValidationError(&errors, sheetName, rowNum, "特殊产品备注", "hanzi", "")
		}
		order.SpecialProductNote = note
	}

	order.PostalCode = getValue(row, idx.PostalCode)
	if order.PostalCode == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "邮编", "required", "")
	}
	order.Country = getValue(row, idx.Country)
	if order.Country == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "国家", "required", "")
	}
	order.Province = getValue(row, idx.Province)
	if order.Province == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "省份", "required", "")
	}
	order.City = getValue(row, idx.City)
	if order.City == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "城市", "required", "")
	}
	order.District = getValue(row, idx.District)

	order.AddressLine1 = getValue(row, idx.AddressLine1)
	order.AddressLine2 = getValue(row, idx.AddressLine2)
	if order.AddressLine1 == "" {
		order.AddressLine1 = defaultAddress
	}

	order.CustomerFullName = getValue(row, idx.CustomerFullName)
	order.CustomerLastName = getValue(row, idx.CustomerLastName)
	order.CustomerFirstName = getValue(row, idx.CustomerFirstName)

	if order.CustomerFullName == "" && (order.CustomerLastName != "" || order.CustomerFirstName != "") {
		order.CustomerFullName = strings.TrimSpace(order.CustomerLastName + " " + order.CustomerFirstName)
	}

	if order.CustomerFullName == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "用户全称", "required", "")
	}
	if order.CustomerLastName == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "用户姓氏", "required", "")
	}
	if order.CustomerFirstName == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "用户名字", "required", "")
	}

	order.PhoneNumber = getValue(row, idx.PhoneNumber)
	if order.PhoneNumber == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "手机号", "required", "")
	}

	order.Email = getValue(row, idx.Email)
	if order.Email == "" && !isPlatform {
		appendValidationError(&errors, sheetName, rowNum, "用户邮箱", "required", "")
	}

	order.TaxNumber = getValue(row, idx.TaxNumber)

	if idx.CurrencyCode >= 0 {
		order.CurrencyCode = getValue(row, idx.CurrencyCode)
	}
	if order.CurrencyCode == "" {
		order.CurrencyCode = "CNY"
	}

	if idx.OrderCreatedAt >= 0 {
		if t, err := parseQueryTime(getValue(row, idx.OrderCreatedAt)); err == nil {
			order.OrderCreatedAt = t
		}
	}
	if order.OrderCreatedAt.IsZero() {
		order.OrderCreatedAt = time.Now()
	}

	if idx.RequiredSignAt >= 0 {
		if t, err := parseQueryTime(getValue(row, idx.RequiredSignAt)); err == nil {
			order.RequiredSignAt = &t
		}
	}
	if idx.PaymentTime >= 0 {
		if t, err := parseQueryTime(getValue(row, idx.PaymentTime)); err == nil {
			order.PaymentTime = &t
		}
	}
	if idx.CompletedAt >= 0 {
		if t, err := parseQueryTime(getValue(row, idx.CompletedAt)); err == nil {
			order.CompletedAt = &t
		}
	}

	if idx.ProductID >= 0 {
		order.ProductID = getValue(row, idx.ProductID)
	}

	if idx.ExpectedRevenue >= 0 {
		if value, err := parseFloat(getValue(row, idx.ExpectedRevenue)); err == nil {
			order.ExpectedRevenue = value
		}
	}
	if order.ExpectedRevenue == 0 {
		order.ExpectedRevenue = order.ProductPrice * float64(order.ExpectedFulfillmentQty)
	}

	if !isPlatform {
		if order.ExpectedFulfillmentQty > 0 {
			order.ItemCount = order.ExpectedFulfillmentQty
		} else {
			order.ItemCount = 1
		}
	} else {
		if order.ExpectedFulfillmentQty > 0 {
			order.ItemCount = order.ExpectedFulfillmentQty
		} else {
			order.ItemCount = 1
		}
	}

	return order, errors
}

func writePlatformHeader(f *excelize.File, sheet string) {
	headers := []string{
		"GSP订单号", "发货仓库", "店铺编号", "负责人", "商品名称", "规格", "货号", "卖家SKU", "平台SKU", "平台SKC", "平台SPU",
		"商品价格", "特殊产品备注", "应履约件数", "邮编", "国家", "省份", "城市", "区", "用户地址1", "用户地址2",
		"用户全称", "用户姓氏", "用户名字", "手机号", "用户邮箱", "税号",
	}
	for i, value := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, value)
	}
}

func writeFactoryHeader(f *excelize.File, sheet string) {
	headers := []string{
		"GSP订单号", "订单创建时间", "店铺编号", "负责人", "商品名称", "规格", "货号", "卖家SKU", "平台SKU", "平台SKC", "平台SPU",
		"商品价格", "特殊产品备注", "应履约件数", "邮编", "国家", "省份", "城市", "区", "用户地址1", "用户地址2",
		"用户全称", "用户姓氏", "用户名字", "手机号", "用户邮箱", "税号",
	}
	for i, value := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, value)
	}
}

func clearSheetData(f *excelize.File, sheet string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}
	for row := len(rows); row >= 2; row-- {
		if err := f.RemoveRow(sheet, row); err != nil {
			return err
		}
	}
	return nil
}

func queryOrdersForExport(c *gin.Context, orderType string) ([]models.OrderInfo, error) {
	query := database.DB.Model(&models.OrderInfo{}).Where("order_type = ?", orderType)
	query = applyOrderFilters(c, query)
	if !isAdminUser(c) {
		userID := currentUserID(c)
		query = query.Where("created_by = ?", userID)
	}

	var orders []models.OrderInfo
	if err := query.Order("id DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func writePlatformRow(f *excelize.File, sheet string, row int, order models.OrderInfo) {
	priceValue := ""
	if order.ProductPrice != 0 {
		priceValue = fmt.Sprintf("%g", order.ProductPrice)
	}

	values := []interface{}{
		order.GSPOrderNo,
		order.ShippingWarehouseCode,
		order.ShopCode,
		order.OwnerName,
		order.ProductName,
		order.Spec,
		order.ItemNo,
		order.SellerSKU,
		order.PlatformSKU,
		order.PlatformSKC,
		order.PlatformSPU,
		priceValue,
		order.SpecialProductNote,
		order.ExpectedFulfillmentQty,
		order.PostalCode,
		order.Country,
		order.Province,
		order.City,
		order.District,
		order.AddressLine1,
		order.AddressLine2,
		order.CustomerFullName,
		order.CustomerLastName,
		order.CustomerFirstName,
		order.PhoneNumber,
		order.Email,
		order.TaxNumber,
	}
	for i, value := range values {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue(sheet, cell, value)
	}
}

func writeFactoryRow(f *excelize.File, sheet string, row int, order models.OrderInfo) {
	createdAt := ""
	if !order.OrderCreatedAt.IsZero() {
		createdAt = order.OrderCreatedAt.Format("2006-01-02 15:04:05")
	}
	priceValue := ""
	if order.ProductPrice != 0 {
		priceValue = fmt.Sprintf("%g", order.ProductPrice)
	}

	values := []interface{}{
		order.GSPOrderNo,
		createdAt,
		order.ShopCode,
		order.OwnerName,
		order.ProductName,
		order.Spec,
		order.ItemNo,
		order.SellerSKU,
		order.PlatformSKU,
		order.PlatformSKC,
		order.PlatformSPU,
		priceValue,
		order.SpecialProductNote,
		order.CurrencyCode,
		order.PostalCode,
		order.Country,
		order.Province,
		order.City,
		order.District,
		order.AddressLine1,
		order.AddressLine2,
		order.CustomerFullName,
		order.CustomerLastName,
		order.CustomerFirstName,
		order.PhoneNumber,
		order.Email,
		order.TaxNumber,
	}
	for i, value := range values {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue(sheet, cell, value)
	}
}

func getValue(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func parseFloat(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty")
	}
	return strconv.ParseFloat(value, 64)
}

func parseInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty")
	}
	return strconv.Atoi(value)
}

func rowIsEmpty(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func upsertOrder(tx *gorm.DB, order *models.OrderInfo) error {
	if order.OrderCreatedAt.IsZero() {
		order.OrderCreatedAt = time.Now()
	}
	if order.ItemCount == 0 {
		order.ItemCount = 1
	}

	var existing models.OrderInfo
	err := tx.Where("gsp_order_no = ? AND order_type = ? AND order_created_at = ?", order.GSPOrderNo, order.OrderType, order.OrderCreatedAt).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(order).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.OrderInfo{}).
				Where("id = ?", order.ID).
				UpdateColumn("updated_at", gorm.Expr("NULL")).Error; err != nil {
				return err
			}
			order.UpdatedAt = time.Time{}
			return nil
		}
		return err
	}

	order.ID = existing.ID
	order.CreatedAt = existing.CreatedAt
	order.CreatedBy = existing.CreatedBy
	return tx.Model(&existing).Updates(order).Error
}

type orderResponse struct {
	models.OrderInfo
	MaterialImageURL          string  `json:"material_image_url"`
	MaterialAttachmentID      uint64  `json:"material_attachment_id"`
	MaterialStorage           string  `json:"material_storage"`
	MaterialFileName          string  `json:"material_file_name"`
	MaterialAssetID           *uint64 `json:"material_asset_id"`
	ShippingLabelURL          string  `json:"shipping_label_url"`
	ShippingLabelAttachmentID uint64  `json:"shipping_label_attachment_id"`
	ShippingLabelStorage      string  `json:"shipping_label_storage"`
	ShippingLabelFileName     string  `json:"shipping_label_file_name"`
	ShippingLabelAssetID      *uint64 `json:"shipping_label_asset_id"`
}

func assignOrderFields(order *models.OrderInfo, params *saveOrderParams) {
	order.GSPOrderNo = params.GSPOrderNo
	if params.OrderType != "" {
		order.OrderType = params.OrderType
	} else if order.OrderType == "" {
		order.OrderType = "platform"
	}
	order.ShippingWarehouseCode = params.ShippingWarehouseCode
	order.ShopCode = params.ShopCode
	order.ProductID = params.ProductID
	order.OwnerName = params.OwnerName
	order.ProductName = params.ProductName
	order.Spec = params.Spec
	order.ItemNo = params.ItemNo
	order.SellerSKU = params.SellerSKU
	order.PlatformSKU = params.PlatformSKU
	order.PlatformSKC = params.PlatformSKC
	order.PlatformSPU = params.PlatformSPU
	order.CurrencyCode = params.CurrencyCode
	order.PostalCode = params.PostalCode
	order.Country = params.Country
	order.Province = params.Province
	order.City = params.City
	order.District = params.District
	order.CustomerFullName = params.CustomerFullName
	order.CustomerLastName = params.CustomerLastName
	order.CustomerFirstName = params.CustomerFirstName
	order.PhoneNumber = params.PhoneNumber
	order.Email = params.Email

	defaultAddress := getConfigValue("default_address")
	if defaultAddress == "" {
		defaultAddress = config.AppConfig.LocalBaseURL
	}

	order.AddressLine1 = strings.TrimSpace(params.AddressLine1)
	order.AddressLine2 = strings.TrimSpace(params.AddressLine2)
	if order.AddressLine1 == "" {
		order.AddressLine1 = defaultAddress
	}

	if params.Status != nil {
		order.Status = *params.Status
	}
	if params.ExpectedFulfillmentQty != nil {
		order.ExpectedFulfillmentQty = *params.ExpectedFulfillmentQty
	}

	if params.ItemCount != nil {
		order.ItemCount = *params.ItemCount
	} else if order.ItemCount == 0 {
		if order.ExpectedFulfillmentQty > 0 {
			order.ItemCount = order.ExpectedFulfillmentQty
		} else {
			order.ItemCount = 1
		}
	}

	if params.ProductPrice != nil {
		order.ProductPrice = *params.ProductPrice
	}
	if params.ExpectedRevenue != nil {
		order.ExpectedRevenue = *params.ExpectedRevenue
	}

	if params.SpecialProductNote != nil {
		order.SpecialProductNote = *params.SpecialProductNote
	}
	if params.TaxNumber != nil {
		order.TaxNumber = *params.TaxNumber
	}

	if params.OrderCreatedAt != "" {
		if t := parseTime(params.OrderCreatedAt); !t.IsZero() {
			order.OrderCreatedAt = t
		}
	}
	if params.PaymentTime != nil {
		order.PaymentTime = parseTimePtr(*params.PaymentTime)
	}
	if params.CompletedAt != nil {
		order.CompletedAt = parseTimePtr(*params.CompletedAt)
	}
	if params.RequiredSignAt != nil {
		order.RequiredSignAt = parseTimePtr(*params.RequiredSignAt)
	}
}

func autoLinkMaterialForOrder(ctx context.Context, tx *gorm.DB, order *models.OrderInfo, operatorID uint) error {
	if order == nil || order.ID == 0 {
		return nil
	}

	orderType := strings.ToLower(strings.TrimSpace(order.OrderType))
	//平台订单，产品素材图文件名需与表格中“订单号”完全一致；
	matchKey := strings.TrimSpace(order.GSPOrderNo)
	//工厂物流订单，产品素材图文件名需与表格中“货号”完全一致；
	if orderType == "factory" {
		matchKey = strings.TrimSpace(order.ItemNo)
	}
	if matchKey == "" {
		return nil
	}

	db := database.DB
	if tx != nil {
		db = tx
	}

	lowerKey := strings.ToLower(matchKey)
	var material models.MaterialAsset
	if err := db.Where("LOWER(TRIM(title)) = ?", lowerKey).Or("LOWER(TRIM(file_name)) LIKE ?", lowerKey+".%").
		First(&material).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("autoLinkMaterialForOrder: no material found for key=%s orderID=%d", matchKey, order.ID)
			return nil
		}
		log.Printf("autoLinkMaterialForOrder: query material failed for key=%s orderID=%d err=%v", matchKey, order.ID, err)
		return err
	}
	log.Printf("autoLinkMaterialForOrder: matched material=%d for key=%s orderID=%d", material.ID, matchKey, order.ID)

	if _, err := upsertOrderMaterialAttachment(ctx, db, order.ID, "material_image", &material, operatorID); err != nil {
		return err
	}
	return nil
}

func validateOrderParams(params *saveOrderParams) error {
	orderType := strings.ToLower(strings.TrimSpace(params.OrderType))
	if orderType != "factory" {
		orderType = "platform"
	}
	params.OrderType = orderType

	sheetName := "平台面单"
	if orderType == "factory" {
		sheetName = "工厂物流"
	}

	shippingSet := loadShippingWarehouseSet()
	defaultAddress := strings.TrimSpace(getConfigValue("default_address"))
	if defaultAddress == "" && config.AppConfig != nil {
		defaultAddress = strings.TrimSpace(config.AppConfig.LocalBaseURL)
	}
	if defaultAddress == "" {
		defaultAddress = "-"
	}

	var validationErrors []validationError
	rowNum := 1
	addErr := func(field, code, extra string) {
		appendValidationError(&validationErrors, sheetName, rowNum, field, code, extra)
	}

	params.GSPOrderNo = strings.TrimSpace(params.GSPOrderNo)
	if params.GSPOrderNo == "" {
		addErr("GSP订单号", "required", "")
	}

	params.ShippingWarehouseCode = strings.TrimSpace(params.ShippingWarehouseCode)
	if orderType == "platform" {
		if params.ShippingWarehouseCode == "" {
			addErr("发货仓库", "required", "")
		} else if shippingSet != nil {
			key := strings.ToUpper(params.ShippingWarehouseCode)
			if _, ok := shippingSet[key]; !ok {
				addErr("发货仓库", "dict", params.ShippingWarehouseCode)
			}
		}
	} else if params.ShippingWarehouseCode != "" && shippingSet != nil {
		key := strings.ToUpper(params.ShippingWarehouseCode)
		if _, ok := shippingSet[key]; !ok {
			addErr("发货仓库", "dict", params.ShippingWarehouseCode)
		}
	}

	params.ShopCode = strings.TrimSpace(params.ShopCode)
	if params.ShopCode == "" {
		addErr("店铺编号", "required", "")
	}

	params.OwnerName = strings.TrimSpace(params.OwnerName)
	if params.OwnerName == "" {
		addErr("负责人", "required", "")
	}

	params.ProductName = strings.TrimSpace(params.ProductName)

	params.Spec = strings.TrimSpace(params.Spec)
	if params.Spec != "" {
		normalized := strings.ReplaceAll(strings.ToLower(params.Spec), " ", "")
		normalized = strings.ReplaceAll(normalized, "x", "*")
		params.Spec = normalized
	}
	if params.Spec == "" {
		addErr("规格", "required", "")
	} else if !specPattern.MatchString(params.Spec) {
		addErr("规格", "spec_format", "")
	}

	params.ItemNo = strings.TrimSpace(params.ItemNo)
	if params.ItemNo == "" {
		addErr("货号", "required", "")
	}

	if params.SellerSKU != "" {
		params.SellerSKU = strings.TrimSpace(params.SellerSKU)
	}
	if params.PlatformSKU != "" {
		params.PlatformSKU = strings.TrimSpace(params.PlatformSKU)
	}
	if params.PlatformSKC != "" {
		params.PlatformSKC = strings.TrimSpace(params.PlatformSKC)
	}
	if params.PlatformSPU != "" {
		params.PlatformSPU = strings.TrimSpace(params.PlatformSPU)
	}

	if params.ProductPrice != nil {
		if *params.ProductPrice < 0 {
			addErr("商品价格", "numeric", "")
		}
	}

	if orderType == "platform" {
		if params.ExpectedFulfillmentQty == nil {
			addErr("应履约件数", "required", "")
		} else if *params.ExpectedFulfillmentQty <= 0 {
			addErr("应履约件数", "numeric", "")
		}
	} else if params.ExpectedFulfillmentQty != nil && *params.ExpectedFulfillmentQty < 0 {
		addErr("应履约件数", "numeric", "")
	}

	if params.SpecialProductNote != nil {
		note := strings.TrimSpace(*params.SpecialProductNote)
		if note == "" {
			params.SpecialProductNote = nil
		} else if !chinesePattern.MatchString(note) {
			addErr("特殊产品备注", "hanzi", "")
		} else {
			params.SpecialProductNote = &note
		}
	}

	params.PostalCode = strings.TrimSpace(params.PostalCode)
	params.Country = strings.TrimSpace(params.Country)
	params.Province = strings.TrimSpace(params.Province)
	params.City = strings.TrimSpace(params.City)
	params.District = strings.TrimSpace(params.District)

	params.AddressLine1 = strings.TrimSpace(params.AddressLine1)
	params.AddressLine2 = strings.TrimSpace(params.AddressLine2)
	if params.AddressLine1 == "" {
		params.AddressLine1 = defaultAddress
	}

	params.CustomerFullName = strings.TrimSpace(params.CustomerFullName)
	params.CustomerLastName = strings.TrimSpace(params.CustomerLastName)
	params.CustomerFirstName = strings.TrimSpace(params.CustomerFirstName)

	params.PhoneNumber = strings.TrimSpace(params.PhoneNumber)
	params.Email = strings.TrimSpace(params.Email)

	if orderType == "factory" {
		if params.PostalCode == "" {
			addErr("邮编", "required", "")
		}
		if params.Country == "" {
			addErr("国家", "required", "")
		}
		if params.Province == "" {
			addErr("省份", "required", "")
		}
		if params.City == "" {
			addErr("城市", "required", "")
		}
		if params.CustomerFullName == "" {
			addErr("用户全称", "required", "")
		}
		if params.CustomerLastName == "" {
			addErr("用户姓氏", "required", "")
		}
		if params.CustomerFirstName == "" {
			addErr("用户名字", "required", "")
		}
		if params.PhoneNumber == "" {
			addErr("手机号", "required", "")
		}
		if params.Email == "" {
			addErr("用户邮箱", "required", "")
		}
	}

	if params.Status == nil {
		addErr("状态", "required", "")
	}

	params.CurrencyCode = strings.TrimSpace(params.CurrencyCode)
	if params.CurrencyCode == "" {
		params.CurrencyCode = "CNY"
	}

	if params.OrderCreatedAt != "" {
		value := strings.TrimSpace(params.OrderCreatedAt)
		if t, err := parseQueryTime(value); err == nil {
			params.OrderCreatedAt = t.Format("2006-01-02 15:04:05")
		} else {
			addErr("订单创建时间", "custom", " 格式应为 YYYY-MM-DD HH:mm:ss")
		}
	}

	normalizeOptionalTime := func(label string, ptr **string) {
		if *ptr == nil {
			return
		}
		value := strings.TrimSpace(**ptr)
		if value == "" {
			*ptr = nil
			return
		}
		if t, err := parseQueryTime(value); err == nil {
			formatted := t.Format("2006-01-02 15:04:05")
			*ptr = &formatted
		} else {
			addErr(label, "custom", " 格式应为 YYYY-MM-DD HH:mm:ss")
		}
	}

	normalizeOptionalTime("支付时间", &params.PaymentTime)
	normalizeOptionalTime("完成时间", &params.CompletedAt)
	normalizeOptionalTime("要求签收时间", &params.RequiredSignAt)

	if len(validationErrors) > 0 {
		formatted := formatValidationErrors(validationErrors)
		if len(formatted) > 0 {
			return errors.New(strings.Join(formatted, "; "))
		}
		return errors.New("订单数据验证失败")
	}

	if params.ExpectedRevenue == nil && params.ProductPrice != nil && params.ExpectedFulfillmentQty != nil {
		value := *params.ProductPrice * float64(*params.ExpectedFulfillmentQty)
		params.ExpectedRevenue = &value
	}

	if params.ItemCount == nil {
		if params.ExpectedFulfillmentQty != nil && *params.ExpectedFulfillmentQty > 0 {
			count := *params.ExpectedFulfillmentQty
			params.ItemCount = &count
		} else {
			count := 1
			params.ItemCount = &count
		}
	}

	return nil
}

func getConfigValue(key string) string {
	var cfg models.Config
	if err := database.DB.Where("`key` = ?", key).First(&cfg).Error; err == nil {
		return cfg.Value
	}
	return ""
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if t, err := parseQueryTime(value); err == nil {
		return t
	}
	return time.Time{}
}

func parseTimePtr(value string) *time.Time {
	if value == "" {
		return nil
	}
	if t, err := parseQueryTime(value); err == nil {
		return &t
	}
	return nil
}

type validationError struct {
	Sheet string
	Row   int
	Field string
	Code  string
	Extra string
}

func appendValidationError(list *[]validationError, sheet string, row int, field, code, extra string) {
	*list = append(*list, validationError{Sheet: sheet, Row: row, Field: field, Code: code, Extra: extra})
}

func formatValidationErrors(errors []validationError) []string {
	if len(errors) == 0 {
		return nil
	}

	sheetGroups := make(map[string]map[string][]validationError)
	for _, err := range errors {
		if _, ok := sheetGroups[err.Sheet]; !ok {
			sheetGroups[err.Sheet] = make(map[string][]validationError)
		}
		key := err.Field + "|" + err.Code + "|" + err.Extra
		sheetGroups[err.Sheet][key] = append(sheetGroups[err.Sheet][key], err)
	}

	var messages []string
	sheetNames := make([]string, 0, len(sheetGroups))
	for sheet := range sheetGroups {
		sheetNames = append(sheetNames, sheet)
	}
	sort.Strings(sheetNames)

	for _, sheet := range sheetNames {
		group := sheetGroups[sheet]
		keys := make([]string, 0, len(group))
		for key := range group {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		var parts []string
		for _, key := range keys {
			errs := group[key]
			kParts := strings.Split(key, "|")
			field := kParts[0]
			code := kParts[1]
			extra := ""
			if len(kParts) > 2 {
				extra = kParts[2]
			}
			rows := make([]int, len(errs))
			for i, e := range errs {
				rows[i] = e.Row
			}
			sort.Ints(rows)
			rowStrings := make([]string, len(rows))
			for i, r := range rows {
				rowStrings[i] = strconv.Itoa(r)
			}
			message := fmt.Sprintf("%s(%s)%s", field, strings.Join(rowStrings, ","), formatValidationSuffix(code, field, extra))
			parts = append(parts, message)
		}

		if len(parts) == 0 {
			continue
		}

		sheetMessage := sheet + "：" + parts[0]
		for i := 1; i < len(parts); i++ {
			sheetMessage += "\n        " + parts[i]
		}
		messages = append(messages, sheetMessage)
	}

	return messages
}

func formatValidationSuffix(code, field, extra string) string {
	switch code {
	case "required":
		return "为空"
	case "numeric":
		return "需为数字"
	case "dict":
		if extra != "" {
			return fmt.Sprintf(" 不在系统字典中[%s]", extra)
		}
		return " 不在系统字典中"
	case "hanzi":
		return " 仅允许输入汉字"
	case "spec_format":
		return " 格式应为 数字*数字"
	case "custom":
		if extra != "" {
			return extra
		}
		return "有误"
	default:
		if extra != "" {
			return extra
		}
		return "有误"
	}
}
