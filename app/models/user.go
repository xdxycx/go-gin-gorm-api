package models

import (
	"gorm.io/gorm"
)

// User 模型映射到数据库中的 users 表
type User struct {
	gorm.Model
	Name  string `json:"name"`
	Email string `json:"email" gorm:"unique"`
}
