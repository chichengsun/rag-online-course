package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Recovery 捕获 panic，避免进程崩溃并返回统一错误响应。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logrus.WithFields(logrus.Fields{
					"panic":  recovered,
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
				}).Error("panic recovered")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
