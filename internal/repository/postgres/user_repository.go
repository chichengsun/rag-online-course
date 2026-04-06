package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"rag-online-course/internal/domain"

	"gorm.io/gorm"
)

// AuthUser 表示登录鉴权场景下的用户查询结果（与 users 表字段对应）。
type AuthUser struct {
	ID           int64
	Email        string
	Username     string
	Name         string
	PasswordHash string
	Role         string
}

// UserRepository 管理用户对象的持久化操作。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储（直接注入 *gorm.DB）。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser 创建用户；username 唯一由 uk_users_username 保证，email 无库级唯一约束。
func (r *UserRepository) CreateUser(ctx context.Context, email, username, name, passwordHash string, role domain.UserRole) (int64, error) {
	var insertedUserID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO users(email, username, name, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, email, username, name, passwordHash, role).Scan(&insertedUserID).Error
	return insertedUserID, err
}

// GetAuthUserByAccount 支持 email/username 两种账号登录方式。
func (r *UserRepository) GetAuthUserByAccount(ctx context.Context, account string) (AuthUser, error) {
	account = strings.TrimSpace(account)
	var authUser AuthUser
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, email, username, name, password_hash, role::text
		FROM users
		WHERE (email = $1 OR username = $1) AND is_active = TRUE
	`, account).Scan(&authUser).Error
	return authUser, err
}

// GetUserProfile 查询当前用户展示信息，映射为领域模型。
func (r *UserRepository) GetUserProfile(ctx context.Context, userID int64) (*domain.User, error) {
	var (
		id        int64
		email     string
		username  string
		name      string
		role      string
		isActive  bool
		createdAt time.Time
		updatedAt time.Time
	)
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, email, username, name, role::text AS role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Row().Scan(&id, &email, &username, &name, &role, &isActive, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &domain.User{
		ID:        id,
		Email:     email,
		Username:  username,
		Name:      name,
		Role:      domain.UserRole(role),
		IsActive:  isActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
