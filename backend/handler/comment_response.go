package handler

import (
	"strconv"
	"time"

	"video_feedsystem/model"
)

// CommentResponse 表示返回给前端的评论结构，ID 使用字符串避免精度丢失。
type CommentResponse struct {
	ID        string    `json:"id"`
	VideoID   string    `json:"video_id"`
	AccountID string    `json:"account_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// newCommentResponse 将评论模型转换为对外响应。
func newCommentResponse(comment *model.Comment) CommentResponse {
	return CommentResponse{
		ID:        strconv.FormatInt(comment.ID, 10),
		VideoID:   strconv.FormatInt(comment.VideoID, 10),
		AccountID: strconv.FormatInt(comment.AccountID, 10),
		Username:  comment.Account.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

// newCommentListResponse 将评论模型列表转换为对外响应列表。
func newCommentListResponse(comments []model.Comment) []CommentResponse {
	result := make([]CommentResponse, 0, len(comments))

	for i := range comments {
		result = append(result, newCommentResponse(&comments[i]))
	}

	return result
}
