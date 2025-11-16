package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 查询服务配置
type Config struct {
	// 服务器配置
	ServerPort            string
	ServerHost            string
	Environment           string
	Debug                 bool
	ServerReadTimeout     int
	ServerWriteTimeout    int
	ServerIdleTimeout     int
	ServerShutdownTimeout int

	// 数据库配置
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// Redis配置
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	RedisEnableTLS     bool
	RedisSkipTLSVerify bool
	RedisDialTimeout   time.Duration
	RedisReadTimeout   time.Duration
	RedisWriteTimeout  time.Duration
	RedisPoolSize      int
	RedisMinIdleConns  int
	RedisMaxRetries    int
	RedisMaxIdleConns  int
	RedisIdleTimeout   time.Duration

	// JWT配置
	JWTSecretKey  string
	JWTAlgorithm  string
	JWTExpiration int // 过期时间（分钟）

	// 向量搜索服务配置
	VectorSearchServiceAddr string
	VectorSearchTimeout     int // 超时时间（秒）
	VectorSearchTopK        int // 默认返回结果数

	// 文档处理服务配置
	DocProcessServiceAddr string
	DocProcessTimeout     int // 超时时间（秒）

	// 嵌入服务配置
	EmbedServiceAddr string
	EmbedTimeout     int // 超时时间（秒）

	// LLM服务配置
	LLMServiceAddr string
	LLMTimeout     int // 超时时间（秒）

	// CORS配置
	CORSAllowOrigins []string

	// 日志配置
	LogLevel string
	LogFile  string

	// 搜索配置
	MaxSearchResults      int
	DefaultSearchLimit    int
	SearchHistoryMaxItems int
	SearchCacheTTL        int // 缓存过期时间（分钟）
	MinQueryLength        int // 最小查询长度

	// 性能配置
	MaxConcurrentSearches int
	SearchWorkerCount     int

	// 安全配置
	EnableRateLimit bool
	RateLimitPerMin int
	MaxQuerySize    int // 最大查询大小（字节）
	AllowedIPs      []string

	// 监控配置
	EnableMetrics bool
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	// 加载.env文件
	_ = godotenv.Load()

	// 解析整数配置项
	redisDB := getEnvAsInt("REDIS_DB", 0)

	// 数据库连接池配置
	dbMaxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	dbMaxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 100)
	dbConnMaxLifetime := time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600)) * time.Second
	dbConnMaxIdleTime := time.Duration(getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 1800)) * time.Second

	// Redis高级配置
	redisEnableTLS := getEnvAsBool("REDIS_ENABLE_TLS", false)
	redisSkipTLSVerify := getEnvAsBool("REDIS_SKIP_TLS_VERIFY", false)
	redisDialTimeout := time.Duration(getEnvAsInt("REDIS_DIAL_TIMEOUT", 5)) * time.Second
	redisReadTimeout := time.Duration(getEnvAsInt("REDIS_READ_TIMEOUT", 30)) * time.Second
	redisWriteTimeout := time.Duration(getEnvAsInt("REDIS_WRITE_TIMEOUT", 30)) * time.Second
	redisPoolSize := getEnvAsInt("REDIS_POOL_SIZE", 10)
	redisMinIdleConns := getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5)
	redisMaxRetries := getEnvAsInt("REDIS_MAX_RETRIES", 3)
	redisMaxIdleConns := getEnvAsInt("REDIS_MAX_IDLE_CONNS", 10)
	redisIdleTimeout := time.Duration(getEnvAsInt("REDIS_IDLE_TIMEOUT", 600)) * time.Second

	maxSearchResults, err := strconv.Atoi(getEnv("MAX_SEARCH_RESULTS", "100"))
	if err != nil {
		maxSearchResults = 100
	}

	defaultSearchLimit, err := strconv.Atoi(getEnv("DEFAULT_SEARCH_LIMIT", "10"))
	if err != nil {
		defaultSearchLimit = 10
	}

	searchHistoryMaxItems, err := strconv.Atoi(getEnv("SEARCH_HISTORY_MAX_ITEMS", "100"))
	if err != nil {
		searchHistoryMaxItems = 100
	}

	jwtExpiration, err := strconv.Atoi(getEnv("JWT_EXPIRATION", "1440"))
	if err != nil {
		jwtExpiration = 1440
	}

	vectorSearchTimeout, err := strconv.Atoi(getEnv("VECTOR_SEARCH_TIMEOUT", "30"))
	if err != nil {
		vectorSearchTimeout = 30
	}

	vectorSearchTopK, err := strconv.Atoi(getEnv("VECTOR_SEARCH_TOPK", "5"))
	if err != nil {
		vectorSearchTopK = 5
	}

	docProcessTimeout, err := strconv.Atoi(getEnv("DOC_PROCESS_TIMEOUT", "60"))
	if err != nil {
		docProcessTimeout = 60
	}

	embedTimeout, err := strconv.Atoi(getEnv("EMBED_TIMEOUT", "30"))
	if err != nil {
		embedTimeout = 30
	}

	llmTimeout, err := strconv.Atoi(getEnv("LLM_TIMEOUT", "60"))
	if err != nil {
		llmTimeout = 60
	}

	searchCacheTTL, err := strconv.Atoi(getEnv("SEARCH_CACHE_TTL", "30"))
	if err != nil {
		searchCacheTTL = 30
	}

	minQueryLength, err := strconv.Atoi(getEnv("MIN_QUERY_LENGTH", "3"))
	if err != nil {
		minQueryLength = 3
	}

	maxConcurrentSearches, err := strconv.Atoi(getEnv("MAX_CONCURRENT_SEARCHES", "10"))
	if err != nil {
		maxConcurrentSearches = 10
	}

	searchWorkerCount, err := strconv.Atoi(getEnv("SEARCH_WORKER_COUNT", "4"))
	if err != nil {
		searchWorkerCount = 4
	}

	rateLimitPerMin, err := strconv.Atoi(getEnv("RATE_LIMIT_PER_MIN", "60"))
	if err != nil {
		rateLimitPerMin = 60
	}

	maxQuerySize, err := strconv.Atoi(getEnv("MAX_QUERY_SIZE", "1000"))
	if err != nil {
		maxQuerySize = 1000
	}

	// 解析布尔配置项
	enableRateLimit := getEnv("ENABLE_RATE_LIMIT", "true") == "true"
	enableMetrics := getEnv("ENABLE_METRICS", "true") == "true"
	debug := getEnv("DEBUG", "false") == "true"

	// 解析CORS配置
	corsOrigins := getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000,http://localhost:3001")
	corsOriginList := strings.Split(corsOrigins, ",")
	for i := range corsOriginList {
		corsOriginList[i] = strings.TrimSpace(corsOriginList[i])
	}

	// 解析允许的IP列表
	allowedIPs := getEnv("ALLOWED_IPS", "")
	allowedIPList := []string{}
	if allowedIPs != "" {
		allowedIPList = strings.Split(allowedIPs, ",")
		for i := range allowedIPList {
			allowedIPList[i] = strings.TrimSpace(allowedIPList[i])
		}
	}

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8082"),
		ServerHost:  getEnv("SERVER_HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       debug,

		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "3306"),
		DBUser:            getEnv("DB_USER", "admin"), // 与其他服务保持一致
		DBPassword:        getEnv("DB_PASSWORD", "password"),
		DBName:            getEnv("DB_NAME", "docu_sage"), // 使用统一的数据库名
		DBMaxIdleConns:    dbMaxIdleConns,
		DBMaxOpenConns:    dbMaxOpenConns,
		DBConnMaxLifetime: dbConnMaxLifetime,
		DBConnMaxIdleTime: dbConnMaxIdleTime,

		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		RedisEnableTLS:     redisEnableTLS,
		RedisSkipTLSVerify: redisSkipTLSVerify,
		RedisDialTimeout:   redisDialTimeout,
		RedisReadTimeout:   redisReadTimeout,
		RedisWriteTimeout:  redisWriteTimeout,
		RedisPoolSize:      redisPoolSize,
		RedisMinIdleConns:  redisMinIdleConns,
		RedisMaxRetries:    redisMaxRetries,
		RedisMaxIdleConns:  redisMaxIdleConns,
		RedisIdleTimeout:   redisIdleTimeout,

		JWTSecretKey:  getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
		JWTAlgorithm:  getEnv("JWT_ALGORITHM", "HS256"),
		JWTExpiration: jwtExpiration,

		VectorSearchServiceAddr: getEnv("VECTOR_SEARCH_SERVICE_ADDR", "http://localhost:8087"),
		VectorSearchTimeout:     vectorSearchTimeout,
		VectorSearchTopK:        vectorSearchTopK,

		DocProcessServiceAddr: getEnv("DOC_PROCESS_SERVICE_ADDR", "http://localhost:8001"), // 与文档处理服务端口保持一致
		DocProcessTimeout:     docProcessTimeout,

		EmbedServiceAddr: getEnv("EMBED_SERVICE_ADDR", "http://localhost:8089"),
		EmbedTimeout:     embedTimeout,

		LLMServiceAddr: getEnv("LLM_SERVICE_ADDR", "http://localhost:8090"),
		LLMTimeout:     llmTimeout,

		CORSAllowOrigins: corsOriginList,

		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFile:  getEnv("LOG_FILE", "query-service.log"),

		MaxSearchResults:      maxSearchResults,
		DefaultSearchLimit:    defaultSearchLimit,
		SearchHistoryMaxItems: searchHistoryMaxItems,
		SearchCacheTTL:        searchCacheTTL,
		MinQueryLength:        minQueryLength,

		MaxConcurrentSearches: maxConcurrentSearches,
		SearchWorkerCount:     searchWorkerCount,

		EnableRateLimit: enableRateLimit,
		RateLimitPerMin: rateLimitPerMin,
		MaxQuerySize:    maxQuerySize,
		AllowedIPs:      allowedIPList,

		EnableMetrics: enableMetrics,
	}, nil
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// SetupLogger 设置日志
func SetupLogger(config *Config) (*zap.Logger, error) {
	var loggerConfig zap.Config

	switch config.LogLevel {
	case "debug":
		loggerConfig = zap.NewDevelopmentConfig()
	case "info":
		loggerConfig = zap.NewProductionConfig()
		loggerConfig.EncoderConfig.TimeKey = "time"
		loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		loggerConfig = zap.NewProductionConfig()
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// DBConnectionURL 获取数据库连接URL
func (cfg *Config) DBConnectionURL() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
}

// RedisURL 获取Redis连接URL
func (cfg *Config) RedisURL() string {
	protocol := "redis"
	if cfg.RedisEnableTLS {
		protocol = "rediss"
	}
	return fmt.Sprintf("%s://:%s@%s/%d", protocol, cfg.RedisPassword, strings.TrimPrefix(cfg.RedisAddr, "http://"), cfg.RedisDB)
}
