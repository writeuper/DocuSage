package middleware

import (
	"docusage/query-service/config"
	"docusage/query-service/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(cfg *config.Config, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := utils.ValidateTokenFromHeader(c, cfg, log)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息设置到上下文中
		utils.SetUserToContext(c, claims)

		log.Debug("User authenticated", zap.Uint("user_id", claims.UserID), zap.String("username", claims.Username))

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		if role != "admin" {
			log.Warn("Admin access denied", zap.String("role", role.(string)))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Admin privileges required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware 跨域中间件
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false

		// 检查来源是否在允许列表中
		for _, allowedOrigin := range cfg.CORSAllowOrigins {
			if allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed || origin == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqURI := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		// 错误信息
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 日志格式
		logger := log.With(
			zap.Int("status", statusCode),
			zap.Duration("latency", latencyTime),
			zap.String("method", reqMethod),
			zap.String("uri", reqURI),
			zap.String("ip", clientIP),
		)

		if statusCode >= 500 {
			logger.Error("Server error", zap.String("error", errorMessage))
		} else if statusCode >= 400 {
			logger.Warn("Client error", zap.String("error", errorMessage))
		} else {
			logger.Info("Request processed")
		}
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(cfg *config.Config, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID或IP
		var key string
		userID, exists := c.Get("user_id")
		if exists {
			key = "ratelimit:user:" + string(userID.(uint))
		} else {
			key = "ratelimit:ip:" + c.ClientIP()
		}

		// 每分钟最多60次请求
		allowed, err := utils.RateLimit(c.Request.Context(), key, 60, time.Minute)
		if err != nil {
			log.Error("Rate limit check failed", zap.Error(err))
			c.Next() // 发生错误时允许请求通过
			return
		}

		if !allowed {
			log.Warn("Rate limit exceeded", zap.String("key", key))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理所有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			log.Error("Request error", zap.Error(err))

			// 根据错误类型返回不同的状态码
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
	}
}
