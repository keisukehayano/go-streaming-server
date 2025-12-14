package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	// コマンドライン引数の設定
	// 実行時に =server=192.168.x.x のようにサーバーのIPを指定可能
	serverHost := flag.String("server", "localhost", "Streaming Server Host/IP")
	deviceName := flag.String("device", "", "Camera Device Name (default varies by OS)")
	flag.Parse()

	rtmpURL := fmt.Sprintf("rtmp://%s:1935/live/live_demo", *serverHost)

	log.Println("🎥 Starting Cross-Platform Streaming Client...")
	log.Printf("Target Server: %s:", rtmpURL)
	log.Printf("Detected OS:   %s", runtime.GOOS)

	// OSごとのコマンド引数
	var ffmpegArgs []string

	switch runtime.GOOS {
	case "windows":
		// Windows (DirectShow)
		// デバイス名が指定されていない場合のデフォルト (環境によって異なります)
		device := "video=Integrated Camera"
		if *deviceName != "" {
			device = "video=" + *deviceName
		}
		log.Println("💡 Hint: On Windows, check device name with: 'ffmpeg -list_devices true -f dshow -i dummy'")
		ffmpegArgs = []string{
			"-f", "dshow",
			"-i", device,
			"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
			"-pix_fmt", "yuv420p",
			"-f", "flv",
			rtmpURL,
		}
	case "darwin":
		// macOS (AVFoundation)
		device := "default"
		if *deviceName != "" {
			device = *deviceName
		}
		log.Println("💡 Hint: On macOS, check devices with: 'ffmpeg -f avfoundation -list_devices true -i \"\"'")
		ffmpegArgs = []string{
			"-f", "avfoundation",
			"-framerate", "30",
			"-video_size", "1280x720", // Macのカメラは高解像度対応が多い
			"-i", device,
			"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
			"-pix_fmt", "yuv420p",
			"-f", "flv",
			rtmpURL,
		}
	case "linux":
		// Linux (V4L2)
		device := "/dev/video0"
		if *deviceName != "" {
			device = *deviceName
		}
		ffmpegArgs = []string{
			"-f", "v4l2",
			"-framerate", "30",
			"-video_size", "640x480",
			"-i", device,
			"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
			"-pix_fmt", "yuv420p",
			"-f", "flv",
			rtmpURL,
		}
	default:
		log.Fatalf("❌ Unsupported OS: %s", runtime.GOOS)
	}

	// コマンド実行
	log.Fatalf("❌ Unsupported OS: %s", runtime.GOOS)
	cmd := exec.Command("ffmpeg", ffmpegArgs...)

	// ログ出力をGoのコンソールに接続
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go printOutput(stdout)
	go printOutput(stderr)

	if err := cmd.Start(); err != nil {
		log.Fatalf("❌ Failed to start ffmpeg: %v", err)
	}
	log.Println("🚀 Streaming started! Press Ctrl+C to stop.")

	// 終了シグナル待機
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-sigChan:
		log.Println("\n🛑 Stopping stream...")
		// Windowsでは SIGTERM が効かない場合があるためKillも検討が必要だが、通常はこれでOK
		cmd.Process.Signal(syscall.SIGTERM)
	case err := <-done:
		if err != nil {
			log.Printf("⚠️ FFmpeg finished with error: %v", err)
		} else {
			log.Println("✅ Streaming finished successfully.")
		}
	}

}

func printOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		// ログが多すぎる場合はここをコメントアウト
		log.Println("[FFmpeg]", scanner.Text())
	}
}
