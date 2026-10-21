package middleware

import (
	"context"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"

	"github.com/cloudwego/hertz/pkg/app"
)

// 根据配置统一控制视频上传、封面上传和视频发布接口
func RequireUploadEnabled(enabled bool) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !enabled {
			httpx.WriteError(ctx, c, apperr.New(apperr.KindForbidden, "当前已暂停视频上传和发布"))
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}
