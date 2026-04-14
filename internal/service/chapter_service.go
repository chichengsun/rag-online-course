package service

import (
	"context"
	"strconv"

	"rag-online-course/internal/repository/postgres"
)

// ChapterService 负责章节对象业务编排。
type ChapterService struct {
	repo        *postgres.ChapterRepository
	sectionRepo *postgres.SectionRepository
}

// NewChapterService 创建章节服务；创建章节后会自动插入「默认节」以便资源有挂载点。
func NewChapterService(repo *postgres.ChapterRepository, sectionRepo *postgres.SectionRepository) *ChapterService {
	return &ChapterService{repo: repo, sectionRepo: sectionRepo}
}

func (s *ChapterService) CreateChapter(ctx context.Context, courseID int64, title string, sortOrder int) (string, error) {
	newChapterID, err := s.repo.CreateChapter(ctx, courseID, title, sortOrder)
	if err != nil {
		return "", err
	}
	if _, secErr := s.sectionRepo.CreateSection(ctx, newChapterID, courseID, "默认节", 1); secErr != nil {
		return "", secErr
	}
	return strconv.FormatInt(newChapterID, 10), nil
}

func (s *ChapterService) ReorderChapter(ctx context.Context, chapterID, teacherID int64, sortOrder int) error {
	return s.repo.ReorderChapter(ctx, chapterID, teacherID, sortOrder)
}

// UpdateChapter 更新章节标题与排序（教师归属校验在仓储层完成）。
func (s *ChapterService) UpdateChapter(ctx context.Context, chapterID, teacherID int64, title string, sortOrder int) error {
	return s.repo.UpdateChapter(ctx, chapterID, teacherID, title, sortOrder)
}

// ListChapters 查询教师指定课程下的章节列表。
func (s *ChapterService) ListChapters(ctx context.Context, courseID, teacherID int64) ([]map[string]any, error) {
	return s.repo.ListChaptersByCourse(ctx, courseID, teacherID)
}

// DeleteChapter 删除章节。
func (s *ChapterService) DeleteChapter(ctx context.Context, chapterID, teacherID int64) error {
	return s.repo.DeleteChapter(ctx, chapterID, teacherID)
}
