// Package learning 存放与 LearningHandler（学生选课与学习）对应的请求/响应 DTO。
package learning

// ListCoursesResp 已发布课程列表返回。
type ListCoursesResp struct {
	Items []map[string]any `json:"items"`
}

// EnrollReq 选课入参（无请求体时由 Handler 填路径与身份）。
type EnrollReq struct {
	CourseID  int64
	StudentID int64
}

// MyCoursesReq 我的课程列表入参。
type MyCoursesReq struct {
	StudentID int64
}

// MyCoursesResp 我的课程列表返回。
type MyCoursesResp struct {
	Items []map[string]any `json:"items"`
}

// CatalogReq 课程目录查询入参。
type CatalogReq struct {
	CourseID  int64
	StudentID int64
}

// CatalogResp 课程目录（章节与资源）；结构与仓储层一致，用 map 表达树形 JSON。
type CatalogResp map[string]any

// UpdateProgressReq 更新学习进度请求体；ResourceID、StudentID 由 Handler 写入。
type UpdateProgressReq struct {
	ResourceID     int64
	StudentID      int64
	WatchedSeconds int     `json:"watched_seconds" binding:"required,min=0"`
	Progress       float64 `json:"progress_percent" binding:"required,min=0,max=100"`
}

// CompleteResourceReq 标记资源学完入参。
type CompleteResourceReq struct {
	ResourceID int64
	StudentID  int64
}
