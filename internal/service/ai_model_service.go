package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"rag-online-course/internal/dto/knowledge"
	"rag-online-course/internal/repository/postgres"
)

// AIModelService 教师 AI 模型配置（问答 / 嵌入 / 重排）的增删改查与连通性探测。
type AIModelService struct {
	repo       *postgres.TeacherAIModelRepository
	httpClient *http.Client
}

// NewAIModelService 创建服务；httpClient 为空时使用 http.DefaultClient。
func NewAIModelService(repo *postgres.TeacherAIModelRepository, httpClient *http.Client) *AIModelService {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &AIModelService{repo: repo, httpClient: httpClient}
}

// List 列出当前教师的模型（响应中不返回 api_key）。
func (s *AIModelService) List(ctx context.Context, teacherID int64) ([]knowledge.AIModelListItem, error) {
	rows, err := s.repo.ListByTeacher(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	out := make([]knowledge.AIModelListItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, knowledge.AIModelListItem{
			ID:         strconv.FormatInt(r.ID, 10),
			Name:       r.Name,
			ModelType:  r.ModelType,
			APIBaseURL: r.APIBaseURL,
			ModelID:    r.ModelID,
			HasAPIKey:  strings.TrimSpace(r.APIKey) != "",
		})
	}
	return out, nil
}

// Create 新建模型配置。
func (s *AIModelService) Create(ctx context.Context, teacherID int64, req knowledge.CreateAIModelReq) (string, error) {
	row := &postgres.TeacherAIModelRow{
		TeacherID:  teacherID,
		Name:       strings.TrimSpace(req.Name),
		ModelType:  req.ModelType,
		APIBaseURL: strings.TrimSpace(req.APIBaseURL),
		ModelID:    strings.TrimSpace(req.ModelID),
		APIKey:     req.APIKey,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return "", err
	}
	return strconv.FormatInt(row.ID, 10), nil
}

// Update 更新模型；api_key 非空时覆盖，为空则保留原密钥。
func (s *AIModelService) Update(ctx context.Context, id, teacherID int64, req knowledge.UpdateAIModelReq) error {
	var keyPtr *string
	if t := strings.TrimSpace(req.APIKey); t != "" {
		keyPtr = &t
	}
	return s.repo.Update(ctx, id, teacherID, strings.TrimSpace(req.Name), strings.TrimSpace(req.APIBaseURL), strings.TrimSpace(req.ModelID), keyPtr)
}

// Delete 删除模型。
func (s *AIModelService) Delete(ctx context.Context, id, teacherID int64) error {
	return s.repo.Delete(ctx, id, teacherID)
}
