package redis

import (
	"context"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"strconv"
)

const hotVideoKey = "video:hot"

// HotVideoScore 表示一个视频及其重建排行榜时使用的热度值。
type HotVideoScore struct {
	VideoID int64
	Score   float64
}

// SetVideoHotScore 覆盖指定视频的最终热度分数，分数不大于零时移出排行榜。
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

// ListHotVideoIDs 按热度从高到低返回得分大于零的视频 ID。
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

// ReplaceHotVideoScores 使用给定数据整体替换热门视频榜。
func ReplaceHotVideoScores(ctx context.Context, scores []HotVideoScore) error {
	members := make([]goredis.Z, 0, len(scores))
	for _, item := range scores {
		members = append(members, goredis.Z{
			Score:  item.Score,
			Member: strconv.FormatInt(item.VideoID, 10),
		})
	}

	pipe := rdb.TxPipeline()
	// 先删除旧排行榜，避免旧数据残留。
	pipe.Del(ctx, hotVideoKey)
	// 没有视频时，只删除旧排行榜，不执行空的 ZAdd。
	if len(members) > 0 {
		pipe.ZAdd(ctx, hotVideoKey, members...)
	}

	// 删除旧榜和写入新榜会作为一组命令执行，避免其他请求刚好在两条命令中间看到一个长期为空的排行榜
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("重建 Redis 热门榜失败: %w", err)
	}

	return nil
}
