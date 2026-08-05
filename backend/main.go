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
	cfg := config.Load()

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
	// 初始化 Redis ，初始成功化后重建热门榜
	if err := redis.InitRedis(cfg.RedisAddr, cfg.RedisPwd); err != nil {
		log.Printf("Redis 初始化失败，服务将继续启动: %v", err)
	} else if err := service.RebuildHotVideos(appCtx); err != nil {
		log.Printf("热门榜重建失败，服务将继续启动: %v", err)
	} else {
		log.Println("热门榜重建完成")
	}
	// 定时周期性地在后台重建热门榜
	startHotVideoRebuild(appCtx)
	// 初始化 RabbitMQ
	if err := mq.InitRabbitMQ(cfg.RabbitMQURL); err != nil {
		log.Printf("RabbitMQ 初始化失败，服务将继续启动: %v", err)
	} else {
		defer mq.Close()
		if err := mq.StartVideoHotRefreshConsumer(service.RefreshHotVideo); err != nil {
			log.Printf("RabbitMQ 热度刷新消费者启动失败: %v", err)
		} else {
			log.Println("RabbitMQ 热度刷新消费者启动完成")
		}
		log.Println("RabbitMQ 初始化完成")
	}

	h := server.Default(
		server.WithHostPorts(cfg.HostPorts),
		server.WithMaxRequestBodySize(220*1024*1024), // 最大请求体为 220 MB
	)

	h.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5189",
			"http://127.0.0.1:5189",
			"https://frp-fun.com:64718",
		},
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
	router.InitRouter(h)
	h.Spin()
}
