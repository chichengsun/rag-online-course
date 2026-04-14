package service

import (
	"context"
	"strconv"
	"strings"

	"rag-online-course/internal/repository/postgres"
)

// CourseService 负责课程对象业务编排（创建、更新）。
type CourseService struct {
	repo *postgres.CourseRepository
}

// NewCourseService 创建课程业务服务。
func NewCourseService(repo *postgres.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

// CreateCourse 创建课程并规范化课程标题。
func (s *CourseService) CreateCourse(ctx context.Context, teacherID int64, title, description string) (string, error) {
	newCourseID, err := s.repo.CreateCourse(
		ctx,
		teacherID,
		normalizeSpaces(title),
		strings.TrimSpace(description),
	)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(newCourseID, 10), nil
}

// UpdateCourse 更新课程基本信息和发布状态。
func (s *CourseService) UpdateCourse(ctx context.Context, courseID, teacherID int64, title, description, status string) error {
	return s.repo.UpdateCourse(
		ctx,
		courseID,
		teacherID,
		normalizeSpaces(title),
		strings.TrimSpace(description),
		status,
	)
}

// ListCourses 教师端分页查询课程列表，支持关键词搜索。
func (s *CourseService) ListCourses(ctx context.Context, teacherID int64, keyword, status, sortBy, sortOrder string, page, pageSize int) ([]map[string]any, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	validStatus := ""
	if status == "draft" || status == "published" || status == "archived" {
		validStatus = status
	}
	validSortBy := "created_at"
	if sortBy == "updated_at" || sortBy == "title" {
		validSortBy = sortBy
	}
	validSortOrder := "desc"
	if strings.EqualFold(sortOrder, "asc") {
		validSortOrder = "asc"
	}
	offset := (page - 1) * pageSize
	return s.repo.ListCoursesByTeacher(
		ctx,
		teacherID,
		strings.TrimSpace(keyword),
		validStatus,
		validSortBy,
		validSortOrder,
		pageSize,
		offset,
	)
}

// DeleteCourse 删除课程。
func (s *CourseService) DeleteCourse(ctx context.Context, courseID, teacherID int64) error {
	return s.repo.DeleteCourse(ctx, courseID, teacherID)
}
