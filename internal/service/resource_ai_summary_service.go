package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	dto "rag-online-course/internal/dto/course"
	"rag-online-course/internal/repository/postgres"
)

// ResourceAISummaryService 为文档类资源生成 AI 摘要：先走解析得到 Markdown，再调用教师配置的问答（qa）模型。
type ResourceAISummaryService struct {
	parseSvc   *ResourceParseService
	resRepo    *postgres.ResourceRepository
	modelRepo  *postgres.TeacherAIModelRepository
	httpClient *http.Client
}

// NewResourceAISummaryService 构造摘要服务；httpClient 为空时使用默认 Client（短超时不适合长文档，由调用方注入长超时实例更佳）。
func NewResourceAISummaryService(
	parseSvc *ResourceParseService,
	resRepo *postgres.ResourceRepository,
	modelRepo *postgres.TeacherAIModelRepository,
	httpClient *http.Client,
) *ResourceAISummaryService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 180 * time.Second}
	}
	return &ResourceAISummaryService{
		parseSvc:   parseSvc,
		resRepo:    resRepo,
		modelRepo:  modelRepo,
		httpClient: httpClient,
	}
}

const summarizeMarkdownMaxRunes = 28000

// StartSummarizeJob 尝试启动异步摘要：若已在 running 则返回当前快照；成功抢占则后台执行并返回 202 语义由 Handler 写入 HTTP 状态码。
func (s *ResourceAISummaryService) StartSummarizeJob(ctx context.Context, resourceID, teacherID int64) (*dto.SummarizeResourceResp, int, error) {
	item, err := s.resRepo.GetResourceByID(ctx, resourceID, teacherID)
	if err != nil {
		return nil, 0, err
	}
	rt := strings.TrimSpace(toString(item["resource_type"]))
	if !isDocumentResourceType(rt) {
		return nil, 0, fmt.Errorf("仅支持文档类资源（pdf/txt/doc/docx/ppt）生成摘要")
	}
	if s.parseSvc == nil {
		return nil, 0, fmt.Errorf("解析服务未初始化")
	}

	started, err := s.resRepo.TryMarkAISummaryRunning(ctx, resourceID, teacherID)
	if err != nil {
		return nil, 0, err
	}
	if !started {
		return s.snapshotSummarizeResp(ctx, resourceID, teacherID)
	}

	go s.runSummarizeJob(resourceID, teacherID)
	return &dto.SummarizeResourceResp{Status: "running"}, http.StatusAccepted, nil
}

// snapshotSummarizeResp 根据当前库内字段组装响应（用于重复点击或轮询对齐）。
func (s *ResourceAISummaryService) snapshotSummarizeResp(ctx context.Context, resourceID, teacherID int64) (*dto.SummarizeResourceResp, int, error) {
	item, err := s.resRepo.GetResourceByID(ctx, resourceID, teacherID)
	if err != nil {
		return nil, 0, err
	}
	st := strings.TrimSpace(toString(item["ai_summary_status"]))
	if st == "" {
		st = "idle"
	}
	out := &dto.SummarizeResourceResp{Status: st}
	if v := toString(item["ai_summary"]); v != "" {
		out.Summary = v
	}
	if v := toString(item["ai_summary_updated_at"]); v != "" {
		out.UpdatedAt = v
	}
	if v := toString(item["ai_summary_error"]); v != "" {
		out.Error = v
	}
	return out, http.StatusOK, nil
}

// runSummarizeJob 在独立协程中执行长耗时解析与模型调用，错误写入 ai_summary_error。
func (s *ResourceAISummaryService) runSummarizeJob(resourceID, teacherID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	defer func() {
		if rec := recover(); rec != nil {
			_ = s.resRepo.SetAISummaryJobFailed(ctx, resourceID, teacherID, fmt.Sprintf("摘要任务异常: %v", rec))
		}
	}()

	summary, err := s.computeSummaryMarkdown(ctx, resourceID, teacherID)
	if err != nil {
		_ = s.resRepo.SetAISummaryJobFailed(ctx, resourceID, teacherID, err.Error())
		return
	}
	if _, err := s.resRepo.UpdateResourceAISummary(ctx, resourceID, teacherID, summary); err != nil {
		_ = s.resRepo.SetAISummaryJobFailed(ctx, resourceID, teacherID, err.Error())
	}
}

// computeSummaryMarkdown 解析文档并调用 LLM，返回待持久化的 Markdown 摘要正文（不写库）。
func (s *ResourceAISummaryService) computeSummaryMarkdown(ctx context.Context, resourceID, teacherID int64) (string, error) {
	parseResp, err := s.parseSvc.ParseResource(ctx, resourceID, teacherID)
	if err != nil {
		return "", err
	}
	if parseResp == nil || parseResp.Status != "ok" {
		msg := "文档解析未成功，无法生成摘要"
		if parseResp != nil {
			if e := strings.TrimSpace(parseResp.Error); e != "" {
				msg = e
			}
		}
		return "", fmt.Errorf("%s", msg)
	}
	body := strings.TrimSpace(parseResp.Markdown)
	if body == "" {
		return "", fmt.Errorf("解析结果为空，无法生成摘要")
	}
	body = truncateRunesForSummary(body, summarizeMarkdownMaxRunes)

	models, err := s.modelRepo.ListByTeacher(ctx, teacherID)
	if err != nil {
		return "", err
	}
	var qa *postgres.TeacherAIModelRow
	for i := range models {
		if models[i].ModelType == "qa" {
			qa = &models[i]
			break
		}
	}
	if qa == nil || strings.TrimSpace(qa.APIKey) == "" {
		return "", fmt.Errorf("未配置可用问答模型（qa）")
	}

	system := "你是课程助教。请阅读用户给出的 Markdown 正文，用中文输出简洁摘要：先给 3～8 条要点列表，再给一段不超过 200 字的综述。不要编造正文中不存在的事实。"
	user := "以下为课程文档解析后的 Markdown 正文，请按要求输出摘要：\n\n" + body
	msgs := []map[string]any{
		{"role": "system", "content": system},
		{"role": "user", "content": user},
	}
	reqBody := map[string]any{
		"model":       strings.TrimSpace(qa.ModelID),
		"messages":    msgs,
		"temperature": 0.3,
	}
	status, respBody, err := s.postJSON(ctx, strings.TrimSpace(qa.APIBaseURL), strings.TrimSpace(qa.APIKey), reqBody)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("摘要模型请求失败（HTTP %d）：%s", status, truncateForSummaryLog(string(respBody), 400))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("解析摘要响应失败: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("摘要模型未返回内容")
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("摘要模型返回空内容")
	}
	return out, nil
}

func isDocumentResourceType(rt string) bool {
	switch rt {
	case "pdf", "txt", "doc", "docx", "ppt":
		return true
	default:
		return false
	}
}

func truncateRunesForSummary(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) > max {
		r = r[:max]
	}
	return string(r) + "\n\n（正文过长已截断，摘要仅基于以上片段）"
}

func truncateForSummaryLog(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

func (s *ResourceAISummaryService) postJSON(ctx context.Context, apiURL, apiKey string, body map[string]any) (int, []byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, b, nil
}
