package db

import (
	"context"
	"video_feedsystem/model"
)

func ListFeed(ctx context.Context, cursor int64, limit int) ([]model.Video, error) {
	var videos []model.Video
	query := DB.WithContext(ctx).Model(&model.Video{}).Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	err := query.Find(&videos).Error
	return videos, err
}
