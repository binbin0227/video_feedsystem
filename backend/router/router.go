package router

import (
	"video_feedsystem/handler"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouter(h *server.Hertz) {
	h.GET("/ping", handler.Ping)

	// 映射成了静态文件并支持分段加载
	h.StaticFS("/uploads", &app.FS{Root: "./.run", AcceptByteRange: true})

	registerAccountRoutes(h)
	registerVideoRoutes(h)
	registerFeedRoutes(h)
	registerCommentRoutes(h)
	registerSocialRoutes(h)
}
