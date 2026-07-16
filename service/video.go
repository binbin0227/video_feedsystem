package service

import (
	"context"
	"errors"
	"strings"
	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/utils"
)

func PublishVideo(ctx context.Context, authorID int64, title, description, playURL, coverURL string) (*model.Video, error) {
	// 1. 去除两边空格
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	playURL = strings.TrimSpace(playURL)
	coverURL = strings.TrimSpace(coverURL)

	// 2. 基础参数校验
	if title == "" || playURL == "" || coverURL == "" {
		return nil, errors.New("视频标题、播放地址和封面地址不能为空")
	}

	// 3. 生成视频 ID
	videoID, err := utils.GenerateID()
	if err != nil {
		return nil, errors.New("生成视频 ID 失败")
	}

	// 4. 打包数据
	video := &model.Video{
		ID:          videoID,
		AuthorID:    authorID,
		Title:       title,
		Description: description,
		PlayURL:     playURL,
		CoverURL:    coverURL,
	}

	// 5. 存入数据库
	if err := db.CreateVideo(ctx, video); err != nil {
		return nil, err
	}
	return video, nil

}
