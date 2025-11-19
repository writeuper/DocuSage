package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type QueryService struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

func NewQueryService(baseURL string, logger *zap.Logger) *QueryService {
	// 确保baseURL包含http://前缀
	if len(baseURL) > 0 && baseURL[:7] != "http://" && baseURL[:8] != "https://" {
		baseURL = "http://" + baseURL
	}

	return &QueryService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 45 * time.Second, // 查询服务可能需要更长的超时时间
		},
		logger: logger,
	}
}

// 执行语义搜索
func (s *QueryService) SemanticSearch(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/query/semantic")
}

// 执行全文搜索
func (s *QueryService) FulltextSearch(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/query/fulltext")
}

// 执行混合搜索
func (s *QueryService) HybridSearch(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/query/hybrid")
}

// 获取搜索历史
func (s *QueryService) GetSearchHistory(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/query/history")
}

// 保存搜索历史
func (s *QueryService) SaveSearchHistory(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/query/history")
}

// 删除搜索历史
func (s *QueryService) DeleteSearchHistory(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/query/history/"+c.Param("id"))
}

// 获取搜索建议
func (s *QueryService) GetSearchSuggestions(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/query/suggestions")
}

// 执行高级搜索
func (s *QueryService) AdvancedSearch(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/query/advanced")
}

// 处理搜索请求
func (s *QueryService) Search(c *gin.Context) {
	// 现在前后端都使用POST方法，直接转发
	s.forwardAuthenticatedRequest(c, "POST", "/api/query/search")
}

// 搜索结果高亮处理
func (s *QueryService) HighlightResults(c *gin.Context) {
	s.forwardRequest(c, "POST", "/api/query/highlight")
}

// 查询统计
func (s *QueryService) GetQueryStats(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/query/stats")
}

// 通用请求转发
func (s *QueryService) forwardRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, false)
}

// 认证请求转发
func (s *QueryService) forwardAuthenticatedRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, true)
}

// 请求转发内部实现
func (s *QueryService) forwardRequestInternal(c *gin.Context, method, path string, requireAuth bool) {
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
		// 修复类型转换错误：将uint转换为string
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

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
	s.logger.Info("Query service request forwarded",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
		zap.String("userID", c.GetString("userID")),
		zap.String("query", c.Query("q")),
	)
}
