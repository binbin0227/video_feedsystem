package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// registerSocialRoutes 注册关注状态、关注操作和关注关系列表接口。
func registerSocialRoutes(h *server.Hertz) {
	social := h.Group("/social")
	{
		authorized := social.Group("", middleware.JWTAuth())
		{
			authorized.POST("/follow", handler.FollowUser)
			authorized.POST("/unfollow", handler.UnfollowUser)
			authorized.GET("/status", handler.GetFollowStatus)
			authorized.GET("/following", handler.GetFollowingList)
			authorized.GET("/followers", handler.GetFollowerList)
		}
	}
}
