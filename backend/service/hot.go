package service

import (
	"context"
	"fmt"

	"video_feedsystem/dal/db"
	"video_feedsystem/dal/redis"
	"video_feedsystem/mq"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const (
	likeHotScore    = 3
	commentHotScore = 2
)

// 根据 MySQL 中的点赞和评论数据重建 Redis 热门榜
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

// 根据 MySQL 当前数据重新计算并覆盖单个视频的热度
func RefreshHotVideo(ctx context.Context, videoID int64) error {
	if videoID <= 0 {
		return fmt.Errorf("视频 ID 不合法")
	}

	row, err := db.GetHotVideoStat(ctx, videoID)
	if err != nil {
		return fmt.Errorf("查询单个视频热度数据失败: %w", err)
	}

	score := row.LikeCount*likeHotScore + row.CommentCount*commentHotScore

	if err := redis.SetVideoHotScore(ctx, videoID, float64(score)); err != nil {
		return fmt.Errorf("刷新 Redis 视频热度失败: %w", err)
	}

	return nil
}

// 发布视频热度刷新消息（ RabbitMQ 不可用时同步刷新作为降级）
func notifyHotVideoRefresh(ctx context.Context, videoID int64) {
	if err := mq.PublishVideoHotRefresh(ctx, videoID); err == nil {
		return
	} else {
		hlog.CtxWarnf(ctx, "发布 RabbitMQ 视频热度刷新消息失败，改为同步刷新，video_id=%d，error=%v", videoID, err)
	}

	if err := RefreshHotVideo(ctx, videoID); err != nil {
		hlog.CtxWarnf(ctx, "同步刷新 Redis 视频热度失败，video_id=%d，error=%v", videoID, err)
	}
}
