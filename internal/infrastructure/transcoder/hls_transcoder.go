package transcoder

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type HLSTranscoder struct{}

func NewHLSTranscoder() *HLSTranscoder {
	return &HLSTranscoder{}
}

// Start は指定されたストリームキー用にFFmpegプロセスを起動し、
// 映像データを書き込むための Writer (標準入力のパイポ) を返します。
func (t *HLSTranscoder) Start(streamKey string) (io.WriteCloser, error) {
	// 1. 保存先ディレクトリ作成: ./hls/{streamKey}
	outputDir := filepath.Join("hls", streamKey)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// HLSプレイリスト(.m3u8)の出力パス
	playListPath := filepath.Join(outputDir, "index.m3u8")

	// 2. FFmpegコマンドの組み立て
	// -y              : 確認なしで上書き
	// -i pipe:0       : 標準入力からデータを受け取る
	// -c:v libx264    : 映像をH.264でエンコード
	// -preset ultrafast : 画質より処理速度を優先（遅延対策）
	// -tune zerolatency : ストリーミング向けの低遅延チューニング
	// -c:a aac        : 音声をAACでエンコード
	// -ar 44100       : 音声サンプリングレート
	// -f hls          : 出力フォーマットはHLS
	// -hls_time 2     : 1つのセグメントファイル(.ts)を約2秒にする
	// -hls_list_size 5: プレイリストには最新5件のみ記載する
	// -hls_flags delete_segments : 古いセグメントファイルを削除する
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", "pipe:0",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-c:a", "aac",
		"-ar", "44100",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "5",
		"-hls_flags", "delete_segments",
		playListPath,
	)

	// 3. 標準入力 (stdin) へのパイプを取得
	// ここにGoプログラムからFLVデータを書き込みます
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// 標準出力・エラー出力を現在のプロセスに流す（デバッグ用）
	// ログが多すぎる場合は cmd.Stderr をコメントアウトしてください
	cmd.Stderr = os.Stderr

	log.Printf("Starting FFmpeg for stream: %s", streamKey)

	// 4. プロセス開始
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %s", err)
	}

	// プロセス終了を監視する非同期処理
	go func() {
		err := cmd.Wait()
		log.Printf("FFmpeg process for [%s] finished. Error: %v", streamKey, err)
	}()
	return stdin, nil

}
