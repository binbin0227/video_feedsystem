package router

import (
	"video_feedsystem/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouter(h *server.Hertz) {
	h.GET("/ping", handler.Ping)
	// URL /uploads/... 对应本地目录 .run/uploads/...
	h.Static("/uploads", "./.run")

	registerAccountRoutes(h)
	registerVideoRoutes(h)
	registerFeedRoutes(h)
	registerCommentRoutes(h)
	registerSocialRoutes(h)
}
