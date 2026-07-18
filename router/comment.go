package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func registerCommentRoutes(h *server.Hertz) {
	comment := h.Group("/comment")
	{
		comment.GET("/list", handler.ListComments)

		authorized := comment.Group("", middleware.JWTAuth())
		{
			authorized.POST("/publish", handler.PublishComment)
			authorized.DELETE("/delete", handler.DeleteComment)
		}
	}
}