package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	dto "rag-online-course/internal/dto/user"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service/auth"
)

var ErrUnauthorized = errors.New("unauthorized")

// UserService 负责用户注册、登录、令牌刷新等用户域业务编排。
type UserService struct {
	userRepo *postgres.UserRepository
	jwt      *auth.JWTService
	sessions *auth.SessionStore
}

// NewUserService 创建用户业务服务。
func NewUserService(userRepo *postgres.UserRepository, jwt *auth.JWTService, sessions *auth.SessionStore) *UserService {
	return &UserService{userRepo: userRepo, jwt: jwt, sessions: sessions}
}

// Register 注册用户并完成用户名规范化、密码哈希等处理。
func (s *UserService) Register(ctx context.Context, registerReq dto.RegisterReq) (dto.RegisterResp, error) {
	var registerResp dto.RegisterResp
	normalizedUsername := normalizeUsername(registerReq.Username)
	if strings.Contains(normalizedUsername, " ") {
		return registerResp, errors.New("username cannot contain spaces")
	}
	passwordHash, err := auth.HashPassword(registerReq.Password)
	if err != nil {
		return registerResp, err
	}
	newUserID, err := s.userRepo.CreateUser(
		ctx,
		strings.TrimSpace(registerReq.Email),
		normalizedUsername,
		normalizeSpaces(registerReq.Name),
		passwordHash,
		registerReq.Role,
	)
	if err != nil {
		return registerResp, err
	}
	registerResp.ID = strconv.FormatInt(newUserID, 10)
	return registerResp, nil
}

// Login 校验账号密码并签发 access/refresh token。
func (s *UserService) Login(ctx context.Context, loginReq dto.LoginReq) (*dto.LoginResp, error) {
	authUser, err := s.userRepo.GetAuthUserByAccount(ctx, loginReq.Account)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if err = auth.ComparePassword(authUser.PasswordHash, loginReq.Password); err != nil {
		return nil, ErrUnauthorized
	}
	userIDStr := strconv.FormatInt(authUser.ID, 10)
	sessionID, err := s.sessions.CreateSession(ctx, userIDStr, authUser.Role, s.jwt.RefreshTTL())
	if err != nil {
		return nil, err
	}
	accessToken, err := s.jwt.GenerateAccessToken(userIDStr, authUser.Role, sessionID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwt.GenerateRefreshToken(userIDStr, authUser.Role, sessionID)
	if err != nil {
		return nil, err
	}
	userPayload := map[string]any{
		"id":       userIDStr,
		"email":    authUser.Email,
		"username": authUser.Username,
		"name":     authUser.Name,
		"role":     authUser.Role,
	}
	return &dto.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userPayload,
	}, nil
}

// Refresh 校验 refresh token 与 Redis 会话后续签 access token。
func (s *UserService) Refresh(ctx context.Context, refreshReq dto.RefreshReq) (dto.RefreshResp, error) {
	var refreshResp dto.RefreshResp
	tokenClaims, err := s.jwt.ParseRefreshToken(refreshReq.RefreshToken)
	if err != nil {
		return refreshResp, ErrUnauthorized
	}
	sessionActive, err := s.sessions.ValidateSession(ctx, tokenClaims.SessionID, tokenClaims.UserID, tokenClaims.Role)
	if err != nil {
		return refreshResp, err
	}
	if !sessionActive {
		return refreshResp, ErrUnauthorized
	}
	_ = s.sessions.TouchSession(ctx, tokenClaims.SessionID, s.jwt.RefreshTTL())
	newAccessToken, err := s.jwt.GenerateAccessToken(tokenClaims.UserID, tokenClaims.Role, tokenClaims.SessionID)
	if err != nil {
		return refreshResp, err
	}
	refreshResp.AccessToken = newAccessToken
	return refreshResp, nil
}

// Me 查询当前用户资料（领域模型转 DTO 响应 map）。
func (s *UserService) Me(ctx context.Context, meReq dto.MeReq) (dto.MeResp, error) {
	profile, err := s.userRepo.GetUserProfile(ctx, meReq.UserID)
	if err != nil {
		return nil, err
	}
	m := map[string]any{
		"id":         strconv.FormatInt(profile.ID, 10),
		"email":      profile.Email,
		"username":   profile.Username,
		"name":       profile.Name,
		"role":       string(profile.Role),
		"is_active":  profile.IsActive,
		"created_at": profile.CreatedAt,
		"updated_at": profile.UpdatedAt,
	}
	return dto.MeResp(m), nil
}

func normalizeUsername(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
