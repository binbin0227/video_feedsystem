package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouter(h *server.Hertz) {
	h.GET("/ping", handler.Ping)
	// URL /uploads/... 对应本地目录 .run/uploads/...
	h.Static("/uploads", "./.run")

	account := h.Group("/account")
	{
		account.POST("/register", handler.Register)
		account.POST("/login", handler.Login)
	}

	video := h.Group("/video")
	{
		video.GET("/list-by-author-id", handler.ListByAuthorID)
		video.GET("/detail", handler.GetVideoDetail)

		authorized := video.Group("", middleware.JWTAuth())
		{
			authorized.POST("/upload-video", handler.UploadVideo)
			authorized.POST("/upload-cover", handler.UploadCover)
			authorized.POST("/publish", handler.PublishVideo)
			authorized.POST("/like", handler.LikeVideo)
		}
	}

	feed := h.Group("/feed")
	{
		feed.GET("/list", handler.ListFeed)
	}
}
