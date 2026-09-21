package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/mq"
)

const (
	outboxBatchSize      = 50
	outboxPollInterval   = time.Second
	outboxEventTimeout   = 5 * time.Second
	outboxRetryBaseDelay = 10 * time.Second
	outboxRetryMaxDelay  = 5 * time.Minute
	outboxEventMaxAge    = 10 * time.Minute
)

func StartOutboxRelay(ctx context.Context) {
	go func() {
		processPendingOutboxEvents(ctx)

		ticker := time.NewTicker(outboxPollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				processPendingOutboxEvents(ctx)
			}
		}
	}()
}

func processPendingOutboxEvents(ctx context.Context) {
	events, err := db.ListPendingOutboxEvents(ctx, time.Now(), outboxBatchSize)
	if err != nil {
		log.Printf("查询待发送 Outbox 事件失败: %v", err)
		return
	}

	for i := range events {
		processOutboxEvent(ctx, &events[i])
	}
}

func processOutboxEvent(ctx context.Context, event *model.OutboxEvent) {
	if !event.CreatedAt.IsZero() && time.Since(event.CreatedAt) >= outboxEventMaxAge {
		if err := db.DeletePendingOutboxEvent(ctx, event.ID); err != nil {
			log.Printf("删除过期 Outbox 事件失败，event_id=%d，error=%v", event.ID, err)
			return
		}
		log.Printf("Outbox 事件已超过最大有效时间，交由热门榜定时重建兜底，event_id=%d", event.ID)
		return
	}

	eventCtx, cancel := context.WithTimeout(ctx, outboxEventTimeout)
	err := publishOutboxEvent(eventCtx, event)
	cancel()

	if err == nil {
		if markErr := db.MarkOutboxEventPublished(ctx, event.ID, time.Now()); markErr != nil {
			log.Printf("标记 Outbox 事件已发送失败，event_id=%d，error=%v", event.ID, markErr)
		}
		return
	}

	nextRetryAt := time.Now().Add(outboxRetryDelay(event.RetryCount))
	if retryErr := db.MarkOutboxEventRetry(ctx, event.ID, nextRetryAt); retryErr != nil {
		log.Printf("更新 Outbox 事件重试状态失败，event_id=%d，error=%v", event.ID, retryErr)
		return
	}

	log.Printf(
		"Outbox 事件发送失败，event_id=%d，retry_count=%d，next_retry_at=%s，error=%v",
		event.ID,
		event.RetryCount+1,
		nextRetryAt.Format(time.RFC3339),
		err,
	)
}

func publishOutboxEvent(ctx context.Context, event *model.OutboxEvent) error {
	switch event.EventType {
	case mq.VideoHotRefreshKey:
		var payload mq.VideoHotRefreshEvent
		if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
			return fmt.Errorf("解析视频热度刷新事件失败: %w", err)
		}
		if payload.VideoID <= 0 {
			return fmt.Errorf("视频 ID 不合法")
		}

		return mq.PublishVideoHotRefresh(ctx, payload.VideoID)
	default:
		return fmt.Errorf("不支持的 Outbox 事件类型: %s", event.EventType)
	}
}

func outboxRetryDelay(retryCount int) time.Duration {
	delay := outboxRetryBaseDelay

	for i := 0; i < retryCount && delay < outboxRetryMaxDelay; i++ {
		delay *= 2
	}

	if delay > outboxRetryMaxDelay {
		return outboxRetryMaxDelay
	}
	return delay
}
