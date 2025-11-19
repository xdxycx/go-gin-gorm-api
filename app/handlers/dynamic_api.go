package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/models"
	"gorm.io/gorm"
)

// RegisterService 处理动态服务注册请求。
// 接收一个包含 Name, Method, Path 和 SQL 的 JSON 对象，并保存到 api_services 表。
func RegisterService(c *gin.Context) {
	var service models.APIService
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "detail": err.Error()})
		return
	}

	// 确保路径以 / 开头
	if !strings.HasPrefix(service.Path, "/") {
		service.Path = "/" + service.Path
	}

	// 将 Method 转换为大写，方便查找
	service.Method = strings.ToUpper(service.Method)

	// 尝试创建服务
	result := config.DB.Create(&service)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务注册失败", "detail": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "服务注册成功", "service": service})
}

// ExecuteService 是动态 SQL 服务的核心执行逻辑。
// 它根据请求的 HTTP 方法和路径，查找对应的 APIService 记录并执行其 SQL。
func ExecuteService(c *gin.Context) {
	reqMethod := c.Request.Method
	// 移除路由前缀 /api/v1/dynamic 后的路径部分
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1/dynamic")
	
	// 如果路径为空，则返回错误或默认首页
	if path == "" || path == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "动态服务路径不能为空"})
		return
	}

	var service models.APIService
	
	// 1. 根据 Method 和 Path 查找注册的服务
	err := config.DB.Where("method = ? AND path = ?", reqMethod, path).First(&service).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到对应的动态服务配置"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询服务配置失败", "detail": err.Error()})
		return
	}

	// 2. 解析请求参数以供 SQL 使用 (简化处理)
	// **警告**: 在实际生产环境中，必须实现安全的 SQL 参数绑定，防止注入。
	// 此处仅作为演示，我们只获取一个名为 `param` 的查询参数或 Body 参数。
	var args []interface{}
	var rawSQL = service.SQL
	
	// 尝试从查询参数或 JSON body 中获取参数
	if reqMethod == http.MethodGet {
		// GET 请求：从 URL 查询参数中获取
		param := c.Query("param")
		if param != "" {
			args = append(args, param)
		}
	} else if reqMethod == http.MethodPost {
		// POST 请求：尝试从 JSON body 中获取，这里只演示获取一个名为 `param` 的字段
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if param, ok := body["param"]; ok {
				args = append(args, param)
			}
		}
	}

	log.Printf("执行 SQL: %s, 参数: %v", rawSQL, args)

	// 3. 执行 SQL 并扫描结果到 []map[string]interface{}
	var results []map[string]interface{}
	// Gorm 的 Raw().Scan() 方法可以执行任意 SQL 并将结果映射到结构体或 map
	db := config.DB.Raw(rawSQL, args...)
	
	// 处理 SQL 执行错误
	if db.Error != nil {
		log.Printf("SQL 执行失败: %v", db.Error)
		// 检查是否是 SQL 错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SQL 执行失败", "detail": db.Error.Error()})
		return
	}

	// 使用 Find() 代替 Scan()，Find() 在 Gorm 内部会处理 Raw() 的结果集
	if err := db.Find(&results).Error; err != nil {
		log.Printf("结果扫描失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "结果处理失败", "detail": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, []interface{}{}) // 返回空数组
		return
	}

	// 4. 返回 JSON 结果
	c.JSON(http.StatusOK, results)
}
