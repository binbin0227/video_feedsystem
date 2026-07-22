package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// registerVideoRoutes 注册公开视频查询以及需要 JWT 的上传、发布和点赞接口。
func registerVideoRoutes(h *server.Hertz) {
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
			authorized.POST("/unlike", handler.UnlikeVideo)
			authorized.GET("/like-status", handler.GetLikeStatus)
			authorized.GET("/liked", handler.GetLikedVideoList)
		}
	}
}
