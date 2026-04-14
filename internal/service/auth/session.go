package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"rag-online-course/internal/config"
)

type SessionStore struct {
	rdb *redis.Client
}

// NewSessionStore 创建 Redis 会话存储。
func NewSessionStore(cfg config.Config) *SessionStore {
	return &SessionStore{
		rdb: redis.NewClient(&redis.Options{
			Addr: cfg.Redis.Addr,
			DB:   cfg.Redis.DB,
		}),
	}
}

// Close 关闭 Redis 连接。
func (s *SessionStore) Close() error {
	return s.rdb.Close()
}

const sessionKeyPrefix = "login_session:"

// CreateSession 创建一个登录会话，并写入 Redis。
// 返回 sessionID，用于写入 JWT claims 并在鉴权时二次校验。
func (s *SessionStore) CreateSession(ctx context.Context, userID, role string, ttl time.Duration) (string, error) {
	if userID == "" {
		return "", errors.New("userID is required")
	}
	sessionID := uuid.NewString()
	// value 形如：userID|role，便于后续校验是否与 JWT 声明一致。
	value := userID + "|" + role
	if err := s.rdb.Set(ctx, sessionKeyPrefix+sessionID, value, ttl).Err(); err != nil {
		return "", err
	}
	return sessionID, nil
}

// ValidateSession 判断会话是否存在且与传入的 userID/role 一致。
func (s *SessionStore) ValidateSession(ctx context.Context, sessionID, userID, role string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}
	val, err := s.rdb.Get(ctx, sessionKeyPrefix+sessionID).Result()
	if err != nil {
		return false, err
	}
	parts := strings.SplitN(val, "|", 2)
	if len(parts) != 2 {
		return false, nil
	}
	return parts[0] == userID && parts[1] == role, nil
}

// TouchSession 刷新会话有效期（用于刷新 token 后保持会话不被踢下线）。
func (s *SessionStore) TouchSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	return s.rdb.Expire(ctx, sessionKeyPrefix+sessionID, ttl).Err()
}

// Flush 清空当前 Redis DB，主要用于测试环境隔离。
func (s *SessionStore) Flush(ctx context.Context) error {
	return s.rdb.FlushDB(ctx).Err()
}
