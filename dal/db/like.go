package db

import (
	"context"
	"errors"
	"video_feedsystem/model"

	"gorm.io/gorm"
)

var ErrLikeNotFound = errors.New("like record not found")

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
