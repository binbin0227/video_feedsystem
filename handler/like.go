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

type LikeRequest struct {
	VideoID string `json:"video_id"`
}

type LikedVideoListResponse struct {
	Videos     []VideoResponse `json:"videos"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

func LikeVideo(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req LikeRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	// 2. 将 videoID 转换成 int64
	videoID, err := parsePositiveInt64String(req.VideoID, "video_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. 解析 accountID
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. service.LikeVideo
	if err := service.LikeVideo(ctx, accountID, videoID); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 5. 返回结果
	c.JSON(consts.StatusOK, map[string]string{"message": "点赞成功"})
}

func UnlikeVideo(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req LikeRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	// 2. 将 videoID 转换成 int64
	videoID, err := parsePositiveInt64String(req.VideoID, "video_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. 解析 accountID
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. service.LikeVideo
	if err := service.UnlikeVideo(ctx, accountID, videoID); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 5. 返回结果
	c.JSON(consts.StatusOK, map[string]string{
		"message": "取消点赞成功",
	})
}

func GetLikeStatus(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 video_id
	videoID, err := parsePositiveInt64Query(c, "video_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 解析 accountID
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. service.CheckLikeStatus
	liked, err := service.CheckLikeStatus(ctx, accountID, videoID)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 返回结果
	c.JSON(consts.StatusOK, map[string]bool{"is_liked": liked})
}

// GetLikedVideoList 分页查询当前用户点赞过的视频。
func GetLikedVideoList(ctx context.Context, c *app.RequestContext) {
	// 1. 获取当前登录用户
	accountID, err := getAccountID(c)
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

	// 3. 查询点赞视频列表
	result, err := service.GetLikedVideoList(ctx, accountID, cursor, limit)
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
	c.JSON(consts.StatusOK, LikedVideoListResponse{
		Videos:     newVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}