package course

// GenerateSectionLessonPlanReq 生成小节教案草案请求：目标为必填，模型可选指定。
type GenerateSectionLessonPlanReq struct {
	// QAModelID 指定教师名下问答模型；为空时自动选择第一个 qa 模型。
	QAModelID *int64 `json:"qa_model_id"`
	// Objectives 教学目标（知识/能力/情感等），至少 1 条。
	Objectives []string `json:"objectives" binding:"required,min=1,dive,required"`
	// TeachingStyle 授课风格偏好（例如：案例驱动、实操优先），可选。
	TeachingStyle string `json:"teaching_style"`
	// DurationMinutes 计划时长（分钟），0 表示不约束。
	DurationMinutes int `json:"duration_minutes"`
	// ExtraHint 补充约束（班级水平、设备条件、禁用方式等），可选。
	ExtraHint string `json:"extra_hint"`
}

// LessonPlanStep 教案步骤：阶段、时长、活动与对应资源引用。
type LessonPlanStep struct {
	Phase           string   `json:"phase"`
	DurationMinutes int      `json:"duration_minutes"`
	Activities      []string `json:"activities"`
	ResourceRefs    []string `json:"resource_refs"`
}

// LessonPlanDraft 小节教案草案正文（结构化 JSON），前端可转渲染或编辑。
type LessonPlanDraft struct {
	Title       string           `json:"title"`
	Objectives  []string         `json:"objectives"`
	Preparation []string         `json:"preparation"`
	Steps       []LessonPlanStep `json:"steps"`
	Assessment  []string         `json:"assessment"`
	Homework    []string         `json:"homework"`
}

// GenerateSectionLessonPlanResp 生成结果，含上下文与教案草案。
type GenerateSectionLessonPlanResp struct {
	CourseTitle   string          `json:"course_title"`
	ChapterTitle  string          `json:"chapter_title"`
	SectionTitle  string          `json:"section_title"`
	ResourceCount int             `json:"resource_count"`
	Plan          LessonPlanDraft `json:"plan"`
}
