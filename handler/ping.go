package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Ping 是一个测试接口，用来确认 Hertz 框架是否正常运行
func Ping(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(consts.StatusOK, utils.H{
		"message": "pong",
		"status":  "Hertz 启动成功",
	})
}
