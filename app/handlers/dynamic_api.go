package handlers

import (
	"log"
	"net/http"
	"strings"
	"fmt" // 【新增】引入 fmt 包

	"github.com/gin-gonic/gin"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/models"
	"go-gin-gorm-api/app/utils"
	"gorm.io/gorm"
)

// RegisterService 处理动态服务注册请求。
func RegisterService(c *gin.Context) {
	var service models.APIService
	if err := c.ShouldBindJSON(&service); err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    400,
			Message: "请求参数错误或缺失: " + err.Error(),
		})
		return
	}

	// 确保路径以 / 开头
	if !strings.HasPrefix(service.Path, "/") {
		service.Path = "/" + service.Path
	}

	// 将 Method 转换为大写
	service.Method = strings.ToUpper(service.Method)

	// 尝试创建服务
	result := config.DB.Create(&service)
	if result.Error != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500,
			Message: "服务注册失败，可能是路径或名称已存在。",
			Data:    gin.H{"detail": result.Error.Error()},
		})
		return
	}

	// 【修改】使用统一的 APIResponse 结构
	c.JSON(http.StatusCreated, utils.APIResponse{
		Code:    0,
		Message: "服务注册成功",
		Data:    service,
	})
}

// ExecuteService 是动态 SQL 服务的核心执行逻辑。
// 它根据请求的 HTTP 方法和路径，查找对应的 APIService 记录并执行其 SQL。
func ExecuteService(c *gin.Context) {
	reqMethod := c.Request.Method
	// 【修改】从 Gin 的通配符路由参数中获取路径，路径包含前导斜杠，例如 "/report"
	path := c.Param("path")
	
	if path == "" {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusNotFound, utils.APIResponse{Code: 404, Message: "动态服务路径未指定"})
		return
	}

	var service models.APIService
	
	// 1. 根据 Method 和 Path 查找注册的服务
	err := config.DB.Where("method = ? AND path = ?", reqMethod, path).First(&service).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 【修改】使用统一的 APIResponse 结构，并提供更详细的错误信息
			c.JSON(http.StatusNotFound, utils.APIResponse{Code: 404, Message: fmt.Sprintf("未找到方法为 %s, 路径为 %s 的动态服务配置", reqMethod, path)})
			return
		}
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "查询服务配置失败", Data: gin.H{"detail": err.Error()}})
		return
	}

	// 2. 收集请求参数以供 SQL 绑定
	// 警告: 使用 GORM 的 Raw().Find() 进行参数绑定以防止 SQL 注入。
	// 用户在编写 SQL 时，必须使用问号 '?' 作为占位符，并且占位符数量和顺序需与传入的参数值数量和顺序一致。
	var args []interface{}
	
	if reqMethod == http.MethodGet {
		// GET 请求：从 URL 查询参数中获取所有参数
		// 注意: Go 语言中 map 遍历顺序不保证一致，但对于 URL Query 参数通常按键名排序。
		// 为了保证SQL执行顺序，用户必须确保SQL占位符的顺序与Query参数的顺序一致。
		for _, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				args = append(args, v[0]) // 仅使用第一个值
			}
		}
	} else if reqMethod == http.MethodPost || reqMethod == http.MethodPut {
		// POST/PUT 请求：从 JSON body 中获取所有参数
		var body map[string]interface{}
		// 尝试绑定 JSON body
		if err := c.ShouldBindJSON(&body); err == nil {
			// 遍历 body map，将所有值收集为参数
			// 警告：JSON 解析器的内部顺序不可控，如果 SQL 有多个占位符，必须确保 JSON 字段的值顺序与 SQL '?' 占位符顺序一致。
			for _, v := range body {
				args = append(args, v)
			}
		} else {
			log.Printf("POST/PUT 请求体绑定 JSON 失败: %v", err)
		}
	}
	
	log.Printf("执行动态服务: Path=%s, Method=%s, SQL=%s, 参数=%v", path, reqMethod, service.SQL, args) // 【修改】增加日志打印收集到的参数

	// 3. 执行 SQL 并扫描结果到 []map[string]interface{}
	var results []map[string]interface{}
	// 【核心安全】使用 Raw().Find() 确保参数 args 被正确绑定到 SQL 中的 '?' 占位符，防止 SQL 注入。
	db := config.DB.Raw(service.SQL, args...) // 【修改】使用 Raw(service.SQL, args...) 传入参数
	
	if db.Error != nil {
		log.Printf("SQL 执行失败: %v", db.Error)
		// 【修改】使用统一的 APIResponse 结构，并提醒检查 SQL 和参数
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500,
			Message: "SQL 执行失败，请检查 SQL 语句和参数数量是否匹配。", 
			Data:    gin.H{"detail": db.Error.Error()},
		})
		return
	}

	// Find() 会将结果集扫描到 results
	if err := db.Find(&results).Error; err != nil {
		log.Printf("结果扫描失败: %v", err)
		
		// 【新增】检查是否是成功的非查询操作 (INSERT, UPDATE, DELETE)
		sqlUpper := strings.ToUpper(service.SQL)
		if db.RowsAffected > 0 && (strings.HasPrefix(sqlUpper, "INSERT") || strings.HasPrefix(sqlUpper, "UPDATE") || strings.HasPrefix(sqlUpper, "DELETE")) {
			// 如果是成功的非查询操作，返回受影响的行数，而不是尝试将空结果扫描到 map
			c.JSON(http.StatusOK, utils.APIResponse{
				Code:    0,
				Message: fmt.Sprintf("SQL 执行成功，影响行数: %d", db.RowsAffected),
				Data:    gin.H{"rows_affected": db.RowsAffected},
			})
			return
		}
		
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500, 
			Message: "结果处理失败。", 
			Data:    gin.H{"detail": err.Error()},
		})
		return
	}
	
	// 对于 SELECT 语句，返回查询结果
	// 【修改】使用统一的 APIResponse 结构
	c.JSON(http.StatusOK, utils.APIResponse{
		Code:    0,
		Message: "查询成功",
		Data:    results,
	})
}
