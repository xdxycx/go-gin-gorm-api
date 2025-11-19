package main

import (
	"log"
	"os"
	
	"go-gin-gorm-api/config"
	"go-gin-gorm-api/router"
)

func main() {
	// 1. 初始化数据库连接
	config.InitDB()
	
	// 2. 运行数据库迁移和填充数据
	config.RunMigrations()
	config.SeedDB()

	// 3. 配置并启动路由
	r := router.SetupRouter()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Gin 应用程序启动中，监听端口: %s", port)
	
	// 启动 Web 服务器
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
