package utils

import (
	"errors"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

// InitSnowflake 初始化全局 ID 生成器。
func InitSnowflake() error {
	sf = sonyflake.NewSonyflake(sonyflake.Settings{})
	if sf == nil {
		return errors.New("雪花算法初始化失败")
	}
	return nil
}

// GenerateID 生成一个 int64 类型的全局唯一 ID。
func GenerateID() (int64, error) {
	if sf == nil {
		return 0, errors.New("雪花算法尚未初始化")
	}

	id, err := sf.NextID()
	if err != nil {
		return 0, err
	}
	return int64(id), nil
}
