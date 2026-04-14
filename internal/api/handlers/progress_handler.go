package handlers

import (
	"net/http"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	dto "rag-online-course/internal/dto/learning"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// ProgressHandler 管理学习进度对象相关接口。
type ProgressHandler struct {
	progressSvc *service.ProgressService
}

// NewProgressHandler 创建学习进度处理器。
func NewProgressHandler(progressSvc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{progressSvc: progressSvc}
}

// UpdateProgress 更新资源学习进度（例如视频播放进度）。
func (h *ProgressHandler) UpdateProgress(c *gin.Context) {
	var updateProgressReq dto.UpdateProgressReq
	if bindErr := c.ShouldBindJSON(&updateProgressReq); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	studentID, parseStudentErr := parseContextUserID(c, middleware.ContextUserID)
	if parseStudentErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid student id")
		return
	}
	updateProgressReq.ResourceID = resourceID
	updateProgressReq.StudentID = studentID
	if err := h.progressSvc.UpdateProgress(c.Request.Context(), updateProgressReq); err != nil {
		if respondPostgresUniqueViolation(c, err, "update_progress") {
			return
		}
		logging.FromContext(c.Request.Context()).WithError(err).Error("update progress failed")
		response.Error(c, http.StatusInternalServerError, "更新学习进度失败，请稍后重试")
		return
	}
	c.Status(http.StatusNoContent)
}

// CompleteResource 标记资源学习完成。
func (h *ProgressHandler) CompleteResource(c *gin.Context) {
	resourceID, parseResourceErr := parsePathID(c, "resourceId")
	if parseResourceErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid resource id")
		return
	}
	studentID, parseStudentErr := parseContextUserID(c, middleware.ContextUserID)
	if parseStudentErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid student id")
		return
	}
	completeResourceReq := dto.CompleteResourceReq{
		ResourceID: resourceID,
		StudentID:  studentID,
	}
	if err := h.progressSvc.CompleteResource(c.Request.Context(), completeResourceReq); err != nil {
		if respondPostgresUniqueViolation(c, err, "complete_resource") {
			return
		}
		logging.FromContext(c.Request.Context()).WithError(err).Error("complete resource failed")
		response.Error(c, http.StatusInternalServerError, "标记完成失败，请稍后重试")
		return
	}
	c.Status(http.StatusNoContent)
}
