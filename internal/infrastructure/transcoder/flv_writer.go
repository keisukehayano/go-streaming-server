package transcoder

import (
	"bufio"
	"encoding/binary"
	"io"
)

// FLVWriter はRTMPメッセージをFLV形式でWriterに書き込みます
// バッファリング機能を追加して書き込みを安定化させます
type FLVWriter struct {
	w *bufio.Writer
}

func NewFLVWriter(w io.Writer) *FLVWriter {
	return &FLVWriter{
		w: bufio.NewWriter(w), // バッファ付きWriterでラップ
	}
}

// WriteHeader はFLVファイルのヘッダー(13バイト)を書き込みます
func (f *FLVWriter) WriteHeader() error {
	// FLV Header (9 bytes) + PreviousTagSize0 (4 bytes) = 13 bytes
	// appendを使わず、固定長配列で定義して確実に0を書き込みます
	header := []byte{
		'F', 'L', 'V', // Signature
		0x01,                   // Version
		0x05,                   // Flags (Audio + Video)
		0x00, 0x00, 0x00, 0x09, // HeaderSize
		0x00, 0x00, 0x00, 0x00, // PreviousTagSize0 (ここが重要！)
	}

	if _, err := f.w.Write(header); err != nil {
		return err
	}

	// ヘッダー書き込み直後に必ずFlushして、FFmpegにデータを到達させる
	return f.w.Flush()
}

// WriteRaw は生のペイロードデータを受け取り、FLVタグとして書き込みます
func (f *FLVWriter) WriteRaw(tagType byte, timestamp uint32, body []byte) error {
	dataSize := uint32(len(body))

	// FLV Tag Header (11 bytes)
	tagHeader := make([]byte, 11)
	tagHeader[0] = tagType

	// Data Size (24 bit)
	tagHeader[1] = byte(dataSize >> 16)
	tagHeader[2] = byte(dataSize >> 8)
	tagHeader[3] = byte(dataSize)

	// Timestamp (24 bit + 8 bit Extended)
	tagHeader[4] = byte(timestamp >> 16)
	tagHeader[5] = byte(timestamp >> 8)
	tagHeader[6] = byte(timestamp)
	tagHeader[7] = byte(timestamp >> 24)

	tagHeader[8] = 0
	tagHeader[9] = 0
	tagHeader[10] = 0

	// 1. Tag Header
	if _, err := f.w.Write(tagHeader); err != nil {
		return err
	}

	// 2. Body
	if _, err := f.w.Write(body); err != nil {
		return err
	}

	// 3. PreviousTagSize
	prevTagSize := uint32(11) + dataSize
	ptsBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(ptsBuf, prevTagSize)

	if _, err := f.w.Write(ptsBuf); err != nil {
		return err
	}

	// タグごとにFlushして遅延を防ぐ（低遅延配信のため）
	return f.w.Flush()
}
