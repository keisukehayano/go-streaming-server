package entity

import (
	"errors"
	"time"
)

// 配信の状態を表す型
type StreamStatus string

const (
	StatusIdle         StreamStatus = "IDLE"
	StatusLive         StreamStatus = "LIVE"
	StatusReconnecting StreamStatus = "RECONNECTING"
)

var ErrStreamAlreadyLive = errors.New("stream is already live")

// Stream 配信の集約(Aggregate)
type Stream struct {
	ID          string
	StreamKey   string
	Status      StreamStatus
	StartedAt   *time.Time
	ViewerCount int
}

// NewStream は初期状態のストリームを生成します。
func NewStream(id, key string) *Stream {
	return &Stream{
		ID:        id,
		StreamKey: key,
		Status:    StatusIdle,
	}
}

// GoLive は配信開始のビジネスロジックです。
// 単なるデータ更新ではなく、状態遷移のルールをここに集約します。
func (s *Stream) GoLive() error {
	if s.Status == StatusLive {
		return ErrStreamAlreadyLive
	}
	now := time.Now()
	s.StartedAt = &now
	s.Status = StatusLive
	return nil
}

// Stop は配信終了です(猶予期間のロジックなどは別途検討可能)
func (s *Stream) Stop() {
	s.Status = StatusIdle
	s.StartedAt = nil
}
