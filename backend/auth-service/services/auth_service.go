package services

import (
	"docusage/auth-service/config"
	"docusage/auth-service/models"
	"docusage/auth-service/utils"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	config *config.Config
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewAuthService 创建认证服务实例
func NewAuthService(cfg *config.Config, db *gorm.DB, redis *redis.Client, logger *zap.Logger) *AuthService {
	return &AuthService{
		config: cfg,
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// Login 用户登录
func (s *AuthService) Login(req *models.LoginRequest) (*models.TokenPair, error) {
	// 查找用户
	var user models.User
	result := s.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("Login attempt failed: User not found", zap.String("username/email", req.Username))
			return nil, errors.New("invalid credentials")
		}
		s.logger.Error("Database error during login", zap.Error(result.Error))
		return nil, errors.New("internal server error")
	}

	// 检查用户锁定状态
	if user.Status == "locked" || !user.LockedUntil.IsZero() && user.LockedUntil.After(time.Now()) {
		s.logger.Warn("Login attempt failed: Account locked", zap.String("username", user.Username))
		return nil, errors.New("account is locked")
	}

	// 检查用户是否激活
	if user.Status != "active" {
		s.logger.Warn("Login attempt failed: Account not active", zap.String("username", user.Username))
		return nil, errors.New("account is not active")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash, s.logger) {
		// 增加失败尝试次数
		user.FailedAttempts++

		// 检查是否需要锁定账户
		if user.FailedAttempts >= s.config.MaxLoginAttempts {
			user.Status = "locked"
			user.LockedUntil = time.Now().Add(time.Duration(s.config.LockoutDurationMinutes) * time.Minute)
			s.logger.Warn("Account locked due to too many failed attempts", zap.String("username", user.Username))
		}

		s.db.Save(&user)
		s.logger.Warn("Login attempt failed: Invalid password", zap.String("username", user.Username))
		return nil, errors.New("invalid credentials")
	}

	// 重置失败尝试次数
	user.FailedAttempts = 0
	user.LastLoginAt = time.Now()
	s.db.Save(&user)

	// 生成令牌对
	tokenPair, err := utils.GenerateTokenPair(&user, s.config, s.logger)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	s.logger.Info("User logged in successfully", zap.Uint("userID", user.ID), zap.String("username", user.Username))
	return tokenPair, nil
}

// Register 用户注册
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	result := s.db.Where("username = ?", req.Username).First(&existingUser)
	if result.Error == nil {
		s.logger.Warn("Registration failed: Username already exists", zap.String("username", req.Username))
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	result = s.db.Where("email = ?", req.Email).First(&existingUser)
	if result.Error == nil {
		s.logger.Warn("Registration failed: Email already exists", zap.String("email", req.Email))
		return nil, errors.New("email already exists")
	}

	// 验证密码复杂度
	if !utils.ValidatePasswordComplexity(req.Password, s.config.MinPasswordLength) {
		s.logger.Warn("Registration failed: Password complexity requirements not met")
		return nil, errors.New("password does not meet complexity requirements")
	}

	// 哈希密码
	hashedPassword, err := utils.HashPassword(req.Password, s.logger)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, errors.New("internal server error")
	}

	// 创建新用户
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         "user", // 默认普通用户
		Status:       "active",
	}

	// 保存用户
	result = s.db.Create(&user)
	if result.Error != nil {
		s.logger.Error("Failed to create user", zap.Error(result.Error))
		return nil, errors.New("internal server error")
	}

	s.logger.Info("User registered successfully", zap.Uint("userID", user.ID), zap.String("username", user.Username))
	return &user, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(refreshToken string) (*models.TokenPair, error) {
	tokenPair, err := utils.RefreshToken(refreshToken, s.config, s.logger)
	if err != nil {
		s.logger.Warn("Token refresh failed", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Token refreshed successfully")
	return tokenPair, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	// 查找用户
	var user models.User
	result := s.db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("Change password failed: User not found", zap.Uint("userID", userID))
			return errors.New("user not found")
		}
		s.logger.Error("Database error during password change", zap.Error(result.Error))
		return errors.New("internal server error")
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash, s.logger) {
		s.logger.Warn("Change password failed: Invalid old password", zap.Uint("userID", userID))
		return errors.New("invalid old password")
	}

	// 验证新密码复杂度
	if !utils.ValidatePasswordComplexity(req.NewPassword, s.config.MinPasswordLength) {
		s.logger.Warn("Change password failed: New password complexity requirements not met", zap.Uint("userID", userID))
		return errors.New("new password does not meet complexity requirements")
	}

	// 哈希新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword, s.logger)
	if err != nil {
		s.logger.Error("Failed to hash new password", zap.Error(err))
		return errors.New("internal server error")
	}

	// 更新密码
	user.PasswordHash = hashedPassword
	result = s.db.Save(&user)
	if result.Error != nil {
		s.logger.Error("Failed to update password", zap.Error(result.Error))
		return errors.New("internal server error")
	}

	s.logger.Info("Password changed successfully", zap.Uint("userID", userID))
	return nil
}

// GetConfig 获取配置信息
func (s *AuthService) GetConfig() *config.Config {
	return s.config
}

// GetUserByID 根据ID获取用户信息
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	result := s.db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("Get user failed: User not found", zap.Uint("userID", userID))
			return nil, errors.New("user not found")
		}
		s.logger.Error("Database error during user retrieval", zap.Error(result.Error))
		return nil, errors.New("internal server error")
	}

	return &user, nil
}

// ListUsers 获取用户列表（管理员功能）
func (s *AuthService) ListUsers(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 计算总数
	result := s.db.Model(&models.User{}).Count(&total)
	if result.Error != nil {
		s.logger.Error("Failed to count users", zap.Error(result.Error))
		return nil, 0, errors.New("internal server error")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	result = s.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users)
	if result.Error != nil {
		s.logger.Error("Failed to list users", zap.Error(result.Error))
		return nil, 0, errors.New("internal server error")
	}

	return users, total, nil
}

// UpdateUser 更新用户信息（管理员功能）
func (s *AuthService) UpdateUser(userID uint, req *models.UpdateUserRequest) (*models.User, error) {
	// 查找用户
	var user models.User
	result := s.db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("Update user failed: User not found", zap.Uint("userID", userID))
			return nil, errors.New("user not found")
		}
		s.logger.Error("Database error during user update", zap.Error(result.Error))
		return nil, errors.New("internal server error")
	}

	// 更新用户信息
	updates := make(map[string]interface{})
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Email != "" && req.Email != user.Email {
		// 检查邮箱是否已被其他用户使用
		var existingUser models.User
		result = s.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser)
		if result.Error == nil {
			s.logger.Warn("Update user failed: Email already exists", zap.String("email", req.Email))
			return nil, errors.New("email already exists")
		}
		updates["email"] = req.Email
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	// 应用更新
	result = s.db.Model(&user).Updates(updates)
	if result.Error != nil {
		s.logger.Error("Failed to update user", zap.Error(result.Error))
		return nil, errors.New("internal server error")
	}

	// 重新获取更新后的用户信息
	result = s.db.First(&user, userID)
	if result.Error != nil {
		s.logger.Error("Failed to retrieve updated user", zap.Error(result.Error))
		return nil, errors.New("internal server error")
	}

	s.logger.Info("User updated successfully", zap.Uint("userID", userID))
	return &user, nil
}
