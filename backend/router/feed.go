package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// registerFeedRoutes 注册公共视频流和登录后的关注视频流接口。
func registerFeedRoutes(h *server.Hertz) {
	feed := h.Group("/feed")
	{
		feed.GET("/list", handler.ListFeed)
		authorized := feed.Group("", middleware.JWTAuth())
		{
			authorized.GET("/following", handler.ListFollowingFeed)
		}
	}
}
