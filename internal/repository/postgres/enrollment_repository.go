package postgres

import (
	"context"

	"gorm.io/gorm"
)

// EnrollmentRepository 管理选课对象持久化。
type EnrollmentRepository struct {
	db *gorm.DB
}

// NewEnrollmentRepository 创建选课仓储。
func NewEnrollmentRepository(db *gorm.DB) *EnrollmentRepository {
	return &EnrollmentRepository{db: db}
}

// EnrollCourse 处理学生选课，重复选课时刷新状态。
func (r *EnrollmentRepository) EnrollCourse(ctx context.Context, courseID, studentID int64) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO course_enrollments(course_id, student_id, status)
		VALUES ($1, $2, 'active')
		ON CONFLICT (course_id, student_id) DO UPDATE SET status = 'active', updated_at = NOW()
	`, courseID, studentID).Error
}

// ListMyCourses 查询学生已选课程。
func (r *EnrollmentRepository) ListMyCourses(ctx context.Context, studentID int64) ([]map[string]any, error) {
	enrollmentRows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.id::text AS id, c.title, c.description, e.enrolled_at
		FROM course_enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1 AND e.status = 'active'
		ORDER BY e.enrolled_at DESC
	`, studentID).Scan(&enrollmentRows).Error
	return enrollmentRows, err
}
