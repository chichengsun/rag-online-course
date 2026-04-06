package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"rag-online-course/internal/dto/knowledge"
)

// TestAIModelConnection 对当前表单配置发起一次最小 HTTP 调用，验证网络与鉴权是否可用（不落库）。
// embedding：OpenAI 兼容 embeddings；qa：chat completions；rerank：常见 rerank JSON（model/query/documents/top_n）。
func (s *AIModelService) TestAIModelConnection(ctx context.Context, teacherID int64, req knowledge.TestAIModelConnectionReq) (*knowledge.TestAIModelConnectionResp, error) {
	apiURL := strings.TrimSpace(req.APIBaseURL)
	modelID := strings.TrimSpace(req.ModelID)
	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return nil, fmt.Errorf("无效的 API 地址")
	}
	key := strings.TrimSpace(req.APIKey)
	if key == "" && req.ExistingModelID != nil && *req.ExistingModelID > 0 {
		row, err := s.repo.GetByID(ctx, *req.ExistingModelID, teacherID)
		if err != nil {
			return nil, err
		}
		key = strings.TrimSpace(row.APIKey)
	}
	if key == "" {
		return nil, fmt.Errorf("请填写 API Key；编辑已保存模型时可不填密钥以使用库中已存值")
	}

	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	var body map[string]any
	switch req.ModelType {
	case "embedding":
		body = map[string]any{"model": modelID, "input": "ok"}
	case "qa":
		body = map[string]any{
			"model":       modelID,
			"messages":    []map[string]string{{"role": "user", "content": "ping"}},
			"max_tokens":  4,
			"temperature": 0,
		}
	case "rerank":
		body = map[string]any{
			"model":     modelID,
			"query":     "ping",
			"documents": []string{"a"},
			"top_n":     1,
		}
	default:
		return nil, fmt.Errorf("未知模型类型")
	}

	status, respBody, err := s.postJSON(ctx, apiURL, key, body)
	if err != nil {
		return &knowledge.TestAIModelConnectionResp{
			OK:         false,
			Message:    err.Error(),
			HTTPStatus: 0,
		}, nil
	}
	if status < 200 || status >= 300 {
		return &knowledge.TestAIModelConnectionResp{
			OK:         false,
			Message:    fmt.Sprintf("HTTP %d：%s", status, truncateAIConnLog(string(respBody), 400)),
			HTTPStatus: status,
		}, nil
	}
	if msg := parseUpstreamErrorJSON(respBody); msg != "" {
		return &knowledge.TestAIModelConnectionResp{
			OK:         false,
			Message:    msg,
			HTTPStatus: status,
		}, nil
	}
	return &knowledge.TestAIModelConnectionResp{
		OK:         true,
		Message:    "连接成功，上游已返回成功状态",
		HTTPStatus: status,
	}, nil
}

func (s *AIModelService) postJSON(ctx context.Context, apiURL, apiKey string, body map[string]any) (int, []byte, error) {
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

func parseUpstreamErrorJSON(b []byte) string {
	var wrap struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(b, &wrap); err != nil || len(wrap.Error) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(wrap.Error, &s); err == nil && s != "" {
		return s
	}
	var obj struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(wrap.Error, &obj); err == nil && obj.Message != "" {
		return obj.Message
	}
	return truncateAIConnLog(string(wrap.Error), 300)
}

func truncateAIConnLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
