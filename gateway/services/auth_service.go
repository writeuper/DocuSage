package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthService 认证服务代理
type AuthService struct {
	baseURL string
	logger  *zap.Logger
	client  *http.Client
}

// NewAuthService 创建认证服务代理实例
func NewAuthService(baseURL string, logger *zap.Logger) *AuthService {
	return &AuthService{
		baseURL: baseURL,
		logger:  logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Login 处理登录请求
func (s *AuthService) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转发请求到认证服务
	resp, err := s.forwardRequest("/api/auth/login", http.MethodPost, req)
	if err != nil {
		s.logger.Error("Failed to forward login request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 复制响应状态码和头部
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)

	// 复制响应体
	io.Copy(c.Writer, resp.Body)
}

// Logout 处理登出请求
func (s *AuthService) Logout(c *gin.Context) {
	// 转发请求到认证服务
	resp, err := s.forwardAuthenticatedRequest(c, "/api/auth/logout", http.MethodPost, nil)
	if err != nil {
		s.logger.Error("Failed to forward logout request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 复制响应状态码和头部
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)

	// 复制响应体
	io.Copy(c.Writer, resp.Body)
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(c *gin.Context) {
	// 转发请求到认证服务
	resp, err := s.forwardAuthenticatedRequest(c, "/api/auth/me", http.MethodGet, nil)
	if err != nil {
		s.logger.Error("Failed to forward get user info request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 复制响应状态码和头部
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)

	// 复制响应体
	io.Copy(c.Writer, resp.Body)
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转发请求到认证服务
	resp, err := s.forwardRequest("/api/auth/refresh", http.MethodPost, req)
	if err != nil {
		s.logger.Error("Failed to forward refresh token request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 复制响应状态码和头部
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)

	// 复制响应体
	io.Copy(c.Writer, resp.Body)
}

// VerifyToken 验证Token
func (s *AuthService) VerifyToken(c *gin.Context) {
	// 转发请求到认证服务
	resp, err := s.forwardRequest("/api/auth/verify", http.MethodGet, nil)
	if err != nil {
		s.logger.Error("Failed to forward verify token request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 复制响应状态码和头部
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)

	// 复制响应体
	io.Copy(c.Writer, resp.Body)
}

// forwardRequest 转发请求到后端服务
func (s *AuthService) forwardRequest(path, method string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	// 如果请求体不为空，则序列化为JSON
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建请求
	req, err := http.NewRequest(method, s.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	return s.client.Do(req)
}

// forwardAuthenticatedRequest 转发携带认证信息的请求
func (s *AuthService) forwardAuthenticatedRequest(c *gin.Context, path, method string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	// 如果请求体不为空，则序列化为JSON
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建请求
	req, err := http.NewRequest(method, s.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 添加认证头
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	return s.client.Do(req)
}
