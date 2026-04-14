package middleware

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"time"

	"rag-online-course/internal/logging"

	"github.com/gin-gonic/gin"
)

const (
	maxBodyLogBytes = 8 * 1024
)

// RequestLogger 记录每个请求的基础访问日志。
// latency_ms 表示从进入本中间件到请求处理链结束的毫秒数。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqBody := readAndResetRequestBody(c)

		respBody := &bytes.Buffer{}
		c.Writer = &responseCaptureWriter{ResponseWriter: c.Writer, body: respBody}
		c.Next()

		fields := map[string]any{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency_ms": time.Since(start).Milliseconds(),
		}
		if reqBody != "" {
			fields["request_body"] = reqBody
		}
		if body := sanitizeBody(limitBody(respBody.String())); body != "" {
			fields["response_body"] = body
		}
		logging.FromContext(c.Request.Context()).WithFields(fields).Info()
	}
}

type responseCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseCaptureWriter) Write(p []byte) (int, error) {
	if w.body.Len() < maxBodyLogBytes {
		rest := maxBodyLogBytes - w.body.Len()
		if len(p) > rest {
			w.body.Write(p[:rest])
		} else {
			w.body.Write(p)
		}
	}
	return w.ResponseWriter.Write(p)
}

func readAndResetRequestBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	ct := c.GetHeader("Content-Type")
	if !strings.Contains(ct, "application/json") && !strings.Contains(ct, "text/") && !strings.Contains(ct, "application/x-www-form-urlencoded") {
		return ""
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
	return sanitizeBody(limitBody(string(raw)))
}

func limitBody(body string) string {
	if body == "" {
		return ""
	}
	body = oneLineBody(body)
	if len(body) <= maxBodyLogBytes {
		return body
	}
	return body[:maxBodyLogBytes] + "...[truncated]"
}

// oneLineBody 将多行或带缩进的文本压缩为单行，便于日志检索。
func oneLineBody(body string) string {
	return strings.Join(strings.Fields(body), " ")
}

func sanitizeBody(body string) string {
	if body == "" {
		return ""
	}
	patterns := []struct {
		key   string
		value string
	}{
		{`"password"\s*:\s*"[^"]*"`, `"password":"***"`},
		{`"token"\s*:\s*"[^"]*"`, `"token":"***"`},
		{`"access_token"\s*:\s*"[^"]*"`, `"access_token":"***"`},
		{`"refresh_token"\s*:\s*"[^"]*"`, `"refresh_token":"***"`},
		{`"authorization"\s*:\s*"[^"]*"`, `"authorization":"***"`},
		{`"api_key"\s*:\s*"[^"]*"`, `"api_key":"***"`},
		{`"secret"\s*:\s*"[^"]*"`, `"secret":"***"`},
	}
	result := body
	for _, p := range patterns {
		re := regexp.MustCompile(p.key)
		result = re.ReplaceAllString(result, p.value)
	}
	return result
}
