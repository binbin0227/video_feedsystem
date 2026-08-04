package handler

import (
	"context"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// RegisterRequest 表示注册接口的 JSON 请求体。
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest 表示登录接口的 JSON 请求体。
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register 处理用户注册请求。
func Register(ctx context.Context, c *app.RequestContext) {
	var req RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}
	if err := service.Register(ctx, req.Username, req.Password); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]string{"message": "账号注册成功！"})
}

// Login 处理用户登录请求并返回 JWT。
func Login(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}
	token, err := service.Login(ctx, req.Username, req.Password)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]string{"token": token})
}

// GetAccountProfile 返回指定账号的主页信息和统计数据。
func GetAccountProfile(ctx context.Context, c *app.RequestContext) {
	accountID, err := parsePositiveInt64Query(c, "account_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	profile, err := service.GetAccountProfile(ctx, accountID)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]any{
		"profile": newAccountProfileResponse(profile),
	})
}

// SearchAccounts 根据用户名关键词搜索用户。
func SearchAccounts(ctx context.Context, c *app.RequestContext) {
	keyword := c.Query("keyword")
	accounts, err := service.SearchAccounts(ctx, keyword)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, AccountSearchResponse{
		Accounts: newAccountSearchListResponse(accounts),
	})
}
