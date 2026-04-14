package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 处理中间件，支持浏览器跨域访问与预检请求。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		// 开发阶段放开为反射 Origin，生产可改为白名单校验。
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Max-Age", "86400")

		if strings.EqualFold(c.Request.Method, http.MethodOptions) {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
