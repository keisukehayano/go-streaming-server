package repository

import (
	"context"

	"github.com/your-username/go-streaming-server/internal/domain/entity"
)

// StreamRepository はデータ永続化に関する契約(インターフェース)です。
// 実装は、internal/infrastuctureで行います。
type StreamRepository interface {
	FindByKey(ctx context.Context, key string) (*entity.Stream, error)
	Save(ctx context.Context, stram *entity.Stream)
	UpdateStatus(ctx context.Context, id string, stream entity.StreamStatus) error
}
