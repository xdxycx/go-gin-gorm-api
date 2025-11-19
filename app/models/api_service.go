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
	
	// Method 是 HTTP 方法 (例如: GET, POST, PUT)
	Method string `gorm:"not null" json:"method" binding:"required,oneof=GET POST PUT DELETE"`
	
	// Path 是服务的 URL 路径 (例如: /api/v1/dynamic/report)
	Path string `gorm:"unique;not null" json:"path" binding:"required"`
	
	// SQL 是要执行的原始 SQL 语句
	// 注意: 必须使用 ? 作为参数占位符，并在请求中传递参数值。
	SQL string `gorm:"not null" json:"sql" binding:"required"`
	
	// ParamKeys 是一个 JSON 数组字符串，定义了 SQL 占位符对应参数的 key 及其顺序。
	// 示例: '["user_id", "username"]'。顺序必须与 SQL 中的 ? 占位符顺序一致。
	ParamKeys string `gorm:"type:text" json:"param_keys"`

	// 【新增】ParamTypes 是一个 JSON 数组字符串，定义了 ParamKeys 中每个参数的预期类型。
	// 示例: '["int", "string", "float", "bool"]'。顺序必须与 ParamKeys 严格一致。
	ParamTypes string `gorm:"type:text" json:"param_types"`
}

// TableName 指定表名为 'api_services'
func (APIService) TableName() string {
	return "api_services"
}
