package service

import (
	"context"
	"errors"
	"time"

	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"gorm.io/gorm"
)

const (
	defaultFollowingOrFollowerLimit = 20
	maxFollowingOrFollowerLimit     = 100
)

type FollowingOrFollowerAccount struct {
	AccountID  int64
	Username   string
	FollowedAt time.Time
}

type FollowingOrFollowerListResult struct {
	Accounts   []FollowingOrFollowerAccount
	NextCursor int64
	HasMore    bool
}

// 关注
func FollowUser(ctx context.Context, followerID, vloggerID int64) error {
	if followerID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if vloggerID <= 0 {
		return apperr.New(apperr.KindInvalid, "目标用户ID不合法")
	}
	if followerID == vloggerID {
		return apperr.New(apperr.KindInvalid, "不能关注自己")
	}

	_, err := db.FindAccountByID(ctx, vloggerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.New(apperr.KindNotFound, "目标用户不存在")
	}
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "查询目标用户失败，请稍后再试", err)
	}

	socialID, err := utils.GenerateID()
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "生成关注记录ID失败", err)
	}
	social := &model.Social{
		ID:         socialID,
		FollowerID: followerID,
		VloggerID:  vloggerID,
	}

	if err := db.CreateFollow(ctx, social); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperr.New(apperr.KindConflict, "请勿重复关注")
		}
		return apperr.Wrap(apperr.KindInternal, "关注失败，请稍后再试", err)
	}

	return nil
}

// 取关
func UnfollowUser(ctx context.Context, followerID, vloggerID int64) error {
	if followerID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if vloggerID <= 0 {
		return apperr.New(apperr.KindInvalid, "目标用户ID不合法")
	}
	if followerID == vloggerID {
		return apperr.New(apperr.KindInvalid, "不能取消关注自己")
	}

	if err := db.DeleteFollow(ctx, followerID, vloggerID); err != nil {
		if errors.Is(err, db.ErrFollowNotFound) {
			return apperr.New(apperr.KindConflict, "尚未关注该用户")
		}
		return apperr.Wrap(apperr.KindInternal, "取消关注失败，请稍后再试", err)
	}

	return nil
}

// CheckFollowStatus 查询当前用户是否已经关注目标账号。
func CheckFollowStatus(ctx context.Context, followerID, vloggerID int64) (bool, error) {
	if followerID <= 0 {
		return false, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if vloggerID <= 0 {
		return false, apperr.New(apperr.KindInvalid, "目标用户ID不合法")
	}

	following, err := db.CheckFollowExist(ctx, followerID, vloggerID)
	if err != nil {
		return false, apperr.Wrap(apperr.KindInternal, "查询关注状态失败，请稍后再试", err)
	}

	return following, nil
}

// GetFollowingList 分页查询当前用户关注的账号。
func GetFollowingList(ctx context.Context, followerID, cursor int64, limit int) (FollowingOrFollowerListResult, error) {
	if followerID <= 0 {
		return FollowingOrFollowerListResult{}, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if cursor < 0 {
		return FollowingOrFollowerListResult{}, apperr.New(apperr.KindInvalid, "cursor 不合法")
	}
	if limit < 0 {
		return FollowingOrFollowerListResult{}, apperr.New(apperr.KindInvalid, "limit 不合法")
	}
	if limit == 0 {
		limit = defaultFollowingOrFollowerLimit
	} else if limit > maxFollowingOrFollowerLimit {
		limit = maxFollowingOrFollowerLimit
	}

	// 多查询一条来判断是否还有下一页。
	rows, err := db.ListFollowingAccounts(ctx, followerID, cursor, limit+1)
	if err != nil {
		return FollowingOrFollowerListResult{}, apperr.Wrap(apperr.KindInternal, "查询关注列表失败，请稍后再试", err)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	accounts := make([]FollowingOrFollowerAccount, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, FollowingOrFollowerAccount{
			AccountID:  row.AccountID,
			Username:   row.Username,
			FollowedAt: row.FollowedAt,
		})
	}

	var nextCursor int64
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].RelationID
	}

	return FollowingOrFollowerListResult{
		Accounts:   accounts,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetFollowerList 分页查询当前用户的粉丝账号。
func GetFollowerList(ctx context.Context, vloggerID, cursor int64, limit int) (FollowingOrFollowerListResult, error) {
	if vloggerID <= 0 {
		return FollowingOrFollowerListResult{}, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if cursor < 0 {
		return FollowingOrFollowerListResult{}, apperr.New(apperr.KindInvalid, "cursor 不合法")
	}
	if limit < 0 {
		return FollowingOrFollowerListResult{}, apperr.New(apperr.KindInvalid, "limit 不合法")
	}
	if limit == 0 {
		limit = defaultFollowingOrFollowerLimit
	} else if limit > maxFollowingOrFollowerLimit {
		limit = maxFollowingOrFollowerLimit
	}

	// 多查询一条来判断是否还有下一页。
	rows, err := db.ListFollowerAccounts(ctx, vloggerID, cursor, limit+1)
	if err != nil {
		return FollowingOrFollowerListResult{}, apperr.Wrap(apperr.KindInternal, "查询粉丝列表失败，请稍后再试", err)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	accounts := make([]FollowingOrFollowerAccount, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, FollowingOrFollowerAccount{
			AccountID:  row.AccountID,
			Username:   row.Username,
			FollowedAt: row.FollowedAt,
		})
	}
	
	var nextCursor int64
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].RelationID
	}

	return FollowingOrFollowerListResult{
		Accounts:   accounts,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}
