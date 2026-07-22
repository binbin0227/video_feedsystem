package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// registerCommentRoutes 注册公开评论列表以及需要 JWT 的发布、删除接口。
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
