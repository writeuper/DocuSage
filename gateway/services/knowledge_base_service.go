package services

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type KnowledgeBaseService struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

func NewKnowledgeBaseService(baseURL string, logger *zap.Logger) *KnowledgeBaseService {
	// 确保baseURL包含http://前缀
	if len(baseURL) > 0 && baseURL[:7] != "http://" && baseURL[:8] != "https://" {
		baseURL = "http://" + baseURL
	}

	return &KnowledgeBaseService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// 获取知识库列表
func (s *KnowledgeBaseService) GetKnowledgeBases(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/knowledge-bases")
}

// 获取知识库详情
func (s *KnowledgeBaseService) GetKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/knowledge-bases/"+c.Param("id"))
}

// 创建知识库
func (s *KnowledgeBaseService) CreateKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/knowledge-bases")
}

// 更新知识库
func (s *KnowledgeBaseService) UpdateKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/knowledge-bases/"+c.Param("id"))
}

// 删除知识库
func (s *KnowledgeBaseService) DeleteKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/knowledge-bases/"+c.Param("id"))
}

// 添加文档到知识库
func (s *KnowledgeBaseService) AddDocumentToKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/knowledge-bases/"+c.Param("id")+"/documents")
}

// 从知识库移除文档
func (s *KnowledgeBaseService) RemoveDocumentFromKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/knowledge-bases/"+c.Param("id")+"/documents/"+c.Param("docId"))
}

// 获取知识库中的文档列表
func (s *KnowledgeBaseService) GetKnowledgeBaseDocuments(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/knowledge-bases/"+c.Param("id")+"/documents")
}

// 知识库统计信息
func (s *KnowledgeBaseService) GetKnowledgeBaseStats(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/knowledge-bases/"+c.Param("id")+"/stats")
}

// 知识库搜索
func (s *KnowledgeBaseService) SearchKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/knowledge-bases/"+c.Param("id")+"/search")
}

// 知识库导入
func (s *KnowledgeBaseService) ImportKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/knowledge-bases/import")
}

// 知识库导出
func (s *KnowledgeBaseService) ExportKnowledgeBase(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/knowledge-bases/"+c.Param("id")+"/export")
}

// 知识库访问控制设置
func (s *KnowledgeBaseService) SetAccessControl(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/knowledge-bases/"+c.Param("id")+"/access")
}

// 通用请求转发
func (s *KnowledgeBaseService) forwardRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, false)
}

// 认证请求转发
func (s *KnowledgeBaseService) forwardAuthenticatedRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, true)
}

// 请求转发内部实现
func (s *KnowledgeBaseService) forwardRequestInternal(c *gin.Context, method, path string, requireAuth bool) {
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
		// 重置请求体，以便后续使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// 创建新的请求
	req, err := http.NewRequestWithContext(c.Request.Context(), method, targetURL, bytes.NewBuffer(body))
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

	// 如果需要认证，从上下文中获取用户信息
	userID := ""
	if requireAuth {
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID = userIDVal.(string)
		req.Header.Set("X-User-ID", userID)

		// 添加用户角色信息
		if role, exists := c.Get("role"); exists {
			req.Header.Set("X-User-Role", role.(string))
		}
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
	loggerFields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
	}

	if userID != "" {
		loggerFields = append(loggerFields, zap.String("userID", userID))
	}

	// 添加知识库ID（如果路径中包含）
	if c.Param("id") != "" {
		loggerFields = append(loggerFields, zap.String("knowledgeBaseID", c.Param("id")))
	}

	s.logger.Info("Knowledge base service request forwarded", loggerFields...)
}
