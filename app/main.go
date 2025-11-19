package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/router"
)

// 【新增】Config 应用程序的配置结构体，包含数据库和应用端口信息
type Config struct {
	DBUser  string
	DBPass  string
	DBHost  string
	DBPort  string
	DBName  string
	AppPort int
}

// 【新增】实现 config.DBConfig 接口方法，用于解耦
func (c *Config) GetDBUser() string { return c.DBUser }
func (c *Config) GetDBPass() string { return c.DBPass }
func (c *Config) GetDBHost() string { return c.DBHost }
func (c *Config) GetDBPort() string { return c.DBPort }
func (c *Config) GetDBName() string { return c.DBName }


// 【修改】loadConfig 从环境变量加载配置
func loadConfig() *Config {
	// 1. 加载 .env 文件 (用于本地开发)
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	appPort := 8080
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		appPort = p
	}

	return &Config{
		DBUser:  os.Getenv("MYSQL_USER"),
		DBPass:  os.Getenv("MYSQL_PASSWORD"),
		DBHost:  os.Getenv("MYSQL_HOST"),
		DBPort:  os.Getenv("MYSQL_PORT"),
		DBName:  os.Getenv("MYSQL_DATABASE"),
		AppPort: appPort,
	}
}

func main() {
	// 【修改】1. 加载配置
	cfg := loadConfig()

	// 【修改】2. 初始化数据库，将配置结构体传递给 InitDatabase
	config.InitDatabase(cfg)

	// 3. 初始化路由
	r := router.InitRouter()

	// 【修改】4. 运行服务
	log.Printf("服务器正在端口 %d 上运行...", cfg.AppPort)
	if err := r.Run(":" + strconv.Itoa(cfg.AppPort)); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
