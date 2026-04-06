package auth

import (
	"testing"

	"rag-online-course/internal/config"
)

func newTestJWTService() *JWTService {
	cfg := config.Config{}
	cfg.JWT.Issuer = "test-issuer"
	cfg.JWT.AccessSecret = "access-secret"
	cfg.JWT.RefreshSecret = "refresh-secret"
	cfg.JWT.AccessTTLMinutes = 10
	cfg.JWT.RefreshTTLMinutes = 60
	return NewJWTService(cfg)
}

func TestJWT_GenerateAndParseAccessToken(t *testing.T) {
	svc := newTestJWTService()

	token, err := svc.GenerateAccessToken("u-1", "student", "s-1")
	if err != nil {
		t.Fatalf("generate access token failed: %v", err)
	}

	claims, err := svc.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse access token failed: %v", err)
	}
	if claims.UserID != "u-1" || claims.Role != "student" || claims.SessionID != "s-1" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWT_ParseWithWrongSecretShouldFail(t *testing.T) {
	svc1 := newTestJWTService()
	token, err := svc1.GenerateAccessToken("u-1", "student", "s-1")
	if err != nil {
		t.Fatalf("generate access token failed: %v", err)
	}

	svc2 := newTestJWTService()
	svc2.accessSecret = []byte("another-secret")
	if _, err = svc2.ParseAccessToken(token); err == nil {
		t.Fatal("expected parse failure with wrong secret")
	}
}
