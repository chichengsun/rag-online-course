package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"rag-online-course/internal/config"
	dto "rag-online-course/internal/dto/course"
	"rag-online-course/internal/repository/postgres"
)

// CourseDesignService 课程设计编排：调用教师配置的 qa 模型生成教学大纲草案，并将确认后的草案落库为章节与节。
type CourseDesignService struct {
	courseRepo          *postgres.CourseRepository
	modelRepo           *postgres.TeacherAIModelRepository
	chapterSvc          *ChapterService
	sectionSvc          *SectionService
	httpClient          *http.Client
	outlineSystemPrompt string
	outlineUserTmpl     *template.Template
}

// NewCourseDesignService 构造服务；httpClient 为空时使用 120s 超时的默认 Client；大纲 user 提示由配置模板解析，非法模板会在启动时 panic。
func NewCourseDesignService(
	cfg config.Config,
	courseRepo *postgres.CourseRepository,
	modelRepo *postgres.TeacherAIModelRepository,
	chapterSvc *ChapterService,
	sectionSvc *SectionService,
	httpClient *http.Client,
) *CourseDesignService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 120 * time.Second}
	}
	tmplSrc := strings.TrimSpace(cfg.CourseDesign.OutlineUserPromptTemplate)
	tmpl, err := template.New("outline_user").Parse(tmplSrc)
	if err != nil {
		panic(fmt.Sprintf("course_design.outline_user_prompt_template 解析失败: %v", err))
	}
	return &CourseDesignService{
		courseRepo:          courseRepo,
		modelRepo:           modelRepo,
		chapterSvc:          chapterSvc,
		sectionSvc:          sectionSvc,
		httpClient:          httpClient,
		outlineSystemPrompt: strings.TrimSpace(cfg.CourseDesign.OutlineSystemPrompt),
		outlineUserTmpl:     tmpl,
	}
}

const (
	outlineMaxChapters        = 15
	outlineMaxSectionsPerCh   = 12
	outlineDescMaxRunes     = 2000
)

// GenerateOutlineDraft 调用问答模型生成章节/节 JSON 草案，不写数据库。
func (s *CourseDesignService) GenerateOutlineDraft(ctx context.Context, courseID, teacherID int64, qaModelID *int64, extraHint string) (*dto.GenerateOutlineDraftResp, error) {
	title, description, err := s.courseRepo.GetCourseForTeacher(ctx, courseID, teacherID)
	if err != nil {
		return nil, err
	}
	qa, err := s.pickQAModel(ctx, teacherID, qaModelID)
	if err != nil {
		return nil, err
	}
	if qa == nil || strings.TrimSpace(qa.APIKey) == "" {
		return nil, fmt.Errorf("未配置可用问答模型（qa）")
	}
	user, err := s.buildOutlineUserPrompt(title, description, strings.TrimSpace(extraHint))
	if err != nil {
		return nil, err
	}
	system := s.outlineSystemPrompt
	if system == "" {
		return nil, fmt.Errorf("配置项 course_design.outline_system_prompt 为空")
	}

	msgs := []map[string]any{
		{"role": "system", "content": system},
		{"role": "user", "content": user},
	}
	reqBody := map[string]any{
		"model":       strings.TrimSpace(qa.ModelID),
		"messages":    msgs,
		"temperature": 0.45,
	}
	status, respBody, err := postJSONOutline(ctx, s.httpClient, strings.TrimSpace(qa.APIBaseURL), strings.TrimSpace(qa.APIKey), reqBody)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("大纲模型请求失败（HTTP %d）：%s", status, truncateOutlineLog(string(respBody), 400))
	}
	content, err := extractChatCompletionContent(respBody)
	if err != nil {
		return nil, err
	}
	jsonStr := stripJSONFence(content)
	var parsed struct {
		Chapters []dto.OutlineChapterDraft `json:"chapters"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("模型返回非合法 JSON：%w", err)
	}
	if err := validateOutlineDraft(parsed.Chapters); err != nil {
		return nil, err
	}
	return &dto.GenerateOutlineDraftResp{Chapters: parsed.Chapters}, nil
}

// outlineUserPromptData 供 config 中 outline_user_prompt_template（text/template）渲染。
type outlineUserPromptData struct {
	Title       string
	Description string
	ExtraHint   string
}

// buildOutlineUserPrompt 按配置模板拼装 user 提示；Description 在描述为空时为「（无）」，超长描述在模板前截断。
func (s *CourseDesignService) buildOutlineUserPrompt(title, description, extraHint string) (string, error) {
	desc := description
	if utf8.RuneCountInString(desc) > outlineDescMaxRunes {
		r := []rune(desc)
		desc = string(r[:outlineDescMaxRunes]) + "…"
	}
	descPrompt := desc
	if descPrompt == "" {
		descPrompt = "（无）"
	}
	data := outlineUserPromptData{
		Title:       title,
		Description: descPrompt,
		ExtraHint:   extraHint,
	}
	var buf strings.Builder
	if err := s.outlineUserTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染大纲 user 提示模板失败: %w", err)
	}
	return buf.String(), nil
}

func validateOutlineDraft(chapters []dto.OutlineChapterDraft) error {
	if len(chapters) == 0 {
		return fmt.Errorf("大纲为空")
	}
	if len(chapters) > outlineMaxChapters {
		return fmt.Errorf("章节数量过多（最多 %d）", outlineMaxChapters)
	}
	for i, ch := range chapters {
		if strings.TrimSpace(ch.Title) == "" {
			return fmt.Errorf("第 %d 章标题为空", i+1)
		}
		if len(ch.Sections) > outlineMaxSectionsPerCh {
			return fmt.Errorf("第 %d 章节数量过多（每章最多 %d）", i+1, outlineMaxSectionsPerCh)
		}
		for j, sec := range ch.Sections {
			if strings.TrimSpace(sec.Title) == "" {
				return fmt.Errorf("第 %d 章第 %d 节标题为空", i+1, j+1)
			}
		}
	}
	return nil
}

// ApplyOutlineDraft 将大纲追加到课程末尾：每章调用 CreateChapter（含默认节），再按草案更新/新增节。
func (s *CourseDesignService) ApplyOutlineDraft(ctx context.Context, courseID, teacherID int64, draft *dto.ApplyOutlineDraftReq) (*dto.ApplyOutlineDraftResp, error) {
	if draft == nil || len(draft.Chapters) == 0 {
		return nil, fmt.Errorf("大纲为空")
	}
	_, _, err := s.courseRepo.GetCourseForTeacher(ctx, courseID, teacherID)
	if err != nil {
		return nil, err
	}
	if err := validateOutlineDraft(draft.Chapters); err != nil {
		return nil, err
	}
	existing, err := s.chapterSvc.ListChapters(ctx, courseID, teacherID)
	if err != nil {
		return nil, err
	}
	maxOrder := 0
	for _, row := range existing {
		n := sortOrderFromChapterMap(row)
		if n > maxOrder {
			maxOrder = n
		}
	}
	var createdChapters, createdSections int
	for i, chDraft := range draft.Chapters {
		sortOrder := maxOrder + i + 1
		chapterIDStr, cerr := s.chapterSvc.CreateChapter(ctx, courseID, strings.TrimSpace(chDraft.Title), sortOrder)
		if cerr != nil {
			return nil, fmt.Errorf("创建章失败「%s」: %w", chDraft.Title, cerr)
		}
		createdChapters++
		chapterID, perr := strconv.ParseInt(chapterIDStr, 10, 64)
		if perr != nil {
			return nil, fmt.Errorf("解析章 id 失败: %w", perr)
		}
		secs, lerr := s.sectionSvc.ListSections(ctx, chapterID, teacherID)
		if lerr != nil {
			return nil, lerr
		}
		if len(secs) == 0 {
			continue
		}
		sections := chDraft.Sections
		firstID, ok := sectionIDFromMap(secs[0])
		if !ok {
			return nil, fmt.Errorf("无法读取默认节 id")
		}
		if len(sections) == 0 {
			createdSections++
			continue
		}
		if err := s.sectionSvc.UpdateSection(ctx, firstID, teacherID, strings.TrimSpace(sections[0].Title), 1); err != nil {
			return nil, err
		}
		createdSections++
		for j := 1; j < len(sections); j++ {
			_, aerr := s.sectionSvc.CreateSection(ctx, chapterID, courseID, teacherID, strings.TrimSpace(sections[j].Title), j+1)
			if aerr != nil {
				return nil, fmt.Errorf("创建节失败「%s」: %w", sections[j].Title, aerr)
			}
			createdSections++
		}
	}
	return &dto.ApplyOutlineDraftResp{
		CreatedChapters: createdChapters,
		CreatedSections: createdSections,
	}, nil
}

func sortOrderFromChapterMap(m map[string]any) int {
	v, ok := m["sort_order"]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(t)
		return n
	default:
		n, _ := strconv.Atoi(fmt.Sprint(v))
		return n
	}
}

func sectionIDFromMap(m map[string]any) (int64, bool) {
	v, ok := m["id"]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case int64:
		return t, true
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		return n, err == nil
	default:
		n, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		return n, err == nil
	}
}

func (s *CourseDesignService) pickQAModel(ctx context.Context, teacherID int64, selected *int64) (*postgres.TeacherAIModelRow, error) {
	if selected != nil {
		row, gErr := s.modelRepo.GetByID(ctx, *selected, teacherID)
		if gErr != nil {
			if errors.Is(gErr, postgres.ErrNotFound) {
				return nil, fmt.Errorf("指定问答模型不存在或无权访问")
			}
			return nil, gErr
		}
		if row.ModelType != "qa" {
			return nil, fmt.Errorf("指定模型不是问答模型（qa）")
		}
		return row, nil
	}
	models, err := s.modelRepo.ListByTeacher(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	var stored postgres.TeacherAIModelRow
	found := false
	for i := range models {
		if models[i].ModelType == "qa" {
			stored = models[i]
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}
	return &stored, nil
}

func postJSONOutline(ctx context.Context, client *http.Client, apiURL, apiKey string, body map[string]any) (int, []byte, error) {
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
	resp, err := client.Do(req)
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

func extractChatCompletionContent(respBody []byte) (string, error) {
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("解析模型响应失败: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("模型未返回内容")
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("模型返回空内容")
	}
	return out, nil
}

func stripJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if nl := strings.IndexByte(s, '\n'); nl >= 0 {
			s = strings.TrimSpace(s[nl+1:])
		}
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = strings.TrimSpace(s[:idx])
		}
	}
	return strings.TrimSpace(s)
}

func truncateOutlineLog(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}
