// compression.go - 압축 유틸리티 함수들
package cqrsx

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/pierrec/lz4/v4"
)

type CompressionType string

const (
	CompressionNone CompressionType = "none"
	CompressionGzip CompressionType = "gzip"
	CompressionLZ4  CompressionType = "lz4"
)

// compressGzip는 데이터를 GZIP으로 압축합니다
func compressGzip(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressGzip는 GZIP 압축된 데이터를 해제합니다
func decompressGzip(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from gzip reader: %w", err)
	}

	return result, nil
}

// compressLZ4는 데이터를 LZ4로 압축합니다
func compressLZ4(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	var buf bytes.Buffer
	writer := lz4.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write to lz4 writer: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close lz4 writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressLZ4는 LZ4 압축된 데이터를 해제합니다
func decompressLZ4(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	reader := lz4.NewReader(bytes.NewReader(data))

	result, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from lz4 reader: %w", err)
	}

	return result, nil
}

// getCompressionRatio는 압축률을 계산합니다
func getCompressionRatio(original, compressed []byte) float64 {
	if len(original) == 0 {
		return 0
	}
	return float64(len(compressed)) / float64(len(original))
}

// shouldCompress는 데이터 압축이 유익한지 판단합니다
func shouldCompress(data []byte, minSize int) bool {
	// 최소 크기 이하이면 압축하지 않음 (오버헤드 고려)
	if len(data) < minSize {
		return false
	}

	// 이미 압축된 것 같은 데이터인지 간단히 확인
	// (연속된 같은 바이트가 적으면 이미 압축된 것으로 간주)
	if isLikelyCompressed(data) {
		return false
	}

	return true
}

// isLikelyCompressed는 데이터가 이미 압축된 것 같은지 확인합니다
func isLikelyCompressed(data []byte) bool {
	if len(data) < 100 {
		return false
	}

	// 샘플링: 처음 100바이트에서 연속된 같은 바이트 개수 확인
	duplicates := 0
	for i := 1; i < 100 && i < len(data); i++ {
		if data[i] == data[i-1] {
			duplicates++
		}
	}

	// 연속 중복이 5% 미만이면 이미 압축된 것으로 간주
	return float64(duplicates)/100.0 < 0.05
}

// CompressionStats는 압축 통계를 나타냅니다
type CompressionStats struct {
	OriginalSize    int64   `json:"originalSize"`
	CompressedSize  int64   `json:"compressedSize"`
	CompressionType string  `json:"compressionType"`
	Ratio           float64 `json:"ratio"`
	TimeTaken       int64   `json:"timeTakenMs"`
}

// CalculateCompressionStats는 압축 통계를 계산합니다
func CalculateCompressionStats(original, compressed []byte, compressionType string, timeTaken int64) CompressionStats {
	return CompressionStats{
		OriginalSize:    int64(len(original)),
		CompressedSize:  int64(len(compressed)),
		CompressionType: compressionType,
		Ratio:           getCompressionRatio(original, compressed),
		TimeTaken:       timeTaken,
	}
}
