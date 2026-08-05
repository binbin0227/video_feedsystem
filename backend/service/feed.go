package service

import (
	"context"
	"video_feedsystem/dal/db"
	"video_feedsystem/dal/redis"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
)

const (
	defaultFeedLimit = 10
	maxFeedLimit     = 50
	hotFeedLimit     = 10
)

// 在视频结构体外带上作者用户名
type FeedVideo struct {
	Video          model.Video
	AuthorUsername string
}

type FeedResult struct {
	Videos     []FeedVideo
	NextCursor int64
	HasMore    bool
}

// 将数据库查询行转换为 Feed 业务对象
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

// 分页查询公共视频流
func GetFeed(ctx context.Context, cursor int64, limit int) (FeedResult, error) {
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

	// 多查询一条来判断是否还有下一页
	rows, err := db.ListFeed(ctx, cursor, limit+1)
	if err != nil {
		return FeedResult{}, apperr.Wrap(apperr.KindInternal, "查询视频流失败，请稍后再试", err)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

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

// 分页查询当前用户关注的人发布的视频
func GetFollowingFeed(ctx context.Context, followerID, cursor int64, limit int) (FeedResult, error) {
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

	// 多查询一条来判断是否还有下一页
	rows, err := db.ListFollowingFeed(ctx, followerID, cursor, limit+1)
	if err != nil {
		return FeedResult{}, apperr.Wrap(apperr.KindInternal, "查询关注流失败，请稍后再试", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
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

// 查询热门视频
func GetHotFeed(ctx context.Context) ([]FeedVideo, error) {
	videoIDs, err := redis.ListHotVideoIDs(ctx, int64(hotFeedLimit))
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "查询热门视频失败，请稍后再试", err)
	}

	if len(videoIDs) == 0 {
		return []FeedVideo{}, nil
	}
	rows, err := db.ListVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "查询热门视频失败，请稍后再试", err)
	}

	// 按排行榜顺序重新排序
	rowMap := make(map[int64]db.FeedVideoRow, len(rows))
	for _, row := range rows {
		rowMap[row.ID] = row
	}
	orderedRows := make([]db.FeedVideoRow, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		row, exists := rowMap[videoID]
		if !exists {
			continue
		}
		orderedRows = append(orderedRows, row)
	}

	return newFeedVideos(orderedRows), nil
}
