package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/router"
)

func main() {
	// 1. 加载 .env 文件 (用于本地开发)
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	// 2. 初始化数据库
	config.InitDatabase()

	// 3. 初始化路由
	r := router.InitRouter()

	// 4. 运行服务
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默认端口
	}

	log.Printf("服务器正在端口 %s 上运行...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
