package utils

// APIResponse 定义统一的 API 响应结构
// 建议所有接口都采用此格式返回数据
type APIResponse struct {
	Code    int         `json:"code"`    // 业务代码, 0 表示成功，非 0 表示错误
	Message string      `json:"message"` // 消息描述
	Data    interface{} `json:"data"`    // 数据载荷 (可以是任意类型)
}
