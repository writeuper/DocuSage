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

type ToolService struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

func NewToolService(baseURL string, logger *zap.Logger) *ToolService {
	return &ToolService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 60 * time.Second, // 工具服务可能需要更长的超时时间
		},
		logger: logger,
	}
}

// 获取可用工具列表
func (s *ToolService) GetTools(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/tools")
}

// 执行特定工具
func (s *ToolService) ExecuteTool(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/tools/"+c.Param("toolId"))
}

// 获取工具执行历史
func (s *ToolService) GetToolExecutionHistory(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/tools/history")
}

// 获取工具详细信息
func (s *ToolService) GetToolInfo(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/tools/"+c.Param("toolId"))
}

// 创建自定义工具
func (s *ToolService) CreateCustomTool(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "POST", "/api/tools/custom")
}

// 更新自定义工具
func (s *ToolService) UpdateCustomTool(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "PUT", "/api/tools/custom/"+c.Param("toolId"))
}

// 删除自定义工具
func (s *ToolService) DeleteCustomTool(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/tools/custom/"+c.Param("toolId"))
}

// 获取工具执行状态
func (s *ToolService) GetToolExecutionStatus(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/tools/executions/"+c.Param("executionId"))
}

// 取消工具执行
func (s *ToolService) CancelToolExecution(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "DELETE", "/api/tools/executions/"+c.Param("executionId"))
}

// 获取工具统计信息
func (s *ToolService) GetToolStats(c *gin.Context) {
	s.forwardAuthenticatedRequest(c, "GET", "/api/tools/stats")
}

// 通用请求转发
func (s *ToolService) forwardRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, false)
}

// 认证请求转发
func (s *ToolService) forwardAuthenticatedRequest(c *gin.Context, method, path string) {
	s.forwardRequestInternal(c, method, path, true)
}

// 请求转发内部实现
func (s *ToolService) forwardRequestInternal(c *gin.Context, method, path string, requireAuth bool) {
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
	s.logger.Info("Tool service request forwarded",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
		zap.String("userID", c.GetString("userID")),
		zap.String("toolId", c.Param("toolId")),
	)
}
