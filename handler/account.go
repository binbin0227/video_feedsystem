package handler

import (
	"context"
	"errors"
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

// 处理用户注册的 http 请求
func Register(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 json
	var req RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{
			"error": "JSON 解析失败",
		})
		return
	}
	if req.Username == "" || req.Password == "" {
		c.JSON(consts.StatusBadRequest, map[string]string{
			"error": "用户名或密码不能为空",
		})
		return
	}

	// 2. 注册
	err := service.Register(ctx, req.Username, req.Password)

	// 3. 返回结果
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			c.JSON(consts.StatusConflict, map[string]string{
				"error": err.Error(),
			})
			return
		}
		c.JSON(consts.StatusInternalServerError, map[string]string{
			"error": "服务器内部错误，请稍后再试",
		})
		return
	}
	c.JSON(consts.StatusOK, map[string]string{
		"message": "账号注册成功！",
	})
}

// 处理用户登录的 http 请求
func Login(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 json
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{
			"error": "JSON 解析失败",
		})
		return
	}

	// 2. 登录
	tokenStr, err := service.Login(ctx, req.Username, req.Password)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 3. 登录成功则返回 token
	c.JSON(consts.StatusOK, map[string]string{
		"token": tokenStr,
	})
}
