// Package user 存放与 UserHandler（认证与个人资料）对应的请求/响应 DTO。
package user

import "rag-online-course/internal/domain"

// RegisterReq 用户注册请求体。
type RegisterReq struct {
	Email    string          `json:"email" binding:"required,email"`
	Username string          `json:"username" binding:"required,min=3,max=64"`
	Name     string          `json:"name" binding:"required,min=2,max=120"`
	Password string          `json:"password" binding:"required,min=6"`
	Role     domain.UserRole `json:"role" binding:"required,oneof=student teacher"`
}

// RegisterResp 注册成功返回。
type RegisterResp struct {
	// ID 对应 users.id（BIGSERIAL），十进制字符串。
	ID string `json:"id"`
}

// LoginReq 登录请求体（account 可为邮箱或用户名）。
type LoginReq struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp 登录成功返回。
type LoginResp struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	User         map[string]any `json:"user"`
}

// RefreshReq 刷新 Access Token 请求体。
type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResp 刷新成功返回。
type RefreshResp struct {
	AccessToken string `json:"access_token"`
}

// MeReq 查询当前用户资料（UserID 由中间件注入，非 JSON）。
type MeReq struct {
	UserID int64
}

// MeResp 当前用户资料（与 users 表查询字段一致，用 map 承接动态列）。
type MeResp map[string]any
