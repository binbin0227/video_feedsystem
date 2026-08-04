package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
	"video_feedsystem/dal/db"
	"video_feedsystem/dal/redis"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

const (
	maxCommentContentLength = 500
	defaultCommentLimit     = 20
	maxCommentLimit         = 100
)

// CommentListResult 表示评论列表的游标分页结果。
type CommentListResult struct {
	Comments   []model.Comment
	NextCursor int64
	HasMore    bool
}

// CreateComment 校验视频和用户后创建评论，并补充响应所需的作者信息。
func CreateComment(ctx context.Context, accountID, videoID int64, content string) (*model.Comment, error) {
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
	_, err := db.FindVideoByID(ctx, videoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.KindNotFound, "视频不存在")
		}
		return nil, apperr.Wrap(apperr.KindInternal, "查询视频失败，请稍后再试", err)
	}
	account, err := db.FindAccountByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.KindUnauthorized, "用户不存在")
		}
		return nil, apperr.Wrap(apperr.KindInternal, "查询用户失败，请稍后再试", err)
	}
	commentID, err := utils.GenerateID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成评论ID失败", err)
	}
	comment := &model.Comment{
		ID:        commentID,
		VideoID:   videoID,
		AccountID: accountID,
		Content:   content,
	}
	if err := db.CreateComment(ctx, comment); err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "发布评论失败，请稍后再试", err)
	}
	// Redis 更新失败不影响已经提交到 MySQL 的评论。
	if err := redis.ChangeVideoHotScore(ctx, videoID, commentHotScore); err != nil {
		hlog.CtxWarnf(ctx, "更新 Redis 视频热度失败，video_id=%d，error=%v", videoID, err)
	}
	// 只补充响应中的作者信息，不让 GORM 再次保存账号关联。
	comment.Account = *account
	return comment, nil
}

// GetCommentList 分页查询指定视频的评论。
func GetCommentList(ctx context.Context, videoID, cursor int64, limit int) (CommentListResult, error) {
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
	_, err := db.FindVideoByID(ctx, videoID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CommentListResult{}, apperr.New(apperr.KindNotFound, "视频不存在")
	}
	if err != nil {
		return CommentListResult{}, apperr.Wrap(apperr.KindInternal, "查询视频失败，请稍后再试", err)
	}
	// 多查询一条来判断是否还有下一页。
	comments, err := db.ListCommentsByVideoID(ctx, videoID, cursor, limit+1)
	if err != nil {
		return CommentListResult{}, apperr.Wrap(apperr.KindInternal, "查询评论失败，请稍后再试", err)
	}
	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}
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

// DeleteComment 校验评论归属后删除当前用户自己的评论。
func DeleteComment(ctx context.Context, accountID, commentID int64) error {
	if accountID <= 0 {
		return apperr.New(apperr.KindUnauthorized, "用户身份无效")
	}
	if commentID <= 0 {
		return apperr.New(apperr.KindInvalid, "评论 ID 不合法")
	}
	comment, err := db.FindCommentByID(ctx, commentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.New(apperr.KindNotFound, "评论不存在")
	}
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "查询评论失败，请稍后再试", err)
	}
	if comment.AccountID != accountID {
		return apperr.New(apperr.KindForbidden, "无权删除该评论")
	}
	if err := db.DeleteCommentByID(ctx, commentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.New(apperr.KindNotFound, "评论不存在")
		}
		return apperr.Wrap(apperr.KindInternal, "删除评论失败，请稍后再试", err)
	}
	// Redis 更新失败不回滚已经删除的 MySQL 评论。
	if err := redis.ChangeVideoHotScore(ctx, comment.VideoID, -commentHotScore); err != nil {
		hlog.CtxWarnf(ctx, "更新 Redis 视频热度失败，video_id=%d，error=%v", comment.VideoID, err)
	}

	return nil
}
