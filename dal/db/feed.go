package db

import (
	"context"
	"video_feedsystem/model"
)

func ListFeed(ctx context.Context, cursor int64, limit int) ([]model.Video, error) {
	var videos []model.Video
	query := DB.WithContext(ctx).
		Model(&model.Video{}).
		Order("id DESC").
		Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	err := query.Find(&videos).Error
	return videos, err
}

func ListFollowingFeed(ctx context.Context, followerID, cursor int64, limit int) ([]model.Video, error) {
	var videos []model.Video

	query := DB.WithContext(ctx).
		Table("videos AS v").
		Where(`
			EXISTS (
				SELECT 1
				FROM socials AS s
				WHERE s.follower_id = ? AND s.vlogger_id = v.author_id
				)`, followerID).
		Order("v.id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("v.id < ?", cursor)
	}

	err := query.Find(&videos).Error
	return videos, err
}
