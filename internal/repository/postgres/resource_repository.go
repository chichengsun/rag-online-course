package postgres

import (
	"context"

	"gorm.io/gorm"
)

// ResourceRepository 管理资源对象持久化。
type ResourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

// ConfirmResource 仅允许课程所属教师在指定节下确认资源入库。
func (r *ResourceRepository) ConfirmResource(ctx context.Context, sectionID, teacherID int64, title, resourceType, objectKey, objectURL, mimeType string, sizeBytes int64, sortOrder int) (int64, error) {
	var insertedResourceID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO chapter_resources(chapter_id, course_id, section_id, title, resource_type, sort_order, object_key, object_url, mime_type, size_bytes)
		SELECT c.id, s.course_id, s.id, $2, $3::resource_type, $4, $5, $6, $7, $8
		FROM course_sections s
		INNER JOIN course_chapters c ON c.id = s.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE s.id = $1 AND co.teacher_id = $9
		RETURNING chapter_resources.id
	`, sectionID, title, resourceType, sortOrder, objectKey, objectURL, mimeType, sizeBytes, teacherID).Scan(&insertedResourceID).Error
	return insertedResourceID, err
}

// ValidateUploadScope 校验节、课程、教师归属是否一致，避免跨课程上传。
func (r *ResourceRepository) ValidateUploadScope(ctx context.Context, sectionID, courseID, teacherID int64) error {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(1)
		FROM course_sections s
		INNER JOIN course_chapters c ON c.id = s.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE s.id = $1 AND s.course_id = $2 AND co.teacher_id = $3
	`, sectionID, courseID, teacherID).Scan(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateResourceTitle 更新资源标题，并校验教师对资源的归属。
func (r *ResourceRepository) UpdateResourceTitle(ctx context.Context, resourceID, teacherID int64, title string) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		UPDATE chapter_resources r
		SET title = $1, updated_at = NOW()
		FROM course_chapters c
		JOIN courses co ON c.course_id = co.id
		WHERE r.id = $2 AND r.chapter_id = c.id AND co.teacher_id = $3
	`, title, resourceID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ReorderResource 调整资源顺序，并校验教师权限（排序在节内唯一）。
func (r *ResourceRepository) ReorderResource(ctx context.Context, resourceID, teacherID int64, sortOrder int) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		UPDATE chapter_resources r
		SET sort_order = $1, updated_at = NOW()
		FROM course_chapters c
		JOIN courses co ON c.course_id = co.id
		WHERE r.id = $2 AND r.chapter_id = c.id AND co.teacher_id = $3
	`, sortOrder, resourceID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListResourcesBySection 查询节下资源列表，并校验教师归属。
func (r *ResourceRepository) ListResourcesBySection(ctx context.Context, sectionID, teacherID int64) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			r.id::text AS id,
			r.chapter_id::text AS chapter_id,
			r.section_id::text AS section_id,
			r.course_id::text AS course_id,
			r.title,
			r.resource_type::text AS resource_type,
			r.sort_order,
			r.object_key,
			r.object_url,
			r.mime_type,
			r.size_bytes,
			r.duration_seconds,
			r.ai_summary,
			r.ai_summary_updated_at,
			r.ai_summary_status,
			r.ai_summary_error,
			r.created_at,
			r.updated_at
		FROM chapter_resources r
		INNER JOIN course_sections s ON s.id = r.section_id
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE r.section_id = $1 AND co.teacher_id = $2
		ORDER BY r.sort_order ASC, r.id ASC
	`, sectionID, teacherID).Scan(&rows).Error
	return rows, err
}

// GetResourceByID 查询教师名下单条资源的完整元数据（含 AI 摘要与任务状态），供详情接口与解析/摘要服务读取。
func (r *ResourceRepository) GetResourceByID(ctx context.Context, resourceID, teacherID int64) (map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			r.id::text AS id,
			r.course_id::text AS course_id,
			r.chapter_id::text AS chapter_id,
			r.section_id::text AS section_id,
			r.title,
			r.resource_type::text AS resource_type,
			r.sort_order,
			r.object_key,
			r.object_url,
			r.mime_type,
			r.size_bytes,
			r.duration_seconds,
			r.ai_summary,
			r.ai_summary_updated_at,
			r.ai_summary_status,
			r.ai_summary_error,
			r.created_at,
			r.updated_at
		FROM chapter_resources r
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE r.id = $1 AND co.teacher_id = $2
	`, resourceID, teacherID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNotFound
	}
	out := rows[0]
	if toString(out["object_key"]) == "" {
		return nil, ErrNotFound
	}
	return out, nil
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

// UpdateResourceAISummary 写入摘要并标记为 succeeded，清空错误信息。
func (r *ResourceRepository) UpdateResourceAISummary(ctx context.Context, resourceID, teacherID int64, summary string) (updatedAt string, err error) {
	rows := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Raw(`
		UPDATE chapter_resources r
		SET ai_summary = $1,
		    ai_summary_updated_at = NOW(),
		    ai_summary_status = 'succeeded',
		    ai_summary_error = NULL,
		    updated_at = NOW()
		FROM course_chapters c
		JOIN courses co ON c.course_id = co.id
		WHERE r.id = $2 AND r.chapter_id = c.id AND co.teacher_id = $3
		RETURNING r.ai_summary_updated_at::text AS updated_at
	`, summary, resourceID, teacherID).Scan(&rows).Error; err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", ErrNotFound
	}
	return toString(rows[0]["updated_at"]), nil
}

// TryMarkAISummaryRunning 仅在当前状态不是 running 时置为 running，用于避免重复后台任务。
func (r *ResourceRepository) TryMarkAISummaryRunning(ctx context.Context, resourceID, teacherID int64) (started bool, err error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE chapter_resources r
		SET ai_summary_status = 'running', ai_summary_error = NULL, updated_at = NOW()
		FROM course_chapters c
		JOIN courses co ON c.course_id = co.id
		WHERE r.id = $1 AND r.chapter_id = c.id AND co.teacher_id = $2
		  AND r.ai_summary_status <> 'running'
	`, resourceID, teacherID)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// SetAISummaryJobFailed 将 running（或其它）任务标记为 failed 并记录错误文案。
func (r *ResourceRepository) SetAISummaryJobFailed(ctx context.Context, resourceID, teacherID int64, errMsg string) error {
	if len(errMsg) > 2000 {
		errMsg = errMsg[:2000] + "…"
	}
	res := r.db.WithContext(ctx).Exec(`
		UPDATE chapter_resources r
		SET ai_summary_status = 'failed',
		    ai_summary_error = $1,
		    updated_at = NOW()
		FROM course_chapters c
		JOIN courses co ON c.course_id = co.id
		WHERE r.id = $2 AND r.chapter_id = c.id AND co.teacher_id = $3
	`, errMsg, resourceID, teacherID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteResource 删除教师名下课程中的资源。
func (r *ResourceRepository) DeleteResource(ctx context.Context, resourceID, teacherID int64) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		DELETE FROM chapter_resources r
		USING course_chapters c, courses co
		WHERE r.id = $1 AND r.chapter_id = c.id AND c.course_id = co.id AND co.teacher_id = $2
	`, resourceID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
