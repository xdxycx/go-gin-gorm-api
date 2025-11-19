package config

import (
	"fmt"
	"log"
	
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-gin-gorm-api/app/models"
)

// DB 存储 GORM 数据库连接实例
var DB *gorm.DB

// 【新增】DBConfig 接口定义了数据库连接所需的配置参数。
// 通过接口隔离，避免 config 包直接依赖 main 包的 Config 结构。
type DBConfig interface {
	GetDBUser() string
	GetDBPass() string
	GetDBHost() string
	GetDBPort() string
	GetDBName() string
}

// 【修改】InitDatabase 初始化数据库连接并自动迁移模型
func InitDatabase(cfg DBConfig) { // 接收 DBConfig 接口
	dbUser := cfg.GetDBUser()
	dbPass := cfg.GetDBPass()
	dbHost := cfg.GetDBHost()
	dbPort := cfg.GetDBPort()
	dbName := cfg.GetDBName()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 开启日志记录，方便调试和问题追踪
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
