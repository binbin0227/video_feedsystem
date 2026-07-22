package db

import (
	"fmt"
	"time"

	"video_feedsystem/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDatabase 连接数据库并自动创建或更新数据表。
func InitDatabase(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}
	// 限制连接总量并定期回收连接，避免长时间运行时无限创建连接或持有失效连接。
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接检查失败: %w", err)
	}

	if err := DB.AutoMigrate(
		&model.Account{},
		&model.Video{},
		&model.Like{},
		&model.Comment{},
		&model.Social{},
	); err != nil {
		return fmt.Errorf("数据库建表失败: %w", err)
	}

	return nil
}
