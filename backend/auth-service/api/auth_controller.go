package api

import (
	"docusage/auth-service/middleware"
	"docusage/auth-service/models"
	"docusage/auth-service/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthController 认证控制器
type AuthController struct {
	authService *services.AuthService
	logger      *zap.Logger
}

// NewAuthController 创建认证控制器实例
func NewAuthController(authService *services.AuthService, logger *zap.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		logger:      logger,
	}
}

// RegisterRoutes 注册路由
func (c *AuthController) RegisterRoutes(router *gin.RouterGroup) {
	// 公开路由
	public := router.Group("/")
	{
		public.POST("/login", c.Login)
		public.POST("/register", c.Register)
		public.POST("/refresh", c.RefreshToken)
		public.POST("/logout", c.Logout)
	}

	// 需要认证的路由
	auth := router.Group("/")
	auth.Use(middleware.AuthMiddleware(c.authService.GetConfig(), c.logger))
	{
		auth.GET("/me", c.GetProfile)
		auth.POST("/change-password", c.ChangePassword)
	}

	// 管理员路由
	admin := router.Group("/admin/")
	admin.Use(middleware.AuthMiddleware(c.authService.GetConfig(), c.logger))
	admin.Use(middleware.AdminMiddleware(c.logger))
	{
		admin.GET("/users", c.ListUsers)
		admin.GET("/users/:id", c.GetUser)
		admin.PUT("/users/:id", c.UpdateUser)
	}
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid login request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	tokenPair, err := c.authService.Login(&req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tokenPair)
}

// Register 用户注册
func (c *AuthController) Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid registration request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user, err := c.authService.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 不返回敏感信息
	response := gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
		"status":    user.Status,
	}

	ctx.JSON(http.StatusCreated, response)
}

// RefreshToken 刷新令牌
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid refresh token request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	tokenPair, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tokenPair)
}

// GetProfile 获取用户个人信息
func (c *AuthController) GetProfile(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")

	user, err := c.authService.GetUserByID(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 不返回敏感信息
	response := gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
		"status":    user.Status,
	}

	ctx.JSON(http.StatusOK, response)
}

// ChangePassword 修改密码
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")

	var req models.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid change password request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := c.authService.ChangePassword(userID.(uint), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *gin.Context) {
	// 在实际应用中，这里可以将令牌加入黑名单
	// 目前简单返回成功消息
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ListUsers 获取用户列表（管理员）
func (c *AuthController) ListUsers(ctx *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := c.authService.ListUsers(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建响应数据
	var userList []gin.H
	for _, user := range users {
		userInfo := gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"full_name":  user.FullName,
			"role":       user.Role,
			"status":     user.Status,
			"created_at": user.CreatedAt,
		}
		if user.LastLoginAt != nil {
			userInfo["last_login_at"] = user.LastLoginAt
		}
		userList = append(userList, userInfo)
	}

	response := gin.H{
		"users":       userList,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	}

	ctx.JSON(http.StatusOK, response)
}

// GetUser 获取指定用户信息（管理员）
func (c *AuthController) GetUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := c.authService.GetUserByID(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"full_name":  user.FullName,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
	if user.LastLoginAt != nil {
		response["last_login_at"] = user.LastLoginAt
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateUser 更新用户信息（管理员）
func (c *AuthController) UpdateUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid update user request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user, err := c.authService.UpdateUser(uint(userID), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"full_name":     user.FullName,
		"role":          user.Role,
		"status":        user.Status,
		"last_login_at": user.LastLoginAt,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}
