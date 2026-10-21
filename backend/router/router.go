package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/storage"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitRouter(h *server.Hertz, uploadEnabled bool) {
	h.GET("/ping", handler.Ping)
	h.GET("/metrics", adaptor.HertzHandler(promhttp.Handler()))

	// 映射成了静态文件并支持分段加载
	h.StaticFS("/uploads", &app.FS{Root: storage.Root(), AcceptByteRange: true})

	registerAccountRoutes(h)
	registerVideoRoutes(h, uploadEnabled)
	registerFeedRoutes(h)
	registerCommentRoutes(h)
	registerSocialRoutes(h)
}
