package db

import (
	"context"
	"video_feedsystem/model"

	"gorm.io/gorm"
)

func CreateComment(ctx context.Context, comment *model.Comment, outboxEvent *model.OutboxEvent) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		if err := tx.Create(outboxEvent).Error; err != nil {
			return err
		}
		return nil
	})
}

// 按评论 ID 倒序分页，并预加载评论作者
func ListCommentsByVideoID(ctx context.Context, videoID, cursor int64, limit int) ([]model.Comment, error) {
	var comments []model.Comment
	query := DB.WithContext(ctx).
		Model(&model.Comment{}).
		Preload("Account").
		Where("video_id = ?", videoID).
		Order("id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	err := query.Find(&comments).Error
	return comments, err
}

// 根据主键查询评论
func FindCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	var comment model.Comment
	err := DB.WithContext(ctx).First(&comment, commentID).Error
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func DeleteCommentByID(ctx context.Context, commentID int64, outboxEvent *model.OutboxEvent) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&model.Comment{}, commentID)
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
