package handlers

import (
	"net/http"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"
)

// RegisterService 注册一个新的查询服务
// POST /admin/services
func RegisterService(c *gin.Context) {
	var service models.ApiService
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存到数据库
	if err := config.DB.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register service: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service registered successfully", 
		"link": "/api/d/" + service.ServiceName,
	})
}

// InvokeService 执行已注册的动态服务
// ANY /api/d/:service_name
func InvokeService(c *gin.Context) {
	serviceName := c.Param("service_name")

	// 1. 查找服务定义
	var service models.ApiService
	if err := config.DB.Where("service_name = ?", serviceName).First(&service).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// 2. 获取请求参数 (支持 Query String 和 JSON Body)
	params := make(map[string]interface{})
	// 绑定 Query 参数 (如 ?id=1)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	// 绑定 Body 参数 (如果是 POST/PUT)
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		var bodyParams map[string]interface{}
		if err := c.ShouldBindJSON(&bodyParams); err == nil {
			for k, v := range bodyParams {
				params[k] = v
			}
		}
	}

	// 3. 执行 Raw SQL
	// 我们使用 map[string]interface{} 来接收结果，因为返回的列是动态的
	var results []map[string]interface{}
	
	// 使用 GORM 的 Named 参数功能 (支持 @name 语法)
	// 注意：这需要 SQL 中使用 @key 占位符，例如: SELECT * FROM users WHERE id = @id
	tx := config.DB.Raw(service.SQL, params).Scan(&results)

	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"sql_error": tx.Error.Error()})
		return
	}

	// 4. 返回 JSON
	c.JSON(http.StatusOK, gin.H{
		"service": service.ServiceName,
		"count":   len(results),
		"data":    results,
	})
}