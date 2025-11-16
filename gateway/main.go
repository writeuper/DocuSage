/*
 * @Author: lixinda
 * @Description: 
 * @File: 
 * @Date: 2025-11-13 20:07:29
 */
package main

import (
	"context"
	"docusage/gateway/config"
	"docusage/gateway/middleware"
	"docusage/gateway/routes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 初始化配置
	cfg := config.LoadConfig()

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// 设置Gin模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	router := gin.New()

	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CorsMiddleware(cfg.AllowOrigins))
	
	// 添加限流中间件
	router.Use(middleware.RateLimitMiddleware(cfg.RateLimitPerUser, cfg.RateLimitPerIP, cfg.RateLimitWindowSec))

	// 注意：Redis连接已移除，限流中间件现在使用内存实现

	// 设置路由
	routes.SetupRoutes(router, cfg, logger)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 在goroutine中启动服务器
	go func() {
		logger.Info("API Gateway starting on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// 设置5秒的超时时间来关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}