package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

var rdb *goredis.Client

func InitRedis(addr string) error {
	rdb = goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return nil
}
