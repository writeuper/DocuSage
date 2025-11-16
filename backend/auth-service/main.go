package main

import (
	"context"
	"docusage/auth-service/api"
	"docusage/auth-service/config"
	"docusage/auth-service/middleware"
	"docusage/auth-service/models"
	"docusage/auth-service/services"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 设置日志
	logger, err := config.SetupLogger(cfg)
	if err != nil {
		fmt.Printf("Failed to setup logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting auth service...")

	// 初始化数据库
	if err := models.InitDB(cfg, logger); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// 初始化Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 测试Redis连接
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logger.Warn("Redis connection failed, continuing without Redis", zap.Error(err))
	} else {
		logger.Info("Connected to Redis")
	}
	defer redisClient.Close()

	// 设置Gin模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由引擎
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware())

	// 创建认证服务
	authService := services.NewAuthService(cfg, models.GetDB(), redisClient, logger)

	// 创建认证控制器并注册路由
	authController := api.NewAuthController(authService, logger)
	apiGroup := router.Group("/api/auth")
	authController.RegisterRoutes(apiGroup)

	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
			"time":    time.Now().Unix(),
		})
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort),
		Handler: router,
	}

	// 启动服务器（非阻塞）
	go func() {
		logger.Info("Starting server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
