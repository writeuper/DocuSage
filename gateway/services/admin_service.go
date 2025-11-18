package services

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminService struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

func NewAdminService(baseURL string, logger *zap.Logger) *AdminService {
	// 确保baseURL包含http://前缀
	if len(baseURL) > 0 && baseURL[:7] != "http://" && baseURL[:8] != "https://" {
		baseURL = "http://" + baseURL
	}

	return &AdminService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// 用户管理
func (s *AdminService) ListUsers(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/users")
}

func (s *AdminService) GetUser(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/users/"+c.Param("id"))
}

func (s *AdminService) CreateUser(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/admin/users")
}

func (s *AdminService) UpdateUser(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/admin/users/"+c.Param("id"))
}

func (s *AdminService) DeleteUser(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/admin/users/"+c.Param("id"))
}

// 角色权限管理
func (s *AdminService) ListRoles(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/roles")
}

func (s *AdminService) CreateRole(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/admin/roles")
}

func (s *AdminService) UpdateRole(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/admin/roles/"+c.Param("id"))
}

// 知识库管理
func (s *AdminService) ListKnowledgeBases(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/knowledge-bases")
}

func (s *AdminService) CreateKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/admin/knowledge-bases")
}

func (s *AdminService) UpdateKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/admin/knowledge-bases/"+c.Param("id"))
}

func (s *AdminService) DeleteKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/admin/knowledge-bases/"+c.Param("id"))
}

// 文档管理
func (s *AdminService) ListAllDocuments(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/documents")
}

func (s *AdminService) ProcessDocuments(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/admin/documents/process")
}

func (s *AdminService) DeleteDocument(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/admin/documents/"+c.Param("id"))
}

// 系统设置
func (s *AdminService) GetSettings(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/settings")
}

func (s *AdminService) UpdateSettings(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/admin/settings")
}

// 统计分析
func (s *AdminService) GetUsageAnalytics(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/analytics/usage")
}

func (s *AdminService) GetSearchAnalytics(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/admin/analytics/search")
}

// 请求转发内部实现
func (s *AdminService) forwardAuthenticatedRequest(c *gin.Context, method, path string) {
	// 创建目标URL
	targetURL := s.baseURL + path

	// 读取请求体
	var body []byte
	if c.Request.Body != nil {
		var err error
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			s.logger.Error("Failed to read request body", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request"})
			return
		}
	}

	// 创建新的请求
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		if key == "Host" || key == "Content-Length" || key == "Accept-Encoding" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 从上下文中获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	req.Header.Set("X-User-ID", userID.(string))

	// 添加用户角色信息
	if role, exists := c.Get("role"); exists {
		req.Header.Set("X-User-Role", role.(string))
	}

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to forward request", zap.String("url", targetURL), zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// 复制响应头
	for key, values := range resp.Header {
		if key == "Transfer-Encoding" || key == "Content-Encoding" {
			continue
		}
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 设置响应状态码
	c.Status(resp.StatusCode)

	// 返回响应体
	c.Writer.Write(respBody)

	// 记录请求日志
	s.logger.Info("Admin service request forwarded",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
		zap.String("userID", userID.(string)),
	)
}
