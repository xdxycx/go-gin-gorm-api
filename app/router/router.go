package router

import (
	"net/http"
	"go-gin-gorm-api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 1. 管理接口：用于注册 SQL 服务
	admin := r.Group("/admin")
	{
		admin.POST("/services", handlers.RegisterService) // 注册新接口
		// 还可以添加 GET /services 查看列表等
	}

	// 2. 动态服务接口：调用已注册的服务
	// 访问模式: /api/d/{service_name}
	dynamic := r.Group("/api/d")
	{
		// 使用 Any 允许 GET 或 POST
		dynamic.Any("/:service_name", handlers.InvokeService)
	}
    
    // 原有的 API
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", handlers.GetUserList)
    }

	return r
}