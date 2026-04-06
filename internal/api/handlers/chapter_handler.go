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

type ChapterHandler struct {
	chapterSvc *service.ChapterService
}

func NewChapterHandler(chapterSvc *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{chapterSvc: chapterSvc}
}

func (h *ChapterHandler) CreateChapter(c *gin.Context) {
	var createChapterReq dto.CreateChapterReq
	if bindErr := c.ShouldBindJSON(&createChapterReq); bindErr != nil {
		logging.FromContext(c.Request.Context()).WithError(bindErr).Warn()
		response.Error(c, http.StatusBadRequest, bindErr.Error())
		return
	}
	courseID, parseCourseErr := parsePathID(c, "courseId")
	if parseCourseErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid course id")
		return
	}
	createChapterReq.CourseID = courseID
	chapterID, err := h.chapterSvc.CreateChapter(c.Request.Context(), createChapterReq.CourseID, createChapterReq.Title, createChapterReq.SortOrder)
	if err != nil {
		if respondPostgresUniqueViolation(c, err, "create_chapter") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "create chapter failed")
		return
	}
	response.Success(c, http.StatusCreated, dto.CreateChapterResp{ID: chapterID})
}

func (h *ChapterHandler) ReorderChapter(c *gin.Context) {
	var reorderChapterReq dto.ReorderChapterReq
	if bindErr := c.ShouldBindJSON(&reorderChapterReq); bindErr != nil {
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
	reorderChapterReq.ChapterID = chapterID
	reorderChapterReq.TeacherID = teacherID
	if err := h.chapterSvc.ReorderChapter(c.Request.Context(), reorderChapterReq.ChapterID, reorderChapterReq.TeacherID, reorderChapterReq.SortOrder); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "reorder chapter failed")
			return
		}
		if respondPostgresUniqueViolation(c, err, "reorder_chapter") {
			return
		}
		response.Error(c, http.StatusInternalServerError, "reorder chapter failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// ListChapters 教师端展示课程章节列表。
func (h *ChapterHandler) ListChapters(c *gin.Context) {
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
	items, err := h.chapterSvc.ListChapters(c.Request.Context(), courseID, teacherID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "list chapters failed")
		return
	}
	response.Success(c, http.StatusOK, dto.ListChaptersResp{Items: items})
}

// DeleteChapter 删除章节。
func (h *ChapterHandler) DeleteChapter(c *gin.Context) {
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
	if err := h.chapterSvc.DeleteChapter(c.Request.Context(), chapterID, teacherID); err != nil {
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNotFound {
			httpStatus = http.StatusNotFound
		}
		response.Error(c, httpStatus, "delete chapter failed")
		return
	}
	c.Status(http.StatusNoContent)
}
