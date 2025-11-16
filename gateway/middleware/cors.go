package middleware

import (
	"github.com/gin-gonic/gin"
)

// CorsMiddleware 创建CORS中间件
func CorsMiddleware(allowOrigins []string) gin.HandlerFunc {
	// 如果没有提供允许的源，则默认允许所有源
	allowOrigin := "*"
	if len(allowOrigins) > 0 && allowOrigins[0] != "" {
		allowOrigin = allowOrigins[0]
	}

	return func(c *gin.Context) {
		// 设置允许的源
		c.Header("Access-Control-Allow-Origin", allowOrigin)

		// 设置允许的请求头
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// 设置允许的请求方法
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// 设置是否允许携带凭证
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// 继续处理请求
		c.Next()
	}
}
