package service

import (
	"strings"
	"testing"

	"rag-online-course/internal/repository/postgres"
)

func TestTruncateTitle(t *testing.T) {
	long := strings.Repeat("测", 100)
	out := truncateTitle(long)
	if len([]rune(out)) != 60 {
		t.Fatalf("expect 60 runes, got %d", len([]rune(out)))
	}
}

func TestBuildReferencesAndContext(t *testing.T) {
	candidates := []postgres.RetrievalChunk{
		{
			ChunkID:      1,
			ResourceID:   2,
			ResourceTitle: "资源A",
			ChunkIndex:   0,
			Content:      "第一段内容",
			Distance:     0.1,
		},
		{
			ChunkID:      3,
			ResourceID:   4,
			ResourceTitle: "资源B",
			ChunkIndex:   1,
			Content:      strings.Repeat("x", 300),
			Distance:     0.4,
		},
	}
	refs, ctx := buildReferencesAndContext(candidates)
	if len(refs) != 2 {
		t.Fatalf("expect 2 refs, got %d", len(refs))
	}
	if !strings.Contains(ctx, "资源:资源A") || !strings.Contains(ctx, "[2]") {
		t.Fatalf("context blocks format unexpected: %s", ctx)
	}
	if refs[0].ChunkID != "1" || refs[0].ResourceID != "2" {
		t.Fatalf("id map failed: %+v", refs[0])
	}
	if refs[0].Score <= refs[1].Score {
		t.Fatalf("score mapping unexpected: ref0=%v ref1=%v", refs[0].Score, refs[1].Score)
	}
	if !strings.HasSuffix(refs[1].Snippet, "...") {
		t.Fatalf("snippet should be truncated with ellipsis")
	}
}
