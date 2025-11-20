package models

import (
    "time"

    "gorm.io/gorm"
)

// Audit 记录动态 SQL 执行的审计信息
type Audit struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    // 请求元信息
    Path      string `gorm:"index;size:191" json:"path"`
    Method    string `gorm:"size:10" json:"method"`
    ClientIP  string `gorm:"size:45" json:"client_ip"`

    // 执行相关
    SQL        string `gorm:"type:text" json:"sql"`
    Args       string `gorm:"type:text" json:"args"`
    DurationMs int64  `json:"duration_ms"`
    Rows       int    `json:"rows"`
    Truncated  bool   `json:"truncated"`
    Error      string `gorm:"type:text" json:"error"`
}

// TableName 指定表名为 'audits'
func (Audit) TableName() string {
    return "audits"
}
