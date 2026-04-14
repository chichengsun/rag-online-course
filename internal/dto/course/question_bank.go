package course

// QuestionBankItem 课程题库题目实体（教师可手工维护，也可由 AI 解析文件后生成）。
type QuestionBankItem struct {
	ID              string `json:"id"`
	CourseID        string `json:"course_id"`
	QuestionType    string `json:"question_type"`
	Stem            string `json:"stem"`
	ReferenceAnswer string `json:"reference_answer"`
	SourceFileName  string `json:"source_file_name,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// ListQuestionBankItemsResp 题库分页/列表响应（page、page_size、total 与 items 一并返回便于前端分页）。
type ListQuestionBankItemsResp struct {
	Items    []QuestionBankItem `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// ParseQuestionBankImportResp AI 解析上传文本后的题目草稿，不入库，供教师编辑后再确认。
type ParseQuestionBankImportResp struct {
	Questions []CreateQuestionBankItemReq `json:"questions"`
}

// ConfirmQuestionBankImportReq 教师确认后的批量入库请求。
type ConfirmQuestionBankImportReq struct {
	SourceFileName string                      `json:"source_file_name"`
	Questions      []CreateQuestionBankItemReq `json:"questions" binding:"required"`
}

// ConfirmQuestionBankImportResp 批量确认入库结果。
type ConfirmQuestionBankImportResp struct {
	CreatedCount int               `json:"created_count"`
	Items        []QuestionBankItem `json:"items"`
}

// CreateQuestionBankItemReq 手工新增题目请求。
type CreateQuestionBankItemReq struct {
	QuestionType    string `json:"question_type" binding:"required"`
	Stem            string `json:"stem" binding:"required"`
	ReferenceAnswer string `json:"reference_answer" binding:"required"`
}

// CreateQuestionBankItemResp 新增题目响应。
type CreateQuestionBankItemResp struct {
	ID string `json:"id"`
}

// UpdateQuestionBankItemReq 更新题目请求。
type UpdateQuestionBankItemReq struct {
	QuestionType    string `json:"question_type" binding:"required"`
	Stem            string `json:"stem" binding:"required"`
	ReferenceAnswer string `json:"reference_answer" binding:"required"`
}

