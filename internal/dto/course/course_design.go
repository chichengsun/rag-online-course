package course

// GenerateOutlineDraftReq 请求 AI 根据课程信息生成章节/节大纲草案（仅生成 JSON，不落库）。
type GenerateOutlineDraftReq struct {
	// QAModelID 指定教师名下问答模型 id；省略则使用列表中第一个 qa 模型。
	QAModelID *int64 `json:"qa_model_id"`
	// ExtraHint 教师补充说明（先修要求、侧重方向、周课时等），可为空。
	ExtraHint string `json:"extra_hint"`
}

// OutlineSectionDraft 大纲中的「节」。
type OutlineSectionDraft struct {
	Title string `json:"title"`
}

// OutlineChapterDraft 大纲中的「章」及其下属节。
type OutlineChapterDraft struct {
	Title    string                `json:"title"`
	Sections []OutlineSectionDraft `json:"sections"`
}

// GenerateOutlineDraftResp AI 生成的大纲草案，供前端预览与二次编辑后再应用。
type GenerateOutlineDraftResp struct {
	Chapters []OutlineChapterDraft `json:"chapters"`
}

// ApplyOutlineDraftReq 将教师确认后的大纲写入课程：在现有章节之后追加新章与节。
type ApplyOutlineDraftReq struct {
	Chapters []OutlineChapterDraft `json:"chapters" binding:"required"`
}

// ApplyOutlineDraftResp 应用大纲后的创建数量统计。
type ApplyOutlineDraftResp struct {
	CreatedChapters int `json:"created_chapters"`
	CreatedSections int `json:"created_sections"`
}
