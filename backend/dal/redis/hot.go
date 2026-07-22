package redis

import (
	"context"
	"fmt"
	"strconv"
)

const hotVideoKey = "video:hot"

// ChangeVideoHotScore 修改视频在热门榜单中的热度分数。
func ChangeVideoHotScore(ctx context.Context, videoID int64, delta float64) error {
	member := strconv.FormatInt(videoID, 10)
	return rdb.ZIncrBy(ctx, hotVideoKey, delta, member).Err()
}

// ListHotVideoIDs 按热度从高到低返回指定数量的视频 ID。
func ListHotVideoIDs(ctx context.Context, limit int64) ([]int64, error) {
	members, err := rdb.ZRevRange(ctx, hotVideoKey, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("查询 Redis 热门榜失败: %w", err)
	}

	videoIDs := make([]int64, 0, len(members))

	for _, member := range members {
		videoID, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("解析热门视频 ID 失败: %w", err)
		}

		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, nil
}
