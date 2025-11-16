package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 存储网关配置
type Config struct {
	Env                string
	ServerPort         string
	AuthServiceAddr    string
	QueryServiceAddr   string
	DocServiceAddr     string
	KnowledgeBaseAddr  string
	ToolServiceAddr    string
	JWTSecretKey       string
	JWTExpirationMin   int
	RateLimitPerIP     int
	RateLimitPerUser   int
	RateLimitWindowSec int
	AllowOrigins       []string
	LogLevel           string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	// 尝试加载.env文件
	_ = godotenv.Load()

	return &Config{
		Env:                getEnv("ENV", "development"),
		ServerPort:         getEnv("SERVER_PORT", "8000"),
		AuthServiceAddr:    getEnv("AUTH_SERVICE_ADDR", "localhost:8001"),
		QueryServiceAddr:   getEnv("QUERY_SERVICE_ADDR", "localhost:8002"),
		DocServiceAddr:     getEnv("DOC_SERVICE_ADDR", "localhost:8003"),
		KnowledgeBaseAddr:  getEnv("KNOWLEDGE_BASE_ADDR", "localhost:8004"),
		ToolServiceAddr:    getEnv("TOOL_SERVICE_ADDR", "localhost:8005"),
		JWTSecretKey:       getEnv("JWT_SECRET_KEY", "default-secret-key"),
		JWTExpirationMin:   getEnvAsInt("JWT_EXPIRATION_MIN", 60),
		RateLimitPerIP:     getEnvAsInt("RATE_LIMIT_PER_IP", 100),
		RateLimitPerUser:   getEnvAsInt("RATE_LIMIT_PER_USER", 300),
		RateLimitWindowSec: getEnvAsInt("RATE_LIMIT_WINDOW_SEC", 60),
		AllowOrigins:       []string{getEnv("ALLOW_ORIGINS", "*")},
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt 获取环境变量并转换为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}