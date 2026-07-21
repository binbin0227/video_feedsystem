package main

import (
	"log"
	"time"

	"video_feedsystem/config"
	"video_feedsystem/dal/db"
	"video_feedsystem/router"
	"video_feedsystem/utils"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

func main() {
	cfg := config.Load()

	if err := db.InitDatabase(cfg.MySQLDSN); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	if err := utils.InitSnowflake(); err != nil {
		log.Fatalf("雪花算法初始化失败: %v", err)
	}
	if err := utils.InitJWT(cfg.JWTSecret); err != nil {
		log.Fatalf("JWT 初始化失败: %v", err)
	}

	h := server.Default(
		server.WithHostPorts(cfg.HostPorts),
		server.WithMaxRequestBodySize(220*1024*1024),
	)

	h.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:5189",
			"http://127.0.0.1:5189",
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

	router.InitRouter(h)
	h.Spin()
}
