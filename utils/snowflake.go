package utils

import (
	"github.com/sony/sonyflake"
	"log"
)

var sf *sonyflake.Sonyflake

// 初始化雪花算法
func InitSnowFlake() {
	var st sonyflake.Settings
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		log.Fatalf("雪花算法初始化失败")
	}
	log.Println("雪花算法初始化成功")
}

func GenerateID()(int64,error){
	id,err:=sf.NextID() // uint64
	if err!=nil{
		return 0,err
	}
	return int64(id),nil
}