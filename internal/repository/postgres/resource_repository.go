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

// ConfirmResource 仅允许课程所属教师确认资源入库。
func (r *ResourceRepository) ConfirmResource(ctx context.Context, chapterID, teacherID int64, title, resourceType, objectKey, objectURL, mimeType string, sizeBytes int64, sortOrder int) (int64, error) {
	var insertedResourceID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO chapter_resources(chapter_id, course_id, title, resource_type, sort_order, object_key, object_url, mime_type, size_bytes)
		SELECT c.id, c.course_id, $2, $3::resource_type, $4, $5, $6, $7, $8
		FROM course_chapters c
		INNER JOIN courses co ON co.id = c.course_id
		WHERE c.id = $1 AND co.teacher_id = $9
		RETURNING chapter_resources.id
	`, chapterID, title, resourceType, sortOrder, objectKey, objectURL, mimeType, sizeBytes, teacherID).Scan(&insertedResourceID).Error
	return insertedResourceID, err
}

// ValidateUploadScope 校验章节、课程、教师归属是否一致，避免跨课程上传。
func (r *ResourceRepository) ValidateUploadScope(ctx context.Context, chapterID, courseID, teacherID int64) error {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(1)
		FROM course_chapters c
		INNER JOIN courses co ON co.id = c.course_id
		WHERE c.id = $1 AND c.course_id = $2 AND co.teacher_id = $3
	`, chapterID, courseID, teacherID).Scan(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// ReorderResource 调整资源顺序，并校验教师权限。
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

// ListResourcesByChapter 查询章节下资源列表，并校验教师归属。
func (r *ResourceRepository) ListResourcesByChapter(ctx context.Context, chapterID, teacherID int64) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			r.id::text AS id,
			r.chapter_id::text AS chapter_id,
			r.course_id::text AS course_id,
			r.title,
			r.resource_type::text AS resource_type,
			r.sort_order,
			r.object_key,
			r.object_url,
			r.mime_type,
			r.size_bytes,
			r.duration_seconds,
			r.created_at,
			r.updated_at
		FROM chapter_resources r
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE r.chapter_id = $1 AND co.teacher_id = $2
		ORDER BY r.sort_order ASC, r.id ASC
	`, chapterID, teacherID).Scan(&rows).Error
	return rows, err
}

// GetResourceByID 查询教师拥有的资源详情；用于生成预览用 URL（如 Office -> PDF 转换）。
func (r *ResourceRepository) GetResourceByID(ctx context.Context, resourceID, teacherID int64) (map[string]any, error) {
	type row struct {
		ObjectKey    string
		ObjectURL    string
		ResourceType string
		MimeType     string
	}

	var rr row
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			r.object_key,
			r.object_url,
			r.resource_type::text AS resource_type,
			r.mime_type
		FROM chapter_resources r
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE r.id = $1 AND co.teacher_id = $2
	`, resourceID, teacherID).Scan(&rr).Error
	if err != nil {
		return nil, err
	}
	if rr.ObjectKey == "" {
		return nil, ErrNotFound
	}

	return map[string]any{
		"object_key":    rr.ObjectKey,
		"object_url":    rr.ObjectURL,
		"resource_type": rr.ResourceType,
		"mime_type":     rr.MimeType,
	}, nil
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
