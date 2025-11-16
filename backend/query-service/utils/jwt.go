package utils

import (
	"context"
	"docusage/query-service/config"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// CustomClaims JWT自定义声明
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ParseToken 解析并验证JWT令牌
func ParseToken(tokenString string, secretKey string) (*CustomClaims, error) {
	// 移除Bearer前缀
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateTokenFromHeader 从请求头验证JWT令牌
func ValidateTokenFromHeader(c *gin.Context, cfg *config.Config, log *zap.Logger) (*CustomClaims, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return nil, fmt.Errorf("authorization header is required")
	}

	claims, err := ParseToken(tokenString, cfg.JWTSecretKey)
	if err != nil {
		log.Warn("Failed to validate token", zap.Error(err), zap.String("ip", c.ClientIP()))
		return nil, fmt.Errorf("invalid or expired token")
	}

	return claims, nil
}

// GenerateToken 生成JWT令牌（用于服务间通信）
func GenerateToken(userID uint, username, email, role string, secretKey string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "query-service",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// SetUserToContext 将用户信息设置到上下文中
func SetUserToContext(c *gin.Context, claims *CustomClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(ctx context.Context) (uint, string, string, string, bool) {
	userID, ok1 := ctx.Value("user_id").(uint)
	username, ok2 := ctx.Value("username").(string)
	email, ok3 := ctx.Value("email").(string)
	role, ok4 := ctx.Value("role").(string)

	return userID, username, email, role, ok1 && ok2 && ok3 && ok4
}
