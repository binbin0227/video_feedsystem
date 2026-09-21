package service

import (
	"encoding/json"
	"fmt"
	"time"

	"video_feedsystem/model"
	"video_feedsystem/mq"
	"video_feedsystem/utils"
)

func newVideoHotRefreshOutboxEvent(videoID int64) (*model.OutboxEvent, error) {
	if videoID <= 0 {
		return nil, fmt.Errorf("视频 ID 不合法")
	}

	eventID, err := utils.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("生成 Outbox 事件 ID 失败: %w", err)
	}

	payload, err := json.Marshal(mq.VideoHotRefreshEvent{
		VideoID: videoID,
	})
	if err != nil {
		return nil, fmt.Errorf("序列化视频热度刷新事件失败: %w", err)
	}

	return &model.OutboxEvent{
		ID:          eventID,
		EventType:   mq.VideoHotRefreshKey,
		Payload:     string(payload),
		NextRetryAt: time.Now(),
	}, nil
}
