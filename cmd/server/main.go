package main

import (
	"log"
	"net"
	"net/http"

	"github.com/your-username/go-streaming-server/internal/domain/entity"
	"github.com/your-username/go-streaming-server/internal/infrastructure/memory"
	"github.com/your-username/go-streaming-server/internal/infrastructure/rtmp"
	"github.com/your-username/go-streaming-server/internal/usecase"
)

func main() {

	// 1. 依存関係のセットアップ
	repo := memory.NewStreamRepository()
	repo.Seed(entity.NewStream("stream-1", "live_demo"))

	startStreamUC := usecase.NewStartStreamUseCase(repo)
	handler := rtmp.NewHandler(startStreamUC)

	// 2. TCPリスナーの作成
	port := ":1935"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", port, err)
	}
	// ハンドラー内でServeするので、ここでCloseはdeferしない（またはServe終了後にClose）
	// 通常はmain終了まで生きるため、このままでもOKですが、行儀よくするなら後述のServe内で管理します。

	log.Printf("RTMP Server is listening on %s", port)

	// ★追加: HLS配信用のHTTPサーバー
	go func() {
		httpPort := ":8080"
		log.Printf("HLS Server is listening on %s", httpPort)

		// hlsディレクトリを公開
		fs := http.FileServer(http.Dir("./hls"))

		// CORSヘッダー（ブラウザから見るために必須）を追加するミドルウェア
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			fs.ServeHTTP(w, r)
		})

		http.Handle("/", handler)
		if err := http.ListenAndServe(httpPort, nil); err != nil {
			log.Printf("HTTP Server Error: %v", err)
		}
	}()

	// 3. ハンドラーにリスナーを渡して実行 (ブロック処理)
	//    Acceptループはハンドラー内部（ライブラリ）で行います
	if err := handler.Serve(listener); err != nil {
		log.Fatalf("Server finished with error: %v", err)
	}

}
