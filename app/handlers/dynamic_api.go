package handlers

import (
	"encoding/json" 
	"errors"        
	"fmt"
	"log"
	"net/http"
	"strconv" 
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/models"
	"go-gin-gorm-api/app/utils"
	"gorm.io/gorm"
)

// RegisterService 处理动态服务注册请求。
// 此函数现在接收并存储 ParamTypes 字段，并检查 ParamKeys 与 ParamTypes 数量的一致性。
func RegisterService(c *gin.Context) {
	var service models.APIService
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    400,
			Message: "请求参数错误或缺失: " + err.Error(),
		})
		return
	}
	
	// 检查 ParamKeys 和 ParamTypes 的数量是否一致
	var paramKeys []string
	var paramTypes []string
	
	if service.ParamKeys != "" && json.Unmarshal([]byte(service.ParamKeys), &paramKeys) != nil {
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "ParamKeys 格式错误 (非 JSON 数组)"})
		return
	}
	if service.ParamTypes != "" && json.Unmarshal([]byte(service.ParamTypes), &paramTypes) != nil {
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "ParamTypes 格式错误 (非 JSON 数组)"})
		return
	}
	
	if len(paramKeys) != len(paramTypes) {
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "ParamKeys 和 ParamTypes 数量不匹配"})
		return
	}

	if !strings.HasPrefix(service.Path, "/") {
		service.Path = "/" + service.Path
	}

	service.Method = strings.ToUpper(service.Method)

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

// ExecuteService 是动态 SQL 服务的核心执行逻辑，已实现强制类型转换。
// 【安全修复】此函数现在只允许执行 SELECT 查询操作。非 SELECT 操作将被阻止并仅记录。
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

	// 2. 解析 ParamKeys 和 ParamTypes 获取参数顺序和类型
	var paramKeys []string
	var paramTypes []string 
	
	if service.ParamKeys != "" {
		if err := json.Unmarshal([]byte(service.ParamKeys), &paramKeys); err != nil {
			c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "服务配置错误：ParamKeys 格式无效"})
			return
		}
	}
	if service.ParamTypes != "" {
		if err := json.Unmarshal([]byte(service.ParamTypes), &paramTypes); err != nil {
			c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "服务配置错误：ParamTypes 格式无效"})
			return
		}
	}

	if len(paramKeys) != len(paramTypes) {
		c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "服务配置错误：ParamKeys 和 ParamTypes 数量不匹配"})
		return
	}

	// 【修改】安全检查：只允许 SELECT 查询操作。非 SELECT 操作将被阻止并仅记录。
	sqlUpper := strings.ToUpper(strings.TrimSpace(service.SQL))
	if !strings.HasPrefix(sqlUpper, "SELECT") {
		log.Printf("Security Alert: Blocked execution of non-SELECT dynamic SQL. Path=%s, Method=%s, SQL=%s", path, reqMethod, service.SQL)

		// 返回成功状态码（HTTP 200），但使用非 0 的业务代码和警告消息，表示操作被安全策略拦截/跳过
		c.JSON(http.StatusOK, utils.APIResponse{
			Code:    1, // 使用非 0 状态码表示操作被安全策略拦截/跳过
			Message: "安全限制: 动态服务只允许执行 SELECT 查询操作。非查询操作已被阻止。",
			Data:    gin.H{"sql_statement_type": strings.Split(sqlUpper, " ")[0]},
		})
		return
	}

	// 3. 收集原始请求参数
	rawParams := make(map[string]interface{})
	
	if reqMethod == http.MethodGet {
		// GET 请求：从 URL 查询参数中获取所有参数 (都是字符串)
		queryParams := c.Request.URL.Query()
		for key := range queryParams {
			if len(queryParams[key]) > 0 {
				rawParams[key] = queryParams[key][0] // 仅使用第一个值
			}
		}
	} else if reqMethod == http.MethodPost || reqMethod == http.MethodPut || reqMethod == http.MethodDelete { 
		// POST/PUT/DELETE 请求：从 JSON body 中获取参数
		if len(paramKeys) > 0 {
			if err := c.ShouldBindJSON(&rawParams); err != nil {
				// 如果绑定失败，可能是 body 为空，或格式错误
				if !errors.Is(err, errors.New("EOF")) { // 忽略 EOF 错误，表示可能没有 body
					c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "请求体解析失败或格式错误", Data: gin.H{"detail": err.Error()}})
					return
				}
			}
		}
	}

	// 4. 严格按照 ParamKeys 和 ParamTypes 顺序进行类型转换和参数收集
	var args []interface{}
	for i, key := range paramKeys {
		expectedType := strings.ToLower(paramTypes[i])
		rawValue, ok := rawParams[key]
		
		if !ok {
			c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: fmt.Sprintf("请求参数缺失: %s", key)})
			return
		}

		var convertedValue interface{}
		var err error
		
		// 统一将原始值转换为字符串以便使用 strconv 进行精确转换
		var strValue string
		switch v := rawValue.(type) {
		case string:
			strValue = v
		case float64: // JSON 中的数字默认为 float64
			// 使用 'f' 格式和 -1 精度以避免科学计数法，并保持原始数字的全部精度
			strValue = strconv.FormatFloat(v, 'f', -1, 64) 
		case bool:
			strValue = strconv.FormatBool(v)
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		
		// 执行类型转换
		switch expectedType {
		case "int", "int64":
			var v int64
			v, err = strconv.ParseInt(strValue, 10, 64)
			convertedValue = v
		case "float", "float64":
			var v float64
			v, err = strconv.ParseFloat(strValue, 64)
			convertedValue = v
		case "bool":
			var v bool
			// ParseBool 接受 1, 0, t, f, T, F, true, false, TRUE, FALSE
			v, err = strconv.ParseBool(strValue)
			convertedValue = v
		case "string":
			convertedValue = strValue
		default:
			// 如果类型未指定或未知，默认使用字符串，并记录警告
			log.Printf("Warning: Unknown type '%s' specified for key '%s'. Defaulting to string.", expectedType, key)
			convertedValue = strValue
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: fmt.Sprintf("参数 '%s' 无法转换为预期类型 '%s'", key, expectedType), Data: gin.H{"error": err.Error()}})
			return
		}

		args = append(args, convertedValue)
	}
	
	log.Printf("执行动态服务: Path=%s, Method=%s, SQL=%s, 参数=%v", path, reqMethod, service.SQL, args)

	// 5. 执行 SQL 并扫描结果
	var results []map[string]interface{}
	// GORM Raw() 方法将确保 args 列表中的参数按顺序绑定到 SQL 语句中的 '?' 占位符
	db := config.DB.Raw(service.SQL, args...) 
	
	if db.Error != nil {
		log.Printf("SQL 执行失败: %v", db.Error)
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500,
			Message: "SQL 执行失败，请检查 SQL 语句、ParamKeys 和 ParamTypes 配置。", 
			Data:    gin.H{"detail": db.Error.Error()},
		})
		return
	}

	// Find() 扫描结果集
	if err := db.Find(&results).Error != nil {
		log.Printf("结果扫描失败: %v", err)
		
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
