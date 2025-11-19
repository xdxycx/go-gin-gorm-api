package router

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go-gin-gorm-api/app/handlers"
)

// InitRouter 初始化 Gin 路由配置
func InitRouter() *gin.Engine {
	// 设置 Gin 模式 (release 模式可以提高性能)
	gin.SetMode(gin.DebugMode)
	
	r := gin.Default()

	// 1. 配置 CORS 跨域
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 生产环境中应限制为特定的域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 根路径健康检查
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to Go Gin Gorm API"})
	})

	// 2. API 路由分组
	v1 := r.Group("/api/v1")
	{
		// 用户管理 (基础示例)
		userRoutes := v1.Group("/users")
		{
			userRoutes.POST("", handlers.CreateUser)
			userRoutes.GET("", handlers.GetUsers)
			userRoutes.GET("/:id", handlers.GetUserByID)
			userRoutes.PUT("/:id", handlers.UpdateUser)
			userRoutes.DELETE("/:id", handlers.DeleteUser)
		}

		// 3. 动态服务注册与执行
		dynamic := v1.Group("/dynamic")
		{
			// POST /api/v1/dynamic/register 用于注册新的动态服务
			dynamic.POST("/register", handlers.RegisterService)
			
			// 【修改】使用 Any() 和通配符 *path 捕获 /api/v1/dynamic/ 之后的任何路径和任何 HTTP 方法
			// 匹配的路径将传递给 ExecuteService，ExecuteService 根据 Path 和 Method 查找对应的 SQL 服务
			log.Println("注册动态服务执行路由 /api/v1/dynamic/*path") // 【修改】增加日志
			dynamic.Any("/*path", handlers.ExecuteService) // 【修改】路由调整为 Any("/*path")
		}
	}

	return r
}
