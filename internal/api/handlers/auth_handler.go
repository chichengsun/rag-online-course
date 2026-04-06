// Package handlers 提供 Gin HTTP 处理器。路径中的 id 参数会在边界层解析为 int64，以匹配 PostgreSQL BIGSERIAL 主键。
package handlers

import (
	"net/http"
	"strings"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	dto "rag-online-course/internal/dto/user"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建用户相关 HTTP 处理器。
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// Register 处理用户注册，支持学生/教师角色创建。
func (h *UserHandler) Register(c *gin.Context) {
	var registerReq dto.RegisterReq
	logger := logging.FromContext(c.Request.Context()).WithFields(map[string]any{
		"username": registerReq.Username,
		"role":     registerReq.Role,
	})
	if bindErr := c.ShouldBindJSON(&registerReq); bindErr != nil {
		logging.FromContext(c.Request.Context()).WithError(bindErr).Warn()
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	logger = logging.FromContext(c.Request.Context()).WithFields(map[string]any{
		"username": registerReq.Username,
		"role":     registerReq.Role,
	})
	registerResp, err := h.userSvc.Register(c.Request.Context(), registerReq)
	if err != nil {
		if respondPostgresUniqueViolation(c, err, "register") {
			return
		}
		logger.WithError(err).Warn()
		// 入库前校验错误（用户名格式等）直接提示；其它异常统一 500，避免泄露内部细节。
		if strings.Contains(strings.ToLower(err.Error()), "username") {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "注册失败，请稍后重试")
		return
	}
	response.Success(c, http.StatusCreated, registerResp)
}

// Login 支持使用 email 或 username 登录。
func (h *UserHandler) Login(c *gin.Context) {
	var loginReq dto.LoginReq
	if bindErr := c.ShouldBindJSON(&loginReq); bindErr != nil {
		logging.FromContext(c.Request.Context()).WithError(bindErr).Warn()
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	logger := logging.FromContext(c.Request.Context()).WithField("account", loginReq.Account)
	loginResp, err := h.userSvc.Login(c.Request.Context(), loginReq)
	if err != nil {
		logger.WithError(err).Warn()
		if err == service.ErrUnauthorized {
			response.Error(c, http.StatusUnauthorized, "invalid account or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, "login failed")
		return
	}
	response.Success(c, http.StatusOK, loginResp)
}

// Refresh 使用 refresh token 续签 access token。
func (h *UserHandler) Refresh(c *gin.Context) {
	var refreshReq dto.RefreshReq
	if bindErr := c.ShouldBindJSON(&refreshReq); bindErr != nil {
		logging.FromContext(c.Request.Context()).WithError(bindErr).Warn()
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	logger := logging.FromContext(c.Request.Context())
	refreshResp, err := h.userSvc.Refresh(c.Request.Context(), refreshReq)
	if err != nil {
		logger.WithError(err).Warn()
		if err == service.ErrUnauthorized {
			response.Error(c, http.StatusUnauthorized, "invalid refresh token")
			return
		}
		response.Error(c, http.StatusInternalServerError, "refresh failed")
		return
	}
	response.Success(c, http.StatusOK, refreshResp)
}

// Me 返回当前登录用户资料。
func (h *UserHandler) Me(c *gin.Context) {
	userID, parseUserErr := parseContextUserID(c, middleware.ContextUserID)
	if parseUserErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	meReq := dto.MeReq{UserID: userID}
	logger := logging.FromContext(c.Request.Context()).WithField("user_id", meReq.UserID)
	profileResp, err := h.userSvc.Me(c.Request.Context(), meReq)
	if err != nil {
		logger.WithError(err).Warn()
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}
	response.Success(c, http.StatusOK, profileResp)
}
