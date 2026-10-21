package middleware

import (
	"context"
	"strings"

	"video_feedsystem/dal/redis"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/utils"

	"github.com/cloudwego/hertz/pkg/app"
)

// 校验 Bearer Token 并把 accountID 写入请求上下文
func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "请求未携带 Token"))
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "Token 格式错误"))
			c.Abort()
			return
		}

		claims, err := utils.ParseAccessToken(parts[1])
		if err != nil {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "Token 已过期或无效"))
			c.Abort()
			return
		}
		if claims.ID == "" || claims.ExpiresAt == nil {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "Token 已过期或无效"))
			c.Abort()
			return
		}

		revoked, err := redis.IsAccessTokenRevoked(ctx, claims.ID)
		if err != nil {
			httpx.WriteError(ctx, c, apperr.Wrap(apperr.KindInternal, "登录状态校验失败，请稍后再试", err))
			c.Abort()
			return
		}
		if revoked {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "Token 已失效，请重新登录"))
			c.Abort()
			return
		}

		c.Set("accountID", claims.AccountID)
		c.Set("accessTokenID", claims.ID)
		c.Set("accessTokenExpiresAt", claims.ExpiresAt.Time)
		c.Next(ctx)
	}
}
