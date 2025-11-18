package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 认证服务配置
type Config struct {
	// 服务器配置
	ServerPort string
	ServerHost string

	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis配置
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// JWT配置
	JWTSecretKey           string
	JWTExpireHours         int
	RefreshTokenExpireDays int

	// CORS配置
	CORSAllowOrigins []string

	// 日志配置
	LogLevel string

	// 安全配置
	MinPasswordLength      int
	MaxLoginAttempts       int
	LockoutDurationMinutes int
	Env                    string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	// 加载.env文件
	_ = godotenv.Load()

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	jwtExpireHours, err := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "2"))
	if err != nil {
		jwtExpireHours = 2
	}

	refreshTokenExpireDays, err := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRE_DAYS", "7"))
	if err != nil {
		refreshTokenExpireDays = 7
	}

	minPasswordLength, err := strconv.Atoi(getEnv("MIN_PASSWORD_LENGTH", "6"))
	if err != nil {
		minPasswordLength = 8
	}

	maxLoginAttempts, err := strconv.Atoi(getEnv("MAX_LOGIN_ATTEMPTS", "5"))
	if err != nil {
		maxLoginAttempts = 5
	}

	lockoutDurationMinutes, err := strconv.Atoi(getEnv("LOCKOUT_DURATION_MINUTES", "30"))
	if err != nil {
		lockoutDurationMinutes = 30
	}

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8081"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "123456"),
		DBName:     getEnv("DB_NAME", "auth_service"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		JWTSecretKey:           getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
		JWTExpireHours:         jwtExpireHours,
		RefreshTokenExpireDays: refreshTokenExpireDays,

		CORSAllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
		},

		LogLevel: getEnv("LOG_LEVEL", "info"),
		Env:      getEnv("ENV", "development"),

		MinPasswordLength:      minPasswordLength,
		MaxLoginAttempts:       maxLoginAttempts,
		LockoutDurationMinutes: lockoutDurationMinutes,
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
