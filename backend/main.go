package main

import (
	"context"
	"log"
	"time"

	"video_feedsystem/config"
	"video_feedsystem/dal/db"
	"video_feedsystem/dal/redis"
	"video_feedsystem/mq"
	"video_feedsystem/router"
	"video_feedsystem/service"
	"video_feedsystem/storage"
	"video_feedsystem/utils"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

const hotVideoRebuildInterval = 10 * time.Minute // 定时重建热门榜的周期

// 定时周期性地在后台重建热门榜
func startHotVideoRebuild(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(hotVideoRebuildInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := service.RebuildHotVideos(ctx); err != nil {
					log.Printf("定时重建热门榜失败: %v", err)
					continue
				}
				log.Println("定时重建热门榜完成")
			}
		}
	}()
}

func main() {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取配置参数
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化上传目录
	if err := storage.Init(cfg.UploadRoot); err != nil {
		log.Fatalf("初始化上传目录失败: %v", err)
	}
	// 初始化数据库
	if err := db.InitDatabase(cfg.MySQLDSN); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	// 初始化雪花算法
	if err := utils.InitSnowflake(); err != nil {
		log.Fatalf("雪花算法初始化失败: %v", err)
	}
	// 初始化 JWT
	if err := utils.InitJWT(cfg.JWTSecret); err != nil {
		log.Fatalf("JWT 初始化失败: %v", err)
	}
	// 初始化 Redis，初始化成功后重建热门榜
	if err := redis.InitRedis(cfg.RedisAddr, cfg.RedisPwd); err != nil {
		log.Fatalf("Redis 初始化失败: %v", err)
	}
	if err := service.RebuildHotVideos(appCtx); err != nil {
		log.Printf("热门榜重建失败，服务将继续启动: %v", err)
	} else {
		log.Println("热门榜重建完成")
	}
	// 定时周期性地在后台重建热门榜
	startHotVideoRebuild(appCtx)
	// 初始化 RabbitMQ
	if err := mq.InitRabbitMQ(cfg.RabbitMQURL); err != nil {
		log.Printf("RabbitMQ 初始化失败，服务将继续启动并在后台重连: %v", err)
	} else {
		log.Println("RabbitMQ 初始化完成")
	}
	defer mq.Close()

	if err := mq.StartVideoHotRefreshConsumer(service.RefreshHotVideo); err != nil {
		log.Printf("RabbitMQ 热度刷新消费者当前未启动，将在连接恢复后自动启动: %v", err)
	} else {
		log.Println("RabbitMQ 热度刷新消费者启动完成")
	}

	service.StartOutboxRelay(appCtx)
	log.Println("Outbox Relay 启动完成")

	h := server.Default(
		server.WithHostPorts(cfg.HostPorts),
		server.WithMaxRequestBodySize(220*1024*1024), // 最大请求体为 220 MB
	)

	h.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{
			"GET",
			"POST",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		MaxAge: 12 * time.Hour,
	}))

	// 挂载路由
	router.InitRouter(h, cfg.UploadEnabled)
	h.Spin()
}
