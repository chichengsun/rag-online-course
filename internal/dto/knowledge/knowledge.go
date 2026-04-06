// Package knowledge 存放课程知识库分块与嵌入相关 DTO。
package knowledge

// ListKnowledgeResourcesResp 课程下可解析资源分页列表（含分块统计）。
type ListKnowledgeResourcesResp struct {
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
	Items    []map[string]any `json:"items"`
}

// ChunkPreviewReq 分块预览参数（基于 docreader 解析后的 Markdown）。
type ChunkPreviewReq struct {
	ChunkSize int `json:"chunk_size" binding:"required,min=1,max=32000"`
	Overlap   int `json:"overlap" binding:"min=0"`
	// ClearPersistedFirst 为 true 时先删除该资源下已落库的分块与向量，再解析预览，避免与历史持久化结果混用。
	ClearPersistedFirst bool `json:"clear_persisted_first"`
}

// UpdateKnowledgeChunkReq 更新已保存分块正文；会清空该条向量与确认状态以便重新确认/嵌入。
type UpdateKnowledgeChunkReq struct {
	Content   string `json:"content" binding:"required"`
	CharStart *int   `json:"char_start,omitempty"`
	CharEnd   *int   `json:"char_end,omitempty"`
}

// ChunkPreviewSegment 预览中的单个分块。
type ChunkPreviewSegment struct {
	Index     int    `json:"index"`
	Content   string `json:"content"`
	CharStart int    `json:"char_start"`
	CharEnd   int    `json:"char_end"`
}

// ChunkPreviewResp 分块预览结果。
type ChunkPreviewResp struct {
	Segments []ChunkPreviewSegment `json:"segments"`
}

// ChunkSaveItem 落库分块一行；chunk_index 由服务端按数组顺序从 0 递增写入。
type ChunkSaveItem struct {
	Content   string         `json:"content" binding:"required"`
	CharStart *int           `json:"char_start,omitempty"`
	CharEnd   *int           `json:"char_end,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// SaveKnowledgeChunksReq 覆盖保存某资源下全部分块（会删除旧分块与向量）；chunks 为空表示清空该资源全部分块。
type SaveKnowledgeChunksReq struct {
	Chunks []ChunkSaveItem `json:"chunks"`
}

// EmbedResourceReq 对已确认且未嵌入的分块调用外部嵌入 API。
type EmbedResourceReq struct {
	ModelID int64 `json:"model_id" binding:"required"`
}

// EmbedResourceResp 嵌入任务摘要。
type EmbedResourceResp struct {
	EmbeddedCount int `json:"embedded_count"`
}

// ConfirmKnowledgeChunksResp 确认分块后的行数。
type ConfirmKnowledgeChunksResp struct {
	ConfirmedCount int64 `json:"confirmed_count"`
}

// ListKnowledgeChunksResp 已保存分块列表。
type ListKnowledgeChunksResp struct {
	Items []map[string]any `json:"items"`
}
