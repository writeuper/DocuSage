package utils

import (
	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
)

// HashPassword 对密码进行哈希加密
func HashPassword(password string, logger *zap.Logger) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return "", err
	}
	return string(hash), nil
}

// CheckPassword 验证密码是否正确
func CheckPassword(password, hashedPassword string, logger *zap.Logger) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		logger.Warn("Password verification failed", zap.Error(err))
		return false
	}
	return true
}

// ValidatePasswordComplexity 验证密码复杂度
func ValidatePasswordComplexity(password string, minLength int) bool {
	// 检查密码长度
	if len(password) < minLength {
		return false
	}
	
	// 这里可以添加更多密码复杂度检查
	// 例如：至少包含一个数字、一个大写字母、一个小写字母等
	
	return true
}