package response

import "github.com/gin-gonic/gin"

// Envelope 定义统一响应体：code/message/data。
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success 输出成功响应。
func Success(c *gin.Context, httpStatus int, data any) {
	c.JSON(httpStatus, Envelope{
		Code:    httpStatus,
		Message: "ok",
		Data:    data,
	})
}

// Error 输出失败响应。
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Envelope{
		Code:    httpStatus,
		Message: message,
	})
}

