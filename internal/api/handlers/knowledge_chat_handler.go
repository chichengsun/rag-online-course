package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	knowdto "rag-online-course/internal/dto/knowledge"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// KnowledgeChatHandler 提供知识库会话管理与问答接口。
type KnowledgeChatHandler struct {
	svc *service.KnowledgeChatService
}

// NewKnowledgeChatHandler 创建知识库对话处理器。
func NewKnowledgeChatHandler(svc *service.KnowledgeChatService) *KnowledgeChatHandler {
	return &KnowledgeChatHandler{svc: svc}
}

// CreateSession POST /teacher/courses/:courseId/knowledge/chats/sessions
func (h *KnowledgeChatHandler) CreateSession(c *gin.Context) {
	courseID, err := parsePathID(c, "courseId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid course id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.CreateChatSessionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	id, err := h.svc.CreateSession(c.Request.Context(), teacherID, courseID, req)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "课程不存在或无权访问")
			return
		}
		response.Error(c, http.StatusInternalServerError, "创建会话失败")
		return
	}
	response.Success(c, http.StatusCreated, knowdto.CreateChatSessionResp{ID: id})
}

// ListSessions GET /teacher/knowledge/chats/sessions
func (h *KnowledgeChatHandler) ListSessions(c *gin.Context) {
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	page, pageSize := 1, 20
	if v := c.Query("page"); v != "" {
		p, e := strconv.Atoi(v)
		if e != nil {
			response.Error(c, http.StatusBadRequest, "invalid page")
			return
		}
		page = p
	}
	if v := c.Query("page_size"); v != "" {
		p, e := strconv.Atoi(v)
		if e != nil {
			response.Error(c, http.StatusBadRequest, "invalid page_size")
			return
		}
		pageSize = p
	}
	var courseID *int64
	if v := c.Query("course_id"); v != "" {
		id, e := strconv.ParseInt(v, 10, 64)
		if e != nil {
			response.Error(c, http.StatusBadRequest, "invalid course_id")
			return
		}
		courseID = &id
	}
	out, err := h.svc.ListSessions(c.Request.Context(), teacherID, courseID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "加载会话列表失败")
		return
	}
	response.Success(c, http.StatusOK, out)
}

// ListMessages GET /teacher/knowledge/chats/sessions/:sessionId/messages
func (h *KnowledgeChatHandler) ListMessages(c *gin.Context) {
	sessionID, err := parsePathID(c, "sessionId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid session id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	page, pageSize := 1, 50
	if v := c.Query("page"); v != "" {
		p, e := strconv.Atoi(v)
		if e != nil {
			response.Error(c, http.StatusBadRequest, "invalid page")
			return
		}
		page = p
	}
	if v := c.Query("page_size"); v != "" {
		p, e := strconv.Atoi(v)
		if e != nil {
			response.Error(c, http.StatusBadRequest, "invalid page_size")
			return
		}
		pageSize = p
	}
	out, err := h.svc.ListMessages(c.Request.Context(), teacherID, sessionID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "加载历史消息失败")
		return
	}
	response.Success(c, http.StatusOK, out)
}

// AskInSession POST /teacher/knowledge/chats/sessions/:sessionId/ask
func (h *KnowledgeChatHandler) AskInSession(c *gin.Context) {
	sessionID, err := parsePathID(c, "sessionId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid session id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.AskInSessionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	out, err := h.svc.AskInSession(c.Request.Context(), teacherID, sessionID, req)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "会话不存在或无权访问")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}

// UpdateSession PATCH /teacher/knowledge/chats/sessions/:sessionId
func (h *KnowledgeChatHandler) UpdateSession(c *gin.Context) {
	sessionID, err := parsePathID(c, "sessionId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid session id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.UpdateChatSessionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	if err := h.svc.UpdateSessionTitle(c.Request.Context(), teacherID, sessionID, req.Title); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "会话不存在或无权访问")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteSession DELETE /teacher/knowledge/chats/sessions/:sessionId
func (h *KnowledgeChatHandler) DeleteSession(c *gin.Context) {
	sessionID, err := parsePathID(c, "sessionId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid session id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.svc.DeleteSession(c.Request.Context(), teacherID, sessionID); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "会话不存在或无权访问")
			return
		}
		response.Error(c, http.StatusInternalServerError, "删除会话失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// AskInSessionStream POST /teacher/knowledge/chats/sessions/:sessionId/ask/stream
func (h *KnowledgeChatHandler) AskInSessionStream(c *gin.Context) {
	sessionID, err := parsePathID(c, "sessionId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid session id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.AskInSessionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()

	emit := func(event string, payload any) {
		b, _ := json.Marshal(payload)
		_, _ = fmt.Fprintf(c.Writer, "event: %s\n", event)
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", string(b))
		c.Writer.Flush()
	}

	err = h.svc.AskInSessionStream(c.Request.Context(), teacherID, sessionID, req, service.ChatStreamCallbacks{
		OnToken: func(token string) {
			emit("token", map[string]any{"token": token})
		},
		OnReferences: func(refs []knowdto.ReferenceItem) {
			emit("references", map[string]any{"references": refs})
		},
		OnDone: func(userMessageID, assistantMessageID string) {
			emit("done", map[string]any{
				"session_id":            strconv.FormatInt(sessionID, 10),
				"user_message_id":       userMessageID,
				"assistant_message_id":  assistantMessageID,
			})
		},
	})
	if err != nil {
		emit("error", map[string]any{"message": err.Error()})
	}
}
