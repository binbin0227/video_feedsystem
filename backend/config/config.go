package config

import "os"

// 保存服务启动所需的数据库、JWT 和监听地址配置。
type Config struct {
	MySQLDSN    string
	RedisAddr   string
	RedisPwd    string
	RabbitMQURL string
	JWTSecret   string
	HostPorts   string
}

func Load() Config {
	return Config{
		MySQLDSN:    getEnv("MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/video_feedsystem?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret:   getEnv("JWT_SECRET", "feedsystem-dev-secret-key"),
		HostPorts:   getEnv("HOST_PORTS", "0.0.0.0:20000"),
		RedisAddr:   getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPwd:    getEnv("REDIS_PWD", "123456"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/"),
	}
}

// 优先读取环境变量，未配置时回退到本地开发默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
