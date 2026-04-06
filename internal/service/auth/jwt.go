package auth

import (
	"errors"
	"time"

	"rag-online-course/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	// SessionID 用于绑定 Redis 登录会话，避免仅凭 JWT 无状态校验。
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

// JWTService 负责签发和解析 Access/Refresh Token。
type JWTService struct {
	issuer        string
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewJWTService 根据配置创建 JWT 服务实例。
func NewJWTService(cfg config.Config) *JWTService {
	return &JWTService{
		issuer:        cfg.JWT.Issuer,
		accessSecret:  []byte(cfg.JWT.AccessSecret),
		refreshSecret: []byte(cfg.JWT.RefreshSecret),
		accessTTL:     time.Duration(cfg.JWT.AccessTTLMinutes) * time.Minute,
		refreshTTL:    time.Duration(cfg.JWT.RefreshTTLMinutes) * time.Minute,
	}
}

// AccessTTL 返回 Access Token 有效期。
func (s *JWTService) AccessTTL() time.Duration { return s.accessTTL }

// RefreshTTL 返回 Refresh Token 有效期。
func (s *JWTService) RefreshTTL() time.Duration { return s.refreshTTL }

// GenerateAccessToken 生成短期访问令牌（绑定登录会话）。
func (s *JWTService) GenerateAccessToken(userID, role, sessionID string) (string, error) {
	return s.generateToken(userID, role, sessionID, s.accessTTL, s.accessSecret)
}

// GenerateRefreshToken 生成长期刷新令牌（绑定登录会话）。
func (s *JWTService) GenerateRefreshToken(userID, role, sessionID string) (string, error) {
	return s.generateToken(userID, role, sessionID, s.refreshTTL, s.refreshSecret)
}

// ParseAccessToken 解析 Access Token。
func (s *JWTService) ParseAccessToken(raw string) (*Claims, error) {
	return s.parseToken(raw, s.accessSecret)
}

// ParseRefreshToken 解析 Refresh Token。
func (s *JWTService) ParseRefreshToken(raw string) (*Claims, error) {
	return s.parseToken(raw, s.refreshSecret)
}

func (s *JWTService) generateToken(userID, role, sessionID string, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *JWTService) parseToken(raw string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
