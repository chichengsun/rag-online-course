package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.OPTIONS("/api/v1/courses", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/courses", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("unexpected allow-origin header: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
