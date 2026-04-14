// Package docreaderhttp 提供对 docreader-http 服务（HTTP REST /v1/read）的客户端，供资源解析编排使用。
package docreaderhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrNotConfigured 表示未配置 docreader 服务地址，调用方应提示运维或跳过解析能力。
var ErrNotConfigured = errors.New("docreader: base_url not configured")

// Client 封装对解析网关的 HTTP 调用，线程安全可并发复用。
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient 根据 BaseURL、可选鉴权头与超时构造客户端；baseURL 为空时后续 V1Read 返回 ErrNotConfigured。
func NewClient(baseURL, internalToken string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 600 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   internalToken,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// V1ReadRequest 与 docreader-http POST /v1/read 的 JSON 体一致；object_key 与 url 二选一。
type V1ReadRequest struct {
	ObjectKey    string `json:"object_key,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	URL          string `json:"url,omitempty"`
	FileName     string `json:"file_name,omitempty"`
	ReturnImages bool   `json:"return_images"`
	UseOCR       bool   `json:"use_ocr"`
}

// ImageItem 表示解析结果中的单张图片（Base64 载荷）。
type ImageItem struct {
	FileName   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	DataBase64 string `json:"data_base64"`
}

// V1ReadResponse 与网关响应 JSON 对齐。
type V1ReadResponse struct {
	Markdown string         `json:"markdown"`
	Images   []ImageItem    `json:"images"`
	Metadata map[string]any `json:"metadata"`
	Error    string         `json:"error"`
}

// V1Read 调用解析服务，HTTP 层错误返回 wrap 后的 error；业务失败时响应体内可能仍带 error 字段由调用方处理。
func (c *Client) V1Read(ctx context.Context, req V1ReadRequest) (*V1ReadResponse, error) {
	if c.baseURL == "" {
		return nil, ErrNotConfigured
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("docreader: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/read", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("docreader: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("X-Internal-Token", c.token)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("docreader: http do: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("docreader: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docreader: http %d: %s", resp.StatusCode, truncate(string(body), 1024))
	}
	var out V1ReadResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("docreader: decode json: %w", err)
	}
	return &out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
