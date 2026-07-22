package router

import (
	"video_feedsystem/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// registerAccountRoutes 注册账号注册、登录、主页和搜索接口。
func registerAccountRoutes(h *server.Hertz) {
	account := h.Group("/account")
	{
		account.POST("/register", handler.Register)
		account.POST("/login", handler.Login)
		account.GET("/profile", handler.GetAccountProfile)
		account.GET("/search", handler.SearchAccounts)
	}
}
