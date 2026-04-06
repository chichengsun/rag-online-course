package minio

import (
	"strings"
	"testing"
)

func TestService_ObjectURL_escapesPathSegments(t *testing.T) {
	s := &Service{
		bucket:        "course-resources",
		endpoint:      "localhost:9000",
		useSSL:        false,
		publicBaseURL: "http://localhost:9000",
	}
	got := s.ObjectURL("courses/1/chapters/2/uuid-1.1 Go 并发.md")
	wantSub := "1.1%20Go%20%E5%B9%B6%E5%8F%91.md"
	if !strings.Contains(got, wantSub) {
		t.Fatalf("ObjectURL should escape path segments; got %q want substring %q", got, wantSub)
	}
	if !strings.HasPrefix(got, "http://localhost:9000/course-resources/") {
		t.Fatalf("unexpected prefix: %s", got)
	}
}

func TestService_ObjectURL_usesEndpointWhenNoPublicBase(t *testing.T) {
	s := &Service{
		bucket:   "b",
		endpoint: "minio:9000",
		useSSL:   false,
	}
	got := s.ObjectURL("a/b")
	if got != "http://minio:9000/b/a/b" {
		t.Fatalf("got %q", got)
	}
}

func TestService_NewObjectKey(t *testing.T) {
	svc := &Service{}
	key := svc.NewObjectKey("course-1", "chapter-2", "lesson.pdf")

	if !strings.HasPrefix(key, "courses/course-1/chapters/chapter-2/") {
		t.Fatalf("unexpected prefix: %s", key)
	}
	if !strings.HasSuffix(key, "-lesson.pdf") {
		t.Fatalf("unexpected suffix: %s", key)
	}
}
