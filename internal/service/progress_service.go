package service

import (
	"context"

	dto "rag-online-course/internal/dto/learning"
	"rag-online-course/internal/repository/postgres"
)

// ProgressService 负责资源学习进度对象业务编排。
type ProgressService struct {
	repo *postgres.ProgressRepository
}

// NewProgressService 创建学习进度业务服务。
func NewProgressService(repo *postgres.ProgressRepository) *ProgressService {
	return &ProgressService{repo: repo}
}

// UpdateProgress 更新学习进度。
func (s *ProgressService) UpdateProgress(ctx context.Context, updateProgressReq dto.UpdateProgressReq) error {
	return s.repo.UpsertResourceProgress(
		ctx,
		updateProgressReq.ResourceID,
		updateProgressReq.StudentID,
		updateProgressReq.WatchedSeconds,
		updateProgressReq.Progress,
	)
}

// CompleteResource 标记资源学习完成。
func (s *ProgressService) CompleteResource(ctx context.Context, completeResourceReq dto.CompleteResourceReq) error {
	return s.repo.CompleteResource(ctx, completeResourceReq.ResourceID, completeResourceReq.StudentID)
}
