package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	RedisAddr        string
	RedisPwd         string
	JWTSecret        string
	StorageDriver    string
	LocalStoragePath string
	LocalBaseURL     string
	COSSecretID      string
	COSSecretKey     string
	COSBucket        string
	COSRegion        string
	COSBaseURL       string
	COSKeyPrefix     string
	COSURLExpires    int
}

var AppConfig *Config

func InitConfig() {
	// 尝试加载.env文件（如果文件不存在也不会报错）
	if err := godotenv.Load(); err != nil {
		// .env文件不存在时使用默认值或环境变量
		log.Println("未找到.env文件，将使用环境变量或默认值")
	}

	AppConfig = &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "3306"),
		DBUser:           getEnv("DB_USER", "root"),
		DBPassword:       getEnv("DB_PASSWORD", "password"),
		DBName:           getEnv("DB_NAME", "haodun"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPwd:         getEnv("REDIS_PWD", ""),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		StorageDriver:    getEnv("STORAGE_DRIVER", "local"),
		LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "./uploads"),
		LocalBaseURL:     getEnv("LOCAL_BASE_URL", ""),
		COSSecretID:      getEnv("COS_SECRET_ID", ""),
		COSSecretKey:     getEnv("COS_SECRET_KEY", ""),
		COSBucket:        getEnv("COS_BUCKET", ""),
		COSRegion:        getEnv("COS_REGION", ""),
		COSBaseURL:       getEnv("COS_BASE_URL", ""),
		COSKeyPrefix:     getEnv("COS_KEY_PREFIX", "orders"),
		COSURLExpires:    getEnvAsInt("COS_URL_EXPIRES", 3600),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}
