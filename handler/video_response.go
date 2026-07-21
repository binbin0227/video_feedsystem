package handler

import (
	"strconv"
	"time"

	"video_feedsystem/model"
	"video_feedsystem/service"
)

// VideoResponse 是返回给前端的视频结构，ID 使用字符串避免前端精度丢失。
type VideoResponse struct {
	ID          string    `json:"id"`
	AuthorID    string    `json:"author_id"`
	AuthorUsername string `json:"author_username,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PlayURL     string    `json:"play_url"`
	CoverURL    string    `json:"cover_url"`
	CreatedAt   time.Time `json:"created_at"`
	LikeCount   int       `json:"like_count"`
	Popularity  int       `json:"popularity"`
}

func newVideoResponse(video *model.Video) VideoResponse {
	return VideoResponse{
		ID:          strconv.FormatInt(video.ID, 10),
		AuthorID:    strconv.FormatInt(video.AuthorID, 10),
		Title:       video.Title,
		Description: video.Description,
		PlayURL:     video.PlayURL,
		CoverURL:    video.CoverURL,
		CreatedAt:   video.CreatedAt,
		LikeCount:   video.LikeCount,
		Popularity:  video.Popularity,
	}
}

func newVideoListResponse(videos []model.Video) []VideoResponse {
	result := make([]VideoResponse, 0, len(videos))
	for i := range videos {
		result = append(result, newVideoResponse(&videos[i]))
	}
	return result
}

func newFeedVideoResponse(item *service.FeedVideo) VideoResponse {
	response := newVideoResponse(&item.Video)
	response.AuthorUsername = item.AuthorUsername
	return response
}

func newFeedVideoListResponse(videos []service.FeedVideo) []VideoResponse {
	result := make([]VideoResponse, 0, len(videos))

	for i := range videos {
		result = append(result, newFeedVideoResponse(&videos[i]))
	}

	return result
}