package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// CourseRepository 管理课程对象持久化。
type CourseRepository struct {
	db *gorm.DB
}

// NewCourseRepository 创建课程仓储（直接注入 *gorm.DB）。
func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

// CreateCourse 创建课程；同一教师下课程标题唯一由联合索引保证。
func (r *CourseRepository) CreateCourse(ctx context.Context, teacherID int64, title, description string) (int64, error) {
	var insertedCourseID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO courses(teacher_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, teacherID, title, description).Scan(&insertedCourseID).Error
	return insertedCourseID, err
}

// UpdateCourse 按教师归属更新课程，防止越权修改。
func (r *CourseRepository) UpdateCourse(ctx context.Context, courseID, teacherID int64, title, description, status string) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		UPDATE courses
		SET title = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4 AND teacher_id = $5
	`, title, description, status, courseID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListPublishedCourses 返回学生可见的发布课程。
func (r *CourseRepository) ListPublishedCourses(ctx context.Context) ([]map[string]any, error) {
	courseRows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT id::text AS id, title, description, cover_image_url
		FROM courses
		WHERE status = 'published'
		ORDER BY created_at DESC
	`).Scan(&courseRows).Error
	return courseRows, err
}

// ListCoursesByTeacher 分页查询教师名下课程列表，并支持关键词/状态筛选与排序。
func (r *CourseRepository) ListCoursesByTeacher(ctx context.Context, teacherID int64, keyword, status, sortBy, sortOrder string, limit, offset int) ([]map[string]any, int64, error) {
	rows := make([]map[string]any, 0)
	var total int64
	searchKeyword := "%" + keyword + "%"
	whereClause := "teacher_id = $1 AND ($2 = '%%' OR title ILIKE $2 OR description ILIKE $2)"
	args := []any{teacherID, searchKeyword}
	if status != "" {
		whereClause += " AND status::text = $3"
		args = append(args, status)
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, strings.ToUpper(sortOrder))
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM courses WHERE %s", whereClause)
	totalErr := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error
	if totalErr != nil {
		return nil, 0, totalErr
	}
	listSQL := fmt.Sprintf(`
		SELECT id::text AS id, title, description, status::text AS status, cover_image_url, created_at, updated_at
		FROM courses
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, len(args)+1, len(args)+2)
	listArgs := append(args, limit, offset)
	listErr := r.db.WithContext(ctx).Raw(listSQL, listArgs...).Scan(&rows).Error
	if listErr != nil {
		return nil, 0, listErr
	}
	return rows, total, nil
}

// GetCourseForTeacher 读取教师名下课程标题与描述，用于课程设计等场景校验归属并拼装 AI 提示。
func (r *CourseRepository) GetCourseForTeacher(ctx context.Context, courseID, teacherID int64) (title string, description string, err error) {
	var row struct {
		Title       string
		Description *string
	}
	q := r.db.WithContext(ctx).Table("courses").
		Select("title", "description").
		Where("id = ? AND teacher_id = ?", courseID, teacherID).
		Take(&row)
	if q.Error != nil {
		if errors.Is(q.Error, gorm.ErrRecordNotFound) {
			return "", "", ErrNotFound
		}
		return "", "", q.Error
	}
	if row.Description != nil {
		description = strings.TrimSpace(*row.Description)
	}
	return strings.TrimSpace(row.Title), description, nil
}

// DeleteCourse 删除教师名下课程。
func (r *CourseRepository) DeleteCourse(ctx context.Context, courseID, teacherID int64) error {
	dbResult := r.db.WithContext(ctx).Exec(`
		DELETE FROM courses
		WHERE id = $1 AND teacher_id = $2
	`, courseID, teacherID)
	if dbResult.Error != nil {
		return dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
