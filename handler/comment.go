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

type PublishCommentRequest struct {
	VideoID string `json:"video_id"`
	Content string `json:"content"`
}

type CommentListResponse struct {
	Comments   []CommentResponse `json:"comments"`
	NextCursor string            `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}

func PublishComment(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req PublishCommentRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	// 2. 解析 video_id
	videoID, err := parsePositiveInt64String(req.VideoID, "video_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. 获取 account ID
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 发布评论
	comment, err := service.CreateComment(ctx, accountID, videoID, req.Content)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 5. 返回创建后的评论
	c.JSON(consts.StatusOK, map[string]any{
		"message": "评论发布成功",
		"comment": newCommentResponse(comment),
	})
}

// ListComments 分页查询指定视频的评论，按最新评论优先返回。
func ListComments(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 video_id
	videoID, err := parsePositiveInt64Query(c, "video_id")
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

	// 3. service.GetCommentList
	result, err := service.GetCommentList(ctx, videoID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	// 5. 返回结果
	c.JSON(consts.StatusOK, CommentListResponse{
		Comments:   newCommentListResponse(result.Comments),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

func DeleteComment(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 comment_id
	commentID, err := parsePositiveInt64Query(c, "comment_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 获取 accountID
	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. service.DeleteComment
	if err := service.DeleteComment(ctx, accountID, commentID); err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 返回结果
	c.JSON(consts.StatusOK, map[string]string{
		"message": "评论删除成功",
	})
}