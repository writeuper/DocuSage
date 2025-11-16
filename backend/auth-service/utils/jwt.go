package utils

import (
	"docusage/auth-service/config"
	"docusage/auth-service/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// CustomClaims 自定义JWT声明
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateTokenPair 生成访问令牌和刷新令牌
func GenerateTokenPair(user *models.User, cfg *config.Config, logger *zap.Logger) (*models.TokenPair, error) {
	// 生成访问令牌
	accessToken, expiresIn, err := generateAccessToken(user, cfg, logger)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := generateRefreshToken(user, cfg, logger)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// generateAccessToken 生成访问令牌
func generateAccessToken(user *models.User, cfg *config.Config, logger *zap.Logger) (string, int64, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpireHours) * time.Hour)
	expiresIn := int64(cfg.JWTExpireHours * 3600)

	claims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecretKey))
	if err != nil {
		logger.Error("Failed to generate access token", zap.Error(err))
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// generateRefreshToken 生成刷新令牌
func generateRefreshToken(user *models.User, cfg *config.Config, logger *zap.Logger) (string, error) {
	expirationTime := time.Now().AddDate(0, 0, cfg.RefreshTokenExpireDays)

	claims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
			Subject:   "refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecretKey))
	if err != nil {
		logger.Error("Failed to generate refresh token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证令牌
func ValidateToken(tokenString string, cfg *config.Config, logger *zap.Logger) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTSecretKey), nil
	})

	if err != nil {
		logger.Error("Failed to parse token", zap.Error(err))
		return nil, err
	}

	if !token.Valid {
		logger.Error("Invalid token")
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func RefreshToken(refreshTokenString string, cfg *config.Config, logger *zap.Logger) (*models.TokenPair, error) {
	// 验证刷新令牌
	claims, err := ValidateToken(refreshTokenString, cfg, logger)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 检查令牌类型
	if claims.Subject != "refresh" {
		return nil, errors.New("invalid refresh token type")
	}

	// 查找用户
	var user models.User
	result := models.GetDB().First(&user, claims.UserID)
	if result.Error != nil {
		logger.Error("User not found", zap.Uint("userID", claims.UserID))
		return nil, errors.New("user not found")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	// 生成新的令牌对
	return GenerateTokenPair(&user, cfg, logger)
}