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
	RelationID int64     `gorm:"column:relation_id"`
	VideoID    int64     `gorm:"column:video_id"`
	AuthorID   int64     `gorm:"column:author_id"`
	Title      string    `gorm:"column:title"`
	Description string   `gorm:"column:description"`
	PlayURL    string    `gorm:"column:play_url"`
	CoverURL   string    `gorm:"column:cover_url"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	LikeCount  int       `gorm:"column:like_count"`
	Popularity int       `gorm:"column:popularity"`
}

func CreateLike(ctx context.Context, like *model.Like) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 向 likes 表插入点赞记录，联合唯一索引会阻止重复点赞
		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// 2. 直接在数据库中执行 like_count + 1，避免并发覆盖
		result := tx.Model(&model.Video{}).Where("id = ?", like.VideoID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")) // 只修改指定字段，而且不会因为点赞而更新视频的 updated_at
		if result.Error != nil {
			return result.Error
		}

		// 视频不存在时撤销刚才插入的点赞记录
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func DeleteLike(ctx context.Context, accountID, videoID int64) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 从 likes 表中删除点赞记录
		result := tx.Where("account_id = ? AND video_id = ?", accountID, videoID).Delete(&model.Like{})
		if result.Error != nil {
			return result.Error
		}

		// 若没有删除任何记录，说明用户原本没有点赞
		if result.RowsAffected == 0 {
			return ErrLikeNotFound
		}

		// 2. 点赞数减一，并避免异常情况下变成负数
		result = tx.Model(&model.Video{}).Where("id = ?", videoID).
			UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END"))
		if result.Error != nil {
			return result.Error
		}

		// 视频不存在时，撤销前面删除点赞记录的操作
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func CheckLikeExist(ctx context.Context, accountID, videoID int64) (bool, error) {
	var count int64
	err := DB.WithContext(ctx).Model(&model.Like{}).
		Where("account_id = ? and video_id = ?", accountID, videoID).
		Count(&count).Error
	return count > 0, err
}

func ListLikedVideos(ctx context.Context, accountID, cursor int64, limit int) ([]LikedVideoRow, error) {
	var rows []LikedVideoRow

	query := DB.WithContext(ctx).
		Table("likes AS l").
		Select(`
			l.id AS relation_id,
			v.id AS video_id,
			v.author_id,
			v.title,
			v.description,
			v.play_url,
			v.cover_url,
			v.created_at,
			v.like_count,
			v.popularity
		`).
		Joins("JOIN videos AS v ON v.id = l.video_id").
		Where("l.account_id = ?", accountID).
		Order("l.id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("l.id < ?", cursor)
	}

	err := query.Scan(&rows).Error
	return rows, err
}