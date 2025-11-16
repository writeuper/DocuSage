package api

import (
	"docusage/query-service/models"
	"docusage/query-service/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// QueryController 查询控制器
type QueryController struct {
	queryService *services.QueryService
	log          *zap.Logger
}

// NewQueryController 创建查询控制器实例
func NewQueryController(queryService *services.QueryService, log *zap.Logger) *QueryController {
	return &QueryController{
		queryService: queryService,
		log:          log,
	}
}

// RegisterRoutes 注册路由
func (ctrl *QueryController) RegisterRoutes(router *gin.RouterGroup) {
	// 公开路由（可选，例如健康检查）
	router.GET("/health", ctrl.HealthCheck)

	// 需要认证的路由
	protected := router.Group("/")
	{
		// 搜索相关
		protected.POST("/search", ctrl.Search)
		protected.GET("/history", ctrl.GetSearchHistory)
		protected.GET("/stats", ctrl.GetQueryStats)
	}

	// 管理员路由
	admin := router.Group("/admin")
	{
		admin.GET("/analytics", ctrl.GetAnalytics)
		admin.GET("/queries", ctrl.ListQueries)
		admin.GET("/queries/:id", ctrl.GetQuery)
	}
}

// Search 搜索接口
func (ctrl *QueryController) Search(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// 绑定请求体
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.log.Warn("Invalid search request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad Request",
			"message": err.Error(),
		})
		return
	}

	// 执行搜索
	response, err := ctrl.queryService.Search(c.Request.Context(), req, userID.(uint))
	if err != nil {
		ctrl.log.Error("Search failed", zap.Error(err), zap.Uint("user_id", userID.(uint)))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search Failed",
			"message": err.Error(),
		})
		return
	}

	// 记录查询日志（异步）
	go func() {
		_ = ctrl.queryService.LogQuery(
			c.Request.Context(),
			userID.(uint),
			req.Query,
			c.ClientIP(),
			c.Request.UserAgent(),
			response.ProcessingTime,
			"success",
		)
	}()

	c.JSON(http.StatusOK, response)
}

// GetSearchHistory 获取搜索历史
func (ctrl *QueryController) GetSearchHistory(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// 获取limit参数
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// 获取搜索历史
	history, err := ctrl.queryService.GetSearchHistory(c.Request.Context(), userID.(uint), limit)
	if err != nil {
		ctrl.log.Error("Failed to get search history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "Failed to retrieve search history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// GetQueryStats 获取查询统计
func (ctrl *QueryController) GetQueryStats(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// 获取统计信息
	stats, err := ctrl.queryService.GetQueryStats(c.Request.Context(), userID.(uint))
	if err != nil {
		ctrl.log.Error("Failed to get query stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "Failed to retrieve query statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAnalytics 获取分析数据
func (ctrl *QueryController) GetAnalytics(c *gin.Context) {
	// 绑定请求参数
	var req models.QueryAnalytics
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.log.Warn("Invalid analytics request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad Request",
			"message": err.Error(),
		})
		return
	}

	// 获取分析数据
	analytics, err := ctrl.queryService.GetAnalytics(c.Request.Context(), req)
	if err != nil {
		ctrl.log.Error("Failed to get analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "Failed to retrieve analytics data",
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// ListQueries 列出查询记录（管理员）
func (ctrl *QueryController) ListQueries(c *gin.Context) {
	// 这里可以实现管理员查看所有查询记录的功能
	c.JSON(http.StatusOK, gin.H{
		"message": "List queries endpoint (admin only)",
	})
}

// GetQuery 获取单个查询记录（管理员）
func (ctrl *QueryController) GetQuery(c *gin.Context) {
	// 这里可以实现管理员查看单个查询记录的功能
	c.JSON(http.StatusOK, gin.H{
		"message": "Get query endpoint (admin only)",
	})
}

// HealthCheck 健康检查
func (ctrl *QueryController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "query-service",
		"message": "Query service is running",
	})
}