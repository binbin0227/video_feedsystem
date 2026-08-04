package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

var rdb *goredis.Client

// InitRedis 创建 Redis 客户端并检查连接是否可用。
func InitRedis(addr string, pwd string) error {
	rdb = goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: pwd,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return nil
}
