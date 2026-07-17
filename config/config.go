package config

import "os"

type Config struct {
	MySQLDSN  string
	JWTSecret string
	HostPorts string
}

// Load 读取环境变量；没有设置时使用本地开发默认值。
func Load() Config {
	return Config{
		MySQLDSN:  getEnv("MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/video_feedsystem?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret: getEnv("JWT_SECRET", "feedsystem-dev-secret-key"),
		HostPorts: getEnv("HOST_PORTS", "0.0.0.0:20000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
