package handlers

import (
	"net/http"

	"rag-online-course/internal/logging"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	knowdto "rag-online-course/internal/dto/knowledge"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// AIModelHandler 教师 AI 模型配置 API。
type AIModelHandler struct {
	svc *service.AIModelService
}

// NewAIModelHandler 创建处理器。
func NewAIModelHandler(svc *service.AIModelService) *AIModelHandler {
	return &AIModelHandler{svc: svc}
}

// ListAIModels GET /teacher/ai-models
func (h *AIModelHandler) ListAIModels(c *gin.Context) {
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	items, err := h.svc.List(c.Request.Context(), teacherID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "加载模型列表失败")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"items": items})
}

// CreateAIModel POST /teacher/ai-models
func (h *AIModelHandler) CreateAIModel(c *gin.Context) {
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.CreateAIModelReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	id, err := h.svc.Create(c.Request.Context(), teacherID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建模型失败")
		return
	}
	response.Success(c, http.StatusCreated, knowdto.CreateAIModelResp{ID: id})
}

// TestAIModelConnection POST /teacher/ai-models/test-connection
func (h *AIModelHandler) TestAIModelConnection(c *gin.Context) {
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.TestAIModelConnectionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	out, err := h.svc.TestAIModelConnection(c.Request.Context(), teacherID, req)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "模型不存在或无权使用已存密钥")
			return
		}
		logging.FromContext(c.Request.Context()).WithError(err).Warn("ai model test connection rejected")
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}

// UpdateAIModel PUT /teacher/ai-models/:modelId
func (h *AIModelHandler) UpdateAIModel(c *gin.Context) {
	modelID, err := parsePathID(c, "modelId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid model id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.UpdateAIModelReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), modelID, teacherID, req); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "模型不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "更新模型失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteAIModel DELETE /teacher/ai-models/:modelId
func (h *AIModelHandler) DeleteAIModel(c *gin.Context) {
	modelID, err := parsePathID(c, "modelId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid model id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), modelID, teacherID); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "模型不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "删除模型失败")
		return
	}
	c.Status(http.StatusNoContent)
}
