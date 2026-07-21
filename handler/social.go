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

// GetFollowingList 分页查询当前用户关注的账号。
func GetFollowingList(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 cursor 和 limit
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

	// 2. 获取 accountID
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. service.GetFollowingList
	result, err := service.GetFollowingList(ctx, followerID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 将下一页游标转换为字符串
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	// 5. 返回结果
	c.JSON(consts.StatusOK, FollowingOrFollowerListResponse{
		Accounts:   newFollowingOrFollowerAccountListResponse(result.Accounts),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// GetFollowerList 分页查询当前用户粉丝的账号。
func GetFollowerList(ctx context.Context, c *app.RequestContext) {
	// 1. cursor 和 limit
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

	// 2. 获取 accountID
	vloggerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. service.GetFollowerList
	result, err := service.GetFollowerList(ctx, vloggerID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 将下一页游标转换为字符串
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	// 5. 返回结果
	c.JSON(consts.StatusOK, FollowingOrFollowerListResponse{
		Accounts:   newFollowingOrFollowerAccountListResponse(result.Accounts),
		HasMore:    result.HasMore,
		NextCursor: nextCursor,
	})
}
