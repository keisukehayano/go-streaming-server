`go run cmd/server/main.go`

```
ffmpeg -re -f lavfi -i testsrc=size=640x360:rate=30 \
  -f lavfi -i sine=frequency=1000 \
  -c:v libx264 -preset ultrafast -c:a aac \
  -f flv rtmp://localhost:1935/live/live_demo
  ```

`http://localhost:8080/live_demo/index.m3u8`


コンパイルと配布（クライアントの作り方）
このコードは、OSごとにコンパイル（ビルド）して実行ファイルとして配布するのが一般的です。

1. Windows用クライアントを作る場合
MacやLinux上で開発している場合、クロスコンパイル機能を使ってWindows用の .exe を作れます。

Bash

# Windows用ビルド (client.exe が生成されます)
GOOS=windows GOARCH=amd64 go build -o client.exe cmd/client/main.go
Windowsでの実行方法: PowerShellなどで実行します。カメラ名がデフォルト（Integrated Camera）と違う場合は -device で指定します。

PowerShell

.\client.exe -server 192.168.1.10 -device "Logitech HD Webcam C270"
※ Windowsにはあらかじめ FFmpegのインストール とPATHの設定が必要です。

2. macOS用クライアントを作る場合
Bash

# macOS (Intel) 用
GOOS=darwin GOARCH=amd64 go build -o client_mac cmd/client/main.go

# macOS (Apple Silicon/M1,M2) 用
GOOS=darwin GOARCH=arm64 go build -o client_mac_m1 cmd/client/main.go
macOSでの実行方法:

Bash

./client_mac -server 192.168.1.10
※ 初回実行時、ターミナルに対して「カメラへのアクセス権限」を求められる場合があります。

3. Linux用クライアントを作る場合
Bash

GOOS=linux GOARCH=amd64 go build -o client_linux cmd/client/main.go
Linuxでの実行方法:

Bash

./client_linux -server 192.168.1.10 -device /dev/video0
注意点：カメラデバイス名の特定について
クライアント側で一番つまづくポイントは「カメラの名前」です。特にWindowsは機種によって名前が異なります。

Windowsの場合: カメラ名がわからない場合、以下のコマンドをユーザーに打ってもらい確認する必要があります。

PowerShell

ffmpeg -list_devices true -f dshow -i dummy
出てきた名前（例: "USB Video Device"）を、-device オプションに渡して実行します。

macOSの場合: 通常は "default" または "0" で動きますが、動かない場合は以下で確認します。

Bash

ffmpeg -f avfoundation -list_devices true -i ""
これで、サーバー（Ubuntu）を中心に、あらゆるOSから映像を送り込めるシステムが完成しました！


ver 2

各OSごとのビルドと配布方法
以下のコマンドでビルドし、実行ファイル と FFmpegバイナリ を同じフォルダに入れて配布してください。

🪟 Windows向け
ビルド:

PowerShell

set GOOS=windows
set GOARCH=amd64
go build -o stream_client.exe cmd/client/main.go
配布物: stream_client.exe + ffmpeg.exe

実行: コマンドプロンプト等で stream_client.exe -server 192.168.x.x

🍎 macOS向け (Apple Silicon)
ビルド:

Bash

GOOS=darwin GOARCH=arm64 go build -o stream_client_mac cmd/client/main.go
配布物: stream_client_mac + ffmpeg (拡張子なし)

実行: ターミナルで ./stream_client_mac -server 192.168.x.x (初回は「開発元未確認」の許可が必要になる場合があります)

🐧 Linux向け
ビルド:

Bash

GOOS=linux GOARCH=amd64 go build -o stream_client_linux cmd/client/main.go
配布物: stream_client_linux + ffmpeg

実行: ./stream_client_linux -server 192.168.x.x
