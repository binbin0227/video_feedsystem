package handler

import (
	"strconv"
	"time"

	"video_feedsystem/service"
)

// FollowingOrFollowerAccountResponse 表示关注或粉丝列表中的单个账号。
type FollowingOrFollowerAccountResponse struct {
	AccountID  string    `json:"account_id"`
	Username   string    `json:"username"`
	FollowedAt time.Time `json:"followed_at"`
}

// FollowingOrFollowerListResponse 表示关注关系的游标分页响应。
type FollowingOrFollowerListResponse struct {
	Accounts   []FollowingOrFollowerAccountResponse `json:"accounts"`
	NextCursor string                               `json:"next_cursor"`
	HasMore    bool                                 `json:"has_more"`
}

// newFollowingOrFollowerAccountListResponse 将关注关系业务结果转换为对外响应。
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
