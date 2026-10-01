package handler

import (
	"net/http"
	"strconv"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/gin-gonic/gin"
)

type PublishCommentRequest struct {
	VideoID string `json:"video_id"`
	Content string `json:"content"`
}

type CommentListResponse struct {
	Comments   []CommentResponse `json:"comments"`
	NextCursor string            `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}

// 发布评论
func PublishComment(c *gin.Context) {
	var req PublishCommentRequest
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

	comment, err := service.CreateComment(c.Request.Context(), accountID, videoID, req.Content)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"message": "评论发布成功",
		"comment": newCommentResponse(comment),
	})
}

// 分页查询指定视频的评论
func ListComments(c *gin.Context) {
	videoID, err := parsePositiveInt64Query(c, "video_id")
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

	result, err := service.GetCommentList(c.Request.Context(), videoID, cursor, limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	c.JSON(http.StatusOK, CommentListResponse{
		Comments:   newCommentListResponse(result.Comments),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 删除当前登录用户自己的评论
func DeleteComment(c *gin.Context) {
	commentID, err := parsePositiveInt64Query(c, "comment_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	if err := service.DeleteComment(c.Request.Context(), accountID, commentID); err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "评论删除成功",
	})
}
