package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/api/response"
	dto "rag-online-course/internal/dto/course"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"

	"github.com/gin-gonic/gin"
)

// QuestionBankHandler 教师端课程题库处理器：支持手工 CRUD 与文件解析导题。
type QuestionBankHandler struct {
	svc *service.QuestionBankService
}

// NewQuestionBankHandler 创建题库处理器。
func NewQuestionBankHandler(svc *service.QuestionBankService) *QuestionBankHandler {
	return &QuestionBankHandler{svc: svc}
}

// ListByCourse GET /teacher/courses/:courseId/question-bank
func (h *QuestionBankHandler) ListByCourse(c *gin.Context) {
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
	page := 1
	pageSize := 20
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			page = v
		}
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			pageSize = v
		}
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	questionType := strings.TrimSpace(c.Query("question_type"))
	out, err := h.svc.ListByCourse(c.Request.Context(), courseID, teacherID, page, pageSize, keyword, questionType)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "course not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "list question bank failed")
		return
	}
	response.Success(c, http.StatusOK, out)
}

// Create POST /teacher/courses/:courseId/question-bank
func (h *QuestionBankHandler) Create(c *gin.Context) {
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
	var req dto.CreateQuestionBankItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	id, err := h.svc.Create(c.Request.Context(), courseID, teacherID, req)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "course not found")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, dto.CreateQuestionBankItemResp{ID: id})
}

// Update PUT /teacher/question-bank/items/:itemId
func (h *QuestionBankHandler) Update(c *gin.Context) {
	itemID, err := parsePathID(c, "itemId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid item id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	var req dto.UpdateQuestionBankItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), itemID, teacherID, req); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "question not found")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete DELETE /teacher/question-bank/items/:itemId
func (h *QuestionBankHandler) Delete(c *gin.Context) {
	itemID, err := parsePathID(c, "itemId")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid item id")
		return
	}
	teacherID, err := parseContextUserID(c, middleware.ContextUserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid teacher id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), itemID, teacherID); err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "question not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "delete question failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// ParseImportFromFile POST /teacher/courses/:courseId/question-bank/import/parse
// 仅调用模型解析上传文本，返回题目草稿，不入库。
func (h *QuestionBankHandler) ParseImportFromFile(c *gin.Context) {
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
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "missing upload file")
		return
	}
	fp, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "open upload file failed")
		return
	}
	defer fp.Close()
	fileBytes, err := io.ReadAll(fp)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "read upload file failed")
		return
	}

	var qaModelID *int64
	if raw := strings.TrimSpace(c.PostForm("qa_model_id")); raw != "" {
		v, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "invalid qa_model_id")
			return
		}
		qaModelID = &v
	}
	out, err := h.svc.ParseImportFromFileText(c.Request.Context(), courseID, teacherID, fileHeader.Filename, fileBytes, qaModelID)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "course not found")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}

// ConfirmImportBatch POST /teacher/courses/:courseId/question-bank/import/confirm
// 将教师确认后的题目列表批量写入题库。
func (h *QuestionBankHandler) ConfirmImportBatch(c *gin.Context) {
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
	var req dto.ConfirmQuestionBankImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.svc.ConfirmImportBatch(c.Request.Context(), courseID, teacherID, req)
	if err != nil {
		if err == postgres.ErrNotFound {
			response.Error(c, http.StatusNotFound, "course not found")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, out)
}
