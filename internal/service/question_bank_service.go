package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"unicode/utf8"

	"rag-online-course/internal/config"
	dto "rag-online-course/internal/dto/course"
	"rag-online-course/internal/repository/postgres"
)

const (
	questionBankMaxFileBytes     = 2 * 1024 * 1024
	questionBankMaxContentRunes  = 12000
	questionBankDefaultPageSize  = 20
	questionBankMaxPageSize      = 100
	questionBankConfirmMaxBatch  = 200
)

var questionTypeAliases = map[string]string{
	"single_choice":   "single_choice",
	"multiple_choice": "multiple_choice",
	"true_false":      "true_false",
	"short_answer":    "short_answer",
	"fill_blank":      "fill_blank",
}

// QuestionBankService 负责课程题库管理：手工 CRUD 与上传文本后 AI 批量出题。
type QuestionBankService struct {
	courseRepo  *postgres.CourseRepository
	modelRepo   *postgres.TeacherAIModelRepository
	questionRepo *postgres.QuestionBankRepository
	httpClient  *http.Client
	systemPrompt string
	userTmpl     *template.Template
}

// NewQuestionBankService 创建题库服务，提示词模板在启动阶段解析，异常时立即暴露配置错误。
func NewQuestionBankService(
	cfg config.Config,
	courseRepo *postgres.CourseRepository,
	modelRepo *postgres.TeacherAIModelRepository,
	questionRepo *postgres.QuestionBankRepository,
	httpClient *http.Client,
) *QuestionBankService {
	tmpl, err := template.New("question_bank_user").Parse(strings.TrimSpace(cfg.QuestionBank.ImportUserPromptTemplate))
	if err != nil {
		panic(fmt.Sprintf("question_bank.import_user_prompt_template 解析失败: %v", err))
	}
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &QuestionBankService{
		courseRepo:   courseRepo,
		modelRepo:    modelRepo,
		questionRepo: questionRepo,
		httpClient:   httpClient,
		systemPrompt: strings.TrimSpace(cfg.QuestionBank.ImportSystemPrompt),
		userTmpl:     tmpl,
	}
}

// ListByCourse 分页返回课程下题目；page、pageSize 非法时回退为默认值；keyword 模糊匹配题干/答案/类型；questionType 精确匹配题型。
func (s *QuestionBankService) ListByCourse(ctx context.Context, courseID, teacherID int64, page, pageSize int, keyword, questionType string) (*dto.ListQuestionBankItemsResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = questionBankDefaultPageSize
	}
	if pageSize > questionBankMaxPageSize {
		pageSize = questionBankMaxPageSize
	}
	offset := (page - 1) * pageSize
	total, err := s.questionRepo.CountByCourse(ctx, courseID, teacherID, keyword, questionType)
	if err != nil {
		return nil, err
	}
	rows, err := s.questionRepo.ListByCoursePaged(ctx, courseID, teacherID, keyword, questionType, pageSize, offset)
	if err != nil {
		return nil, err
	}
	items := make([]dto.QuestionBankItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.QuestionBankItem{
			ID:              strings.TrimSpace(fmt.Sprint(row["id"])),
			CourseID:        strings.TrimSpace(fmt.Sprint(row["course_id"])),
			QuestionType:    strings.TrimSpace(fmt.Sprint(row["question_type"])),
			Stem:            strings.TrimSpace(fmt.Sprint(row["stem"])),
			ReferenceAnswer: strings.TrimSpace(fmt.Sprint(row["reference_answer"])),
			SourceFileName:  strings.TrimSpace(fmt.Sprint(row["source_file_name"])),
			CreatedAt:       strings.TrimSpace(fmt.Sprint(row["created_at"])),
			UpdatedAt:       strings.TrimSpace(fmt.Sprint(row["updated_at"])),
		})
	}
	return &dto.ListQuestionBankItemsResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Create 手工新增单题，强制校验核心字段。
func (s *QuestionBankService) Create(ctx context.Context, courseID, teacherID int64, req dto.CreateQuestionBankItemReq) (string, error) {
	qType, stem, answer, err := sanitizeQuestion(req.QuestionType, req.Stem, req.ReferenceAnswer)
	if err != nil {
		return "", err
	}
	id, err := s.questionRepo.Create(ctx, courseID, teacherID, qType, stem, answer, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprint(id), nil
}

// Update 修改题目内容。
func (s *QuestionBankService) Update(ctx context.Context, itemID, teacherID int64, req dto.UpdateQuestionBankItemReq) error {
	qType, stem, answer, err := sanitizeQuestion(req.QuestionType, req.Stem, req.ReferenceAnswer)
	if err != nil {
		return err
	}
	return s.questionRepo.Update(ctx, itemID, teacherID, qType, stem, answer)
}

// Delete 删除题目。
func (s *QuestionBankService) Delete(ctx context.Context, itemID, teacherID int64) error {
	return s.questionRepo.Delete(ctx, itemID, teacherID)
}

type questionImportPromptData struct {
	CourseTitle string
	FileName    string
	Content     string
}

// parsedQuestionList 与题库导入模型返回的 JSON 对齐；options 为可选，用于将选项合并进题干。
type parsedQuestionList struct {
	Questions []struct {
		QuestionType    string   `json:"question_type"`
		Stem            string   `json:"stem"`
		ReferenceAnswer string   `json:"reference_answer"`
		Options         []string `json:"options"`
	} `json:"questions"`
}

// ParseImportFromFileText 调用 QA 模型解析上传文本为题目草稿，不入库。
func (s *QuestionBankService) ParseImportFromFileText(ctx context.Context, courseID, teacherID int64, fileName string, fileBytes []byte, qaModelID *int64) (*dto.ParseQuestionBankImportResp, error) {
	courseTitle, _, err := s.courseRepo.GetCourseForTeacher(ctx, courseID, teacherID)
	if err != nil {
		return nil, err
	}
	drafts, err := s.parseUploadToDraftReqs(ctx, teacherID, courseTitle, fileName, fileBytes, qaModelID)
	if err != nil {
		return nil, err
	}
	if len(drafts) == 0 {
		return nil, fmt.Errorf("未生成可展示题目，请检查文件内容或提示词")
	}
	return &dto.ParseQuestionBankImportResp{Questions: drafts}, nil
}

// ConfirmImportBatch 将教师确认后的题目批量写入题库，单次上限由 questionBankConfirmMaxBatch 约束。
func (s *QuestionBankService) ConfirmImportBatch(ctx context.Context, courseID, teacherID int64, req dto.ConfirmQuestionBankImportReq) (*dto.ConfirmQuestionBankImportResp, error) {
	_, _, err := s.courseRepo.GetCourseForTeacher(ctx, courseID, teacherID)
	if err != nil {
		return nil, err
	}
	if len(req.Questions) == 0 {
		return nil, fmt.Errorf("题目列表不能为空")
	}
	if len(req.Questions) > questionBankConfirmMaxBatch {
		return nil, fmt.Errorf("单次最多确认入库 %d 道题", questionBankConfirmMaxBatch)
	}
	src := strings.TrimSpace(req.SourceFileName)
	items := make([]dto.QuestionBankItem, 0, len(req.Questions))
	for _, q := range req.Questions {
		qType, stem, answer, vErr := sanitizeQuestion(q.QuestionType, q.Stem, q.ReferenceAnswer)
		if vErr != nil {
			return nil, fmt.Errorf("题目校验失败：%v", vErr)
		}
		insertedID, cErr := s.questionRepo.Create(ctx, courseID, teacherID, qType, stem, answer, src)
		if cErr != nil {
			return nil, cErr
		}
		items = append(items, dto.QuestionBankItem{
			ID:              fmt.Sprint(insertedID),
			CourseID:        fmt.Sprint(courseID),
			QuestionType:    qType,
			Stem:            stem,
			ReferenceAnswer: answer,
			SourceFileName:  src,
		})
	}
	return &dto.ConfirmQuestionBankImportResp{
		CreatedCount: len(items),
		Items:        items,
	}, nil
}

// parseUploadToDraftReqs 校验文件、调用模型并转为 Create 请求体形态的草稿（不写库）。
func (s *QuestionBankService) parseUploadToDraftReqs(ctx context.Context, teacherID int64, courseTitle, fileName string, fileBytes []byte, qaModelID *int64) ([]dto.CreateQuestionBankItemReq, error) {
	if len(fileBytes) == 0 {
		return nil, fmt.Errorf("上传文件为空")
	}
	if len(fileBytes) > questionBankMaxFileBytes {
		return nil, fmt.Errorf("上传文件过大（最大 2MB）")
	}
	content := strings.TrimSpace(string(fileBytes))
	if !utf8.ValidString(content) {
		return nil, fmt.Errorf("仅支持 UTF-8 编码文本文件")
	}
	if utf8.RuneCountInString(content) > questionBankMaxContentRunes {
		r := []rune(content)
		content = string(r[:questionBankMaxContentRunes]) + "..."
	}
	if content == "" {
		return nil, fmt.Errorf("文件内容为空")
	}
	model, err := s.pickQAModel(ctx, teacherID, qaModelID)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, fmt.Errorf("未配置可用问答模型（qa）")
	}
	userPrompt, err := s.renderImportPrompt(questionImportPromptData{
		CourseTitle: strings.TrimSpace(courseTitle),
		FileName:    strings.TrimSpace(fileName),
		Content:     content,
	})
	if err != nil {
		return nil, err
	}
	if s.systemPrompt == "" {
		return nil, fmt.Errorf("配置项 question_bank.import_system_prompt 为空")
	}
	reqBody := map[string]any{
		"model": strings.TrimSpace(model.ModelID),
		"messages": []map[string]any{
			{"role": "system", "content": s.systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	}
	status, respBody, err := postJSONOutline(ctx, s.httpClient, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), reqBody)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("题库导入模型请求失败（HTTP %d）：%s", status, truncateOutlineLog(string(respBody), 400))
	}
	contentJSON, err := extractChatCompletionContent(respBody)
	if err != nil {
		return nil, err
	}
	var parsed parsedQuestionList
	if err := json.Unmarshal([]byte(stripJSONFence(contentJSON)), &parsed); err != nil {
		return nil, fmt.Errorf("模型返回非合法题库 JSON：%w", err)
	}
	if len(parsed.Questions) == 0 {
		return nil, fmt.Errorf("未解析出任何题目")
	}
	out := make([]dto.CreateQuestionBankItemReq, 0, len(parsed.Questions))
	for _, q := range parsed.Questions {
		qTypeRaw := strings.TrimSpace(strings.ToLower(q.QuestionType))
		qTypeNorm, typeOK := questionTypeAliases[qTypeRaw]
		stemIn := strings.TrimSpace(q.Stem)
		if typeOK {
			stemIn = mergeChoiceStemIntoStem(stemIn, q.Options, qTypeNorm)
		}
		qType, stem, answer, vErr := sanitizeQuestion(q.QuestionType, stemIn, q.ReferenceAnswer)
		if vErr != nil {
			continue
		}
		out = append(out, dto.CreateQuestionBankItemReq{
			QuestionType:    qType,
			Stem:            stem,
			ReferenceAnswer: answer,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("未生成可入库题目，请检查文件内容或提示词")
	}
	return out, nil
}

// mergeChoiceStemIntoStem 将 options 格式化为 A./B./C. 等形式并追加到题干后，保证选择题题干含完整选项。
// 非选择题类型或非空 options 时才会拼接；题干已含选项时模型可能仍传 options，重复内容由提示词约束减少。
func mergeChoiceStemIntoStem(stem string, options []string, qType string) string {
	if qType != "single_choice" && qType != "multiple_choice" && qType != "true_false" {
		return stem
	}
	clean := make([]string, 0, len(options))
	for _, o := range options {
		t := strings.TrimSpace(o)
		if t != "" {
			clean = append(clean, t)
		}
	}
	if len(clean) == 0 {
		return stem
	}
	var b strings.Builder
	b.WriteString(strings.TrimSpace(stem))
	b.WriteString("\n\n")
	const labels = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i, opt := range clean {
		if i > 0 {
			b.WriteByte('\n')
		}
		var label string
		if i < len(labels) {
			label = string(labels[i])
		} else {
			label = fmt.Sprintf("%d", i+1)
		}
		b.WriteString(label)
		b.WriteString(". ")
		b.WriteString(opt)
	}
	return b.String()
}

func sanitizeQuestion(questionType, stem, answer string) (string, string, string, error) {
	qTypeRaw := strings.TrimSpace(strings.ToLower(questionType))
	qType, ok := questionTypeAliases[qTypeRaw]
	s := strings.TrimSpace(stem)
	a := strings.TrimSpace(answer)
	if qTypeRaw == "" || s == "" || a == "" {
		return "", "", "", fmt.Errorf("题目类型、题干、参考答案不能为空")
	}
	if !ok {
		return "", "", "", fmt.Errorf("题目类型非法，仅支持：single_choice / multiple_choice / true_false / short_answer / fill_blank")
	}
	return qType, s, a, nil
}

func (s *QuestionBankService) renderImportPrompt(data questionImportPromptData) (string, error) {
	var buf bytes.Buffer
	if err := s.userTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染题库导入提示词失败: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}

func (s *QuestionBankService) pickQAModel(ctx context.Context, teacherID int64, selected *int64) (*postgres.TeacherAIModelRow, error) {
	if selected != nil {
		row, err := s.modelRepo.GetByID(ctx, *selected, teacherID)
		if err != nil {
			if err == postgres.ErrNotFound {
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
