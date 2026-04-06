package domain

import "time"

// Course 表示教师创建的课程。ID、TeacherID 对应 BIGINT/BIGSERIAL。
type Course struct {
	ID          int64        `json:"id"`
	TeacherID   int64        `json:"teacher_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      CourseStatus `json:"status"`
	CoverURL    string       `json:"cover_image_url"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Chapter 表示课程下的章节；同一课程内 sort_order、title 分别有唯一约束。
type Chapter struct {
	ID        int64     `json:"id"`
	CourseID  int64     `json:"course_id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Enrollment 表示学生选课关系。
type Enrollment struct {
	ID        int64              `json:"id"`
	CourseID  int64              `json:"course_id"`
	StudentID int64              `json:"student_id"`
	Status    EnrollmentStatus   `json:"status"`
	Enrolled  time.Time          `json:"enrolled_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}
