package router

import (
	"time"
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/gin-gonic/gin"
)

func registerCommentRoutes(h *gin.Engine) {
	comment := h.Group("/comment")
	{
		comment.GET("/list", handler.ListComments)

		authorized := comment.Group("", middleware.JWTAuth())
		{
			authorized.POST("/publish", middleware.RateLimitByAccount("comment", 20, time.Minute), handler.PublishComment)
			authorized.DELETE("/delete", handler.DeleteComment)
		}
	}
}
