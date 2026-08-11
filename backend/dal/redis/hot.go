package redis

import (
	"context"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"strconv"
)

const hotVideoKey = "video:hot"

type HotVideoScore struct {
	VideoID int64
	Score   float64
}

// 覆盖指定视频的最终热度分数，分数 < 0 时移出排行榜
func SetVideoHotScore(ctx context.Context, videoID int64, score float64) error {
	member := strconv.FormatInt(videoID, 10)

	if score <= 0 {
		if err := rdb.ZRem(ctx, hotVideoKey, member).Err(); err != nil {
			return fmt.Errorf("从 Redis 热门榜移除视频失败: %w", err)
		}
		return nil
	}

	if err := rdb.ZAdd(ctx, hotVideoKey, goredis.Z{
		Score:  score,
		Member: member,
	}).Err(); err != nil {
		return fmt.Errorf("覆盖 Redis 视频热度失败: %w", err)
	}

	return nil
}

// 按热度从高到低返回分数 > 0 的 videoID。
func ListHotVideoIDs(ctx context.Context, limit int64) ([]int64, error) {
	members, err := rdb.ZRevRangeByScore(ctx, hotVideoKey, &goredis.ZRangeBy{
		Max:    "+inf",
		Min:    "(0",
		Offset: 0,
		Count:  limit,
	}).Result()
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

// 替换热门视频榜
func ReplaceHotVideoScores(ctx context.Context, scores []HotVideoScore) error {
	members := make([]goredis.Z, 0, len(scores))
	for _, item := range scores {
		members = append(members, goredis.Z{
			Score:  item.Score,
			Member: strconv.FormatInt(item.VideoID, 10),
		})
	}

	pipe := rdb.TxPipeline()
	// 先删除旧排行榜
	pipe.Del(ctx, hotVideoKey)
	if len(members) > 0 {
		pipe.ZAdd(ctx, hotVideoKey, members...)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("重建 Redis 热门榜失败: %w", err)
	}

	return nil
}
