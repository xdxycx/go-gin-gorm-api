package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/models"
	"go-gin-gorm-api/app/utils"
	"gorm.io/gorm"
)

// allowedQueryPrefixes 定义了允许通过动态服务执行的只读查询语句前缀。
// 仅允许 SELECT, WITH, EXPLAIN 和 DESCRIBE/DESC 等不会修改数据库状态的语句。
var allowedQueryPrefixes = []string{"SELECT", "WITH", "EXPLAIN", "DESCRIBE", "DESC "}

// isAllowedQuery 检查 SQL 语句是否以允许的只读前缀开始。
func isAllowedQuery(sql string) bool {
	sqlUpper := strings.ToUpper(strings.TrimSpace(sql))
	for _, prefix := range allowedQueryPrefixes {
		if strings.HasPrefix(sqlUpper, prefix) {
			return true
		}
	}
	return false
}


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
// 【安全修复】此函数现在只允许执行预定义的只读查询操作。非查询操作将被阻止并仅记录。
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

	// 【安全检查】使用辅助函数检查是否为允许的只读查询
	if !isAllowedQuery(service.SQL) {
		sqlUpper := strings.ToUpper(strings.TrimSpace(service.SQL))
		
		log.Printf("Security Alert: Blocked execution of write/unauthorized dynamic SQL. Path=%s, Method=%s, SQL=%s", path, reqMethod, service.SQL)

		// 返回成功状态码（HTTP 200），但使用非 0 的业务代码和警告消息，表示操作被安全策略拦截/跳过
		c.JSON(http.StatusOK, utils.APIResponse{
			Code:    1, // 使用非 0 状态码表示操作被安全策略拦截/跳过
			Message: fmt.Sprintf("安全限制: 动态服务只允许执行 %v 查询操作。非查询操作已被阻止。", allowedQueryPrefixes),
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
				// 忽略 EOF 错误，表示请求体为空，但这通常意味着参数缺失，后续检查会捕获
				if !errors.Is(err, errors.New("EOF")) { 
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
		case float64: // JSON 解析数字默认是 float64
			// 转换为字符串时，使用 -1 精度，以确保保留原始数字的所有有效位，避免科学计数法或精度丢失。
			strValue = strconv.FormatFloat(v, 'f', -1, 64) 
		case bool:
			strValue = strconv.FormatBool(v)
		default:
			// 兜底：尝试将其他类型转换为字符串
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
			// ParseBool 接受多种格式 (t, f, 1, 0, true, false)
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

	// 5. 执行 SQL 并扫描结果（带超时与行数限制），并写入审计表
	var results []map[string]interface{}

	// 从环境变量读取可配置项，提供默认值
	maxRows := 1000
	if v := os.Getenv("DYNAMIC_MAX_ROWS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxRows = n
		}
	}

	timeoutSec := 5
	if v := os.Getenv("DYNAMIC_QUERY_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSec = n
		}
	}

	queryTimeout := time.Duration(timeoutSec) * time.Second

	ctx, cancel := context.WithTimeout(c.Request.Context(), queryTimeout)
	defer cancel()

	start := time.Now()

	// 使用带上下文的 DB 执行查询
	db := config.DB.WithContext(ctx).Raw(service.SQL, args...)
	if db.Error != nil {
		log.Printf("SQL 执行失败: %v", db.Error)
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500,
			Message: "SQL 执行失败，请检查 SQL 语句、ParamKeys 和 ParamTypes 配置。",
			Data:    gin.H{"detail": db.Error.Error()},
		})
		return
	}

	if err := db.Find(&results).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("SQL 执行超时: Path=%s, Method=%s, SQL=%s, err=%v", path, reqMethod, service.SQL, err)
			c.JSON(http.StatusOK, utils.APIResponse{Code: 2, Message: "查询超时，已取消执行"})
			return
		}

		log.Printf("结果扫描失败: %v", err)
		c.JSON(http.StatusInternalServerError, utils.APIResponse{
			Code:    500,
			Message: "结果处理失败。",
			Data:    gin.H{"detail": err.Error()},
		})
		return
	}

	duration := time.Since(start)
	truncated := false
	rows := len(results)
	if rows > maxRows {
		results = results[:maxRows]
		truncated = true
	}

	// 持久化审计记录
	argsBytes, _ := json.Marshal(args)
	audit := models.Audit{
		Path:       path,
		Method:     reqMethod,
		ClientIP:   c.ClientIP(),
		SQL:        service.SQL,
		Args:       string(argsBytes),
		DurationMs: duration.Milliseconds(),
		Rows:       rows,
		Truncated:  truncated,
	}
	if err := config.DB.Create(&audit).Error; err != nil {
		log.Printf("AUDIT WRITE FAILED: %v", err)
	}

	// 返回查询结果（包含截断提示）
	resp := utils.APIResponse{Code: 0, Message: "查询成功", Data: results}
	if truncated {
		resp.Message = "查询成功（结果已被限制为最大行数）"
		resp.Data = gin.H{"rows_returned": len(results), "truncated": true, "data": results}
	}
	c.JSON(http.StatusOK, resp)
}
