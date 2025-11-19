package handlers

import (
	"encoding/json" 
	"errors"        
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/models"
	"go-gin-gorm-api/app/utils"
	"gorm.io/gorm"
)

// RegisterService 处理动态服务注册请求。
// 逻辑不变，现在支持存储 ParamKeys 字段。
func RegisterService(c *gin.Context) {
	var service models.APIService
	if err := c.ShouldBindJSON(&service); err != nil {
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
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500,
			Message: "服务注册失败，可能是路径或名称已存在。",
			Data:    gin.H{"detail": result.Error.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Code:    0,
		Message: "服务注册成功",
		Data:    service,
	})
}

// ExecuteService 是动态 SQL 服务的核心执行逻辑，已修复参数顺序问题。
func ExecuteService(c *gin.Context) {
	reqMethod := c.Request.Method
	path := c.Param("path")
	
	if path == "" {
		c.JSON(http.StatusNotFound, utils.APIResponse{Code: 404, Message: "动态服务路径未指定"})
		return
	}

	var service models.APIService
	
	// 1. 根据 Method 和 Path 查找注册的服务
	err := config.DB.Where("method = ? AND path = ?", reqMethod, path).First(&service).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { 
			c.JSON(http.StatusNotFound, utils.APIResponse{Code: 404, Message: fmt.Sprintf("未找到方法为 %s, 路径为 %s 的动态服务配置", reqMethod, path)})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "查询服务配置失败", Data: gin.H{"detail": err.Error()}})
		return
	}

	// 2. 解析 ParamKeys 获取参数顺序 (修复缺陷的关键步骤)
	var paramKeys []string
	if service.ParamKeys != "" {
		if err := json.Unmarshal([]byte(service.ParamKeys), &paramKeys); err != nil {
			log.Printf("ParamKeys JSON 解析失败: %v", err)
			c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "服务配置错误：ParamKeys 格式无效"})
			return
		}
	}

	// 3. 收集请求参数以供 SQL 绑定 (严格按照 ParamKeys 的顺序)
	var args []interface{}
	
	if reqMethod == http.MethodGet {
		// GET 请求：从 URL 查询参数中获取参数
		queryParams := c.Request.URL.Query()
		for _, key := range paramKeys {
			values := queryParams[key]
			if len(values) == 0 {
				c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: fmt.Sprintf("GET 请求参数缺失: %s", key)})
				return
			}
			// 严格按 ParamKeys 顺序添加参数
			args = append(args, values[0]) 
		}
	} else if reqMethod == http.MethodPost || reqMethod == http.MethodPut || reqMethod == http.MethodDelete { 
		// POST/PUT/DELETE 请求：从 JSON body 中获取参数
		if len(paramKeys) > 0 {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "请求体解析失败或格式错误", Data: gin.H{"detail": err.Error()}})
				return
			}

			for _, key := range paramKeys {
				value, ok := body[key]
				if !ok {
					c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: fmt.Sprintf("请求参数缺失: %s", key)})
					return
				}
				// 严格按 ParamKeys 顺序添加参数
				args = append(args, value)
			}
		}
	}
	
	log.Printf("执行动态服务: Path=%s, Method=%s, SQL=%s, 参数=%v", path, reqMethod, service.SQL, args)

	// 4. 执行 SQL 并扫描结果
	var results []map[string]interface{}
	// 【核心安全】使用 Raw().Find() 确保参数 args 被正确绑定到 SQL 中的 '?' 占位符
	db := config.DB.Raw(service.SQL, args...)
	
	if db.Error != nil {
		log.Printf("SQL 执行失败: %v", db.Error)
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500,
			Message: "SQL 执行失败，请检查 SQL 语句和 ParamKeys 中参数数量、顺序是否匹配。", 
			Data:    gin.H{"detail": db.Error.Error()},
		})
		return
	}

	// Find() 扫描结果集
	if err := db.Find(&results).Error != nil {
		log.Printf("结果扫描失败: %v", err)
		
		// 检查是否是成功的非查询操作 (INSERT, UPDATE, DELETE)
		sqlUpper := strings.ToUpper(service.SQL)
		if db.RowsAffected > 0 && (strings.HasPrefix(sqlUpper, "INSERT") || strings.HasPrefix(sqlUpper, "UPDATE") || strings.HasPrefix(sqlUpper, "DELETE")) {
			// 如果是成功的非查询操作，返回受影响的行数
			c.JSON(http.StatusOK, utils.APIResponse{
				Code:    0,
				Message: fmt.Sprintf("SQL 执行成功，影响行数: %d", db.RowsAffected),
				Data:    gin.H{"rows_affected": db.RowsAffected},
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500, 
			Message: "结果处理失败。", 
			Data:    gin.H{"detail": err.Error()},
		})
		return
	}
	
	// 对于 SELECT 语句，返回查询结果
	c.JSON(http.StatusOK, utils.APIResponse{
		Code:    0,
		Message: "查询成功",
		Data:    results,
	})
}
