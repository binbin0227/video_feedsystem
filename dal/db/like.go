package db

import (
	"context"
	"video_feedsystem/model"

	"gorm.io/gorm"
)

func CreateLike(ctx context.Context, like *model.Like) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 向 likes 表插入点赞记录，联合唯一索引会阻止重复点赞。
		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// 2. 直接在数据库中执行 like_count + 1，避免并发覆盖。
		result := tx.Model(&model.Video{}).Where("id = ?", like.VideoID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")) // 只修改指定字段，而且不会因为点赞而更新视频的 updated_at
		if result.Error != nil {
			return result.Error
		}

		// 视频不存在时撤销刚才插入的点赞记录。
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
