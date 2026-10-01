package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/gin-gonic/gin"
)

func registerFeedRoutes(h *gin.Engine) {
	feed := h.Group("/feed")
	{
		feed.GET("/hot", handler.ListHotFeed)
		feed.GET("/list", handler.ListFeed)
		authorized := feed.Group("", middleware.JWTAuth())
		{
			authorized.GET("/following", handler.ListFollowingFeed)
		}
	}
}
