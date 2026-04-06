package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"rag-online-course/internal/repository/postgres"
)

// RerankService 对候选分块执行重排；未配置模型时可降级直出召回顺序。
type RerankService struct {
	httpClient HTTPJSONClient
}

// HTTPJSONClient 定义重排/模型调用最小能力，便于服务复用。
type HTTPJSONClient interface {
	PostJSON(ctx context.Context, apiURL, apiKey string, body map[string]any) (status int, respBody []byte, err error)
}

// NewRerankService 创建重排服务。
func NewRerankService(httpClient HTTPJSONClient) *RerankService {
	return &RerankService{httpClient: httpClient}
}

// RerankCandidates 使用给定模型对候选分块重排；失败时返回原顺序与错误。
func (s *RerankService) RerankCandidates(
	ctx context.Context,
	model *postgres.TeacherAIModelRow,
	query string,
	candidates []postgres.RetrievalChunk,
) ([]postgres.RetrievalChunk, error) {
	if model == nil || len(candidates) == 0 {
		return candidates, nil
	}
	docs := make([]string, 0, len(candidates))
	for _, it := range candidates {
		docs = append(docs, it.Content)
	}
	body := map[string]any{
		"model":     strings.TrimSpace(model.ModelID),
		"query":     query,
		"documents": docs,
		"top_n":     len(candidates),
	}
	status, respBody, err := s.httpClient.PostJSON(ctx, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), body)
	if err != nil {
		return candidates, err
	}
	if status < 200 || status >= 300 {
		return candidates, fmt.Errorf("rerank http %d: %s", status, truncateForLog(string(respBody), 256))
	}
	type rerankResp struct {
		Results []struct {
			Index int     `json:"index"`
			Score float64 `json:"relevance_score"`
		} `json:"results"`
		Data []struct {
			Index int     `json:"index"`
			Score float64 `json:"score"`
		} `json:"data"`
		Error any `json:"error"`
	}
	var parsed rerankResp
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return candidates, err
	}
	if parsed.Error != nil {
		return candidates, fmt.Errorf("rerank upstream error: %v", parsed.Error)
	}
	type scored struct {
		chunk postgres.RetrievalChunk
		score float64
	}
	tmp := make([]scored, 0, len(candidates))
	if len(parsed.Results) > 0 {
		for _, item := range parsed.Results {
			if item.Index < 0 || item.Index >= len(candidates) {
				continue
			}
			tmp = append(tmp, scored{chunk: candidates[item.Index], score: item.Score})
		}
	} else if len(parsed.Data) > 0 {
		for _, item := range parsed.Data {
			if item.Index < 0 || item.Index >= len(candidates) {
				continue
			}
			tmp = append(tmp, scored{chunk: candidates[item.Index], score: item.Score})
		}
	}
	if len(tmp) == 0 {
		return candidates, nil
	}
	sort.Slice(tmp, func(i, j int) bool { return tmp[i].score > tmp[j].score })
	out := make([]postgres.RetrievalChunk, 0, len(tmp))
	for _, it := range tmp {
		out = append(out, it.chunk)
	}
	return out, nil
}
