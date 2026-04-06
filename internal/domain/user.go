package domain

import "time"

// User 表示系统用户（学生或教师）。
// ID 对应 users.id（BIGSERIAL）；HTTP/API 中常以十进制字符串传递。
// Email 在库中无唯一约束，仅 username 有唯一索引，业务上可按需校验邮箱。
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
