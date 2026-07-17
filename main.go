package main

import (
	"log"

	"video_feedsystem/config"
	"video_feedsystem/dal/db"
	"video_feedsystem/router"
	"video_feedsystem/utils"

	"github.com/cloudwego/hertz/pkg/app/server"
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

	router.InitRouter(h)
	h.Spin()
}
