package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"rag-online-course/internal/config"
	dto "rag-online-course/internal/dto/course"
	"rag-online-course/internal/repository/postgres"
)

// SectionLessonPlanService 基于小节目标与资源上下文生成结构化教案草案。
type SectionLessonPlanService struct {
	sectionRepo    *postgres.SectionRepository
	resourceRepo   *postgres.ResourceRepository
	modelRepo      *postgres.TeacherAIModelRepository
	httpClient     *http.Client
	systemPrompt   string
	userPromptTmpl *template.Template
}

// NewSectionLessonPlanService 创建小节教案服务；提示词由配置注入并在启动时解析模板。
func NewSectionLessonPlanService(
	cfg config.Config,
	sectionRepo *postgres.SectionRepository,
	resourceRepo *postgres.ResourceRepository,
	modelRepo *postgres.TeacherAIModelRepository,
	httpClient *http.Client,
) *SectionLessonPlanService {
	tmpl, err := template.New("lesson_plan_user").Parse(strings.TrimSpace(cfg.CourseDesign.LessonPlanUserPromptTemplate))
	if err != nil {
		panic(fmt.Sprintf("course_design.lesson_plan_user_prompt_template 解析失败: %v", err))
	}
	return &SectionLessonPlanService{
		sectionRepo:    sectionRepo,
		resourceRepo:   resourceRepo,
		modelRepo:      modelRepo,
		httpClient:     httpClient,
		systemPrompt:   strings.TrimSpace(cfg.CourseDesign.LessonPlanSystemPrompt),
		userPromptTmpl: tmpl,
	}
}

type lessonPlanPromptResource struct {
	Title   string
	Type    string
	Summary string
}

type lessonPlanPromptData struct {
	CourseTitle     string
	ChapterTitle    string
	SectionTitle    string
	Objectives      []string
	TeachingStyle   string
	DurationMinutes int
	ExtraHint       string
	Resources       []lessonPlanPromptResource
}

// GenerateSectionLessonPlan 生成小节教案草案（JSON 结构），不落库。
func (s *SectionLessonPlanService) GenerateSectionLessonPlan(
	ctx context.Context,
	sectionID, teacherID int64,
	req dto.GenerateSectionLessonPlanReq,
) (*dto.GenerateSectionLessonPlanResp, error) {
	sec, err := s.sectionRepo.GetSectionDetail(ctx, sectionID, teacherID)
	if err != nil {
		return nil, err
	}
	courseTitle := strings.TrimSpace(fmt.Sprint(sec["course_title"]))
	chapterTitle := strings.TrimSpace(fmt.Sprint(sec["chapter_title"]))
	sectionTitle := strings.TrimSpace(fmt.Sprint(sec["section_title"]))
	if courseTitle == "" || chapterTitle == "" || sectionTitle == "" {
		return nil, fmt.Errorf("小节上下文信息不完整")
	}
	qaModel, err := s.pickQAModel(ctx, teacherID, req.QAModelID)
	if err != nil {
		return nil, err
	}
	if qaModel == nil || strings.TrimSpace(qaModel.APIKey) == "" {
		return nil, fmt.Errorf("未配置可用问答模型（qa）")
	}
	resRows, err := s.resourceRepo.ListResourcesBySection(ctx, sectionID, teacherID)
	if err != nil {
		return nil, err
	}
	promptData := lessonPlanPromptData{
		CourseTitle:     courseTitle,
		ChapterTitle:    chapterTitle,
		SectionTitle:    sectionTitle,
		Objectives:      normalizeObjectives(req.Objectives),
		TeachingStyle:   strings.TrimSpace(req.TeachingStyle),
		DurationMinutes: req.DurationMinutes,
		ExtraHint:       strings.TrimSpace(req.ExtraHint),
		Resources:       toLessonPlanPromptResources(resRows),
	}
	if len(promptData.Objectives) == 0 {
		return nil, fmt.Errorf("objectives 不能为空")
	}
	userPrompt, err := s.renderUserPrompt(promptData)
	if err != nil {
		return nil, err
	}
	if s.systemPrompt == "" {
		return nil, fmt.Errorf("配置项 course_design.lesson_plan_system_prompt 为空")
	}
	body := map[string]any{
		"model": strings.TrimSpace(qaModel.ModelID),
		"messages": []map[string]any{
			{"role": "system", "content": s.systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.35,
	}
	status, respBody, err := postJSONOutline(ctx, s.httpClient, strings.TrimSpace(qaModel.APIBaseURL), strings.TrimSpace(qaModel.APIKey), body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("教案模型请求失败（HTTP %d）：%s", status, truncateOutlineLog(string(respBody), 400))
	}
	content, err := extractChatCompletionContent(respBody)
	if err != nil {
		return nil, err
	}
	jsonStr := stripJSONFence(content)
	var plan dto.LessonPlanDraft
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("模型返回非合法教案 JSON：%w", err)
	}
	if err := validateLessonPlan(plan); err != nil {
		return nil, err
	}
	return &dto.GenerateSectionLessonPlanResp{
		CourseTitle:   courseTitle,
		ChapterTitle:  chapterTitle,
		SectionTitle:  sectionTitle,
		ResourceCount: len(resRows),
		Plan:          plan,
	}, nil
}

func (s *SectionLessonPlanService) renderUserPrompt(data lessonPlanPromptData) (string, error) {
	var b bytes.Buffer
	if err := s.userPromptTmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("渲染教案 user 提示失败: %w", err)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *SectionLessonPlanService) pickQAModel(ctx context.Context, teacherID int64, selected *int64) (*postgres.TeacherAIModelRow, error) {
	if selected != nil {
		row, err := s.modelRepo.GetByID(ctx, *selected, teacherID)
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				return nil, fmt.Errorf("指定问答模型不存在或无权访问")
			}
			return nil, err
		}
		if row.ModelType != "qa" {
			return nil, fmt.Errorf("指定模型不是问答模型（qa）")
		}
		return row, nil
	}
	rows, err := s.modelRepo.ListByTeacher(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if rows[i].ModelType == "qa" {
			m := rows[i]
			return &m, nil
		}
	}
	return nil, nil
}

func normalizeObjectives(raw []string) []string {
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		t := strings.TrimSpace(v)
		if t == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

func toLessonPlanPromptResources(rows []map[string]any) []lessonPlanPromptResource {
	const maxSummaryRunes = 280
	out := make([]lessonPlanPromptResource, 0, len(rows))
	for _, row := range rows {
		title := strings.TrimSpace(fmt.Sprint(row["title"]))
		rType := strings.TrimSpace(fmt.Sprint(row["resource_type"]))
		summary := strings.TrimSpace(fmt.Sprint(row["ai_summary"]))
		if summary == "" {
			summary = "暂无 AI 摘要，可仅根据标题与类型推断使用方式。"
		}
		r := []rune(summary)
		if len(r) > maxSummaryRunes {
			summary = string(r[:maxSummaryRunes]) + "..."
		}
		out = append(out, lessonPlanPromptResource{
			Title:   title,
			Type:    rType,
			Summary: summary,
		})
	}
	return out
}

func validateLessonPlan(plan dto.LessonPlanDraft) error {
	if strings.TrimSpace(plan.Title) == "" {
		return fmt.Errorf("教案标题为空")
	}
	if len(plan.Objectives) == 0 {
		return fmt.Errorf("教案目标为空")
	}
	if len(plan.Steps) == 0 {
		return fmt.Errorf("教案步骤为空")
	}
	for i, st := range plan.Steps {
		if strings.TrimSpace(st.Phase) == "" {
			return fmt.Errorf("第 %d 个步骤缺少 phase", i+1)
		}
		if len(st.Activities) == 0 {
			return fmt.Errorf("第 %d 个步骤 activities 为空", i+1)
		}
	}
	return nil
}
