package postgres

import (
	"context"

	"gorm.io/gorm"
)

// SectionRepository 管理「节」对象持久化；节隶属于章节，资源挂在节下。
type SectionRepository struct {
	db *gorm.DB
}

// NewSectionRepository 创建节仓储。
func NewSectionRepository(db *gorm.DB) *SectionRepository {
	return &SectionRepository{db: db}
}

// CreateSection 在指定章节下创建节（调用方已保证章节有效，例如章节刚写入后的默认节）。
func (r *SectionRepository) CreateSection(ctx context.Context, chapterID, courseID int64, title string, sortOrder int) (int64, error) {
	var insertedID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO course_sections(chapter_id, course_id, title, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, chapterID, courseID, title, sortOrder).Scan(&insertedID).Error
	return insertedID, err
}

// CreateSectionForTeacher 创建节并校验章节属于该教师名下课程，避免跨课程篡改章节 ID。
func (r *SectionRepository) CreateSectionForTeacher(ctx context.Context, chapterID, courseID, teacherID int64, title string, sortOrder int) (int64, error) {
	var insertedID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO course_sections(chapter_id, course_id, title, sort_order)
		SELECT c.id, c.course_id, $3, $4
		FROM course_chapters c
		INNER JOIN courses co ON co.id = c.course_id
		WHERE c.id = $1 AND c.course_id = $2 AND co.teacher_id = $5
		RETURNING id
	`, chapterID, courseID, title, sortOrder, teacherID).Scan(&insertedID).Error
	if err != nil {
		return 0, err
	}
	if insertedID == 0 {
		return 0, ErrNotFound
	}
	return insertedID, nil
}

// ListSectionsByChapter 查询章节下节列表，并校验课程归属教师。
func (r *SectionRepository) ListSectionsByChapter(ctx context.Context, chapterID, teacherID int64) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT s.id::text AS id, s.chapter_id::text AS chapter_id, s.course_id::text AS course_id,
		       s.title, s.sort_order, s.created_at, s.updated_at
		FROM course_sections s
		INNER JOIN course_chapters c ON c.id = s.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE s.chapter_id = $1 AND co.teacher_id = $2
		ORDER BY s.sort_order ASC, s.id ASC
	`, chapterID, teacherID).Scan(&rows).Error
	return rows, err
}

// UpdateSection 更新节标题与排序，并校验教师是否拥有该课程。
func (r *SectionRepository) UpdateSection(ctx context.Context, sectionID, teacherID int64, title string, sortOrder int) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		UPDATE course_sections s
		SET title = $1, sort_order = $2, updated_at = NOW()
		FROM course_chapters c
		JOIN courses co ON c.course_id = co.id
		WHERE s.id = $3 AND s.chapter_id = c.id AND co.teacher_id = $4
	`, title, sortOrder, sectionID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ReorderSection 调整节顺序，并校验教师是否拥有该课程。
func (r *SectionRepository) ReorderSection(ctx context.Context, sectionID, teacherID int64, sortOrder int) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		UPDATE course_sections s
		SET sort_order = $1, updated_at = NOW()
		FROM course_chapters c
		JOIN courses co ON c.course_id = co.id
		WHERE s.id = $2 AND s.chapter_id = c.id AND co.teacher_id = $3
	`, sortOrder, sectionID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSection 删除教师名下章节中的节（节下资源随外键级联删除）。
func (r *SectionRepository) DeleteSection(ctx context.Context, sectionID, teacherID int64) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		DELETE FROM course_sections s
		USING course_chapters c, courses co
		WHERE s.id = $1 AND s.chapter_id = c.id AND c.course_id = co.id AND co.teacher_id = $2
	`, sectionID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetSectionScope 返回节所属 chapter_id、course_id，并校验教师归属；用于资源上传作用域校验。
func (r *SectionRepository) GetSectionScope(ctx context.Context, sectionID, courseID, teacherID int64) (chapterID int64, err error) {
	var chID int64
	err = r.db.WithContext(ctx).Raw(`
		SELECT s.chapter_id
		FROM course_sections s
		INNER JOIN course_chapters c ON c.id = s.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE s.id = $1 AND s.course_id = $2 AND co.teacher_id = $3
	`, sectionID, courseID, teacherID).Scan(&chID).Error
	if err != nil {
		return 0, err
	}
	if chID == 0 {
		return 0, ErrNotFound
	}
	return chID, nil
}

// GetSectionDetail 查询教师可访问的小节所属课程/章节/小节标题，供教案生成拼装上下文。
func (r *SectionRepository) GetSectionDetail(ctx context.Context, sectionID, teacherID int64) (map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			co.id::text AS course_id,
			co.title AS course_title,
			c.id::text AS chapter_id,
			c.title AS chapter_title,
			s.id::text AS section_id,
			s.title AS section_title
		FROM course_sections s
		INNER JOIN course_chapters c ON c.id = s.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE s.id = $1 AND co.teacher_id = $2
		LIMIT 1
	`, sectionID, teacherID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNotFound
	}
	return rows[0], nil
}
