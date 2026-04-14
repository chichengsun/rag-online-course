package handlers

import (
	"net/http"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// EnrollmentHandler 管理选课对象相关接口。
type EnrollmentHandler struct {
	enrollmentSvc *service.EnrollmentService
}

// NewEnrollmentHandler 创建选课处理器。
func NewEnrollmentHandler(enrollmentSvc *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{enrollmentSvc: enrollmentSvc}
}

// ListCourses 返回已发布课程列表（供学生选课）。
func (h *EnrollmentHandler) ListCourses(c *gin.Context) {
	listCoursesResp, err := h.enrollmentSvc.ListPublishedCourses(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "list courses failed")
		return
	}
	response.Success(c, http.StatusOK, listCoursesResp)
}

// Enroll 将当前学生加入课程。
func (h *EnrollmentHandler) Enroll(c *gin.Context) {
	courseID, parseCourseErr := parsePathID(c, "courseId")
	if parseCourseErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid course id")
		return
	}
	studentID, parseStudentErr := parseContextUserID(c, middleware.ContextUserID)
	if parseStudentErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid student id")
		return
	}
	if err := h.enrollmentSvc.Enroll(c.Request.Context(), courseID, studentID); err != nil {
		if respondPostgresUniqueViolation(c, err, "enroll") {
			return
		}
		logging.FromContext(c.Request.Context()).WithError(err).Error("enroll failed")
		response.Error(c, http.StatusInternalServerError, "选课失败，请稍后重试")
		return
	}
	c.Status(http.StatusNoContent)
}

// MyCourses 返回当前学生已选课程。
func (h *EnrollmentHandler) MyCourses(c *gin.Context) {
	studentID, parseStudentErr := parseContextUserID(c, middleware.ContextUserID)
	if parseStudentErr != nil {
		response.Error(c, http.StatusBadRequest, "invalid student id")
		return
	}
	myCoursesResp, err := h.enrollmentSvc.MyCourses(c.Request.Context(), studentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "list my courses failed")
		return
	}
	response.Success(c, http.StatusOK, myCoursesResp)
}
