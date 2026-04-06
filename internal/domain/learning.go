package domain

import "time"

// LearningProgress 表示学生在课程维度的总体学习进度。
// LastResourceID 对应可空 BIGINT，无记录时为 nil。
type LearningProgress struct {
	ID                 int64      `json:"id"`
	CourseID           int64      `json:"course_id"`
	StudentID          int64      `json:"student_id"`
	LastResourceID     *int64     `json:"last_resource_id,omitempty"`
	CompletedResources int        `json:"completed_resources"`
	TotalResources     int        `json:"total_resources"`
	ProgressPercent    float64    `json:"progress_percent"`
	LastLearnedAt      *time.Time `json:"last_learned_at,omitempty"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
