package service

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"gorm.io/gorm"
)

const (
	maxVideoTitleLength       = 128
	maxVideoDescriptionLength = 2000
	maxVideoURLLength         = 255
	defaultAuthorVideoLimit   = 20
	maxAuthorVideoLimit       = 100
)

type AuthorVideoListResult struct {
	Videos     []model.Video
	NextCursor int64
	HasMore    bool
}

func validateUploadPath(uploadPath, category string, authorID int64) error {
	cleanedPath := path.Clean(uploadPath)
	expectedPrefix := fmt.Sprintf("/uploads/%s/%d/", category, authorID)

	if cleanedPath != uploadPath {
		return apperr.New(apperr.KindInvalid, "上传文件路径不合法")
	}

	if !strings.HasPrefix(cleanedPath, expectedPrefix) {
		return apperr.New(apperr.KindInvalid, "上传文件不属于当前用户")
	}

	return nil
}

// PublishVideo 校验视频信息并写入数据库。
func PublishVideo(ctx context.Context, authorID int64, title, description, playURL, coverURL string) (*model.Video, error) {
	// 1. 参数校验
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
	if utf8.RuneCountInString(description) > maxVideoDescriptionLength {
		return nil, apperr.New(apperr.KindInvalid, "视频描述不能超过2000个字符")
	}
	if len(playURL) > maxVideoURLLength || len(coverURL) > maxVideoURLLength {
		return nil, apperr.New(apperr.KindInvalid, "视频或封面路径过长")
	}

	// 2. 文件路径路径校验
	if err := validateUploadPath(playURL, "videos", authorID); err != nil {
		return nil, err
	}
	if err := validateUploadPath(coverURL, "covers", authorID); err != nil {
		return nil, err
	}
	if strings.ToLower(path.Ext(playURL)) != ".mp4" {
		return nil, apperr.New(apperr.KindInvalid, "视频路径格式不合法")
	}
	coverExt := strings.ToLower(path.Ext(coverURL))
	if coverExt != ".jpg" && coverExt != ".jpeg" && coverExt != ".png" {
		return nil, apperr.New(apperr.KindInvalid, "封面路径格式不合法")
	}

	// 3. 相同上传文件只允许发布一次；客户端因网络或页面跳转失败重试时直接返回原记录。
	existingVideo, err := db.FindVideoByAuthorAndMedia(ctx, authorID, playURL, coverURL)
	if err == nil {
		return existingVideo, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Wrap(apperr.KindInternal, "检查视频发布状态失败，请稍后再试", err)
	}

	// 4. 生成视频 ID
	videoID, err := utils.GenerateID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成视频ID失败", err)
	}

	// 5. 打包并存入数据库
	video := &model.Video{
		ID:          videoID,
		AuthorID:    authorID,
		Title:       title,
		Description: description,
		PlayURL:     playURL,
		CoverURL:    coverURL,
	}
	if err := db.CreateVideo(ctx, video); err != nil {
		// 唯一索引解决并发请求竞态；另一个请求已经写入时仍返回同一条视频。
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			existingVideo, findErr := db.FindVideoByAuthorAndMedia(ctx, authorID, playURL, coverURL)
			if findErr == nil {
				return existingVideo, nil
			}
		}
		return nil, apperr.Wrap(apperr.KindInternal, "视频发布失败，请稍后再试", err)
	}
	return video, nil
}

// ListByAuthorID 查询指定作者的视频列表；没有数据时返回空数组。
func ListByAuthorID(ctx context.Context, authorID, cursor int64, limit int) (AuthorVideoListResult, error) {
	// 1. 校验参数
	if authorID <= 0 {
		return AuthorVideoListResult{}, apperr.New(apperr.KindInvalid, "账号ID不合法")
	}
	if cursor < 0 {
		return AuthorVideoListResult{}, apperr.New(apperr.KindInvalid, "cursor 不合法")
	}
	if limit < 0 {
		return AuthorVideoListResult{}, apperr.New(apperr.KindInvalid, "limit 不合法")
	}
	if limit == 0 {
		limit = defaultAuthorVideoLimit
	} else if limit > maxAuthorVideoLimit {
		limit = maxAuthorVideoLimit
	}

	// 2. 确认作者存在
	_, err := db.FindAccountByID(ctx, authorID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthorVideoListResult{}, apperr.New(apperr.KindNotFound, "用户不存在")
	}
	if err != nil {
		return AuthorVideoListResult{}, apperr.Wrap(apperr.KindInternal, "查询用户失败，请稍后再试", err)
	}

	// 3. 多查询一条，判断是否还有下一页
	videos, err := db.ListByAuthorID(ctx, authorID, cursor, limit+1)
	if err != nil {
		return AuthorVideoListResult{}, apperr.Wrap(apperr.KindInternal, "查询用户发布的视频失败，请稍后再试", err)
	}

	hasMore := len(videos) > limit
	if hasMore {
		videos = videos[:limit]
	}

	// 4. 使用最后一个视频的 ID 作为下一页游标
	var nextCursor int64
	if hasMore && len(videos) > 0 {
		nextCursor = videos[len(videos)-1].ID
	}

	return AuthorVideoListResult{
		Videos:     videos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetVideoDetail 查询单个视频，不存在时返回 404 类业务错误。
func GetVideoDetail(ctx context.Context, videoID int64) (*model.Video, error) {
	// 1. 校验参数
	if videoID <= 0 {
		return nil, apperr.New(apperr.KindInvalid, "视频ID不合法")
	}

	// 2. 查询视频并预加载作者，供详情响应返回作者用户名。
	video, err := db.FindVideoWithAuthorByID(ctx, videoID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.New(apperr.KindNotFound, "视频不存在")
	}
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "查询视频详情失败，请稍后再试", err)
	}
	return video, nil
}
