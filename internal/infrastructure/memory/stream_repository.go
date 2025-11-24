package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/your-username/go-streaming-server/internal/domain/entity"
)

// メモリ上でデータを管理するリポジトリの実装
type InMemoryStreamRepository struct {
	store map[string]*entity.Stream
	mu    sync.RWMutex // 並行処理用にロック機構えお用意
}

func NewStreamRepository() *InMemoryStreamRepository {
	return &InMemoryStreamRepository{
		store: make(map[string]*entity.Stream),
	}
}

// テスト用に初期データを仕込むヘルパーメソッド
func (r *InMemoryStreamRepository) Seed(stream *entity.Stream) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.store[stream.StreamKey] = stream
}

// インターフェースの実装: FindByKey
func (r *InMemoryStreamRepository) FindByKey(ctx context.Context, key string) (*entity.Stream, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if s, ok := r.store[key]; ok {
		// ポインタをそのまま返すと副作用が怖いので、本来はコピーを返すべきですが
		// 学習用として簡易的に実装します
		return s, nil
	}
	return nil, errors.New("stream not found")
}

// インタフェースの実装: Save
func (r *InMemoryStreamRepository) Save(ctx context.Context, stream *entity.Stream) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// インタフェースの実装: UpdateもInsertも同じ処理
	r.store[stream.StreamKey] = stream
	return nil
}

// インターフェースの実装: UpdateStatus (今回はSaveで代用するので空実装でも可)
func (r *InMemoryStreamRepository) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error {
	// 実際の実装は省略
	return nil
}
