package domain

type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleTeacher UserRole = "teacher"
)

type CourseStatus string

const (
	CourseDraft     CourseStatus = "draft"
	CoursePublished CourseStatus = "published"
	CourseArchived  CourseStatus = "archived"
)

type ResourceType string

const (
	ResourcePPT   ResourceType = "ppt"
	ResourcePDF   ResourceType = "pdf"
	ResourceTXT   ResourceType = "txt"
	ResourceVideo ResourceType = "video"
	ResourceDOC   ResourceType = "doc"
	ResourceDOCX  ResourceType = "docx"
	ResourceAudio ResourceType = "audio"
)

type EnrollmentStatus string

const (
	EnrollActive    EnrollmentStatus = "active"
	EnrollDropped   EnrollmentStatus = "dropped"
	EnrollCompleted EnrollmentStatus = "completed"
)
