package handler

import (
	"context"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register 处理用户注册请求。
func Register(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	// 2. 调用 service.Register
	if err := service.Register(ctx, req.Username, req.Password); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. 注册成功
	c.JSON(consts.StatusOK, map[string]string{"message": "账号注册成功！"})
}

func Login(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	// 2. 调用 service.Login 并返回 token
	token, err := service.Login(ctx, req.Username, req.Password)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. 登录成功，向前端返回 token
	c.JSON(consts.StatusOK, map[string]string{"token": token})
}
