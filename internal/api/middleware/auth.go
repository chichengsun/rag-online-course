package middleware

import (
	"net/http"
	"strings"

	"rag-online-course/internal/api/response"
	"rag-online-course/internal/service/auth"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "userID"
	ContextRole   = "role"
)

// RequireAuth 校验 Bearer Token，并把用户身份写入请求上下文。
// RequireAuth 基于 JWT + Redis Session 双重校验。
// JWT 提供无状态的签名校验；Redis Session 用于“可集中吊销/共享会话”。
func RequireAuth(jwtSvc *auth.JWTService, sessionStore *auth.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}
		raw := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtSvc.ParseAccessToken(raw)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		active, err := sessionStore.ValidateSession(c.Request.Context(), claims.SessionID, claims.UserID, claims.Role)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "session validate failed")
			c.Abort()
			return
		}
		if !active {
			response.Error(c, http.StatusUnauthorized, "session expired")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// RequireRole 在鉴权通过后进行角色级访问控制。
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *gin.Context) {
		role := c.GetString(ContextRole)
		if !allowed[role] {
			response.Error(c, http.StatusForbidden, "permission denied")
			c.Abort()
			return
		}
		c.Next()
	}
}
