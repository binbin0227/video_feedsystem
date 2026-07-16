package db

import (
	"context"
	"video_feedsystem/model"
)

// 将新视频信息写入数据库
func CreateVideo(ctx context.Context, video *model.Video) error {
	return DB.WithContext(ctx).Create(video).Error
}

// 根据 AuthorID 查出 ta 发布的所有视频
func ListByAuthorID(ctx context.Context, authorID int64) ([]model.Video, error) {
	var videos []model.Video
	err := DB.WithContext(ctx).
		Model(&model.Video{}).
		Where("author_id = ?", authorID).
		Order("created_at desc").
		Find(&videos).Error
	return videos, err
}
