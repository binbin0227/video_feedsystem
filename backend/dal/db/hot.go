package db

import "context"

type HotVideoStatRow struct {
	VideoID      int64 `gorm:"column:video_id"`
	LikeCount    int64 `gorm:"column:like_count"`
	CommentCount int64 `gorm:"column:comment_count"`
}

func ListHotVideoStats(ctx context.Context) ([]HotVideoStatRow, error) {
	var rows []HotVideoStatRow

	err := DB.WithContext(ctx).
		Table("videos AS v").
		Select("v.id AS video_id, v.like_count, COUNT(c.id) AS comment_count").
		Joins("LEFT JOIN comments AS c ON c.video_id = v.id").
		Group("v.id, v.like_count").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	return rows, nil
}
