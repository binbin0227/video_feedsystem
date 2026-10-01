package middleware

import (
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// 根据配置统一控制视频上传、封面上传和视频发布接口
func RequireUploadEnabled(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			httpx.WriteError(c, apperr.New(apperr.KindForbidden, "当前已暂停视频上传和发布"))
			c.Abort()
			return
		}

		c.Next()
	}
}
