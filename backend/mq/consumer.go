package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// VideoHotRefreshHandler 表示处理视频热度刷新事件的业务函数。
type VideoHotRefreshHandler func(ctx context.Context, videoID int64) error

// StartVideoHotRefreshConsumer 启动视频热度刷新消费者。
func StartVideoHotRefreshConsumer(handler VideoHotRefreshHandler) error {
	if handler == nil {
		return fmt.Errorf("视频热度刷新处理函数不能为空")
	}
	if conn == nil {
		return fmt.Errorf("RabbitMQ 尚未初始化")
	}
	if consumerCh != nil && !consumerCh.IsClosed() {
		return fmt.Errorf("视频热度刷新消费者已经启动")
	}

	newConsumerCh, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("创建 RabbitMQ 消费 Channel 失败: %w", err)
	}

	// 消费者也声明自己依赖的交换机、队列和绑定。
	if err := declareTopology(newConsumerCh); err != nil {
		_ = newConsumerCh.Close()
		return err
	}

	// 最多允许当前消费者持有 10 条尚未确认的消息。
	if err := newConsumerCh.Qos(10, 0, false); err != nil {
		_ = newConsumerCh.Close()
		return fmt.Errorf("设置 RabbitMQ 消费者预取数量失败: %w", err)
	}

	deliveries, err := newConsumerCh.Consume(
		VideoHotRefreshQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = newConsumerCh.Close()
		return fmt.Errorf("启动 RabbitMQ 消费者失败: %w", err)
	}

	consumerCh = newConsumerCh

	go consumeVideoHotRefreshDeliveries(deliveries, handler)
	return nil
}

// consumeVideoHotRefreshDeliveries 持续读取 RabbitMQ 推送的消息。
func consumeVideoHotRefreshDeliveries(deliveries <-chan amqp.Delivery, handler VideoHotRefreshHandler) {
	for delivery := range deliveries {
		handleVideoHotRefreshDelivery(delivery, handler)
	}

	log.Println("RabbitMQ 视频热度刷新消费者已停止")
}

// handleVideoHotRefreshDelivery 处理单条视频热度刷新消息。
func handleVideoHotRefreshDelivery(delivery amqp.Delivery, handler VideoHotRefreshHandler) {
	var event VideoHotRefreshEvent

	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		log.Printf("解析视频热度刷新消息失败，body=%s，error=%v", string(delivery.Body), err)

		if nackErr := delivery.Nack(false, false); nackErr != nil {
			log.Printf("拒绝非法 RabbitMQ 消息失败: %v", nackErr)
		}
		return
	}

	if event.VideoID <= 0 {
		log.Printf("视频热度刷新消息中的 video_id 不合法：%d", event.VideoID)

		if nackErr := delivery.Nack(false, false); nackErr != nil {
			log.Printf("拒绝非法 RabbitMQ 消息失败: %v", nackErr)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := handler(ctx, event.VideoID)
	cancel()

	if err != nil {
		// 第一次失败重新入队一次；已经重投过仍失败，则不再无限重试。
		requeue := !delivery.Redelivered

		log.Printf(
			"刷新视频热度失败，video_id=%d，redelivered=%t，requeue=%t，error=%v",
			event.VideoID,
			delivery.Redelivered,
			requeue,
			err,
		)

		if nackErr := delivery.Nack(false, requeue); nackErr != nil {
			log.Printf("Nack RabbitMQ 消息失败: %v", nackErr)
		}
		return
	}

	if err := delivery.Ack(false); err != nil {
		log.Printf("确认 RabbitMQ 消息失败，video_id=%d，error=%v", event.VideoID, err)
	}
}
