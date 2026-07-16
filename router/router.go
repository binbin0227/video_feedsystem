package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouter(h *server.Hertz) {

	h.Static("/static", "./.run/uploads")

	account := h.Group("/account")
	{
		account.POST("/register", handler.Register)
		account.POST("/login", handler.Login)
	}
	video := h.Group("/video", middleware.JWTAuth())
	{
		// 公共接口
		// video.GET("/list-by-author-id", ...)
		// video.GET("/detail", ...)

		// 需要登录的接口
		authorized := video.Group("", middleware.JWTAuth())
		{
			authorized.POST("/upload-video", handler.UploadVideo)
			authorized.POST("/upload-cover", handler.UploadCover)
			authorized.POST("/publish", handler.PublishVideo)
		}
	}
	// feed := h.Group("/feed")
}
