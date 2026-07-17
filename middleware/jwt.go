package middleware

import (
	"context"
	"strings"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/utils"

	"github.com/cloudwego/hertz/pkg/app"
)

func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 1. 从 HTTP Header 中获取 Authorization: Bearer <token>
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

		// 2. 检查 token 有效性
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindUnauthorized, "Token 已过期或无效"))
			c.Abort()
			return
		}

		// 3. 验证成功，将 accountID 写入 c
		c.Set("accountID", claims.AccountID)
		c.Next(ctx)
	}
}
