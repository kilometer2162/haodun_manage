package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"haodun_manage/backend/config"
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"haodun_manage/backend/utils"
)

type storageSettingsResponse struct {
	StorageDriver    string `json:"storage_driver"`
	LocalStoragePath string `json:"local_storage_path"`
	LocalBaseURL     string `json:"local_base_url"`
	COSKeyPrefix     string `json:"cos_key_prefix"`
}

type updateStorageSettingsRequest struct {
	StorageDriver    string `json:"storage_driver" binding:"required"`
	LocalStoragePath string `json:"local_storage_path"`
	LocalBaseURL     string `json:"local_base_url"`
	COSKeyPrefix     string `json:"cos_key_prefix"`
}

func GetStorageSettings(c *gin.Context) {
	resp := storageSettingsResponse{
		StorageDriver:    utils.GetStorageDriver(),
		LocalStoragePath: config.AppConfig.LocalStoragePath,
		LocalBaseURL:     config.AppConfig.LocalBaseURL,
		COSKeyPrefix:     strings.Trim(config.AppConfig.COSKeyPrefix, "/"),
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func UpdateStorageSettings(c *gin.Context) {
	var req updateStorageSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driver := utils.NormalizeStorageDriver(req.StorageDriver)
	if driver == utils.StorageDriverCOS {
		if config.AppConfig.COSSecretID == "" || config.AppConfig.COSSecretKey == "" || config.AppConfig.COSBucket == "" || config.AppConfig.COSRegion == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "COS配置不完整，无法启用COS存储"})
			return
		}
	}

	localPath := strings.TrimSpace(req.LocalStoragePath)
	if localPath == "" {
		localPath = config.AppConfig.LocalStoragePath
	}
	localBaseURL := strings.TrimSpace(req.LocalBaseURL)
	if driver == utils.StorageDriverLocal && localBaseURL == "" {
		localBaseURL = config.AppConfig.LocalBaseURL
	}
	cosKeyPrefix := strings.Trim(strings.TrimSpace(req.COSKeyPrefix), "/")
	if cosKeyPrefix == "" {
		cosKeyPrefix = config.AppConfig.COSKeyPrefix
	}

	config.AppConfig.StorageDriver = driver
	config.AppConfig.LocalStoragePath = localPath
	config.AppConfig.LocalBaseURL = localBaseURL
	config.AppConfig.COSKeyPrefix = cosKeyPrefix

	if err := upsertConfigValue("storage_driver", driver, "存储驱动", "storage"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新存储驱动失败"})
		return
	}
	if err := upsertConfigValue("local_storage_path", localPath, "本地存储路径", "storage"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新本地存储路径失败"})
		return
	}
	if err := upsertConfigValue("local_base_url", localBaseURL, "本地访问URL前缀", "storage"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新本地URL前缀失败"})
		return
	}
	if err := upsertConfigValue("cos_key_prefix", cosKeyPrefix, "COS对象前缀", "storage"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新COS前缀失败"})
		return
	}

	resp := storageSettingsResponse{
		StorageDriver:    driver,
		LocalStoragePath: localPath,
		LocalBaseURL:     localBaseURL,
		COSKeyPrefix:     cosKeyPrefix,
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func upsertConfigValue(key, value, label, group string) error {
	cfgType := "text"
	sort := 0
	description := ""

	switch key {
	case "storage_driver":
		cfgType = "select"
		sort = 1
		description = "附件存储方式: local 或 cos"
	case "local_storage_path":
		sort = 2
		description = "本地磁盘保存附件的路径"
	case "local_base_url":
		sort = 3
		description = "前端访问本地附件的URL前缀"
	case "cos_key_prefix":
		sort = 4
		description = "上传到COS的对象路径前缀"
	}

	assign := models.Config{
		Value:       value,
		Label:       label,
		Type:        cfgType,
		Group:       group,
		Description: description,
		Sort:        sort,
		Status:      1,
	}
	return database.DB.Where(models.Config{Key: key}).
		Assign(assign).
		FirstOrCreate(&models.Config{}).Error
}
