package main

import (
	"context"
	"fmt"
	"log"

	"github.com/your-username/go-streaming-server/internal/domain/entity"
	"github.com/your-username/go-streaming-server/internal/infrastructure/memory"
	"github.com/your-username/go-streaming-server/internal/usecase"
)

func main() {
	// 1. インフラ層(リポジトリ)の初期化
	// ここを将来 MySQLRepositoryに変えるだけで、DBが切り替わります。
	repo := memory.NewStreamRepository()

	// テストデータ(配信枠)を用意
	testLey := "Live_demo_key"
	repo.Seed(entity.NewStream("stream-1", testLey))

	// 2. ユースケース層の初期化(インフラを注入)
	stratStreamUC := usecase.NewStartStreamUseCase(repo)

	// 3. 実行してみる
	ctx := context.Background()
	fmt.Println("--- 配信開始リクエスト受信 ---")

	err := stratStreamUC.Run(ctx, usecase.StartStrreamInput{
		StreamKey: testLey,
	})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// 念のため、本当にステータスが変わったか確認
	savedStream, _ := repo.FindByKey(ctx, testLey)
	fmt.Printf("現在のの状態: %s (開始時刻: %v)\n", savedStream.Status, savedStream.StraemAt)
}
