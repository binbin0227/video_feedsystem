package handler

import (
	"strconv"
	"time"

	"video_feedsystem/service"
)

type FollowingOrFollowerAccountResponse struct {
	AccountID  string    `json:"account_id"`
	Username   string    `json:"username"`
	FollowedAt time.Time `json:"followed_at"`
}

type FollowingOrFollowerListResponse struct {
	Accounts   []FollowingOrFollowerAccountResponse `json:"accounts"`
	NextCursor string                               `json:"next_cursor"`
	HasMore    bool                                 `json:"has_more"`
}

func newFollowingOrFollowerAccountListResponse(accounts []service.FollowingOrFollowerAccount) []FollowingOrFollowerAccountResponse {
	result := make([]FollowingOrFollowerAccountResponse, 0, len(accounts))

	for _, account := range accounts {
		result = append(result, FollowingOrFollowerAccountResponse{
			AccountID:  strconv.FormatInt(account.AccountID, 10),
			Username:   account.Username,
			FollowedAt: account.FollowedAt,
		})
	}

	return result
}
