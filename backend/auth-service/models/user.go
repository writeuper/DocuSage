package models

import (
	"gorm.io/gorm"
	"time"
)

// User 用户模型
type User struct {
	gorm.Model
	ID uint `gorm:"primarykey" json:"id"`

	Username       string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email          string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash   string     `gorm:"size:255;not null" json:"-"`
	FullName       string     `gorm:"size:100" json:"full_name"`
	Role           string     `gorm:"size:20;default:'user'" json:"role"`     // admin, user
	Status         string     `gorm:"size:20;default:'active'" json:"status"` // active, inactive, locked
	LastLoginAt    *time.Time `json:"last_login_at"`
	FailedAttempts int        `gorm:"default:0" json:"failed_attempts"`
	LockedUntil    *time.Time `json:"locked_until"`

	// 关系字段将在后续与其他服务集成时添加
}

// Role 角色模型
type Role struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Permissions string    `gorm:"type:json" json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Resource    string    `gorm:"size:50" json:"resource"`
	Action      string    `gorm:"size:50" json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TokenPair 令牌对模型
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
}
