package handler

import (
	"context"
	"strconv"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type FeedResponse struct {
	Videos     []VideoResponse `json:"videos"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

func ListFeed(ctx context.Context, c *app.RequestContext) {
	// 1. 读取 cursor limit
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

	// 2. service.GetFeed
	result, err := service.GetFeed(ctx, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. 处理 nextCursor
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	// 4. 返回结果
	c.JSON(consts.StatusOK, FeedResponse{
		Videos:     newVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// ListFollowingFeed 分页返回当前用户的关注流。
func ListFollowingFeed(ctx context.Context, c *app.RequestContext) {
	// 1. 获取当前登录用户
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 解析 cursor 和 limit
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

	// 3. 查询关注流
	result, err := service.GetFollowingFeed(ctx, followerID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 将游标转换为字符串
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	// 5. 返回结果
	c.JSON(consts.StatusOK, FeedResponse{
		Videos:     newVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}
