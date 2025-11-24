package usecase

import (
	"context"
	"fmt"

	"github.com/your-username/go-streaming-server/internal/domain/repository"
)

// ユーザーからの入力データ(DTD)
type StartStrreamInput struct {
	StreamKey string
}

// アプリケーションロジック
type StartStreamUseCase struct {
	// 具体的なDBではなく、インターフェースに依存する(依存性逆転の原則)
	repo repository.StreamRepository
}

// コンストラクタでリポジトリを注入(Dependency Injection)
func NewStartStreamUseCase(repo repository.StreamRepository) *StartStreamUseCase {
	return &StartStreamUseCase{
		repo: repo,
	}
}

func (u *StartStreamUseCase) Run(ctx context.Context, input StartStrreamInput) error {
	// 1. ドメインオブジェクトの取得(リポジトリ経由)
	stream, err := u.repo.FindByKey(ctx, input.StreamKey)
	if err != nil {
		// 本来はここで"NotFound"なら新規作成するなどの分岐が入りますが、
		// 今回は簡単のため、見つからなければエラーとします。
		return fmt.Errorf("failed to find stram: %w", err)
	}

	// 2. ドメインロジックの実行(ビジネスルールの適用)
	// "配信開始できるか？"の判断は useCase ではなく Entity が行います。
	if err := stream.GoLive(); err != nil {
		return fmt.Errorf("faied to go live: %w", err)
	}

	// 3. 状態の保存
	if err := u.repo.Save(ctx, stream); err != nil {
		return fmt.Errorf("failed to save stream: %w", err)
	}

	fmt.Printf("✅ Stream [%s] is now LIVE! (Viewers: %d)\n", stream.ID, stream.ViewerCount)
	return nil
}
