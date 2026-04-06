package postgres

import (
	"context"

	"gorm.io/gorm"
)

// CatalogRepository 管理课程目录读取。
type CatalogRepository struct {
	db *gorm.DB
}

// NewCatalogRepository 创建目录仓储。
func NewCatalogRepository(db *gorm.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

// GetCatalog 返回课程目录，仅对已选课学生开放；资源行含 course_id 与库表一致。
func (r *CatalogRepository) GetCatalog(ctx context.Context, courseID, studentID int64) (map[string]any, error) {
	var studentHasEnrollment bool
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXISTS(
			SELECT 1 FROM course_enrollments
			WHERE course_id = $1 AND student_id = $2 AND status = 'active'
		)
	`, courseID, studentID).Scan(&studentHasEnrollment).Error
	if err != nil {
		return nil, err
	}
	if !studentHasEnrollment {
		return nil, ErrNoCourseAccess
	}

	catalog := map[string]any{"course_id": courseID, "chapters": []map[string]any{}}
	chapterRows := make([]map[string]any, 0)
	err = r.db.WithContext(ctx).Raw(`
		SELECT id::text AS id, title, sort_order
		FROM course_chapters
		WHERE course_id = $1
		ORDER BY sort_order ASC
	`, courseID).Scan(&chapterRows).Error
	if err != nil {
		return nil, err
	}

	for _, chapterRow := range chapterRows {
		chapterIDStr, errConv := scalarIDToString(chapterRow["id"])
		if errConv != nil || chapterIDStr == "" {
			return nil, ErrInvalidID
		}
		resourceRows := make([]map[string]any, 0)
		resourceScanErr := r.db.WithContext(ctx).Raw(`
			SELECT id::text AS id, course_id::text AS course_id, title, resource_type::text AS resource_type, sort_order, object_url, mime_type, size_bytes
			FROM chapter_resources
			WHERE chapter_id = $1
			ORDER BY sort_order ASC
		`, chapterIDStr).Scan(&resourceRows).Error
		if resourceScanErr != nil {
			return nil, resourceScanErr
		}
		chapterRow["resources"] = resourceRows
	}
	catalog["chapters"] = chapterRows
	return catalog, nil
}
