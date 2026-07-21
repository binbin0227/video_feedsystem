package db

import (
	"context"

	"video_feedsystem/model"
)

// CreateVideo 将新视频写入数据库。
func CreateVideo(ctx context.Context, video *model.Video) error {
	return DB.WithContext(ctx).Create(video).Error
}

// FindVideoByAuthorAndMedia 根据作者、视频路径和封面路径查询已经发布的记录，用于防止重复发布。
func FindVideoByAuthorAndMedia(ctx context.Context, authorID int64, playURL, coverURL string) (*model.Video, error) {
	var video model.Video
	err := DB.WithContext(ctx).
		Where("author_id = ? AND play_url = ? AND cover_url = ?", authorID, playURL, coverURL).
		First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// ListByAuthorID 按发布时间倒序查询作者的所有视频。
func ListByAuthorID(ctx context.Context, authorID, cursor int64, limit int) ([]model.Video, error) {
	var videos []model.Video

	query := DB.WithContext(ctx).
		Model(&model.Video{}).
		Where("author_id = ?", authorID).
		Order("id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	err := query.Find(&videos).Error
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

// FindVideoWithAuthorByID 查询视频详情并预加载作者，只供需要展示作者信息的场景使用。
func FindVideoWithAuthorByID(ctx context.Context, videoID int64) (*model.Video, error) {
	var video model.Video
	err := DB.WithContext(ctx).Preload("Author").First(&video, videoID).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}
