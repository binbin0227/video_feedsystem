package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	videoDetailBaseTTL   = 10 * time.Minute
	videoDetailJitterTTL = 2 * time.Minute
	videoNotFoundTTL     = 1 * time.Minute
)

// videoDetailTTL 生成带随机抖动的视频详情缓存过期时间。
func videoDetailTTL() time.Duration {
	return videoDetailBaseTTL + time.Duration(rand.Int63n(int64(videoDetailJitterTTL)))
}

// VideoDetailCache 保存视频详情接口需要的缓存字段。
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
	NotFound       bool      `json:"not_found,omitempty"`
}

// videoDetailKey 生成视频详情缓存的 Redis Key。
func videoDetailKey(videoID int64) string {
	return "video:detail:" + strconv.FormatInt(videoID, 10)
}

// GetVideoDetailCache 查询视频详情缓存，并区分缓存未命中和 Redis 错误。
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

// SetVideoDetailCache 写入视频详情缓存并设置过期时间。
func SetVideoDetailCache(ctx context.Context, cache *VideoDetailCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("序列化视频详情缓存失败: %w", err)
	}

	if err := rdb.Set(ctx, videoDetailKey(cache.ID), data, videoDetailTTL()).Err(); err != nil {
		return fmt.Errorf("写入视频详情缓存失败: %w", err)
	}

	return nil
}

// DeleteVideoDetailCache 删除指定视频的详情缓存。
func DeleteVideoDetailCache(ctx context.Context, videoID int64) error {
	if err := rdb.Del(ctx, videoDetailKey(videoID)).Err(); err != nil {
		return fmt.Errorf("删除视频详情缓存失败: %w", err)
	}

	return nil
}

// SetVideoNotFoundCache 短暂记录视频不存在，避免重复请求穿透到 MySQL。
func SetVideoNotFoundCache(ctx context.Context, videoID int64) error {
	cache := VideoDetailCache{
		ID:       videoID,
		NotFound: true,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("序列化视频空值缓存失败: %w", err)
	}

	if err := rdb.Set(ctx, videoDetailKey(videoID), data, videoNotFoundTTL).Err(); err != nil {
		return fmt.Errorf("写入视频空值缓存失败: %w", err)
	}

	return nil
}
