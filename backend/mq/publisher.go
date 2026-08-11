package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// 发布视频热门分数刷新事件
func PublishVideoHotRefresh(ctx context.Context, videoID int64) error {
	event := VideoHotRefreshEvent{
		VideoID: videoID,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化视频热度刷新事件失败: %w", err)
	}

	message := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Type:         VideoHotRefreshKey,
		Body:         body,
	}

	publishMu.Lock()
	defer publishMu.Unlock()

	if publishCh == nil || publishCh.IsClosed() {
		return fmt.Errorf("RabbitMQ 发布 Channel 未初始化或已经关闭")
	}

	confirmation, err := publishCh.PublishWithDeferredConfirmWithContext(
		ctx,
		VideoEventExchange,
		VideoHotRefreshKey,
		false,
		false,
		message,
	)
	if err != nil {
		return fmt.Errorf("发布视频热度刷新事件失败: %w", err)
	}
	if confirmation == nil {
		return fmt.Errorf("RabbitMQ 发布 Channel 未开启确认模式")
	}

	confirmCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	acked, err := confirmation.WaitContext(confirmCtx)
	if err != nil {
		return fmt.Errorf("等待 RabbitMQ 发布确认失败: %w", err)
	}
	if !acked {
		return fmt.Errorf("RabbitMQ 拒绝确认视频热度刷新消息")
	}

	return nil
}
