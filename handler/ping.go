package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Ping 用于确认服务是否正常启动。
func Ping(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, map[string]string{
		"message": "pong",
		"status":  "Hertz 启动成功",
	})
}
