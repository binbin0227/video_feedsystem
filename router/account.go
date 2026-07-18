package router

import (
	"video_feedsystem/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func registerAccountRoutes(h *server.Hertz) {
	account := h.Group("/account")
	{
		account.POST("/register", handler.Register)
		account.POST("/login", handler.Login)
	}
}