package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"haodun_manage/backend/config"
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
)

const (
	StorageDriverLocal = "local"
	StorageDriverCOS   = "cos"
)

var (
	safeNameRegex = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
)

func GetStorageDriver() string {
	driver := StorageDriverLocal
	if config.AppConfig != nil {
		driver = NormalizeStorageDriver(config.AppConfig.StorageDriver)
	}

	if database.DB != nil {
		var cfg models.Config
		if err := database.DB.Select("value").Where("`key` = ? AND status = 1", "storage_driver").First(&cfg).Error; err == nil {
			if strings.TrimSpace(cfg.Value) != "" {
				driver = NormalizeStorageDriver(cfg.Value)
			}
		}
	}

	return driver
}

func NormalizeStorageDriver(driver string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case StorageDriverCOS:
		return StorageDriverCOS
	default:
		return StorageDriverLocal
	}
}

func BuildAttachmentKey(orderID uint64, fileName string) string {
	prefix := "orders"
	if config.AppConfig != nil {
		if v := strings.Trim(config.AppConfig.COSKeyPrefix, "/"); v != "" {
			prefix = v
		}
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	name := strings.TrimSuffix(fileName, ext)
	name = sanitizeFileName(name)
	if name == "" {
		name = randomSuffix()
	}

	timePart := time.Now().UTC().Format("20060102T150405")
	objectName := fmt.Sprintf("%s_%s%s", timePart, name, ext)

	return path.Join(prefix, fmt.Sprintf("%d", orderID), objectName)
}

func BuildMaterialObjectKey(materialCode string, fileName string) string {
	if materialCode == "" {
		materialCode = randomSuffix()
	}
	sanitizedCode := sanitizeFileName(strings.ToLower(materialCode))
	if sanitizedCode == "" {
		sanitizedCode = randomSuffix()
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	name := strings.TrimSuffix(fileName, ext)
	name = sanitizeFileName(name)
	if name == "" {
		name = randomSuffix()
	}

	timePart := time.Now().UTC().Format("20060102T150405")
	objectName := fmt.Sprintf("%s_%s%s", timePart, name, ext)

	return path.Join("materials", sanitizedCode, objectName)
}

func UploadAttachment(ctx context.Context, storage string, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	switch NormalizeStorageDriver(storage) {
	case StorageDriverCOS:
		return UploadToCOS(ctx, objectKey, reader, size, contentType)
	case StorageDriverLocal:
		return uploadToLocal(objectKey, reader)
	default:
		return "", fmt.Errorf("unsupported storage driver: %s", storage)
	}
}

func DeleteAttachment(ctx context.Context, storage string, objectKey string) error {
	switch NormalizeStorageDriver(storage) {
	case StorageDriverCOS:
		return DeleteFromCOS(ctx, objectKey)
	case StorageDriverLocal:
		return deleteFromLocal(objectKey)
	default:
		return fmt.Errorf("unsupported storage driver: %s", storage)
	}
}

func BuildAttachmentURL(storage, objectKey string) (string, error) {
	switch NormalizeStorageDriver(storage) {
	case StorageDriverCOS:
		return BuildCOSObjectURL(objectKey)
	case StorageDriverLocal:
		return buildLocalURL(objectKey), nil
	default:
		return "", fmt.Errorf("unsupported storage driver: %s", storage)
	}
}

func GenerateAttachmentDownloadURL(ctx context.Context, storage, objectKey string) (string, error) {
	switch NormalizeStorageDriver(storage) {
	case StorageDriverCOS:
		return GenerateCOSPresignedURL(ctx, objectKey, "", 0)
	case StorageDriverLocal:
		url := buildLocalURL(objectKey)
		if url == "" {
			return "", nil
		}
		return url, nil
	default:
		return "", fmt.Errorf("unsupported storage driver: %s", storage)
	}
}

func GetLocalFilePath(objectKey string) (string, error) {
	return getLocalFullPath(objectKey)
}

func uploadToLocal(objectKey string, reader io.Reader) (string, error) {
	fullPath, err := getLocalFullPath(objectKey)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}

	tmpPath := fullPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	if err := os.Rename(tmpPath, fullPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	return buildLocalURL(objectKey), nil
}

func deleteFromLocal(objectKey string) error {
	fullPath, err := getLocalFullPath(objectKey)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func getLocalFullPath(objectKey string) (string, error) {
	base := "./uploads"
	if config.AppConfig != nil && strings.TrimSpace(config.AppConfig.LocalStoragePath) != "" {
		base = config.AppConfig.LocalStoragePath
	}
	base = filepath.Clean(base)

	cleanObject := strings.TrimLeft(objectKey, "/")
	cleanObject = filepath.FromSlash(cleanObject)
	fullPath := filepath.Join(base, cleanObject)

	baseWithSep := base + string(os.PathSeparator)
	if fullPath != base && !strings.HasPrefix(fullPath, baseWithSep) {
		return "", errors.New("invalid object key path")
	}

	return fullPath, nil
}

func buildLocalURL(objectKey string) string {
	if config.AppConfig == nil {
		return ""
	}
	baseURL := strings.TrimSpace(config.AppConfig.LocalBaseURL)
	if baseURL == "" {
		return ""
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return baseURL + "/" + strings.TrimLeft(path.Clean("/"+strings.ReplaceAll(objectKey, "\\", "/")), "/")
}

func sanitizeFileName(name string) string {
	name = safeNameRegex.ReplaceAllString(strings.ToLower(name), "_")
	name = strings.Trim(name, "_")
	return name
}

func randomSuffix() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
