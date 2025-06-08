// compression_test.go - 압축 기능 테스트
package cqrsx

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressGzip(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool // 압축이 동작해야 하는지
	}{
		{
			name:     "Empty data",
			input:    []byte{},
			expected: true,
		},
		{
			name:     "Small data",
			input:    []byte("hello world"),
			expected: true,
		},
		{
			name:     "Repetitive data (good compression)",
			input:    bytes.Repeat([]byte("test data "), 100),
			expected: true,
		},
		{
			name:     "Random-like data",
			input:    []byte("abcdefghijklmnopqrstuvwxyz1234567890!@#$%^&*()"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			compressed, err := compressGzip(tt.input)
			
			// Then
			if tt.expected {
				assert.NoError(t, err)
				if len(tt.input) > 0 {
					// 빈 데이터가 아니면 압축 결과가 있어야 함
					assert.NotEmpty(t, compressed)
				}
				
				// 압축 해제 테스트
				decompressed, err := decompressGzip(compressed)
				assert.NoError(t, err)
				assert.Equal(t, tt.input, decompressed)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCompressLZ4(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Empty data",
			input: []byte{},
		},
		{
			name:  "Small data",
			input: []byte("hello world"),
		},
		{
			name:  "Large repetitive data",
			input: bytes.Repeat([]byte("compress this text "), 500),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			compressed, err := compressLZ4(tt.input)
			
			// Then
			assert.NoError(t, err)
			
			// 압축 해제 테스트
			decompressed, err := decompressLZ4(compressed)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decompressed)
		})
	}
}

func TestCompressionEfficiency(t *testing.T) {
	// Given - 반복되는 데이터 (압축에 유리함)
	originalData := bytes.Repeat([]byte("This is test data for compression efficiency testing. "), 100)
	
	// When
	gzipCompressed, err := compressGzip(originalData)
	require.NoError(t, err)
	
	lz4Compressed, err := compressLZ4(originalData)
	require.NoError(t, err)
	
	// Then
	gzipRatio := getCompressionRatio(originalData, gzipCompressed)
	lz4Ratio := getCompressionRatio(originalData, lz4Compressed)
	
	t.Logf("Original size: %d bytes", len(originalData))
	t.Logf("GZIP compressed: %d bytes (ratio: %.2f)", len(gzipCompressed), gzipRatio)
	t.Logf("LZ4 compressed: %d bytes (ratio: %.2f)", len(lz4Compressed), lz4Ratio)
	
	// 압축률이 50% 이하여야 함 (반복 데이터이므로)
	assert.Less(t, gzipRatio, 0.5, "GZIP should compress repetitive data significantly")
	assert.Less(t, lz4Ratio, 0.8, "LZ4 should provide some compression")
}

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		minSize  int
		expected bool
	}{
		{
			name:     "Too small data",
			data:     []byte("small"),
			minSize:  100,
			expected: false,
		},
		{
			name:     "Large enough data",
			data:     bytes.Repeat([]byte("test "), 50),
			minSize:  100,
			expected: true,
		},
		{
			name:     "Already compressed-like data",
			data:     make([]byte, 200), // 모두 0인 데이터는 압축됨
			minSize:  100,
			expected: false, // isLikelyCompressed가 false를 반환할 것
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldCompress(tt.data, tt.minSize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCompressionStats(t *testing.T) {
	// Given
	original := bytes.Repeat([]byte("test data "), 100)
	compressed := bytes.Repeat([]byte("comp "), 50) // 더 작은 압축 데이터
	compressionType := "gzip"
	timeTaken := int64(150) // 150ms
	
	// When
	stats := CalculateCompressionStats(original, compressed, compressionType, timeTaken)
	
	// Then
	assert.Equal(t, int64(len(original)), stats.OriginalSize)
	assert.Equal(t, int64(len(compressed)), stats.CompressedSize)
	assert.Equal(t, compressionType, stats.CompressionType)
	assert.Equal(t, timeTaken, stats.TimeTaken)
	assert.Less(t, stats.Ratio, 1.0) // 압축되었으므로 비율이 1보다 작아야 함
}

func BenchmarkCompressionAlgorithms(b *testing.B) {
	// 테스트 데이터 준비
	smallData := bytes.Repeat([]byte("small test data "), 10)
	largeData := bytes.Repeat([]byte("large test data for compression benchmarking "), 1000)
	
	b.Run("GZIP Small Data", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			compressGzip(smallData)
		}
	})
	
	b.Run("LZ4 Small Data", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			compressLZ4(smallData)
		}
	})
	
	b.Run("GZIP Large Data", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			compressGzip(largeData)
		}
	})
	
	b.Run("LZ4 Large Data", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			compressLZ4(largeData)
		}
	})
}
