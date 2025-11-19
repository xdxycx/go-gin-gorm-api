package models

import (
	"time"

	"gorm.io/gorm"
)

// APIService 定义了动态 API 服务的注册模型
// 它存储了 API 服务的元数据，包括要执行的 SQL 语句。
type APIService struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Name 是服务的友好名称
	Name string `gorm:"unique;not null" json:"name" binding:"required"`
	
	// Method 是 HTTP 方法 (例如: GET, POST)
	Method string `gorm:"not null" json:"method" binding:"required,oneof=GET POST"`
	
	// Path 是服务的 URL 路径 (例如: /api/v1/report)
	Path string `gorm:"unique;not null" json:"path" binding:"required"`
	
	// SQL 是要执行的原始 SQL 语句
	// 注意: 实际项目中应加入参数校验和安全措施防止 SQL 注入
	SQL string `gorm:"not null" json:"sql" binding:"required"`
}

// TableName 指定表名为 'api_services'
func (APIService) TableName() string {
	return "api_services"
}
