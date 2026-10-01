package router

import (
	"net/http"
	"path/filepath"

	"video_feedsystem/handler"
	"video_feedsystem/storage"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitRouter(h *gin.Engine, uploadEnabled bool) {
	h.GET("/ping", handler.Ping)
	h.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 映射成了静态文件并支持分段加载
	h.StaticFS("/uploads", http.Dir(filepath.Join(storage.Root(), "uploads")))

	registerAccountRoutes(h)
	registerVideoRoutes(h, uploadEnabled)
	registerFeedRoutes(h)
	registerCommentRoutes(h)
	registerSocialRoutes(h)
}
