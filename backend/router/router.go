package router

import (
	"video_feedsystem/handler"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// InitRouter 注册健康检查、静态媒体目录和各业务模块路由。
func InitRouter(h *server.Hertz) {
	h.GET("/ping", handler.Ping)
	h.StaticFS("/uploads", &app.FS{Root: "./.run", AcceptByteRange: true})

	registerAccountRoutes(h)
	registerVideoRoutes(h)
	registerFeedRoutes(h)
	registerCommentRoutes(h)
	registerSocialRoutes(h)
}
