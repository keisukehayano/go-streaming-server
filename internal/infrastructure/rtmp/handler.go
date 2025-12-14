package rtmp

import (
	"context"
	"io"
	"log"
	"net"
	"strings"

	// ★重要：パッケージ名 rtmp と衝突しないようにエイリアス(rtmplib)をつける
	rtmplib "github.com/yutopp/go-rtmp"
	"github.com/yutopp/go-rtmp/message"

	"github.com/your-username/go-streaming-server/internal/infrastructure/transcoder"
	"github.com/your-username/go-streaming-server/internal/usecase"
)

// Handler は RTMP 接続全体の管理と依存性の注入を担当します
type Handler struct {
	startStreamUC *usecase.StartStreamUseCase
	transcoder    *transcoder.HLSTranscoder
}

func NewHandler(startUC *usecase.StartStreamUseCase) *Handler {
	return &Handler{
		startStreamUC: startUC,
		transcoder:    transcoder.NewHLSTranscoder(),
	}
}

// Serve はリスナーを受け取り、RTMPサーバーを開始します(ブロッキング)
func (h *Handler) Serve(listener net.Listener) error {
	// yutopp/go-rtmp のサーバー設定
	config := &rtmplib.ServerConfig{
		// 接続が来るたびに呼ばれるファクトリー関数
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmplib.ConnConfig) {
			return conn, &rtmplib.ConnConfig{
				// 接続ごとの処理を行うハンドラー(ConnHandler)を生成して渡す
				Handler: &ConnHandler{
					parent: h,
					conn:   conn,
				},
				Logger: nil,
			}
		},
	}
	srv := rtmplib.NewServer(config)
	return srv.Serve(listener)
}

// --- 以下、接続ごとの処理を行う内部ハンドラー ---

// ConnHandler は1つの接続に対するコールバックを処理します
// rtmp.DefaultHandler を埋め込むことで、未実装のメソッドをデフォルト処理させます
type ConnHandler struct {
	rtmplib.DefaultHandler
	parent *Handler
	conn   net.Conn
	// 配信開始
	ffmpegStdin io.WriteCloser
	flvWriter   *transcoder.FLVWriter
	isLive      bool
}

// OnServe は接続確立時に呼ばれます
func (h *ConnHandler) OnServe(conn *rtmplib.Conn) {
	// 接続開始時の処理があればここに記述
}

func (h *ConnHandler) OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error {
	return nil
}

// OnConnect : 接続時のログ出し（エラーにならないようCommandObjectへのアクセスは避ける）
func (h *ConnHandler) OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error {
	log.Printf("Client Connected: %+v", cmd)
	return nil
}

// OnPublish : 配信開始リクエスト
func (h *ConnHandler) OnPublish(ctx *rtmplib.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
	streamKey := cmd.PublishingName
	if idx := strings.Index(streamKey, "?"); idx != -1 {
		streamKey = streamKey[:idx]
	}

	log.Printf("Client requested PUBLISH. Key: %s", streamKey)

	// 1. [Control Plane] UseCase実行
	input := usecase.StartStreamInput{StreamKey: streamKey}
	if err := h.parent.startStreamUC.Run(context.Background(), input); err != nil {
		log.Printf("❌ Auth Failed: %v", err)
		return err
	}

	// 2. [Infrastructure] FFmpegプロセス起動
	pipe, err := h.parent.transcoder.Start(streamKey)
	if err != nil {
		log.Printf("❌ Transcoder Start Failed: %v", err)
		return err
	}

	h.ffmpegStdin = pipe
	h.flvWriter = transcoder.NewFLVWriter(pipe)

	if err := h.flvWriter.WriteHeader(); err != nil {
		h.ffmpegStdin.Close()
		return err
	}

	h.isLive = true
	log.Printf("✅ Stream [%s] Started! Piping data...", streamKey)

	return nil
}

func (h *ConnHandler) OnAudio(timestamp uint32, payload io.Reader) error {
	if !h.isLive || h.flvWriter == nil {
		_, err := io.Copy(io.Discard, payload)
		return err
	}
	data, err := io.ReadAll(payload)
	if err != nil {
		return err
	}
	return h.flvWriter.WriteRaw(0x08, timestamp, data)
}

func (h *ConnHandler) OnVideo(timestamp uint32, payload io.Reader) error {
	if !h.isLive || h.flvWriter == nil {
		_, err := io.Copy(io.Discard, payload)
		return err
	}
	data, err := io.ReadAll(payload)
	if err != nil {
		return err
	}
	return h.flvWriter.WriteRaw(0x09, timestamp, data)
}

func (h *ConnHandler) OnClose() {
	if h.ffmpegStdin != nil {
		log.Println("Closing FFmpeg pipe...")
		h.ffmpegStdin.Close()
	}
}
