package service

import (
	"context"
	"fmt"

	"video_feedsystem/dal/db"
	"video_feedsystem/dal/redis"
)

const (
	likeHotScore    = 3
	commentHotScore = 2
)

// RebuildHotVideos 根据 MySQL 中的点赞和评论数据重建 Redis 热门榜。
func RebuildHotVideos(ctx context.Context) error {
	rows, err := db.ListHotVideoStats(ctx)
	if err != nil {
		return fmt.Errorf("查询热门榜重建数据失败: %w", err)
	}
	scores := make([]redis.HotVideoScore, 0, len(rows))

	for _, row := range rows {
		score := row.LikeCount*likeHotScore + row.CommentCount*commentHotScore

		// 没有任何互动的视频暂时不放进热门榜
		if score <= 0 {
			continue
		}

		scores = append(scores, redis.HotVideoScore{
			VideoID: row.VideoID,
			Score:   float64(score),
		})
	}
	if err := redis.ReplaceHotVideoScores(ctx, scores); err != nil {
		return fmt.Errorf("写入 Redis 热门榜失败: %w", err)
	}

	return nil
}
