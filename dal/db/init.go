package db

import (
	"log"
	"video_feedsystem/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	// 连接数据库
	dsn := "root:123456@tcp(127.0.0.1:3306)/video_feedsystem?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动建表
	err = DB.AutoMigrate(
		&model.Account{},
		&model.Video{},
		&model.Like{},
		&model.Comment{},
		&model.Social{},
	)
	if err != nil {
		log.Fatalf("数据库建表失败: %v", err)
	}

	log.Println("数据库连接成功")
}
