package handlers

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestPostgresUniqueViolationResponse_stringFallback(t *testing.T) {
	err := errors.New(
		`ERROR: duplicate key value violates unique constraint "uk_chapter_resources_course_chapter_title" (SQLSTATE 23505)`,
	)
	st, msg, ok := postgresUniqueViolationResponse(err)
	if !ok || st != http.StatusConflict {
		t.Fatalf("unexpected: ok=%v st=%d msg=%q", ok, st, msg)
	}
	if !strings.Contains(msg, "同名资源") {
		t.Fatalf("message should mention duplicate title: %q", msg)
	}
}

func TestExtractUniqueConstraintNameFromText(t *testing.T) {
	s := `duplicate key value violates unique constraint "chapter_resources_chapter_id_sort_order_key"`
	if got := extractUniqueConstraintNameFromText(s); got != "chapter_resources_chapter_id_sort_order_key" {
		t.Fatalf("got %q", got)
	}
}
