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

type FeedVideo struct {
	Video          model.Video
	AuthorUsername string
}

type FeedResult struct {
	Videos     []FeedVideo
	NextCursor int64
	HasMore    bool
}

func newFeedVideos(rows []db.FeedVideoRow) []FeedVideo {
	videos := make([]FeedVideo, 0, len(rows))

	for _, row := range rows {
		videos = append(videos, FeedVideo{
			Video:          row.Video,
			AuthorUsername: row.AuthorUsername,
		})
	}

	return videos
}

func GetFeed(ctx context.Context, cursor int64, limit int) (FeedResult, error) {
	// 1. 校验参数
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
	rows, err := db.ListFeed(ctx, cursor, limit+1)
	if err != nil {
		return FeedResult{}, apperr.Wrap(apperr.KindInternal, "查询视频流失败，请稍后再试", err)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	// 3. 还有剩余视频且刚当前返回列表不为空才返回 cursor
	var nextCursor int64
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].ID
	}

	// 4. 返回结果
	return FeedResult{
		Videos:     newFeedVideos(rows),
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetFollowingFeed 分页查询当前用户关注的人发布的视频。
func GetFollowingFeed(ctx context.Context, followerID, cursor int64, limit int) (FeedResult, error) {
	// 1. 校验参数
	if followerID <= 0 {
		return FeedResult{}, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if cursor < 0 {
		return FeedResult{}, apperr.New(apperr.KindInvalid, "cursor 不合法")
	}
	if limit < 0 {
		return FeedResult{}, apperr.New(apperr.KindInvalid, "limit 不合法")
	}
	if limit == 0 {
		limit = defaultFeedLimit
	} else if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	// 2. 多查询一条，判断是否还有下一页
	rows, err := db.ListFollowingFeed(ctx, followerID, cursor, limit+1)
	if err != nil {
		return FeedResult{}, apperr.Wrap(apperr.KindInternal, "查询关注流失败，请稍后再试", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	// 3. 使用当前页最后一个视频的 ID 作为下一页游标
	var nextCursor int64
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].ID
	}

	return FeedResult{
		Videos:     newFeedVideos(rows),
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
