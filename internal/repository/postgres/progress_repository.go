package postgres

import (
	"context"

	"gorm.io/gorm"
)

// ProgressRepository 管理资源学习进度持久化。
type ProgressRepository struct {
	db *gorm.DB
}

// NewProgressRepository 创建进度仓储。
func NewProgressRepository(db *gorm.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// UpsertResourceProgress 写入或更新资源学习进度。
func (r *ProgressRepository) UpsertResourceProgress(ctx context.Context, resourceID, studentID int64, watchedSeconds int, progressPercent float64) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO resource_learning_records(resource_id, student_id, started_at, watched_seconds, progress_percent, updated_at)
		VALUES ($1, $2, NOW(), $3, $4, NOW())
		ON CONFLICT(resource_id, student_id)
		DO UPDATE SET watched_seconds = $3, progress_percent = $4, updated_at = NOW()
	`, resourceID, studentID, watchedSeconds, progressPercent).Error
}

// CompleteResource 标记资源学习完成状态。
func (r *ProgressRepository) CompleteResource(ctx context.Context, resourceID, studentID int64) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO resource_learning_records(resource_id, student_id, started_at, completed_at, watched_seconds, progress_percent, is_completed, updated_at)
		VALUES ($1, $2, NOW(), NOW(), 0, 100, TRUE, NOW())
		ON CONFLICT(resource_id, student_id)
		DO UPDATE SET completed_at = NOW(), progress_percent = 100, is_completed = TRUE, updated_at = NOW()
	`, resourceID, studentID).Error
}
