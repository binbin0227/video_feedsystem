package handler

import (
	"strconv"
	"time"

	"video_feedsystem/service"
)

// AccountProfileResponse 表示用户主页接口的响应字段。
type AccountProfileResponse struct {
	AccountID         string    `json:"account_id"`
	Username          string    `json:"username"`
	CreatedAt         time.Time `json:"created_at"`
	VideoCount        int64     `json:"video_count"`
	ReceivedLikeCount int64     `json:"received_like_count"`
	FollowingCount    int64     `json:"following_count"`
	FollowerCount     int64     `json:"follower_count"`
}

// AccountSearchItemResponse 表示单个用户搜索结果。
type AccountSearchItemResponse struct {
	AccountID         string `json:"account_id"`
	Username          string `json:"username"`
	ReceivedLikeCount int64  `json:"received_like_count"`
	FollowerCount     int64  `json:"follower_count"`
}

// AccountSearchResponse 表示用户搜索接口的响应结构。
type AccountSearchResponse struct {
	Accounts []AccountSearchItemResponse `json:"accounts"`
}

// newAccountProfileResponse 将 Service 结果转换为对外响应，并把 int64 ID 转为字符串。
func newAccountProfileResponse(profile *service.AccountProfile) AccountProfileResponse {
	return AccountProfileResponse{
		AccountID:         strconv.FormatInt(profile.AccountID, 10),
		Username:          profile.Username,
		CreatedAt:         profile.CreatedAt,
		VideoCount:        profile.VideoCount,
		ReceivedLikeCount: profile.ReceivedLikeCount,
		FollowingCount:    profile.FollowingCount,
		FollowerCount:     profile.FollowerCount,
	}
}

// newAccountSearchListResponse 将用户搜索业务结果转换为对外响应。
func newAccountSearchListResponse(accounts []service.AccountSearchItem) []AccountSearchItemResponse {
	result := make([]AccountSearchItemResponse, 0, len(accounts))

	for _, account := range accounts {
		result = append(result, AccountSearchItemResponse{
			AccountID:         strconv.FormatInt(account.AccountID, 10),
			Username:          account.Username,
			ReceivedLikeCount: account.ReceivedLikeCount,
			FollowerCount:     account.FollowerCount,
		})
	}

	return result
}
