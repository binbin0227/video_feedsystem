package router

import (
	"video_feedsystem/handler"
	"video_feedsystem/middleware"

	"github.com/gin-gonic/gin"
)

func registerAccountRoutes(h *gin.Engine) {
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
