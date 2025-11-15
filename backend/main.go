package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"haodun_manage/backend/config"
	"haodun_manage/backend/database"
	"haodun_manage/backend/router"
	"haodun_manage/backend/utils"

	"github.com/gin-gonic/gin"
)

// loadEnv 加载 .env 文件到环境变量
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("无法打开 .env 文件: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// 解析 key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// 移除引号
		value = strings.Trim(value, `"'`)
		
		// 设置环境变量
		os.Setenv(key, value)
	}
	
	return scanner.Err()
}

func main() {
	// 加载 .env 文件
	if err := loadEnv(".env"); err != nil {
		fmt.Printf("加载 .env 文件失败: %v\n", err)
		fmt.Println("请确保 .env 文件存在，或使用环境变量")
	}

	// 初始化配置
	config.InitConfig()

	// 初始化数据库
	database.InitDB()

	// 初始化Redis
	database.InitRedis()

	// 初始化路由
	r := gin.Default()

	// 配置CORS
	r.Use(utils.CORSMiddleware())

	// 注册路由
	router.SetupRoutes(r)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("服务器启动在端口 %s\n", port)
	r.Run(":" + port)
}
