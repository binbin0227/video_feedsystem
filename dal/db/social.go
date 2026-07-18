package db

import (
	"context"
	"errors"

	"video_feedsystem/model"
)

var ErrFollowNotFound = errors.New("follow relation not found")

func CreateFollow(ctx context.Context, social *model.Social) error {
	return DB.WithContext(ctx).Create(social).Error
}

func DeleteFollow(ctx context.Context, followerID, vloggerID int64) error {
	result := DB.WithContext(ctx).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Delete(&model.Social{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFollowNotFound
	}

	return nil
}

func CheckFollowExist(ctx context.Context, followerID, vloggerID int64) (bool, error) {
	var count int64
	err := DB.WithContext(ctx).
		Model(&model.Social{}).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Count(&count).Error

	return count > 0, err
}