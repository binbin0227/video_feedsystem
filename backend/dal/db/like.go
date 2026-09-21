package db

import (
	"context"
	"errors"
	"time"
	"video_feedsystem/model"

	"gorm.io/gorm"
)

var ErrLikeNotFound = errors.New("like record not found")

type LikedVideoRow struct {
	RelationID     int64     `gorm:"column:relation_id"`
	VideoID        int64     `gorm:"column:video_id"`
	AuthorID       int64     `gorm:"column:author_id"`
	AuthorUsername string    `gorm:"column:author_username"`
	Title          string    `gorm:"column:title"`
	Description    string    `gorm:"column:description"`
	PlayURL        string    `gorm:"column:play_url"`
	CoverURL       string    `gorm:"column:cover_url"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	LikeCount      int       `gorm:"column:like_count"`
}

func CreateLike(ctx context.Context, like *model.Like, outboxEvent *model.OutboxEvent) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(like).Error; err != nil {
			return err
		}

		result := tx.Model(&model.Video{}).Where("id = ?", like.VideoID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Create(outboxEvent).Error; err != nil {
			return err
		}

		return nil
	})
}

func DeleteLike(ctx context.Context, accountID, videoID int64, outboxEvent *model.OutboxEvent) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("account_id = ? AND video_id = ?", accountID, videoID).
			Delete(&model.Like{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrLikeNotFound
		}

		result = tx.Model(&model.Video{}).Where("id = ?", videoID).
			UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Create(outboxEvent).Error; err != nil {
			return err
		}

		return nil
	})
}

// 判断指定账号是否已经点赞该视频
func CheckLikeExist(ctx context.Context, accountID, videoID int64) (bool, error) {
	var count int64
	err := DB.WithContext(ctx).Model(&model.Like{}).
		Where("account_id = ? and video_id = ?", accountID, videoID).
		Count(&count).Error
	return count > 0, err
}

// 按点赞关系 ID 倒序查询当前用户点赞过的视频
func ListLikedVideos(ctx context.Context, accountID, cursor int64, limit int) ([]LikedVideoRow, error) {
	var rows []LikedVideoRow

	query := DB.WithContext(ctx).
		Table("likes AS l").
		Select(`
			l.id AS relation_id,
			v.id AS video_id,
			v.author_id,
			a.username AS author_username,
			v.title,
			v.description,
			v.play_url,
			v.cover_url,
			v.created_at,
			v.like_count
		`).
		Joins("JOIN videos AS v ON v.id = l.video_id").
		Joins("JOIN accounts AS a ON a.id = v.author_id").
		Where("l.account_id = ?", accountID).
		Order("l.id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("l.id < ?", cursor)
	}

	err := query.Scan(&rows).Error
	return rows, err
}
