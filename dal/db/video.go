package db

import (
	"context"

	"video_feedsystem/model"
)

// CreateVideo 将新视频写入数据库。
func CreateVideo(ctx context.Context, video *model.Video) error {
	return DB.WithContext(ctx).Create(video).Error
}

// ListByAuthorID 按发布时间倒序查询作者的所有视频。
func ListByAuthorID(ctx context.Context, authorID int64) ([]model.Video, error) {
	var videos []model.Video
	err := DB.WithContext(ctx).
		Where("author_id = ?", authorID).
		Order("created_at DESC").
		Find(&videos).Error
	return videos, err
}

// FindVideoByID 根据主键查询一个视频。
func FindVideoByID(ctx context.Context, videoID int64) (*model.Video, error) {
	var video model.Video
	err := DB.WithContext(ctx).First(&video, videoID).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}
