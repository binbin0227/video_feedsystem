package router

import (
	"time"
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func registerVideoRoutes(h *server.Hertz) {
	video := h.Group("/video")
	{
		video.GET("/list-by-author-id", handler.ListByAuthorID)
		video.GET("/detail", handler.GetVideoDetail)

		authorized := video.Group("", middleware.JWTAuth())
		{
			authorized.POST("/upload-video", middleware.RateLimitByAccount("upload", 10, time.Minute), handler.UploadVideo)
			authorized.POST("/upload-cover", middleware.RateLimitByAccount("upload", 10, time.Minute), handler.UploadCover)
			authorized.POST("/publish", middleware.RateLimitByAccount("publish_video", 10, time.Minute), handler.PublishVideo)
			authorized.POST("/like", middleware.RateLimitByAccount("like_operation", 120, time.Minute), handler.LikeVideo)
			authorized.POST("/unlike", middleware.RateLimitByAccount("like_operation", 120, time.Minute), handler.UnlikeVideo)
			authorized.GET("/like-status", handler.GetLikeStatus)
			authorized.GET("/liked", handler.GetLikedVideoList)
		}
	}
}
