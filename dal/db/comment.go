package db

import (
	"context"
	"video_feedsystem/model"

	"gorm.io/gorm"
)

func CreateComment(ctx context.Context, comment *model.Comment) error {
	return DB.WithContext(ctx).Create(comment).Error
}

func ListCommentsByVideoID(ctx context.Context, videoID, cursor int64, limit int) ([]model.Comment, error) {
	var comments []model.Comment
	query := DB.WithContext(ctx).
		Model(&model.Comment{}).
		Where("video_id = ?", videoID).
		Order("id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	err := query.Find(&comments).Error
	return comments, err
}

func FindCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	var comment model.Comment
	err := DB.WithContext(ctx).First(&comment, commentID).Error
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func DeleteCommentByID(ctx context.Context, commentID int64) error {
	result := DB.WithContext(ctx).Delete(&model.Comment{}, commentID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}