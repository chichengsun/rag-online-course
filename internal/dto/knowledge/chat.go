package knowledge

// CreateChatSessionReq 创建知识库对话会话；title 为空时由后端按首问自动生成。
type CreateChatSessionReq struct {
	Title string `json:"title"`
}

// CreateChatSessionResp 创建会话返回。
type CreateChatSessionResp struct {
	ID string `json:"id"`
}

// UpdateChatSessionReq 更新会话标题。
type UpdateChatSessionReq struct {
	Title string `json:"title" binding:"required,max=256"`
}

// ListChatSessionsResp 会话分页列表。
type ListChatSessionsResp struct {
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
	Items    []map[string]any `json:"items"`
}

// ListChatMessagesResp 会话消息分页列表（按创建时间升序）。
type ListChatMessagesResp struct {
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
	Items    []map[string]any `json:"items"`
}

// AskInSessionReq 在已有会话中继续提问。
type AskInSessionReq struct {
	Question  string `json:"question" binding:"required"`
	TopK      int    `json:"top_k"`
	UseRerank *bool  `json:"use_rerank"`
	// QAModelID 可选指定问答模型 ID；未传则走默认 qa 模型。
	QAModelID *int64 `json:"qa_model_id"`
	// SemanticMinScore 语义检索最小分数阈值（0~1）。
	SemanticMinScore *float64 `json:"semantic_min_score"`
	// KeywordMinScore 关键词检索最小分数阈值（0~1）。
	KeywordMinScore *float64 `json:"keyword_min_score"`
}

// ReferenceItem 回答引用的来源分块。
type ReferenceItem struct {
	// CitationNo 为回答中的引用序号（如 [3]）。
	CitationNo   int     `json:"citation_no"`
	ChunkID      string  `json:"chunk_id"`
	ResourceID   string  `json:"resource_id"`
	ResourceTitle string `json:"resource_title"`
	ChunkIndex   int     `json:"chunk_index"`
	Score        float64 `json:"score"`
	Snippet      string  `json:"snippet"`
	// FullContent 为分块完整内容，用于“查看全文”。
	FullContent  string  `json:"full_content"`
}

// AskInSessionResp 一次问答返回。
type AskInSessionResp struct {
	SessionID          string          `json:"session_id"`
	UserMessageID      string          `json:"user_message_id"`
	AssistantMessageID string          `json:"assistant_message_id"`
	Answer             string          `json:"answer"`
	References         []ReferenceItem `json:"references"`
}
