package handler

import (
	"net/http"
	"strconv"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/gin-gonic/gin"
)

type FeedResponse struct {
	Videos     []VideoResponse `json:"videos"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

// 分页返回公共视频流
func ListFeed(c *gin.Context) {
	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	result, err := service.GetFeed(c.Request.Context(), cursor, limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}
	c.JSON(http.StatusOK, FeedResponse{
		Videos:     newFeedVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 分页返回当前用户的关注流
func ListFollowingFeed(c *gin.Context) {
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	result, err := service.GetFollowingFeed(c.Request.Context(), followerID, cursor, limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}
	c.JSON(http.StatusOK, FeedResponse{
		Videos:     newFeedVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 返回全站热度最高的 10 个视频
func ListHotFeed(c *gin.Context) {
	videos, err := service.GetHotFeed(c.Request.Context())
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, map[string]any{
		"videos": newFeedVideoListResponse(videos),
	})
}
