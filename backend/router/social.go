package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/gin-gonic/gin"
)

func registerSocialRoutes(h *gin.Engine) {
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
