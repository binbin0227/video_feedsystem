package handler

import (
	"net/http"
	"strconv"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/gin-gonic/gin"
)

type LikeRequest struct {
	VideoID string `json:"video_id"`
}

type LikedVideoListResponse struct {
	Videos     []VideoResponse `json:"videos"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

// 点赞视频
func LikeVideo(c *gin.Context) {
	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	videoID, err := parsePositiveInt64String(req.VideoID, "video_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	if err := service.LikeVideo(c.Request.Context(), accountID, videoID); err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "点赞成功"})
}

// 取消视频点赞
func UnlikeVideo(c *gin.Context) {
	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	videoID, err := parsePositiveInt64String(req.VideoID, "video_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	if err := service.UnlikeVideo(c.Request.Context(), accountID, videoID); err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "取消点赞成功",
	})
}

// 返回当前用户对视频的点赞状态
func GetLikeStatus(c *gin.Context) {
	videoID, err := parsePositiveInt64Query(c, "video_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	liked, err := service.CheckLikeStatus(c.Request.Context(), accountID, videoID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]bool{"is_liked": liked})
}

// 分页查询当前用户点赞过的视频
func GetLikedVideoList(c *gin.Context) {
	accountID, err := getAccountID(c)
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

	result, err := service.GetLikedVideoList(c.Request.Context(), accountID, cursor, limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	c.JSON(http.StatusOK, LikedVideoListResponse{
		Videos:     newVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}
