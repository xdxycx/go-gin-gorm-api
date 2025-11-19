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
		AllowOrigins:     []string{"*"}, // 允许所有来源，实际应用中应限制
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// MaxAge: 12 * time.Hour,
	}))

	// 根路径健康检查
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to Go Gin Gorm API"})
	})

	// 2. 基础 API 路由分组
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

		// 动态服务注册路由
		dynamicRegister := v1.Group("/dynamic")
		{
			// POST 用于注册新的动态服务
			dynamicRegister.POST("/register", handlers.RegisterService)
		}

		// 3. 【核心】动态服务执行路由 (使用通配符捕获所有未匹配的路径)
		// 注意：这必须放在最后，作为兜底路由
		dynamicExecute := v1.Group("/dynamic")
		{
			log.Println("注册动态服务执行路由 /api/v1/dynamic/*")
			// 使用 *path 捕获 /api/v1/dynamic/ 之后的所有路径
			// 路由匹配时，:path 参数在 ExecuteService 中被忽略，而是使用 Request.URL.Path 提取
			dynamicExecute.Any("/*path", handlers.ExecuteService)
		}
	}

	return r
}
