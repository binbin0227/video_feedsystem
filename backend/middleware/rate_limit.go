package middleware

import (
	"context"
	"fmt"
	"time"
	"video_feedsystem/dal/redis"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// RateLimitByAccount 根据当前登录用户限制指定操作的请求频率。
func RateLimitByAccount(action string, limit int64, window time.Duration) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		value, exist := c.Get("accountID")
		if !exist {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "用户未登录"))
			c.Abort()
			return
		}

		accountID, ok := value.(int64)
		if !ok {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindInternal, "服务器内部错误，请稍后再试"))
			c.Abort()
			return
		}

		key := fmt.Sprintf("rate_limit:%s:%d", action, accountID)
		allowed, err := redis.AllowRequest(ctx, key, limit, window)
		if err != nil {
			hlog.CtxWarnf(ctx, "Redis 限流失败，放行请求，key=%s，error=%v", key, err)
			c.Next(ctx)
			return
		}

		if !allowed {
			httpx.WriteError(ctx, c, apperr.New(
				apperr.KindTooManyRequests,
				"请求过于频繁，请稍后再试",
			))
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}
