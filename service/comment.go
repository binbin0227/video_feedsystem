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
	maxCommentContentLength = 500
	defaultCommentLimit     = 20
	maxCommentLimit         = 100
)

type CommentListResult struct {
	Comments   []model.Comment
	NextCursor int64
	HasMore    bool
}

func CreateComment(ctx context.Context, accountID, videoID int64, content string) (*model.Comment, error) {
	// 1. 校验合法性
	content = strings.TrimSpace(content)
	if accountID <= 0 {
		return nil, apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if videoID <= 0 {
		return nil, apperr.New(apperr.KindInvalid, "视频ID不合法")
	}
	if content == "" {
		return nil, apperr.New(apperr.KindInvalid, "评论内容不能为空")
	}
	if utf8.RuneCountInString(content) > maxCommentContentLength {
		return nil, apperr.New(apperr.KindInvalid, "评论内容不能超过500个字符")
	}

	// 2. 确认视频存在
	_, err := db.FindVideoByID(ctx, videoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.KindNotFound, "视频不存在")
		}
		return nil, apperr.Wrap(apperr.KindInternal, "查询视频失败，请稍后再试", err)
	}

	// 3. 查询评论作者，保证发布成功后的响应可以立即返回用户名
	account, err := db.FindAccountByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.KindUnauthorized, "用户不存在")
		}
		return nil, apperr.Wrap(apperr.KindInternal, "查询用户失败，请稍后再试", err)
	}

	// 4. 生成 commentID
	commentID, err := utils.GenerateID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成评论ID失败", err)
	}

	// 5. 打包
	comment := &model.Comment{
		ID:        commentID,
		VideoID:   videoID,
		AccountID: accountID,
		Content:   content,
	}

	// 6. db.CreateComment
	if err := db.CreateComment(ctx, comment); err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "发布评论失败，请稍后再试", err)
	}

	// 7. 只为响应补充作者信息，不让 GORM 在创建评论时重复保存账号关联
	comment.Account = *account

	// 8. 返回评论
	return comment, nil
}

func GetCommentList(ctx context.Context, videoID, cursor int64, limit int) (CommentListResult, error) {
	// 1. 校验参数
	if videoID <= 0 {
		return CommentListResult{}, apperr.New(apperr.KindInvalid, "视频 ID 不合法")
	}
	if cursor < 0 {
		return CommentListResult{}, apperr.New(apperr.KindInvalid, "cursor 不合法")
	}
	if limit < 0 {
		return CommentListResult{}, apperr.New(apperr.KindInvalid, "limit 不合法")
	}
	if limit == 0 {
		limit = defaultCommentLimit
	} else if limit > maxCommentLimit {
		limit = maxCommentLimit
	}

	// 2. 确认视频存在
	_, err := db.FindVideoByID(ctx, videoID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CommentListResult{}, apperr.New(apperr.KindNotFound, "视频不存在")
	}
	if err != nil {
		return CommentListResult{}, apperr.Wrap(apperr.KindInternal, "查询视频失败，请稍后再试", err)
	}

	// 3. 多查询一条，用来判断是否还有下一页
	comments, err := db.ListCommentsByVideoID(ctx, videoID, cursor, limit+1)
	if err != nil {
		return CommentListResult{}, apperr.Wrap(apperr.KindInternal, "查询评论失败，请稍后再试", err)
	}
	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	// 4. 最后一条评论的 ID 作为下一页游标
	var nextCursor int64
	if hasMore && len(comments) > 0 {
		nextCursor = comments[len(comments)-1].ID
	}

	return CommentListResult{
		Comments:   comments,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func DeleteComment(ctx context.Context, accountID, commentID int64) error {
	// 1. 校验参数
	if accountID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if commentID <= 0 {
		return apperr.New(apperr.KindInvalid, "评论 ID 不合法")
	}

	// 2. db.FindCommentByID
	comment, err := db.FindCommentByID(ctx, commentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.New(apperr.KindNotFound, "评论不存在")
	}
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "查询评论失败，请稍后再试", err)
	}

	// 3. 检查评论所有权
	if comment.AccountID != accountID {
		return apperr.New(apperr.KindForbidden, "无权删除该评论")
	}

	// 4. db.DeleteCommentByID
	if err := db.DeleteCommentByID(ctx, commentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.New(apperr.KindNotFound, "评论不存在")
		}
		return apperr.Wrap(apperr.KindInternal, "删除评论失败，请稍后再试", err)
	}

	return nil
}
