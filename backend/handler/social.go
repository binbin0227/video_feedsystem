package handler

import (
	"context"
	"strconv"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type FollowRequest struct {
	VloggerID string `json:"vlogger_id"`
}

// 让当前登录用户关注目标账号
func FollowUser(ctx context.Context, c *app.RequestContext) {
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
	
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	if err := service.FollowUser(ctx, followerID, vloggerID); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, map[string]string{
		"message": "关注成功",
	})
}

// 取消当前登录用户对目标账号的关注
func UnfollowUser(ctx context.Context, c *app.RequestContext) {
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

	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	if err := service.UnfollowUser(ctx, followerID, vloggerID); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, map[string]string{
		"message": "取消关注成功",
	})
}

// 返回当前登录用户对目标账号的关注状态
func GetFollowStatus(ctx context.Context, c *app.RequestContext) {
	vloggerID, err := parsePositiveInt64Query(c, "vlogger_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	following, err := service.CheckFollowStatus(ctx, followerID, vloggerID)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, map[string]bool{
		"is_following": following,
	})
}

// 分页查询当前用户关注的账号
func GetFollowingList(ctx context.Context, c *app.RequestContext) {
	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	result, err := service.GetFollowingList(ctx, followerID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	c.JSON(consts.StatusOK, FollowingOrFollowerListResponse{
		Accounts:   newFollowingOrFollowerAccountListResponse(result.Accounts),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 分页查询当前用户粉丝的账号
func GetFollowerList(ctx context.Context, c *app.RequestContext) {
	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	vloggerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	result, err := service.GetFollowerList(ctx, vloggerID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	c.JSON(consts.StatusOK, FollowingOrFollowerListResponse{
		Accounts:   newFollowingOrFollowerAccountListResponse(result.Accounts),
		HasMore:    result.HasMore,
		NextCursor: nextCursor,
	})
}
