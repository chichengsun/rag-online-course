package handlers

import (
	"net/http"
	"strings"

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
}

// NewResourceHandler 创建资源处理器；parseSvc 可为 nil 时解析路由不应注册（当前由 DI 始终注入）。
func NewResourceHandler(resourceSvc *service.ResourceService, parseSvc *service.ResourceParseService) *ResourceHandler {
	return &ResourceHandler{resourceSvc: resourceSvc, parseSvc: parseSvc}
}

func (h *ResourceHandler) InitUpload(c *gin.Context) {
	var initUploadReq dto.InitUploadReq
	if bindErr := c.ShouldBindJSON(&initUploadReq); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	chapterID, parseChapterErr := parsePathID(c, "chapterId")
	if parseChapterErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid chapter id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	initUploadReq.ChapterID = chapterID
	objectKey, uploadURL, expireSeconds, err := h.resourceSvc.InitUpload(
		c.Request.Context(),
		initUploadReq.CourseID,
		initUploadReq.ChapterID,
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
	chapterID, parseChapterErr := parsePathID(c, "chapterId")
	if parseChapterErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid chapter id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	confirmResourceReq.ChapterID = chapterID
	confirmResourceReq.TeacherID = teacherID
	resourceID, err := h.resourceSvc.ConfirmResource(
		c.Request.Context(),
		confirmResourceReq.ChapterID,
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

// ListResources 教师端展示章节资源列表。
func (h *ResourceHandler) ListResources(c *gin.Context) {
	chapterID, parseChapterErr := parsePathID(c, "chapterId")
	if parseChapterErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid chapter id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	items, err := h.resourceSvc.ListResources(c.Request.Context(), chapterID, teacherID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "list resources failed")
		return
	}
	response.Success(c, http.StatusOK, dto.ListResourcesResp{Items: items})
}

// PreviewResourceURL 获取资源预览用 URL（Office 会在服务端生成 PDF 预览）。
func (h *ResourceHandler) PreviewResourceURL(c *gin.Context) {
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
	previewURL, err := h.resourceSvc.GetResourcePreviewURL(c.Request.Context(), resourceID, teacherID)
	if err != nil {
		log := logging.FromContext(c.Request.Context()).WithError(err).WithField("office_preview_output", extractOfficePreviewOutput(err))
		log.Error("get preview url failed")
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			httpStatus = http.StatusNotFound
		}
		if httpStatus == http.StatusInternalServerError {
			response.Error(c, httpStatus, "预览转换失败，请确认已安装/可用 Office 转换环境（LibreOffice）")
			return
		}
		response.Error(c, httpStatus, "get preview url failed")
		return
	}
	response.Success(c, http.StatusOK, dto.PreviewResourceURLResp{PreviewURL: previewURL})
}

func extractOfficePreviewOutput(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	const key = "output="
	idx := strings.Index(s, key)
	if idx < 0 {
		return ""
	}
	out := s[idx+len(key):]
	// 避免日志过大，只保留前 2KB。
	if len(out) > 2*1024 {
		out = out[:2*1024] + "...(truncated)"
	}
	return out
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
