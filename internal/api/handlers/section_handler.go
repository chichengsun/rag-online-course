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

// SectionHandler 教师端「节」管理：节位于章节与资源之间。
type SectionHandler struct {
	sectionSvc    *service.SectionService
	lessonPlanSvc *service.SectionLessonPlanService
}

// NewSectionHandler 创建节处理器。
func NewSectionHandler(sectionSvc *service.SectionService, lessonPlanSvc *service.SectionLessonPlanService) *SectionHandler {
	return &SectionHandler{sectionSvc: sectionSvc, lessonPlanSvc: lessonPlanSvc}
}

// CreateSection 在指定章节下创建节。
func (h *SectionHandler) CreateSection(c *gin.Context) {
	courseID, errCourse := parsePathID(c, "courseId")
	if errCourse != nil {
		response.Error(c, http.StatusBadRequest, "invalid course id")
		return
	}
	chapterID, errChapter := parsePathID(c, "chapterId")
	if errChapter != nil {
		response.Error(c, http.StatusBadRequest, "invalid chapter id")
		return
	}
	var req dto.CreateSectionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		logging.FromContext(c.Request.Context()).WithError(bindErr).Warn()
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	teacherID, errTeacher := parseContextUserID(c, middleware.ContextUserID)
	if errTeacher != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	id, err := h.sectionSvc.CreateSection(c.Request.Context(), chapterID, courseID, teacherID, req.Title, req.SortOrder)
	if err != nil {
		if respondPostgresUniqueViolation(c, err, "create_section") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "create section failed")
		return
	}
	response.Success(c, http.StatusCreated, dto.CreateSectionResp{ID: id})
}

// ListSections 列出章节下的节。
func (h *SectionHandler) ListSections(c *gin.Context) {
	chapterID, errChapter := parsePathID(c, "chapterId")
	if errChapter != nil {
		response.Error(c, http.StatusBadRequest, "invalid chapter id")
		return
	}
	teacherID, errTeacher := parseContextUserID(c, middleware.ContextUserID)
	if errTeacher != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	items, err := h.sectionSvc.ListSections(c.Request.Context(), chapterID, teacherID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "list sections failed")
		return
	}
	response.Success(c, http.StatusOK, dto.ListSectionsResp{Items: items})
}

// ReorderSection 调整节顺序。
func (h *SectionHandler) ReorderSection(c *gin.Context) {
	var req dto.ReorderSectionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	sectionID, errSec := parsePathID(c, "sectionId")
	if errSec != nil {
		response.Error(c, http.StatusBadRequest, "invalid section id")
		return
	}
	teacherID, errTeacher := parseContextUserID(c, middleware.ContextUserID)
	if errTeacher != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	req.SectionID = sectionID
	req.TeacherID = teacherID
	if err := h.sectionSvc.ReorderSection(c.Request.Context(), req.SectionID, req.TeacherID, req.SortOrder); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "reorder section failed")
			return
		}
		if respondPostgresUniqueViolation(c, err, "reorder_section") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "reorder section failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateSection PUT /teacher/sections/:sectionId — 更新节标题与排序。
func (h *SectionHandler) UpdateSection(c *gin.Context) {
	var req dto.UpdateSectionReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	sectionID, errSec := parsePathID(c, "sectionId")
	if errSec != nil {
		response.Error(c, http.StatusBadRequest, "invalid section id")
		return
	}
	teacherID, errTeacher := parseContextUserID(c, middleware.ContextUserID)
	if errTeacher != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	req.SectionID = sectionID
	req.TeacherID = teacherID
	if err := h.sectionSvc.UpdateSection(c.Request.Context(), req.SectionID, req.TeacherID, req.Title, req.SortOrder); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "update section failed")
			return
		}
		if respondPostgresUniqueViolation(c, err, "update_section") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "update section failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteSection 删除节（节下资源随外键级联删除）。
func (h *SectionHandler) DeleteSection(c *gin.Context) {
	sectionID, errSec := parsePathID(c, "sectionId")
	if errSec != nil {
		response.Error(c, http.StatusBadRequest, "invalid section id")
		return
	}
	teacherID, errTeacher := parseContextUserID(c, middleware.ContextUserID)
	if errTeacher != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.sectionSvc.DeleteSection(c.Request.Context(), sectionID, teacherID); err != nil {
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			httpStatus = http.StatusNotFound
		}
		response.Error(c, httpStatus, "delete section failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// GenerateLessonPlanDraft POST /teacher/sections/:sectionId/lesson-plan/generate — 根据小节目标与资源生成结构化教案草案。
func (h *SectionHandler) GenerateLessonPlanDraft(c *gin.Context) {
	sectionID, errSec := parsePathID(c, "sectionId")
	if errSec != nil {
		response.Error(c, http.StatusBadRequest, "invalid section id")
		return
	}
	teacherID, errTeacher := parseContextUserID(c, middleware.ContextUserID)
	if errTeacher != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req dto.GenerateSectionLessonPlanReq
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	out, err := h.lessonPlanSvc.GenerateSectionLessonPlan(c.Request.Context(), sectionID, teacherID, req)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "section not found")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}
