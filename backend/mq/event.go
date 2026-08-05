package mq

// VideoHotRefreshEvent 表示某个视频需要重新计算热门分数。
type VideoHotRefreshEvent struct {
	VideoID int64 `json:"video_id"`
}
