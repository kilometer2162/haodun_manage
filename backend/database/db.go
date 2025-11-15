package database

import (
	"fmt"
	"strings"

	"haodun_manage/backend/config"
	"haodun_manage/backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect database: %v", err))
	}

	// 自动迁移
	DB.AutoMigrate(
		&models.Department{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Resource{},
		&models.Log{},
		&models.IPAccess{},
		&models.DictType{},
		&models.DictItem{},
		&models.Config{},
		&models.Notice{},
		&models.NoticeRead{},
		&models.OrderInfo{},
		&models.OrderAttachment{},
		&models.MaterialFolder{},
		&models.MaterialAsset{},
	)

	// 初始化默认数据
	initDefaultData()
	// 从系统参数加载动态配置
	loadStorageSettingsFromDB()
}

func initDefaultData() {
	// 创建默认部门
	defaultDept := models.Department{}
	DB.Where(models.Department{Name: "管理部门"}).
		Attrs(models.Department{
			Description: "管理部门",
			Status:      1,
			Sort:        0,
		}).
		FirstOrCreate(&defaultDept)

	// 创建默认管理员角色
	adminRole := models.Role{}
	DB.Where(models.Role{Name: "admin"}).
		Attrs(models.Role{Description: "系统管理员"}).
		FirstOrCreate(&adminRole)

	// 创建默认管理员用户
	adminUser := models.User{}
	DB.Where(models.User{Username: "admin"}).
		Attrs(models.User{
			Password:     hashPassword("admin123"),
			Email:        "admin@example.com",
			RoleID:       adminRole.ID,
			Status:       1,
			DepartmentID: defaultDept.ID,
			EmployeeType: "internal",
		}).
		FirstOrCreate(&adminUser)

	// 创建默认站点名称配置
	siteNameConfig := models.Config{}
	DB.Where(models.Config{Key: "site_name"}).
		Attrs(models.Config{
			Value:       "管理系统",
			Label:       "站点名称",
			Type:        "text",
			Group:       "system",
			Description: "系统展示名称",
			Sort:        1,
			Status:      1,
		}).
		FirstOrCreate(&siteNameConfig)

	// 默认地址配置
	DB.Where(models.Config{Key: "default_address"}).
		Attrs(models.Config{
			Value:       "默认地址待更新",
			Label:       "默认地址",
			Type:        "text",
			Group:       "order",
			Description: "当订单未填写收件地址时的默认地址",
			Sort:        5,
			Status:      1,
		}).
		FirstOrCreate(&models.Config{})

	// 存储配置
	DB.Where(models.Config{Key: "storage_driver"}).
		Attrs(models.Config{
			Value:       config.AppConfig.StorageDriver,
			Label:       "存储驱动",
			Type:        "select",
			Group:       "storage",
			Description: "附件存储方式: local 或 cos",
			Sort:        1,
			Status:      1,
		}).
		FirstOrCreate(&models.Config{})

	defaultMaterialFolder := models.MaterialFolder{}
	DB.Where(models.MaterialFolder{Name: "默认文件夹"}).
		Attrs(models.MaterialFolder{Path: "默认文件夹"}).
		FirstOrCreate(&defaultMaterialFolder)

	DB.Where(models.Config{Key: "local_storage_path"}).
		Attrs(models.Config{
			Value:       config.AppConfig.LocalStoragePath,
			Label:       "本地存储路径",
			Type:        "text",
			Group:       "storage",
			Description: "本地磁盘保存附件的路径",
			Sort:        2,
			Status:      1,
		}).
		FirstOrCreate(&models.Config{})

	DB.Where(models.Config{Key: "local_base_url"}).
		Attrs(models.Config{
			Value:       config.AppConfig.LocalBaseURL,
			Label:       "本地访问URL前缀",
			Type:        "text",
			Group:       "storage",
			Description: "前端访问本地附件的URL前缀",
			Sort:        3,
			Status:      1,
		}).
		FirstOrCreate(&models.Config{})

	DB.Where(models.Config{Key: "cos_key_prefix"}).
		Attrs(models.Config{
			Value:       config.AppConfig.COSKeyPrefix,
			Label:       "COS对象前缀",
			Type:        "text",
			Group:       "storage",
			Description: "上传到COS的对象路径前缀",
			Sort:        4,
			Status:      1,
		}).
		FirstOrCreate(&models.Config{})
}

func loadStorageSettingsFromDB() {
	if config.AppConfig == nil {
		return
	}

	type kv struct {
		Key   string
		Value string
	}

	var items []kv
	if err := DB.Model(&models.Config{}).
		Select("`key`, `value`").
		Where("`key` IN ?", []string{
			"storage_driver",
			"local_storage_path",
			"local_base_url",
			"cos_key_prefix",
		}).Scan(&items).Error; err != nil {
		return
	}

	values := make(map[string]string, len(items))
	for _, item := range items {
		values[item.Key] = item.Value
	}

	if v := values["storage_driver"]; v != "" {
		config.AppConfig.StorageDriver = normalizeStorageDriver(v)
	}
	if v := values["local_storage_path"]; v != "" {
		config.AppConfig.LocalStoragePath = v
	}
	if v := values["local_base_url"]; v != "" {
		config.AppConfig.LocalBaseURL = v
	}
	if v := values["cos_key_prefix"]; v != "" {
		config.AppConfig.COSKeyPrefix = v
	}
}

func normalizeStorageDriver(driver string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	if driver == "cos" {
		return "cos"
	}
	return "local"
}

func hashPassword(password string) string {
	// 使用bcrypt进行密码哈希
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return password
	}
	return string(bytes)
}
