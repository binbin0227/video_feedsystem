package db

import (
	"context"

	"video_feedsystem/model"
)

type FeedVideoRow struct {
	model.Video
	AuthorUsername string `gorm:"column:author_username"`
}

func ListFeed(ctx context.Context, cursor int64, limit int) ([]FeedVideoRow, error) {
	var rows []FeedVideoRow

	query := DB.WithContext(ctx).
		Table("videos AS v").
		Select("v.*, a.username AS author_username").
		Joins("JOIN accounts AS a ON a.id = v.author_id").
		Order("v.id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("v.id < ?", cursor)
	}

	err := query.Scan(&rows).Error
	return rows, err
}

func ListFollowingFeed(ctx context.Context, followerID, cursor int64, limit int) ([]FeedVideoRow, error) {
	var rows []FeedVideoRow

	query := DB.WithContext(ctx).
		Table("videos AS v").
		Select("v.*, a.username AS author_username").
		Joins("JOIN accounts AS a ON a.id = v.author_id").
		Where(`
			EXISTS (
				SELECT 1
				FROM socials AS s
				WHERE s.follower_id = ?
				  AND s.vlogger_id = v.author_id
			)
		`, followerID).
		Order("v.id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("v.id < ?", cursor)
	}

	err := query.Scan(&rows).Error
	return rows, err
}
