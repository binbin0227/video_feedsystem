package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const videoDetailTTL = 10 * time.Minute

type VideoDetailCache struct {
	ID             int64     `json:"id"`
	AuthorID       int64     `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	PlayURL        string    `json:"play_url"`
	CoverURL       string    `json:"cover_url"`
	CreatedAt      time.Time `json:"created_at"`
	LikeCount      int       `json:"like_count"`
}

func videoDetailKey(videoID int64) string {
	return "video:detail:" + strconv.FormatInt(videoID, 10) // 如 video:detail:12345
}

func GetVideoDetailCache(ctx context.Context, videoID int64) (*VideoDetailCache, bool, error) {
	data, err := rdb.Get(ctx, videoDetailKey(videoID)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("查询视频详情缓存失败: %w", err)
	}

	var cache VideoDetailCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false, fmt.Errorf("解析视频详情缓存失败: %w", err)
	}

	return &cache, true, nil
}

func SetVideoDetailCache(ctx context.Context, cache *VideoDetailCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("序列化视频详情缓存失败: %w", err)
	}

	if err := rdb.Set(ctx, videoDetailKey(cache.ID), data, videoDetailTTL).Err(); err != nil {
		return fmt.Errorf("写入视频详情缓存失败: %w", err)
	}

	return nil
}

func DeleteVideoDetailCache(ctx context.Context, videoID int64) error {
	if err := rdb.Del(ctx, videoDetailKey(videoID)).Err(); err != nil {
		return fmt.Errorf("删除视频详情缓存失败: %w", err)
	}

	return nil
}
