package handlers

import (
	"net/http"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// CatalogHandler 管理课程目录对象相关接口。
type CatalogHandler struct {
	catalogSvc *service.CatalogService
}

// NewCatalogHandler 创建目录处理器。
func NewCatalogHandler(catalogSvc *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogSvc: catalogSvc}
}

// Catalog 返回课程目录（章节/资源），默认按排序字段升序。
func (h *CatalogHandler) Catalog(c *gin.Context) {
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
	catalogResp, err := h.catalogSvc.Catalog(c.Request.Context(), courseID, studentID)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if err == postgres.ErrNoCourseAccess {
			httpStatus = http.StatusForbidden
		}
		response.Error(c, httpStatus, err.Error())
		return
	}
	response.Success(c, http.StatusOK, catalogResp)
}
