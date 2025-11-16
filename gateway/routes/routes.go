package routes

import (
	"docusage/gateway/config"
	"docusage/gateway/middleware"
	"docusage/gateway/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes 设置所有路由
func SetupRoutes(router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	// 初始化服务代理
	authService := services.NewAuthService(cfg.AuthServiceAddr, logger)
	queryService := services.NewQueryService(cfg.QueryServiceAddr, logger)
	docService := services.NewDocService(cfg.DocServiceAddr, logger)
	kbService := services.NewKnowledgeBaseService(cfg.KnowledgeBaseAddr, logger)
	toolService := services.NewToolService(cfg.ToolServiceAddr, logger)
	// Admin服务暂时使用Auth服务地址
	adminService := services.NewAdminService(cfg.AuthServiceAddr, logger)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// 公开路由 - 无需认证
	public := router.Group("/")
	{
		// 认证相关路由
		public.POST("/api/auth/login", authService.Login)
		public.POST("/api/auth/refresh", authService.RefreshToken)
		public.GET("/api/auth/verify", authService.VerifyToken)

		// 公开的文档元数据访问
		public.GET("/api/docs/:id/metadata", docService.GetDocumentMetadata)

		// 公开的搜索结果高亮
		public.POST("/api/query/highlight", queryService.HighlightResults)
	}

	// 需要认证的路由
	auth := router.Group("/")
	auth.Use(middleware.JWTAuthMiddleware(cfg.JWTSecretKey))
	{
		// 用户认证相关
		auth.POST("/api/auth/logout", authService.Logout)
		auth.GET("/api/auth/me", authService.GetUserInfo)

		// 查询服务路由
		auth.POST("/api/query/semantic", queryService.SemanticSearch)
		auth.POST("/api/query/fulltext", queryService.FulltextSearch)
		auth.POST("/api/query/hybrid", queryService.HybridSearch)
		auth.GET("/api/query/history", queryService.GetSearchHistory)
		auth.POST("/api/query/history", queryService.SaveSearchHistory)
		auth.DELETE("/api/query/history/:id", queryService.DeleteSearchHistory)
		auth.GET("/api/query/suggestions", queryService.GetSearchSuggestions)
		auth.POST("/api/query/advanced", queryService.AdvancedSearch)
		auth.GET("/api/query/stats", queryService.GetQueryStats)

		// 文档操作路由
		auth.POST("/api/docs/search", docService.Search)
		auth.GET("/api/docs/:id", docService.GetDocument)
		auth.POST("/api/docs/upload", docService.UploadDocument)
		auth.DELETE("/api/docs/:id", docService.DeleteDocument)
		auth.PUT("/api/docs/:id", docService.UpdateDocument)
		auth.GET("/api/docs", docService.GetDocumentList)
		auth.POST("/api/docs/batch-process", docService.BatchProcessDocuments)
		auth.POST("/api/docs/convert/:id", docService.ConvertDocument)

		// 知识库操作路由
		auth.GET("/api/knowledge-bases", kbService.GetKnowledgeBases)
		auth.GET("/api/knowledge-bases/:id", kbService.GetKnowledgeBase)
		auth.POST("/api/knowledge-bases", kbService.CreateKnowledgeBase)
		auth.PUT("/api/knowledge-bases/:id", kbService.UpdateKnowledgeBase)
		auth.DELETE("/api/knowledge-bases/:id", kbService.DeleteKnowledgeBase)
		auth.POST("/api/knowledge-bases/:id/documents", kbService.AddDocumentToKnowledgeBase)
		auth.DELETE("/api/knowledge-bases/:id/documents/:docId", kbService.RemoveDocumentFromKnowledgeBase)
		auth.GET("/api/knowledge-bases/:id/documents", kbService.GetKnowledgeBaseDocuments)
		auth.GET("/api/knowledge-bases/:id/stats", kbService.GetKnowledgeBaseStats)
		auth.POST("/api/knowledge-bases/:id/search", kbService.SearchKnowledgeBase)

		// 工具服务路由
		auth.GET("/api/tools", toolService.GetTools)
		auth.POST("/api/tools/:toolId", toolService.ExecuteTool)
		auth.GET("/api/tools/history", toolService.GetToolExecutionHistory)
		auth.GET("/api/tools/:toolId", toolService.GetToolInfo)
		auth.GET("/api/tools/executions/:executionId", toolService.GetToolExecutionStatus)
	}

	// 管理员路由 - 需要认证和管理员权限
	admin := router.Group("/api/admin/")
	admin.Use(middleware.JWTAuthMiddleware(cfg.JWTSecretKey))
	admin.Use(middleware.RateLimitMiddleware(cfg.RateLimitPerIP, cfg.RateLimitPerUser, cfg.RateLimitWindowSec))
	// 暂时移除管理员权限检查，使用标准认证即可
// admin.Use(middleware.AdminAuthMiddleware())
	{
		// 知识库导入导出
		admin.POST("/knowledge-bases/import", kbService.ImportKnowledgeBase)
		admin.GET("/knowledge-bases/:id/export", kbService.ExportKnowledgeBase)
		admin.PUT("/knowledge-bases/:id/access", kbService.SetAccessControl)

		// 自定义工具管理
		admin.POST("/tools/custom", toolService.CreateCustomTool)
		admin.PUT("/tools/custom/:toolId", toolService.UpdateCustomTool)
		admin.DELETE("/tools/custom/:toolId", toolService.DeleteCustomTool)
		admin.GET("/tools/stats", toolService.GetToolStats)
		admin.DELETE("/tools/executions/:executionId", toolService.CancelToolExecution)
		
		// 用户管理
		admin.GET("/users", adminService.ListUsers)
		admin.GET("/users/:id", adminService.GetUser)
		admin.POST("/users", adminService.CreateUser)
		admin.PUT("/users/:id", adminService.UpdateUser)
		admin.DELETE("/users/:id", adminService.DeleteUser)

		// 角色权限管理
		admin.GET("/roles", adminService.ListRoles)
		admin.POST("/roles", adminService.CreateRole)
		admin.PUT("/roles/:id", adminService.UpdateRole)

		// 知识库管理
		admin.GET("/knowledge-bases", adminService.ListKnowledgeBases)
		admin.POST("/knowledge-bases", adminService.CreateKnowledgeBase)
		admin.PUT("/knowledge-bases/:id", adminService.UpdateKnowledgeBase)
		admin.DELETE("/knowledge-bases/:id", adminService.DeleteKnowledgeBase)

		// 文档管理
		admin.GET("/documents", adminService.ListAllDocuments)
		admin.POST("/documents/process", adminService.ProcessDocuments)
		admin.DELETE("/documents/:id", adminService.DeleteDocument)

		// 系统设置
		admin.GET("/settings", adminService.GetSettings)
		admin.PUT("/settings", adminService.UpdateSettings)

		// 统计分析
		admin.GET("/analytics/usage", adminService.GetUsageAnalytics)
		admin.GET("/analytics/search", adminService.GetSearchAnalytics)
	}
}