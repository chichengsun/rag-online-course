package service

import (
	"context"
	"strconv"

	"rag-online-course/internal/repository/postgres"
)

// SectionService 负责「节」对象业务编排（隶属于章节，其下挂资源）。
type SectionService struct {
	repo *postgres.SectionRepository
}

// NewSectionService 创建节服务。
func NewSectionService(repo *postgres.SectionRepository) *SectionService {
	return &SectionService{repo: repo}
}

// CreateSection 在章节下新增节（校验教师对课程的归属）。
func (s *SectionService) CreateSection(ctx context.Context, chapterID, courseID, teacherID int64, title string, sortOrder int) (string, error) {
	id, err := s.repo.CreateSectionForTeacher(ctx, chapterID, courseID, teacherID, title, sortOrder)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

// ListSections 查询教师指定章节下的节列表。
func (s *SectionService) ListSections(ctx context.Context, chapterID, teacherID int64) ([]map[string]any, error) {
	return s.repo.ListSectionsByChapter(ctx, chapterID, teacherID)
}

// ReorderSection 调整节顺序。
func (s *SectionService) ReorderSection(ctx context.Context, sectionID, teacherID int64, sortOrder int) error {
	return s.repo.ReorderSection(ctx, sectionID, teacherID, sortOrder)
}

// UpdateSection 更新节标题与排序。
func (s *SectionService) UpdateSection(ctx context.Context, sectionID, teacherID int64, title string, sortOrder int) error {
	return s.repo.UpdateSection(ctx, sectionID, teacherID, title, sortOrder)
}

// DeleteSection 删除节（节下资源随数据库外键级联删除）。
func (s *SectionService) DeleteSection(ctx context.Context, sectionID, teacherID int64) error {
	return s.repo.DeleteSection(ctx, sectionID, teacherID)
}
