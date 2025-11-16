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

type DocService struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

func NewDocService(baseURL string, logger *zap.Logger) *DocService {
	return &DocService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// 文档搜索处理
func (s *DocService) Search(c *gin.Context) {
	s.forwardRequest(c, "POST", "/api/docs/search")
}

// 获取文档详情
func (s *DocService) GetDocument(c *gin.Context) {
	s.forwardRequest(c, "GET", "/api/docs/"+c.Param("id"))
}

// 上传文档
func (s *DocService) UploadDocument(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/docs/upload")
}

// 删除文档
func (s *DocService) DeleteDocument(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/docs/"+c.Param("id"))
}

// 更新文档
func (s *DocService) UpdateDocument(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/docs/"+c.Param("id"))
}

// 获取文档列表
func (s *DocService) GetDocumentList(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/docs")
}

// 批量处理文档
func (s *DocService) BatchProcessDocuments(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/docs/batch-process")
}

// 文档转码
func (s *DocService) ConvertDocument(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/docs/convert/"+c.Param("id"))
}

// 获取文档元数据
func (s *DocService) GetDocumentMetadata(c *gin.Context) {
	s.forwardRequest(c, "GET", "/api/docs/"+c.Param("id")+"/metadata")
}

// 通用请求转发
func (s *DocService) forwardRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, false)
}

// 认证请求转发
func (s *DocService) forwardAuthenticatedRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, true)
}

// 请求转发内部实现
func (s *DocService) forwardRequestInternal(c *gin.Context, method, path string, requireAuth bool) {
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

	// 如果需要认证，从上下文中获取用户信息
	if requireAuth {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		req.Header.Set("X-User-ID", userID.(string))
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
	s.logger.Info("Document service request forwarded",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
		zap.String("userID", c.GetString("userID")),
	)
}
