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

// FeedVideo 在视频模型之外携带作者用户名。
type FeedVideo struct {
	Video          model.Video
	AuthorUsername string
}

// FeedResult 表示视频流的游标分页结果。
type FeedResult struct {
	Videos     []FeedVideo
	NextCursor int64
	HasMore    bool
}

// newFeedVideos 将 DAL 查询行转换为 Feed 业务对象。
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

// GetFeed 分页查询公共视频流。
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

// GetHotFeed 查询热门视频
func GetHotFeed(ctx context.Context) ([]FeedVideo,error) {
	// 1. 从 Redis 查询热度最高的 10 个视频 ID
	videoIDs,err:=redis.ListHotVideoIDs(ctx,int64(hotFeedLimit))
	if err!=nil{
		return nil,apperr.Wrap(apperr.KindInternal,"查询热门视频失败，请稍后再试",err)
	}

	// Redis 热门榜没有数据
	if len(videoIDs) == 0 {
		return []FeedVideo{}, nil
	}

	// 2. db.ListVideosByIDs 查询完整视频信息
	rows,err:=db.ListVideosByIDs(ctx,videoIDs)
	if err!=nil{
		return nil,apperr.Wrap(apperr.KindInternal,"查询热门视频失败，请稍后再试",err)
	}

	// 3. 将 MySQL 查询结果按照视频 ID 放入 map
	rowMap:=make(map[int64]db.FeedVideoRow,len(rows))
	for _,row := range rows{
		rowMap[row.ID]=row
	}

	// 4. 按照 Redis 返回的热度顺序重新排列
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