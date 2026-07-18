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

func FollowUser(ctx context.Context, followerID, vloggerID int64) error {
	// 1. 校验参数
	if followerID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if vloggerID <= 0 {
		return apperr.New(apperr.KindInvalid, "目标用户ID不合法")
	}
	if followerID == vloggerID {
		return apperr.New(apperr.KindInvalid, "不能关注自己")
	}

	// 2. 确认目标用户存在
	_, err := db.FindAccountByID(ctx, vloggerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.New(apperr.KindNotFound, "目标用户不存在")
	}
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "查询目标用户失败，请稍后再试", err)
	}

	// 3. 生成 socialID
	socialID, err := utils.GenerateID()
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "生成关注记录ID失败", err)
	}

	// 4. 打包
	social := &model.Social{
		ID:         socialID,
		FollowerID: followerID,
		VloggerID:  vloggerID,
	}

	// 5. db.CreateFollow
	if err := db.CreateFollow(ctx, social); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperr.New(apperr.KindConflict, "请勿重复关注")
		}
		return apperr.Wrap(apperr.KindInternal, "关注失败，请稍后再试", err)
	}

	return nil
}

func UnfollowUser(ctx context.Context, followerID, vloggerID int64) error {
	// 1. 校验参数
	if followerID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if vloggerID <= 0 {
		return apperr.New(apperr.KindInvalid, "目标用户ID不合法")
	}
	if followerID == vloggerID {
		return apperr.New(apperr.KindInvalid, "不能取消关注自己")
	}

	// 2. db.DeleteFollow
	if err := db.DeleteFollow(ctx, followerID, vloggerID); err != nil {
		if errors.Is(err, db.ErrFollowNotFound) {
			return apperr.New(apperr.KindConflict, "尚未关注该用户")
		}
		return apperr.Wrap(apperr.KindInternal, "取消关注失败，请稍后再试", err)
	}

	return nil
}

func CheckFollowStatus(ctx context.Context, followerID, vloggerID int64) (bool, error) {
	// 1. 校验参数
	if followerID <= 0 {
		return false, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if vloggerID <= 0 {
		return false, apperr.New(apperr.KindInvalid, "目标用户ID不合法")
	}

	// 2. db.CheckFollowExist
	following, err := db.CheckFollowExist(ctx, followerID, vloggerID)
	if err != nil {
		return false, apperr.Wrap(apperr.KindInternal, "查询关注状态失败，请稍后再试", err)
	}

	return following, nil
}