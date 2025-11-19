package config

import (
	"fmt"
	"log"
	"os"
	"time"
        "go-gin-gorm-api/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	// 从环境变量读取配置
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)

	var err error

	// 循环重试连接，等待 MySQL 启动
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("成功连接到 MySQL 数据库!")
			return
		}
		log.Printf("连接数据库失败，重试中... (%d/5): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("无法连接到数据库: %v", err)
}

// Migrate 和 Seed 放在这里，便于初始化
func RunMigrations() {
	if err := DB.AutoMigrate(&models.User{}, &models.ApiService{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移成功，users表已准备就绪")
}

func SeedDB() {
	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		DB.Create(&models.User{Name: "Alice", Email: "alice@example.com"})
		DB.Create(&models.User{Name: "Bob", Email: "bob@example.com"})
		log.Println("已创建测试用户数据")
	}
}
