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

// CourseDesignHandler 教师端「课程设计」HTTP 入口：AI 大纲草案生成与应用。
type CourseDesignHandler struct {
	svc *service.CourseDesignService
}

// NewCourseDesignHandler 创建处理器。
func NewCourseDesignHandler(svc *service.CourseDesignService) *CourseDesignHandler {
	return &CourseDesignHandler{svc: svc}
}

// GenerateOutlineDraft POST /teacher/courses/:courseId/design/outline-draft/generate
func (h *CourseDesignHandler) GenerateOutlineDraft(c *gin.Context) {
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
	var req dto.GenerateOutlineDraftReq
	_ = c.ShouldBindJSON(&req)
	out, genErr := h.svc.GenerateOutlineDraft(c.Request.Context(), courseID, teacherID, req.QAModelID, req.ExtraHint)
	if genErr != nil {
		if genErr == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "course not found")
			return
		}
		logging.FromContext(c.Request.Context()).WithError(genErr).WithFields(map[string]any{
			"course_id": courseID,
		}).Warn("generate outline draft failed")
		response.Error(c, http.StatusBadRequest, genErr.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}

// ApplyOutlineDraft POST /teacher/courses/:courseId/design/outline-draft/apply
func (h *CourseDesignHandler) ApplyOutlineDraft(c *gin.Context) {
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
	var req dto.ApplyOutlineDraftReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	out, applyErr := h.svc.ApplyOutlineDraft(c.Request.Context(), courseID, teacherID, &req)
	if applyErr != nil {
		if applyErr == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "course not found")
			return
		}
		logging.FromContext(c.Request.Context()).WithError(applyErr).WithFields(map[string]any{
			"course_id": courseID,
		}).Warn("apply outline draft failed")
		response.Error(c, http.StatusBadRequest, applyErr.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}
