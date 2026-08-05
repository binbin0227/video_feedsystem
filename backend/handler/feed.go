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

// 分页返回公共视频流
func ListFeed(ctx context.Context, c *app.RequestContext) {
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
	result, err := service.GetFeed(ctx, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}
	c.JSON(consts.StatusOK, FeedResponse{
		Videos:     newFeedVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 分页返回当前用户的关注流
func ListFollowingFeed(ctx context.Context, c *app.RequestContext) {
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
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
	result, err := service.GetFollowingFeed(ctx, followerID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}
	c.JSON(consts.StatusOK, FeedResponse{
		Videos:     newFeedVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 返回全站热度最高的 10 个视频
func ListHotFeed(ctx context.Context, c *app.RequestContext) {
	videos, err := service.GetHotFeed(ctx)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]any{
		"videos": newFeedVideoListResponse(videos),
	})
}
