// Package course 存放与 CourseHandler（教师端课程与资源）对应的请求/响应 DTO。
package course

// CreateCourseReq 创建课程请求体；TeacherID 由 Handler 从上下文写入。
type CreateCourseReq struct {
	TeacherID   int64
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// CreateCourseResp 创建课程返回。
type CreateCourseResp struct {
	// ID 对应 courses.id（BIGSERIAL），十进制字符串。
	ID string `json:"id"`
}

// ListCoursesReq 教师课程分页查询入参；TeacherID 由 Handler 注入。
type ListCoursesReq struct {
	TeacherID int64
	Page      int
	PageSize  int
}

// ListCoursesResp 教师课程分页查询返回。
type ListCoursesResp struct {
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
	Items    []map[string]any `json:"items"`
}

// UpdateCourseReq 更新课程请求体；CourseID、TeacherID 由 Handler 写入。
type UpdateCourseReq struct {
	CourseID    int64
	TeacherID   int64
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required,oneof=draft published archived"`
}

// CreateChapterReq 创建章节请求体；CourseID 由 Handler 写入。
type CreateChapterReq struct {
	CourseID  int64
	Title     string `json:"title" binding:"required"`
	SortOrder int    `json:"sort_order" binding:"required,min=1"`
}

// CreateChapterResp 创建章节返回。
type CreateChapterResp struct {
	// ID 对应 chapters.id（BIGSERIAL），十进制字符串。
	ID string `json:"id"`
}

// ListChaptersReq 章节列表查询入参；TeacherID 由 Handler 注入。
type ListChaptersReq struct {
	CourseID  int64
	TeacherID int64
}

// ListChaptersResp 章节列表查询返回。
type ListChaptersResp struct {
	Items []map[string]any `json:"items"`
}

// ReorderChapterReq 调整章节顺序请求体；ChapterID、TeacherID 由 Handler 写入。
type ReorderChapterReq struct {
	ChapterID int64
	TeacherID int64
	SortOrder int `json:"sort_order" binding:"required,min=1"`
}

// InitUploadReq 初始化资源上传请求体；ChapterID 由路径参数写入。
type InitUploadReq struct {
	CourseID     int64 `json:"course_id" binding:"required"`
	ChapterID    int64
	FileName     string `json:"file_name" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required,oneof=ppt pdf txt video doc docx audio"`
}

// InitUploadResp 预签名上传信息返回。
type InitUploadResp struct {
	ObjectKey     string `json:"object_key"`
	UploadURL     string `json:"upload_url"`
	ExpireSeconds int    `json:"expire_seconds"`
}

// ConfirmResourceReq 确认资源入库请求体；ChapterID、TeacherID 由 Handler 写入。
type ConfirmResourceReq struct {
	ChapterID    int64
	TeacherID    int64
	Title        string `json:"title" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required,oneof=ppt pdf txt video doc docx audio"`
	SortOrder    int    `json:"sort_order" binding:"required,min=1"`
	ObjectKey    string `json:"object_key" binding:"required"`
	MimeType     string `json:"mime_type" binding:"required"`
	SizeBytes    int64  `json:"size_bytes" binding:"required,min=0"`
}

// ConfirmResourceResp 确认资源返回。
type ConfirmResourceResp struct {
	// ID 对应 chapter_resources.id（BIGSERIAL），十进制字符串。
	ID string `json:"id"`
}

// ListResourcesReq 章节资源列表查询入参；TeacherID 由 Handler 注入。
type ListResourcesReq struct {
	ChapterID int64
	TeacherID int64
}

// ListResourcesResp 章节资源列表查询返回。
type ListResourcesResp struct {
	Items []map[string]any `json:"items"`
}

// PreviewResourceURLResp 预览 URL 返回。
type PreviewResourceURLResp struct {
	PreviewURL string `json:"preview_url"`
}

// ParseResourceResp 教师触发解析后的响应；Markdown 可能较大，由调用方按需展示或仅展示摘要。
type ParseResourceResp struct {
	Status     string         `json:"status"`
	Markdown   string         `json:"markdown,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Error      string         `json:"error,omitempty"`
	ImageCount int            `json:"image_count"`
}

// ReorderResourceReq 调整资源顺序请求体；ResourceID、TeacherID 由 Handler 写入。
type ReorderResourceReq struct {
	ResourceID int64
	TeacherID  int64
	SortOrder  int `json:"sort_order" binding:"required,min=1"`
}
