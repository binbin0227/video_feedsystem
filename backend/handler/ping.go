package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 确认服务是否正常启动
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"message": "pong",
		"status":  "Gin 启动成功",
	})
}
