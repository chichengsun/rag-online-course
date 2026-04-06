package service

import (
	"context"
	"strconv"

	"rag-online-course/internal/repository/postgres"
)

// ChapterService 负责章节对象业务编排。
type ChapterService struct {
	repo *postgres.ChapterRepository
}

func NewChapterService(repo *postgres.ChapterRepository) *ChapterService {
	return &ChapterService{repo: repo}
}

func (s *ChapterService) CreateChapter(ctx context.Context, courseID int64, title string, sortOrder int) (string, error) {
	newChapterID, err := s.repo.CreateChapter(ctx, courseID, title, sortOrder)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(newChapterID, 10), nil
}

func (s *ChapterService) ReorderChapter(ctx context.Context, chapterID, teacherID int64, sortOrder int) error {
	return s.repo.ReorderChapter(ctx, chapterID, teacherID, sortOrder)
}

// ListChapters 查询教师指定课程下的章节列表。
func (s *ChapterService) ListChapters(ctx context.Context, courseID, teacherID int64) ([]map[string]any, error) {
	return s.repo.ListChaptersByCourse(ctx, courseID, teacherID)
}

// DeleteChapter 删除章节。
func (s *ChapterService) DeleteChapter(ctx context.Context, chapterID, teacherID int64) error {
	return s.repo.DeleteChapter(ctx, chapterID, teacherID)
}
