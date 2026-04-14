package handlers

import (
	"net/http"
	"strconv"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	dto "rag-online-course/internal/dto/course"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	courseSvc *service.CourseService
}

// NewCourseHandler 创建课程建设相关 HTTP 处理器。
func NewCourseHandler(courseSvc *service.CourseService) *CourseHandler {
	return &CourseHandler{courseSvc: courseSvc}
}

// CreateCourse 创建课程，并在应用层规范化标题格式。
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var createCourseReq dto.CreateCourseReq
	if bindErr := c.ShouldBindJSON(&createCourseReq); bindErr != nil {
		logging.FromContext(c.Request.Context()).WithError(bindErr).Warn()
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	teacherID, parseErr := parseContextUserID(c, middleware.ContextUserID)
	if parseErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	createCourseReq.TeacherID = teacherID
	logger := logging.FromContext(c.Request.Context()).WithFields(map[string]any{
		"teacher_id": createCourseReq.TeacherID,
		"title":      createCourseReq.Title,
	})
	courseID, err := h.courseSvc.CreateCourse(
		c.Request.Context(),
		createCourseReq.TeacherID,
		createCourseReq.Title,
		createCourseReq.Description,
	)
	if err != nil {
		if respondPostgresUniqueViolation(c, err, "create_course") {
			return
		}
		logger.WithError(err).Warn()
		response.Error(c, http.StatusInternalServerError, "create course failed")
		return
	}
	response.Success(c, http.StatusCreated, dto.CreateCourseResp{ID: courseID})
}

// UpdateCourse 更新课程基本信息和发布状态。
func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	var updateCourseReq dto.UpdateCourseReq
	if bindErr := c.ShouldBindJSON(&updateCourseReq); bindErr != nil {
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	courseID, parseCourseErr := parsePathID(c, "courseId")
	if parseCourseErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid course id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	updateCourseReq.CourseID = courseID
	updateCourseReq.TeacherID = teacherID
	if err := h.courseSvc.UpdateCourse(
		c.Request.Context(),
		updateCourseReq.CourseID,
		updateCourseReq.TeacherID,
		updateCourseReq.Title,
		updateCourseReq.Description,
		updateCourseReq.Status,
	); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "update course failed")
			return
		}
		if respondPostgresUniqueViolation(c, err, "update_course") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "update course failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// ListCourses 教师端分页展示课程列表。
func (h *CourseHandler) ListCourses(c *gin.Context) {
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	page := 1
	pageSize := 10
	keyword := c.Query("keyword")
	status := c.Query("status")
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")
	if pageRaw := c.Query("page"); pageRaw != "" {
		parsedPage, parsePageErr := strconv.Atoi(pageRaw)
		if parsePageErr != nil {
			response.Error(c, http.StatusBadRequest, "invalid page")
			return
		}
		page = parsedPage
	}
	if pageSizeRaw := c.Query("page_size"); pageSizeRaw != "" {
		parsedPageSize, parsePageSizeErr := strconv.Atoi(pageSizeRaw)
		if parsePageSizeErr != nil {
			response.Error(c, http.StatusBadRequest, "invalid page_size")
			return
		}
		pageSize = parsedPageSize
	}
	items, total, err := h.courseSvc.ListCourses(c.Request.Context(), teacherID, keyword, status, sortBy, sortOrder, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "list courses failed")
		return
	}
	response.Success(c, http.StatusOK, dto.ListCoursesResp{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	})
}

// DeleteCourse 删除课程。
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	courseID, parseCourseErr := parsePathID(c, "courseId")
	if parseCourseErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid course id")
		return
	}
	teacherID, parseTeacherErr := parseContextUserID(c, middleware.ContextUserID)
	if parseTeacherErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.courseSvc.DeleteCourse(c.Request.Context(), courseID, teacherID); err != nil {
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			httpStatus = http.StatusNotFound
		}
		response.Error(c, httpStatus, "delete course failed")
		return
	}
	c.Status(http.StatusNoContent)
}
