package handler

import (
	"strconv"
	"time"

	"video_feedsystem/model"
)

type CommentResponse struct {
	ID        string    `json:"id"`
	VideoID   string    `json:"video_id"`
	AccountID string    `json:"account_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

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

func newCommentListResponse(comments []model.Comment) []CommentResponse {
	result := make([]CommentResponse, 0, len(comments))

	for i := range comments {
		result = append(result, newCommentResponse(&comments[i]))
	}

	return result
}
