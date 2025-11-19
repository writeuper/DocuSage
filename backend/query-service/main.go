package main

import (
	"context"
	"docusage/query-service/api"
	"docusage/query-service/config"
	"docusage/query-service/middleware"
	"docusage/query-service/models"
	"docusage/query-service/services"
	"docusage/query-service/utils"
	"fmt"
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
	// 注意：不需要全局上下文，各组件可以创建自己的上下文

	// 加载环境变量文件
	if err := godotenv.Load(); err != nil {
		fmt.Printf("警告: 无法加载.env文件: %v\n", err)
	}

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 设置日志
	log, err := config.SetupLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting query service",
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.Environment),
		zap.Bool("debug", cfg.Debug),
	)

	// 初始化数据库
	log.Info("Initializing database connection...")
	if err := models.InitDB(cfg, log); err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func() {
		if err := models.CloseDB(); err != nil {
			log.Error("Error closing database connection", zap.Error(err))
		} else {
			log.Info("Database connection closed")
		}
	}()

	// 初始化Redis
	log.Info("Initializing Redis connection...")
	if err := utils.InitRedis(cfg, log); err != nil {
		log.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer utils.CloseRedis()

	// 设置Gin模式
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	router := gin.New()

	// 添加中间件
	log.Info("Setting up middlewares...")
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(log))
	//router.Use(middleware.CORSMiddleware(cfg))

	// 条件性地添加速率限制中间件
	if cfg.EnableRateLimit {
		router.Use(middleware.RateLimitMiddleware(cfg, log))
		log.Info("Rate limit middleware enabled", zap.Int("per_minute", cfg.RateLimitPerMin))
	} else {
		log.Info("Rate limit middleware disabled")
	}

	router.Use(middleware.ErrorHandlerMiddleware(log))

	// 初始化服务
	log.Info("Initializing services...")
	queryService := services.NewQueryService(models.GetDB(), cfg, log)

	// 初始化控制器
	log.Info("Initializing controllers...")
	queryController := api.NewQueryController(queryService, log)

	// 注册路由
	log.Info("Registering routes...")
	apiGroup := router.Group("/api/query")
	queryController.RegisterRoutes(apiGroup)

	// 注册认证中间件到受保护的路由
	//authMiddleware := middleware.AuthMiddleware(cfg, log)
	//adminMiddleware := middleware.AdminMiddleware(log)

	// 应用中间件到相应的路由组
	//protectedRoutes := apiGroup.Group("/")
	//protectedRoutes.Use(authMiddleware)

	//adminRoutes := apiGroup.Group("/admin")
	//adminRoutes.Use(authMiddleware, adminMiddleware)

	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		// 进行健康检查，包括数据库和Redis连接
		dbHealthy := models.CheckDBHealth()
		redisHealthy := utils.CheckRedisHealth()

		status := http.StatusOK
		if !dbHealthy || !redisHealthy {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":  map[bool]string{true: "ok", false: "degraded"}[dbHealthy && redisHealthy],
			"service": "query-service",
			"version": "1.0.0",
			"components": gin.H{
				"database": dbHealthy,
				"redis":    redisHealthy,
			},
			"timestamp": time.Now().Unix(),
		})
	})

	// 指标路由（如果启用）
	if cfg.EnableMetrics {
		router.GET("/metrics", func(c *gin.Context) {
			// 获取Redis统计信息作为查询服务的活动指标
			redisStats := utils.GetRedisStats()
			c.JSON(http.StatusOK, gin.H{
				"service":        "query-service",
				"active_queries": 0, // 临时设置为0，实际项目中可能需要实现查询计数
				"cache_stats":    redisStats,
				"timestamp":      time.Now().Unix(),
			})
		})
		log.Info("Metrics endpoint enabled")
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.ServerIdleTimeout) * time.Second,
	}

	// 优雅关闭
	go func() {
		// 监听中断信号
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		log.Info("Received termination signal", zap.String("signal", sig.String()))
		log.Info("Shutting down server...")

		// 创建带有超时的上下文用于优雅关闭
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.ServerShutdownTimeout)*time.Second)
		defer shutdownCancel()

		// 启动关闭过程
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("Server forced to shutdown", zap.Error(err))
		} else {
			log.Info("Server shutdown gracefully")
		}
	}()

	// 启动服务器（异步）
	log.Info("Server starting",
		zap.String("addr", srv.Addr),
		zap.Int("read_timeout", cfg.ServerReadTimeout),
		zap.Int("write_timeout", cfg.ServerWriteTimeout),
		zap.Int("idle_timeout", cfg.ServerIdleTimeout),
		zap.Int("shutdown_timeout", cfg.ServerShutdownTimeout),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server", zap.Error(err))
	}

	log.Info("Server process completed")
}
