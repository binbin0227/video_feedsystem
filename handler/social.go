package handler

import (
	"context"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type FollowRequest struct {
	VloggerID string `json:"vlogger_id"`
}

func FollowUser(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req FollowRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}
	vloggerID, err := parsePositiveInt64String(req.VloggerID, "vlogger_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 读取 accountID
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. service.FollowUser
	if err := service.FollowUser(ctx, followerID, vloggerID); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 返回结果
	c.JSON(consts.StatusOK, map[string]string{
		"message": "关注成功",
	})
}

func UnfollowUser(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req FollowRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}
	vloggerID, err := parsePositiveInt64String(req.VloggerID, "vlogger_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 读取 accountID
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. service.UnfollowUser
	if err := service.UnfollowUser(ctx, followerID, vloggerID); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 返回结果
	c.JSON(consts.StatusOK, map[string]string{
		"message": "取消关注成功",
	})
}

func GetFollowStatus(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	vloggerID, err := parsePositiveInt64Query(c, "vlogger_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 读取 accountID
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. service.CheckFollowStatus
	following, err := service.CheckFollowStatus(ctx, followerID, vloggerID)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 返回结果
	c.JSON(consts.StatusOK, map[string]bool{
		"is_following": following,
	})
}