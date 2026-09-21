package mq

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	VideoEventExchange   = "video.events"
	VideoHotRefreshQueue = "video.hot.refresh"
	VideoHotRefreshKey   = "video.hot.refresh"

	rabbitMQReconnectBaseDelay = time.Second
	rabbitMQReconnectMaxDelay  = 30 * time.Second
)

var (
	conn            *amqp.Connection
	publishCh       *amqp.Channel
	consumerCh      *amqp.Channel
	consumerHandler VideoHotRefreshHandler
	resourceMu      sync.Mutex
	managerCancel   context.CancelFunc
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
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("RabbitMQ 连接地址不能为空")
	}

	resourceMu.Lock()
	if managerCancel != nil {
		resourceMu.Unlock()
		return fmt.Errorf("RabbitMQ 已经初始化")
	}
	managerCtx, cancel := context.WithCancel(context.Background())
	managerCancel = cancel
	resourceMu.Unlock()

	newConn, newPublishCh, err := openRabbitMQPublisher(url)
	if err != nil {
		go maintainRabbitMQConnection(managerCtx, url, nil, nil)
		return err
	}

	resourceMu.Lock()
	if managerCtx.Err() != nil {
		resourceMu.Unlock()
		_ = newPublishCh.Close()
		_ = newConn.Close()
		return fmt.Errorf("RabbitMQ 初始化已取消")
	}
	conn = newConn
	publishCh = newPublishCh
	resourceMu.Unlock()

	go maintainRabbitMQConnection(managerCtx, url, newConn, newPublishCh)
	return nil
}

func openRabbitMQPublisher(url string) (*amqp.Connection, *amqp.Channel, error) {
	newConn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	newPublishCh, err := newConn.Channel()
	if err != nil {
		_ = newConn.Close()
		return nil, nil, fmt.Errorf("创建 RabbitMQ Channel 失败: %w", err)
	}

	if err := declareTopology(newPublishCh); err != nil {
		_ = newPublishCh.Close()
		_ = newConn.Close()
		return nil, nil, err
	}

	if err := newPublishCh.Confirm(false); err != nil {
		_ = newPublishCh.Close()
		_ = newConn.Close()
		return nil, nil, fmt.Errorf("开启 RabbitMQ 发布确认模式失败: %w", err)
	}

	return newConn, newPublishCh, nil
}

func maintainRabbitMQConnection(ctx context.Context, url string, currentConn *amqp.Connection, currentPublishCh *amqp.Channel) {
	for {
		if currentConn != nil && currentPublishCh != nil {
			connClosed := currentConn.NotifyClose(make(chan *amqp.Error, 1))
			publishClosed := currentPublishCh.NotifyClose(make(chan *amqp.Error, 1))

			select {
			case <-ctx.Done():
				return
			case closeErr := <-connClosed:
				logRabbitMQClosed("连接", closeErr)
			case closeErr := <-publishClosed:
				logRabbitMQClosed("发布 Channel", closeErr)
				_ = currentConn.Close()
			}
		}

	reconnectLoop:
		for delay := rabbitMQReconnectBaseDelay; ; delay = nextRabbitMQReconnectDelay(delay) {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return
			case <-timer.C:
			}

			newConn, newPublishCh, err := openRabbitMQPublisher(url)
			if err != nil {
				log.Printf("RabbitMQ 重连失败，%s 后继续重试: %v", nextRabbitMQReconnectDelay(delay), err)
				continue
			}

			resourceMu.Lock()
			handler := consumerHandler
			resourceMu.Unlock()

			var newConsumerCh *amqp.Channel
			var deliveries <-chan amqp.Delivery
			if handler != nil {
				newConsumerCh, deliveries, err = openVideoHotRefreshConsumer(newConn)
				if err != nil {
					_ = newPublishCh.Close()
					_ = newConn.Close()
					log.Printf("RabbitMQ 重连后恢复消费者失败，%s 后继续重试: %v", nextRabbitMQReconnectDelay(delay), err)
					continue
				}
			}

			resourceMu.Lock()
			if ctx.Err() != nil {
				resourceMu.Unlock()
				if newConsumerCh != nil {
					_ = newConsumerCh.Close()
				}
				_ = newPublishCh.Close()
				_ = newConn.Close()
				return
			}
			oldConn := conn
			oldPublishCh := publishCh
			oldConsumerCh := consumerCh
			conn = newConn
			publishCh = newPublishCh
			consumerCh = newConsumerCh
			resourceMu.Unlock()

			closeRabbitMQResources(oldConn, oldPublishCh, oldConsumerCh)

			if deliveries != nil {
				go consumeVideoHotRefreshDeliveries(deliveries, handler)
			}

			log.Println("RabbitMQ 重连成功，发布和消费链路已经恢复")
			currentConn = newConn
			currentPublishCh = newPublishCh
			break reconnectLoop
		}
	}
}

func nextRabbitMQReconnectDelay(delay time.Duration) time.Duration {
	delay *= 2
	if delay > rabbitMQReconnectMaxDelay {
		return rabbitMQReconnectMaxDelay
	}
	return delay
}

func logRabbitMQClosed(resource string, closeErr *amqp.Error) {
	if closeErr == nil {
		log.Printf("RabbitMQ %s已关闭，准备自动重连", resource)
		return
	}
	log.Printf("RabbitMQ %s异常关闭，准备自动重连: %v", resource, closeErr)
}

func closeRabbitMQResources(oldConn *amqp.Connection, oldPublishCh, oldConsumerCh *amqp.Channel) {
	if oldConsumerCh != nil && !oldConsumerCh.IsClosed() {
		_ = oldConsumerCh.Close()
	}
	if oldPublishCh != nil && !oldPublishCh.IsClosed() {
		_ = oldPublishCh.Close()
	}
	if oldConn != nil && !oldConn.IsClosed() {
		_ = oldConn.Close()
	}
}

func Close() {
	resourceMu.Lock()
	cancel := managerCancel
	managerCancel = nil
	oldConn := conn
	oldPublishCh := publishCh
	oldConsumerCh := consumerCh
	conn = nil
	publishCh = nil
	consumerCh = nil
	consumerHandler = nil
	resourceMu.Unlock()

	if cancel != nil {
		cancel()
	}
	closeRabbitMQResources(oldConn, oldPublishCh, oldConsumerCh)
}
