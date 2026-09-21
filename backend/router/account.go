package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func registerAccountRoutes(h *server.Hertz) {
	account := h.Group("/account")
	{
		account.POST("/register", handler.Register)
		account.POST("/login", handler.Login)
		account.POST("/refresh", handler.RefreshAuthTokens)
		account.GET("/profile", handler.GetAccountProfile)
		account.GET("/search", handler.SearchAccounts)

		authorized := account.Group("", middleware.JWTAuth())
		authorized.POST("/logout", handler.Logout)
	}
}
