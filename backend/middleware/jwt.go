package middleware

import (
	"context"
	"strings"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/utils"

	"github.com/cloudwego/hertz/pkg/app"
)

// JWTAuth 校验 Bearer Token，并把账号 ID 写入请求上下文供后续 Handler 使用。
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
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "Token 已过期或无效"))
			c.Abort()
			return
		}
		c.Set("accountID", claims.AccountID)
		c.Next(ctx)
	}
}
