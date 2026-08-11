package mq

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	VideoEventExchange   = "video.events"
	VideoHotRefreshQueue = "video.hot.refresh"
	VideoHotRefreshKey   = "video.hot.refresh"
)

var (
	conn       *amqp.Connection
	publishCh  *amqp.Channel
	consumerCh *amqp.Channel
	publishMu  sync.Mutex
)

// 声明视频事件交换机、热门刷新队列、绑定关系。
func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		VideoEventExchange,
		"direct",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,
	); err != nil {
		return fmt.Errorf("声明 RabbitMQ 交换机失败: %w", err)
	}

	if _, err := ch.QueueDeclare(
		VideoHotRefreshQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	); err != nil {
		return fmt.Errorf("声明 RabbitMQ 队列失败: %w", err)
	}

	if err := ch.QueueBind(
		VideoHotRefreshQueue,
		VideoHotRefreshKey,
		VideoEventExchange,
		false, // noWait
		nil); err != nil {
		return fmt.Errorf("绑定 RabbitMQ 队列失败: %w", err)
	}

	return nil
}

func InitRabbitMQ(url string) error {
	newConn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	newPublishCh, err := newConn.Channel()
	if err != nil {
		_ = newConn.Close()
		return fmt.Errorf("创建 RabbitMQ Channel 失败: %w", err)
	}

	if err := declareTopology(newPublishCh); err != nil {
		_ = newPublishCh.Close()
		_ = newConn.Close()
		return err
	}

	if err := newPublishCh.Confirm(false); err != nil {
		_ = newPublishCh.Close()
		_ = newConn.Close()
		return fmt.Errorf("开启 RabbitMQ 发布确认模式失败: %w", err)
	}

	conn = newConn
	publishCh = newPublishCh
	return nil
}

func Close() {
	if consumerCh != nil {
		_ = consumerCh.Close()
	}

	publishMu.Lock()
	if publishCh != nil {
		_ = publishCh.Close()
	}
	publishMu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}
