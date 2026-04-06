package middleware

import (
	"errors"
	"net/http"

	"rag-online-course/internal/api/response"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 统一处理上下文中的业务错误，避免错误响应风格分散。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 已写响应则不覆盖。
		if c.Writer.Written() {
			return
		}
		if len(c.Errors) == 0 {
			return
		}

		last := c.Errors.Last()
		status := mapHTTPStatus(last.Err)
		logging.FromContext(c.Request.Context()).
			WithFields(map[string]any{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"status": status,
			}).
			WithError(last.Err).
			Error()
		if status >= http.StatusInternalServerError {
			response.Error(c, status, "internal server error")
			return
		}
		response.Error(c, status, last.Error())
	}
}

func mapHTTPStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusInternalServerError
	case errors.Is(err, postgres.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, postgres.ErrNoCourseAccess):
		return http.StatusForbidden
	case errors.Is(err, service.ErrUnauthorized):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
