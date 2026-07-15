package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"video_feedsystem/dal/db"
	"video_feedsystem/router"
	"video_feedsystem/utils"
)

func main() {
	db.InitDatabase()
	utils.InitSnowFlake()
	
	h := server.Default(server.WithHostPorts("0.0.0.0:8080"))
	router.InitRouter(h)
	h.Spin()
}
