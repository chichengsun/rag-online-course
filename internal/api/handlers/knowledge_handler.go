package handlers

import (
	"net/http"
	"strconv"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	knowdto "rag-online-course/internal/dto/knowledge"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// KnowledgeHandler 课程知识库：资源列表、分块预览/保存、确认、嵌入。
type KnowledgeHandler struct {
	svc *service.KnowledgeService
}

// NewKnowledgeHandler 创建处理器。
func NewKnowledgeHandler(svc *service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{svc: svc}
}

// ListKnowledgeResources GET /teacher/courses/:courseId/knowledge/resources
func (h *KnowledgeHandler) ListKnowledgeResources(c *gin.Context) {
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
	out, err := h.svc.ListKnowledgeResources(c.Request.Context(), courseID, teacherID, page, pageSize)
	if err != nil {
		logging.FromContext(c.Request.Context()).WithError(err).Error("list knowledge resources failed")
		response.Error(c, http.StatusInternalServerError, "加载知识库资源失败")
		return
	}
	response.Success(c, http.StatusOK, out)
}

// ChunkPreview POST /teacher/resources/:resourceId/knowledge/chunk-preview
func (h *KnowledgeHandler) ChunkPreview(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.ChunkPreviewReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	out, err := h.svc.ChunkPreview(c.Request.Context(), resourceID, teacherID, req.ChunkSize, req.Overlap, req.ClearPersistedFirst)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "资源不存在或无权访问")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}

// SaveKnowledgeChunks PUT /teacher/resources/:resourceId/knowledge/chunks
func (h *KnowledgeHandler) SaveKnowledgeChunks(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.SaveKnowledgeChunksReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	if req.Chunks == nil {
		req.Chunks = []knowdto.ChunkSaveItem{}
	}
	if err := h.svc.SaveKnowledgeChunks(c.Request.Context(), resourceID, teacherID, req.Chunks); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "资源不存在或无权访问")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ConfirmKnowledgeChunks POST /teacher/resources/:resourceId/knowledge/chunks/confirm
func (h *KnowledgeHandler) ConfirmKnowledgeChunks(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	n, err := h.svc.ConfirmKnowledgeChunks(c.Request.Context(), resourceID, teacherID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "确认分块失败")
		return
	}
	response.Success(c, http.StatusOK, knowdto.ConfirmKnowledgeChunksResp{ConfirmedCount: n})
}

// ListKnowledgeChunks GET /teacher/resources/:resourceId/knowledge/chunks
func (h *KnowledgeHandler) ListKnowledgeChunks(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	items, err := h.svc.ListKnowledgeChunks(c.Request.Context(), resourceID, teacherID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "加载分块失败")
		return
	}
	response.Success(c, http.StatusOK, knowdto.ListKnowledgeChunksResp{Items: items})
}

// ClearKnowledgeChunks DELETE /teacher/resources/:resourceId/knowledge/chunks
func (h *KnowledgeHandler) ClearKnowledgeChunks(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.svc.ClearKnowledgeChunks(c.Request.Context(), resourceID, teacherID); err != nil {
		response.Error(c, http.StatusInternalServerError, "清空分块失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateKnowledgeChunk PATCH /teacher/resources/:resourceId/knowledge/chunks/:chunkId
func (h *KnowledgeHandler) UpdateKnowledgeChunk(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	chunkID, err := parsePathID(c, "chunkId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid chunk id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.UpdateKnowledgeChunkReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	if err := h.svc.UpdateKnowledgeChunk(c.Request.Context(), resourceID, teacherID, chunkID, req); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "分块不存在或无权访问")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteKnowledgeChunk DELETE /teacher/resources/:resourceId/knowledge/chunks/:chunkId
func (h *KnowledgeHandler) DeleteKnowledgeChunk(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	chunkID, err := parsePathID(c, "chunkId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid chunk id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.svc.DeleteKnowledgeChunk(c.Request.Context(), resourceID, teacherID, chunkID); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "分块不存在或无权访问")
			return
		}
		response.Error(c, http.StatusInternalServerError, "删除分块失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// EmbedResource POST /teacher/resources/:resourceId/knowledge/embed
func (h *KnowledgeHandler) EmbedResource(c *gin.Context) {
	resourceID, err := parsePathID(c, "resourceId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req knowdto.EmbedResourceReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	out, err := h.svc.EmbedResource(c.Request.Context(), resourceID, teacherID, req.ModelID)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "模型或分块不存在")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}
