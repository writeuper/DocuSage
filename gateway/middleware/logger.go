package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware 创建日志中间件
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latency := endTime.Sub(startTime)

		// 请求方法
		method := c.Request.Method

		// 请求路由
		path := c.Request.URL.Path

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 用户ID (如果存在)
		userID := c.GetString("userID")

		// 日志字段
		logFields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		}

		// 添加用户ID (如果存在)
		if userID != "" {
			logFields = append(logFields, zap.String("user_id", userID))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error("Server error", logFields...)
		case statusCode >= 400:
			logger.Warn("Client error", logFields...)
		default:
			logger.Info("Request processed", logFields...)
		}
	}
}
