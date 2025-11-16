package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 内存限流实现
type memoryRateLimiter struct {
	mutex       sync.RWMutex
	requests    map[string][]time.Time
	cleanupTime time.Time
}

var limiter = &memoryRateLimiter{
	requests: make(map[string][]time.Time),
}

// RateLimitMiddleware 创建基于内存的限流中间件
func RateLimitMiddleware(limitPerIP, limitPerUser, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()
		ipKey := "rate_limit:ip:" + clientIP

		// 检查IP限流
		if exceeded := checkMemoryRateLimit(ipKey, limitPerIP, windowSeconds); exceeded {
			c.JSON(429, gin.H{"error": "Too many requests from your IP"})
			c.Abort()
			return
		}

		// 检查用户限流 (如果已认证)
		if userID := c.GetString("userID"); userID != "" {
			userKey := "rate_limit:user:" + userID
			if exceeded := checkMemoryRateLimit(userKey, limitPerUser, windowSeconds); exceeded {
				c.JSON(429, gin.H{"error": "Too many requests for your account"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// checkMemoryRateLimit 基于内存的限流检查
func checkMemoryRateLimit(key string, limit, windowSeconds int) bool {
	now := time.Now()
	windowStart := now.Add(-time.Duration(windowSeconds) * time.Second)

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	// 定期清理过期记录
	if now.Sub(limiter.cleanupTime) > time.Minute {
		cleanupExpiredRequests(windowStart)
		limiter.cleanupTime = now
	}

	// 获取现有请求记录
	requests := limiter.requests[key]
	var validRequests []time.Time

	// 过滤有效的请求（在时间窗口内）
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// 检查是否超出限制
	if len(validRequests) >= limit {
		return true // 超出限制
	}

	// 添加当前请求
	validRequests = append(validRequests, now)
	limiter.requests[key] = validRequests

	return false // 未超出限制
}

// cleanupExpiredRequests 清理所有过期的请求记录
func cleanupExpiredRequests(windowStart time.Time) {
	for key, requests := range limiter.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		if len(validRequests) > 0 {
			limiter.requests[key] = validRequests
		} else {
			delete(limiter.requests, key) // 删除空记录以节省内存
		}
	}
}
