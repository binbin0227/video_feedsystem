package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"gorm.io/gorm"
)

const (
	maxVideoTitleLength = 128
	maxVideoURLLength   = 255
)

// PublishVideo 校验视频信息并写入数据库。
func PublishVideo(ctx context.Context, authorID int64, title, description, playURL, coverURL string) (*model.Video, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	playURL = strings.TrimSpace(playURL)
	coverURL = strings.TrimSpace(coverURL)

	if authorID <= 0 {
		return nil, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if title == "" || playURL == "" || coverURL == "" {
		return nil, apperr.New(apperr.KindInvalid, "标题、视频路径、封面路径不能为空")
	}
	if utf8.RuneCountInString(title) > maxVideoTitleLength {
		return nil, apperr.New(apperr.KindInvalid, "视频标题不能超过128个字符")
	}
	if len(playURL) > maxVideoURLLength || len(coverURL) > maxVideoURLLength {
		return nil, apperr.New(apperr.KindInvalid, "视频或封面路径过长")
	}

	videoID, err := utils.GenerateID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成视频ID失败", err)
	}

	video := &model.Video{
		ID:          videoID,
		AuthorID:    authorID,
		Title:       title,
		Description: description,
		PlayURL:     playURL,
		CoverURL:    coverURL,
	}

	if err := db.CreateVideo(ctx, video); err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "视频发布失败，请稍后再试", err)
	}
	return video, nil
}

// ListByAuthorID 查询指定作者的视频列表；没有数据时返回空数组。
func ListByAuthorID(ctx context.Context, authorID int64) ([]model.Video, error) {
	if authorID <= 0 {
		return nil, apperr.New(apperr.KindInvalid, "账号ID不合法")
	}

	videos, err := db.ListByAuthorID(ctx, authorID)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "查询用户发布的视频失败，请稍后再试", err)
	}
	return videos, nil
}

// GetVideoDetail 查询单个视频，不存在时返回 404 类业务错误。
func GetVideoDetail(ctx context.Context, videoID int64) (*model.Video, error) {
	// 1. 合法性校验
	if videoID <= 0 {
		return nil, apperr.New(apperr.KindInvalid, "视频ID不合法")
	}

	// 2. db.FindVideoByID
	video, err := db.FindVideoByID(ctx, videoID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.New(apperr.KindNotFound, "视频不存在")
	}
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "查询视频详情失败，请稍后再试", err)
	}
	return video, nil
}
