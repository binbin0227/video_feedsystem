package redis

import (
	"context"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"time"
)

var rateLimitScript = goredis.NewScript(`
	local current = redis.call("INCR", KEYS[1])

	if current == 1 then
		redis.call("PEXPIRE", KEYS[1], ARGV[2])
	end

	if current > tonumber(ARGV[1]) then
		return 0
	end

	return 1
`)

// 在固定时间窗口内增加请求次数，并判断是否允许本次请求
func AllowRequest(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("限流 Key 不能为空")
	}
	if limit <= 0 {
		return false, fmt.Errorf("限流次数必须大于 0")
	}
	if window <= 0 {
		return false, fmt.Errorf("限流时间窗口必须大于 0")
	}

	result, err := rateLimitScript.Run(
		ctx,
		rdb,
		[]string{key},
		limit,
		window.Milliseconds(),
	).Int()
	if err != nil {
		return false, fmt.Errorf("执行 Redis 限流脚本失败: %w", err)
	}

	return result == 1, nil
}
