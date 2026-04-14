package domain

import "time"

// Resource 表示节下的学习资源（文档/视频等）。
// CourseID、ChapterID 为冗余列；同一节内 title、sort_order 分别有唯一约束。
type Resource struct {
	ID              int64        `json:"id"`
	CourseID        int64        `json:"course_id"`
	ChapterID       int64        `json:"chapter_id"`
	SectionID       int64        `json:"section_id"`
	Title           string       `json:"title"`
	ResourceType    ResourceType `json:"resource_type"`
	SortOrder       int          `json:"sort_order"`
	ObjectKey       string       `json:"object_key"`
	ObjectURL       string       `json:"object_url"`
	MimeType        string       `json:"mime_type"`
	SizeBytes       int64        `json:"size_bytes"`
	DurationSeconds int          `json:"duration_seconds"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// ResourceLearningRecord 表示学生在资源维度的学习记录。
type ResourceLearningRecord struct {
	ID              int64     `json:"id"`
	ResourceID      int64     `json:"resource_id"`
	StudentID       int64     `json:"student_id"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
	WatchedSeconds  int       `json:"watched_seconds"`
	ProgressPercent float64   `json:"progress_percent"`
	IsCompleted     bool      `json:"is_completed"`
	UpdatedAt       time.Time `json:"updated_at"`
}
