package middleware

import (
	"fmt"
	"log"
	"time"
	"video_feedsystem/dal/redis"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// 限制当前登录用户指定操作的请求频率
func RateLimitByAccount(action string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		value, exist := c.Get("accountID")
		if !exist {
			httpx.WriteError(c, apperr.New(apperr.KindUnauthorized, "用户未登录"))
			c.Abort()
			return
		}

		accountID, ok := value.(int64)
		if !ok {
			httpx.WriteError(c, apperr.New(apperr.KindInternal, "服务器内部错误，请稍后再试"))
			c.Abort()
			return
		}

		key := fmt.Sprintf("rate_limit:%s:%d", action, accountID)
		allowed, err := redis.AllowRequest(ctx, key, limit, window)
		if err != nil {
			log.Printf("Redis 限流失败，放行请求，key=%s，error=%v", key, err)
			c.Next()
			return
		}

		if !allowed {
			httpx.WriteError(c, apperr.New(
				apperr.KindTooManyRequests,
				"请求过于频繁，请稍后再试",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
