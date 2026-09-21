package db

import (
	"context"
	"time"

	"video_feedsystem/model"

	"gorm.io/gorm"
)

func ListPendingOutboxEvents(ctx context.Context, now time.Time, limit int) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent

	err := DB.WithContext(ctx).
		Where("published_at IS NULL").
		Where("next_retry_at <= ?", now).
		Order("id ASC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

func MarkOutboxEventPublished(ctx context.Context, eventID int64, publishedAt time.Time) error {
	result := DB.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ? AND published_at IS NULL", eventID).
		Update("published_at", publishedAt)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func MarkOutboxEventRetry(ctx context.Context, eventID int64, nextRetryAt time.Time) error {
	result := DB.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ? AND published_at IS NULL", eventID).
		Updates(map[string]any{
			"retry_count":   gorm.Expr("retry_count + 1"),
			"next_retry_at": nextRetryAt,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func DeletePendingOutboxEvent(ctx context.Context, eventID int64) error {
	return DB.WithContext(ctx).
		Where("id = ? AND published_at IS NULL", eventID).
		Delete(&model.OutboxEvent{}).Error
}
