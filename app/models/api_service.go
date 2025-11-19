package models

import (
	"gorm.io/gorm"
)

// ApiService 定义了动态注册的接口服务
type ApiService struct {
	gorm.Model
	ServiceName string `gorm:"uniqueIndex;not null" json:"service_name"` // 服务唯一标识，如 "get_user_orders"
	Desc        string `json:"desc"`                                     // 服务描述
	SQL         string `gorm:"type:text;not null" json:"sql"`            // 自定义查询语句，使用 @param 作为占位符
	Method      string `gorm:"default:'GET'" json:"method"`              // HTTP 方法: GET/POST
}