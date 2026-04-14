package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"rag-online-course/internal/dto/knowledge"
	"rag-online-course/internal/pkg/chunktext"
	"rag-online-course/internal/repository/postgres"

	"github.com/sirupsen/logrus"
)

// KnowledgeService 编排课程知识库：解析预览分块、落库、确认与向量嵌入。
type KnowledgeService struct {
	chunkRepo  *postgres.EmbeddingChunkRepository
	parseSvc   *ResourceParseService
	modelRepo  *postgres.TeacherAIModelRepository
	httpClient *http.Client
}

// NewKnowledgeService 构造知识库服务；httpClient 为空时使用 http.DefaultClient。
func NewKnowledgeService(
	chunkRepo *postgres.EmbeddingChunkRepository,
	parseSvc *ResourceParseService,
	modelRepo *postgres.TeacherAIModelRepository,
	httpClient *http.Client,
) *KnowledgeService {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &KnowledgeService{
		chunkRepo:  chunkRepo,
		parseSvc:   parseSvc,
		modelRepo:  modelRepo,
		httpClient: httpClient,
	}
}

func clampKnowledgePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// ListKnowledgeResources 分页返回课程下可解析文档类资源及分块统计。
func (s *KnowledgeService) ListKnowledgeResources(ctx context.Context, courseID, teacherID int64, page, pageSize int) (*knowledge.ListKnowledgeResourcesResp, error) {
	page, pageSize = clampKnowledgePage(page, pageSize)
	offset := (page - 1) * pageSize
	items, total, err := s.chunkRepo.ListByCourse(ctx, courseID, teacherID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &knowledge.ListKnowledgeResourcesResp{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	}, nil
}

// ChunkPreview 调用解析服务得到 Markdown 后按参数切分；clearPersistedFirst 为 true 时先清空该资源已落库分块与向量，再生成预览（不落库）。
func (s *KnowledgeService) ChunkPreview(ctx context.Context, resourceID, teacherID int64, chunkSize, overlap int, clearPersistedFirst bool) (*knowledge.ChunkPreviewResp, error) {
	if clearPersistedFirst {
		if err := s.chunkRepo.ClearAllChunksForResource(ctx, resourceID, teacherID); err != nil {
			return nil, err
		}
	}
	parseResp, err := s.parseSvc.ParseResource(ctx, resourceID, teacherID)
	if err != nil {
		return nil, err
	}
	if parseResp.Status != "ok" {
		msg := strings.TrimSpace(parseResp.Error)
		if msg == "" {
			msg = "解析未成功"
		}
		return nil, fmt.Errorf("%s", msg)
	}
	segments := chunktext.Split(parseResp.Markdown, chunkSize, overlap)
	out := make([]knowledge.ChunkPreviewSegment, 0, len(segments))
	for _, seg := range segments {
		out = append(out, knowledge.ChunkPreviewSegment{
			Index:     seg.Index,
			Content:   seg.Content,
			CharStart: seg.CharStart,
			CharEnd:   seg.CharEnd,
		})
	}
	return &knowledge.ChunkPreviewResp{Segments: out}, nil
}

// ClearKnowledgeChunks 删除该资源下全部分块与向量（教师归属）。
func (s *KnowledgeService) ClearKnowledgeChunks(ctx context.Context, resourceID, teacherID int64) error {
	return s.chunkRepo.ClearAllChunksForResource(ctx, resourceID, teacherID)
}

// UpdateKnowledgeChunk 更新单条已保存分块；会清空该条向量与确认状态。
func (s *KnowledgeService) UpdateKnowledgeChunk(ctx context.Context, resourceID, teacherID, chunkID int64, req knowledge.UpdateKnowledgeChunkReq) error {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return fmt.Errorf("分块内容不能为空")
	}
	return s.chunkRepo.UpdateChunkContent(ctx, chunkID, resourceID, teacherID, content, req.CharStart, req.CharEnd)
}

// DeleteKnowledgeChunk 删除单条已保存分块并重排 chunk_index。
func (s *KnowledgeService) DeleteKnowledgeChunk(ctx context.Context, resourceID, teacherID, chunkID int64) error {
	return s.chunkRepo.DeleteChunkAndRenumber(ctx, chunkID, resourceID, teacherID)
}

// SaveKnowledgeChunks 覆盖写入资源分块（删除旧数据含已嵌入向量）。
func (s *KnowledgeService) SaveKnowledgeChunks(ctx context.Context, resourceID, teacherID int64, chunks []knowledge.ChunkSaveItem) error {
	dbChunks := make([]struct {
		Index    int
		Content  string
		CharS    *int
		CharE    *int
		MetaJSON []byte
	}, 0, len(chunks))
	for i, ch := range chunks {
		content := strings.TrimSpace(ch.Content)
		if content == "" {
			return fmt.Errorf("分块 %d 内容不能为空", i)
		}
		dbChunks = append(dbChunks, struct {
			Index    int
			Content  string
			CharS    *int
			CharE    *int
			MetaJSON []byte
		}{
			Index:    i,
			Content:  content,
			CharS:    ch.CharStart,
			CharE:    ch.CharEnd,
			MetaJSON: postgres.MetaJSONBytes(ch.Metadata),
		})
	}
	return s.chunkRepo.ReplaceDraftChunks(ctx, resourceID, teacherID, dbChunks)
}

// ConfirmKnowledgeChunks 将当前未确认且未嵌入的分块标记为已确认，供后续嵌入。
func (s *KnowledgeService) ConfirmKnowledgeChunks(ctx context.Context, resourceID, teacherID int64) (int64, error) {
	return s.chunkRepo.ConfirmDraftChunks(ctx, resourceID, teacherID)
}

// ListKnowledgeChunks 列出资源下已保存分块。
func (s *KnowledgeService) ListKnowledgeChunks(ctx context.Context, resourceID, teacherID int64) ([]map[string]any, error) {
	return s.chunkRepo.ListChunksForResource(ctx, resourceID, teacherID)
}

// EmbedResource 对已确认、未嵌入分块批量调用 OpenAI 兼容 embeddings 接口并写回 pgvector。
func (s *KnowledgeService) EmbedResource(ctx context.Context, resourceID, teacherID, modelID int64) (*knowledge.EmbedResourceResp, error) {
	m, err := s.modelRepo.GetByID(ctx, modelID, teacherID)
	if err != nil {
		return nil, err
	}
	if m.ModelType != "embedding" {
		return nil, fmt.Errorf("所选模型类型不是 embedding")
	}
	if strings.TrimSpace(m.APIKey) == "" {
		return nil, fmt.Errorf("该嵌入模型未配置 API Key，请在模型管理中补充")
	}
	pending, err := s.chunkRepo.ListChunksPendingEmbed(ctx, resourceID, teacherID)
	if err != nil {
		return nil, err
	}
	if len(pending) == 0 {
		return nil, fmt.Errorf("没有待嵌入的分块：请先保存分块并点击确认")
	}

	const batchSize = 32
	embedded := 0
	for i := 0; i < len(pending); i += batchSize {
		j := i + batchSize
		if j > len(pending) {
			j = len(pending)
		}
		batch := pending[i:j]
		inputs := make([]string, len(batch))
		for k := range batch {
			inputs[k] = batch[k].Content
		}
		vecs, err := s.callOpenAICompatEmbeddings(ctx, strings.TrimSpace(m.APIBaseURL), strings.TrimSpace(m.ModelID), strings.TrimSpace(m.APIKey), inputs)
		if err != nil {
			logrus.WithError(err).WithField("resource_id", resourceID).Warn("embedding batch failed")
			return nil, err
		}
		if len(vecs) != len(batch) {
			return nil, fmt.Errorf("嵌入接口返回向量数量与分块不一致")
		}
		for k := range batch {
			lit := floatsToVectorLiteral(vecs[k])
			if err := s.chunkRepo.UpdateChunkEmbedding(ctx, batch[k].ID, teacherID, lit); err != nil {
				return nil, err
			}
			embedded++
		}
	}
	return &knowledge.EmbedResourceResp{EmbeddedCount: embedded}, nil
}

type openAICompatEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *KnowledgeService) callOpenAICompatEmbeddings(ctx context.Context, url, model, apiKey string, inputs []string) ([][]float64, error) {
	body := map[string]any{
		"model": model,
		"input": inputs,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("嵌入 HTTP %d: %s", resp.StatusCode, truncateForLog(string(respBody), 512))
	}
	var parsed openAICompatEmbeddingResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("解析嵌入响应失败: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("嵌入接口错误: %s", parsed.Error.Message)
	}
	if len(parsed.Data) != len(inputs) {
		return nil, fmt.Errorf("嵌入接口返回 data 长度 %d，期望 %d", len(parsed.Data), len(inputs))
	}
	out := make([][]float64, len(inputs))
	for _, item := range parsed.Data {
		if item.Index < 0 || item.Index >= len(out) {
			return nil, fmt.Errorf("嵌入响应 index 非法: %d", item.Index)
		}
		out[item.Index] = item.Embedding
	}
	for i, v := range out {
		if len(v) == 0 {
			return nil, fmt.Errorf("嵌入响应缺少 index=%d 的向量", i)
		}
	}
	return out, nil
}

func floatsToVectorLiteral(v []float64) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(x, 'f', -1, 64))
	}
	b.WriteByte(']')
	return b.String()
}

func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
