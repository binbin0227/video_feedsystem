package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouter(h *server.Hertz) {
	account := h.Group("/account")
	video := h.Group("/video")
	feed := h.Group("/feed")
}