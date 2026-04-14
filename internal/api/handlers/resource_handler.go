package handlers

import (
	"net/http"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	dto "rag-online-course/internal/dto/course"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

type ResourceHandler struct {
	resourceSvc *service.ResourceService
	parseSvc    *service.ResourceParseService
	summarySvc  *service.ResourceAISummaryService
}

// NewResourceHandler 创建资源处理器；parseSvc/summarySvc 可为 nil，对应能力将返回 503。
func NewResourceHandler(resourceSvc *service.ResourceService, parseSvc *service.ResourceParseService, summarySvc *service.ResourceAISummaryService) *ResourceHandler {
	return &ResourceHandler{resourceSvc: resourceSvc, parseSvc: parseSvc, summarySvc: summarySvc}
}

func (h *ResourceHandler) InitUpload(c *gin.Context) {
	var initUploadReq dto.InitUploadReq
	if bindErr := c.ShouldBindJSON(&initUploadReq); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	sectionID, parseSectionErr := parsePathID(c, "sectionId")
	if parseSectionErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid section id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	initUploadReq.SectionID = sectionID
	objectKey, uploadURL, expireSeconds, err := h.resourceSvc.InitUpload(
		c.Request.Context(),
		initUploadReq.CourseID,
		initUploadReq.SectionID,
		teacherID,
		initUploadReq.FileName,
	)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			httpStatus = http.StatusNotFound
		}
		response.Error(c, httpStatus, "create presigned url failed")
		return
	}
	response.Success(c, http.StatusOK, dto.InitUploadResp{
		ObjectKey:     objectKey,
		UploadURL:     uploadURL,
		ExpireSeconds: expireSeconds,
	})
}

func (h *ResourceHandler) ConfirmResource(c *gin.Context) {
	var confirmResourceReq dto.ConfirmResourceReq
	if bindErr := c.ShouldBindJSON(&confirmResourceReq); bindErr != nil {
		logging.FromContext(c.Request.Context()).WithError(bindErr).Warn()
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	sectionID, parseSectionErr := parsePathID(c, "sectionId")
	if parseSectionErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid section id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	confirmResourceReq.SectionID = sectionID
	confirmResourceReq.TeacherID = teacherID
	resourceID, err := h.resourceSvc.ConfirmResource(
		c.Request.Context(),
		confirmResourceReq.SectionID,
		confirmResourceReq.TeacherID,
		confirmResourceReq.Title,
		confirmResourceReq.ResourceType,
		confirmResourceReq.ObjectKey,
		confirmResourceReq.MimeType,
		confirmResourceReq.SizeBytes,
		confirmResourceReq.SortOrder,
	)
	if err != nil {
		if respondPostgresUniqueViolation(c, err, "confirm_resource") {
			return
		}
		logging.FromContext(c.Request.Context()).WithError(err).Error("confirm resource failed")
		response.Error(c, http.StatusInternalServerError, "资源入库失败，请稍后重试或联系管理员")
		return
	}
	response.Success(c, http.StatusCreated, dto.ConfirmResourceResp{ID: resourceID})
}

func (h *ResourceHandler) ReorderResource(c *gin.Context) {
	var reorderResourceReq dto.ReorderResourceReq
	if bindErr := c.ShouldBindJSON(&reorderResourceReq); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	reorderResourceReq.ResourceID = resourceID
	reorderResourceReq.TeacherID = teacherID
	if err := h.resourceSvc.ReorderResource(c.Request.Context(), reorderResourceReq.ResourceID, reorderResourceReq.TeacherID, reorderResourceReq.SortOrder); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "reorder resource failed")
			return
		}
		if respondPostgresUniqueViolation(c, err, "reorder_resource") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "reorder resource failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateResource PUT /teacher/resources/:resourceId — 更新资源标题（不含排序，排序见 ReorderResource）。
func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	var req dto.UpdateResourceTitleReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	req.ResourceID = resourceID
	req.TeacherID = teacherID
	if err := h.resourceSvc.UpdateResourceTitle(c.Request.Context(), req.ResourceID, req.TeacherID, req.Title); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "update resource failed")
			return
		}
		if respondPostgresUniqueViolation(c, err, "update_resource") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "update resource failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// ListResources 教师端展示某节下的资源列表。
func (h *ResourceHandler) ListResources(c *gin.Context) {
	sectionID, parseSectionErr := parsePathID(c, "sectionId")
	if parseSectionErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid section id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	items, err := h.resourceSvc.ListResources(c.Request.Context(), sectionID, teacherID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "list resources failed")
		return
	}
	response.Success(c, http.StatusOK, dto.ListResourcesResp{Items: items})
}

// GetResourceDetail 返回资源元数据（含可选 AI 摘要），供预览页等按需拉取。
func (h *ResourceHandler) GetResourceDetail(c *gin.Context) {
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	detail, err := h.resourceSvc.GetTeacherResourceDetail(c.Request.Context(), resourceID, teacherID)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			httpStatus = http.StatusNotFound
		}
		response.Error(c, httpStatus, "get resource detail failed")
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// SummarizeResource 对文档类资源触发解析并调用问答模型生成摘要，写入 ai_summary。
func (h *ResourceHandler) SummarizeResource(c *gin.Context) {
	if h.summarySvc == nil {
		response.Error(c, http.StatusServiceUnavailable, "摘要服务未初始化")
		return
	}
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	resp, httpStatus, err := h.summarySvc.StartSummarizeJob(c.Request.Context(), resourceID, teacherID)
	if err != nil {
		logging.FromContext(c.Request.Context()).WithError(err).WithField("resource_id", resourceID).Warn("summarize job start failed")
		code := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			code = http.StatusNotFound
		}
		response.Error(c, code, err.Error())
		return
	}
	response.Success(c, httpStatus, resp)
}

// ParseResource 教师触发多模态解析：经 docreader-http 返回 Markdown/元数据；持久化分块与向量由 resource_embedding_chunks 管线完成。
func (h *ResourceHandler) ParseResource(c *gin.Context) {
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if h.parseSvc == nil {
		response.Error(c, http.StatusServiceUnavailable, "解析服务未初始化")
		return
	}
	out, err := h.parseSvc.ParseResource(c.Request.Context(), resourceID, teacherID)
	if err != nil {
		logging.FromContext(c.Request.Context()).WithError(err).WithField("resource_id", resourceID).Error("parse resource failed")
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "资源不存在或无权访问")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}

// DeleteResource 删除资源。
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.resourceSvc.DeleteResource(c.Request.Context(), resourceID, teacherID); err != nil {
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			httpStatus = http.StatusNotFound
		}
		response.Error(c, httpStatus, "delete resource failed")
		return
	}
	c.Status(http.StatusNoContent)
}
