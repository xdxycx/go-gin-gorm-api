package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-gin-gorm-api/app/models"
)

// DB 存储 GORM 数据库连接实例
var DB *gorm.DB

// InitDatabase 初始化数据库连接并自动迁移模型
func InitDatabase() {
	// 从环境变量中读取数据库配置
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbName := os.Getenv("MYSQL_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 开启日志记录，方便调试
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	log.Println("数据库连接成功！")

	// 自动迁移所有模型
	err = DB.AutoMigrate(
		&models.User{},
		&models.APIService{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成。")
}
