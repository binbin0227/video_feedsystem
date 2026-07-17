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

func LikeVideo(ctx context.Context, accountID, videoID int64) error {
	// 1. 合法性校验
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
