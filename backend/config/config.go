package config

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// 保存服务启动所需的数据库、JWT 和监听地址配置。
type Config struct {
	MySQLDSN    string
	RedisAddr   string
	RedisPwd    string
	RabbitMQURL string
	JWTSecret   string
	HostPorts   string
	CORSOrigins []string
	UploadRoot  string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("读取 .env 文件失败: %v", err)
	}

	cfg := Config{
		MySQLDSN:    getEnv("MYSQL_DSN", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		HostPorts:   getEnv("HOST_PORTS", "0.0.0.0:8080"),
		RedisAddr:   getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPwd:    getEnv("REDIS_PWD", ""),
		RabbitMQURL: getEnv("RABBITMQ_URL", ""),
		CORSOrigins: splitCSV(getEnv("CORS_ORIGINS", "")),
		UploadRoot:  getEnv("UPLOAD_ROOT", "./.run"),
	}

	if cfg.MySQLDSN == "" {
		return Config{}, errors.New("MYSQL_DSN 不能为空")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET 不能为空")
	}
	if cfg.RedisPwd == "" {
		return Config{}, errors.New("REDIS_PWD 不能为空")
	}
	if cfg.RabbitMQURL == "" {
		return Config{}, errors.New("RABBITMQ_URL 不能为空")
	}
	if len(cfg.CORSOrigins) == 0 {
		return Config{}, errors.New("CORS_ORIGINS 不能为空")
	}

	return cfg, nil
}

// 优先读取环境变量，未配置时回退到本地开发默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}

	return result
}
