package service

import (
	"context"
	"errors"
	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"gorm.io/gorm"
)

const (
	defaultLikedVideoLimit = 20
	maxLikedVideoLimit     = 100
)

type LikedVideoListResult struct {
	Videos     []model.Video
	NextCursor int64
	HasMore    bool
}

func LikeVideo(ctx context.Context, accountID, videoID int64) error {
	// 1. 校验参数
	if accountID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}

	if videoID <= 0 {
		return apperr.New(apperr.KindInvalid, "视频ID不合法")
	}

	// 2. 生成点赞记录 ID
	likeID, err := utils.GenerateID()
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "生成点赞记录失败", err)
	}

	// 3. 打包 like
	like := &model.Like{
		ID:        likeID,
		VideoID:   videoID,
		AccountID: accountID,
	}

	// 4. db.CreateLike
	err = db.CreateLike(ctx, like)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperr.New(apperr.KindConflict, "请勿重复点赞")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.New(apperr.KindNotFound, "视频不存在")
		}
		return apperr.Wrap(apperr.KindInternal, "点赞失败，请稍后再试", err)
	}
	return nil
}

func UnlikeVideo(ctx context.Context, accountID, videoID int64) error {
	// 1. 校验合法性
	if accountID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if videoID <= 0 {
		return apperr.New(apperr.KindInvalid, "视频ID不合法")
	}

	// 2. db.DeleteLike
	err := db.DeleteLike(ctx, accountID, videoID)
	if err != nil {
		if errors.Is(err, db.ErrLikeNotFound) {
			return apperr.New(apperr.KindConflict, "尚未点赞")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.New(apperr.KindNotFound, "视频不存在")
		}
		return apperr.Wrap(apperr.KindInternal, "取消点赞失败，请稍后再试", err)
	}

	return nil
}

func CheckLikeStatus(ctx context.Context, accountID, videoID int64) (bool, error) {
	// 1. 校验合法性
	if accountID <= 0 {
		return false, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if videoID <= 0 {
		return false, apperr.New(apperr.KindInvalid, "视频ID不合法")
	}

	// 2. db.CheckLikeExist
	liked, err := db.CheckLikeExist(ctx, accountID, videoID)
	if err != nil {
		return false, apperr.Wrap(apperr.KindInternal, "查询点赞状态失败，请稍后再试", err)
	}

	return liked, nil
}


func GetLikedVideoList(ctx context.Context, accountID, cursor int64, limit int) (LikedVideoListResult, error) {
	// 1. 校验参数
	if accountID <= 0 {
		return LikedVideoListResult{}, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if cursor < 0 {
		return LikedVideoListResult{}, apperr.New(apperr.KindInvalid, "cursor 不合法")
	}
	if limit < 0 {
		return LikedVideoListResult{}, apperr.New(apperr.KindInvalid, "limit 不合法")
	}
	if limit == 0 {
		limit = defaultLikedVideoLimit
	} else if limit > maxLikedVideoLimit {
		limit = maxLikedVideoLimit
	}

	// 2. 多查询一条，判断是否还有下一页
	rows, err := db.ListLikedVideos(ctx, accountID, cursor, limit+1)
	if err != nil {
		return LikedVideoListResult{}, apperr.Wrap(apperr.KindInternal, "查询点赞视频列表失败，请稍后再试", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	// 3. 将 DAL 的 JOIN 查询结果转换为 Video
	videos := make([]model.Video, 0, len(rows))
	for _, row := range rows {
		videos = append(videos, model.Video{
			ID:          row.VideoID,
			AuthorID:    row.AuthorID,
			Title:       row.Title,
			Description: row.Description,
			PlayURL:     row.PlayURL,
			CoverURL:    row.CoverURL,
			CreatedAt:   row.CreatedAt,
			LikeCount:   row.LikeCount,
			Popularity:  row.Popularity,
		})
	}

	// 4. 使用最后一条点赞关系的 ID 作为游标
	var nextCursor int64
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].RelationID
	}

	return LikedVideoListResult{
		Videos:     videos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}