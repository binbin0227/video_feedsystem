package service

import (
	"context"
	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
)

const (
	defaultFeedLimit = 10
	maxFeedLimit     = 50
)

type FeedResult struct {
	Videos     []model.Video
	NextCursor int64
	HasMore    bool
}

func GetFeed(ctx context.Context, cursor int64, limit int) (FeedResult, error) {
	// 1. 合法性校验
	if cursor < 0 {
		return FeedResult{}, apperr.New(apperr.KindInvalid, "cursor 不合法")
	}
	if limit < 0 {
		return FeedResult{}, apperr.New(apperr.KindInvalid, "limit 不合法")
	} else if limit == 0 {
		limit = defaultFeedLimit
	} else if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	// 2. db.ListFeed 多查一个来判断还有没有剩余视频
	videos, err := db.ListFeed(ctx, cursor, limit+1)
	if err != nil {
		return FeedResult{}, apperr.Wrap(apperr.KindInternal, "查询视频流失败，请稍后再试", err)
	}
	hasMore := len(videos) > limit
	if hasMore {
		videos = videos[:limit]
	}

	// 3. 还有剩余视频且刚当前返回列表不为空才返回 cursor
	var nextCursor int64
	if hasMore && len(videos) > 0 {
		nextCursor = videos[len(videos)-1].ID
	}

	// 4. 返回结果
	return FeedResult{
		Videos:     videos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
