package postgres

import (
	"context"

	"gorm.io/gorm"
)

// ChapterRepository 管理章节对象持久化。
type ChapterRepository struct {
	db *gorm.DB
}

func NewChapterRepository(db *gorm.DB) *ChapterRepository {
	return &ChapterRepository{db: db}
}

// CreateChapter 创建课程章节；(course_id,sort_order) 与 (course_id,title) 唯一由索引保证。
func (r *ChapterRepository) CreateChapter(ctx context.Context, courseID int64, title string, sortOrder int) (int64, error) {
	var insertedChapterID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO course_chapters(course_id, title, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id
	`, courseID, title, sortOrder).Scan(&insertedChapterID).Error
	return insertedChapterID, err
}

// UpdateChapter 更新章节标题与排序，并校验教师是否拥有该课程。
func (r *ChapterRepository) UpdateChapter(ctx context.Context, chapterID, teacherID int64, title string, sortOrder int) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		UPDATE course_chapters c
		SET title = $1, sort_order = $2, updated_at = NOW()
		FROM courses co
		WHERE c.id = $3 AND c.course_id = co.id AND co.teacher_id = $4
	`, title, sortOrder, chapterID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ReorderChapter 调整章节顺序，并校验教师是否拥有该课程。
func (r *ChapterRepository) ReorderChapter(ctx context.Context, chapterID, teacherID int64, sortOrder int) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		UPDATE course_chapters c
		SET sort_order = $1, updated_at = NOW()
		FROM courses co
		WHERE c.id = $2 AND c.course_id = co.id AND co.teacher_id = $3
	`, sortOrder, chapterID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListChaptersByCourse 查询课程下章节列表，并校验教师归属。
func (r *ChapterRepository) ListChaptersByCourse(ctx context.Context, courseID, teacherID int64) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.id::text AS id, c.course_id::text AS course_id, c.title, c.sort_order, c.created_at, c.updated_at
		FROM course_chapters c
		INNER JOIN courses co ON co.id = c.course_id
		WHERE c.course_id = $1 AND co.teacher_id = $2
		ORDER BY c.sort_order ASC, c.id ASC
	`, courseID, teacherID).Scan(&rows).Error
	return rows, err
}

// DeleteChapter 删除教师名下课程中的章节。
func (r *ChapterRepository) DeleteChapter(ctx context.Context, chapterID, teacherID int64) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		DELETE FROM course_chapters c
		USING courses co
		WHERE c.id = $1 AND c.course_id = co.id AND co.teacher_id = $2
	`, chapterID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
