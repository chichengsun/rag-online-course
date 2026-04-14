package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parsePathID 将路径参数解析为 int64 主键。
func parsePathID(c *gin.Context, key string) (int64, error) {
	return strconv.ParseInt(c.Param(key), 10, 64)
}

// parseContextUserID 将鉴权中间件注入的用户 ID 解析为 int64 主键。
func parseContextUserID(c *gin.Context, key string) (int64, error) {
	return strconv.ParseInt(c.GetString(key), 10, 64)
}
