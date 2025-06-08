// encryption_test.go - 암호화 기능 테스트
package cqrsx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESEncryptor_EncryptDecrypt(t *testing.T) {
	// Given
	encryptor := NewAESEncryptor("test-passphrase-123")
	
	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "Simple text",
			plaintext: "hello world",
		},
		{
			name:      "JSON data",
			plaintext: `{"name": "test", "value": 123, "sensitive": "secret"}`,
		},
		{
			name:      "Large text",
			plaintext: string(make([]byte, 10000)), // 10KB 데이터
		},
		{
			name:      "Unicode text",
			plaintext: "안녕하세요 世界 🌍",
		},
		{
			name:      "Special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encrypted)

			decrypted, err := encryptor.Decrypt(string(encrypted))
			require.NoError(t, err)

			// Then
			assert.Equal(t, tt.plaintext, decrypted)
			
			// 암호화된 데이터는 원본과 달라야 함
			assert.NotEqual(t, tt.plaintext, string(encrypted))
		})
	}
}

func TestAESEncryptor_ErrorCases(t *testing.T) {
	encryptor := NewAESEncryptor("test-passphrase")
	
	t.Run("Empty plaintext", func(t *testing.T) {
		_, err := encryptor.Encrypt("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plaintext cannot be empty")
	})
	
	t.Run("Empty ciphertext", func(t *testing.T) {
		_, err := encryptor.Decrypt("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ciphertext cannot be empty")
	})
	
	t.Run("Invalid ciphertext", func(t *testing.T) {
		_, err := encryptor.Decrypt("invalid-base64")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode base64")
	})
	
	t.Run("Corrupted ciphertext", func(t *testing.T) {
		// 유효한 base64이지만 잘못된 암호화 데이터
		_, err := encryptor.Decrypt("dGVzdA==") // "test"의 base64
		assert.Error(t, err)
	})
}

func TestAESEncryptor_DifferentKeys(t *testing.T) {
	// Given
	encryptor1 := NewAESEncryptor("passphrase1")
	encryptor2 := NewAESEncryptor("passphrase2")
	plaintext := "secret message"
	
	// When
	encrypted, err := encryptor1.Encrypt(plaintext)
	require.NoError(t, err)
	
	// Then - 다른 키로는 복호화할 수 없어야 함
	_, err = encryptor2.Decrypt(string(encrypted))
	assert.Error(t, err)
}

func TestAESEncryptor_Consistency(t *testing.T) {
	// Given
	encryptor := NewAESEncryptor("consistent-key")
	plaintext := "consistent data"
	
	// When - 같은 데이터를 여러 번 암호화
	encrypted1, err1 := encryptor.Encrypt(plaintext)
	encrypted2, err2 := encryptor.Encrypt(plaintext)
	
	// Then
	require.NoError(t, err1)
	require.NoError(t, err2)
	
	// 같은 평문이라도 매번 다른 암호문이 생성되어야 함 (nonce 때문에)
	assert.NotEqual(t, string(encrypted1), string(encrypted2))
	
	// 하지만 둘 다 올바르게 복호화되어야 함
	decrypted1, err := encryptor.Decrypt(string(encrypted1))
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)
	
	decrypted2, err := encryptor.Decrypt(string(encrypted2))
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestNoOpEncryptor(t *testing.T) {
	// Given
	encryptor := NewNoOpEncryptor()
	plaintext := "no encryption needed"
	
	// When
	encrypted, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	
	decrypted, err := encryptor.Decrypt(string(encrypted))
	require.NoError(t, err)
	
	// Then - NoOp이므로 원본과 같아야 함
	assert.Equal(t, plaintext, string(encrypted))
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptorInterface(t *testing.T) {
	// 인터페이스 구현 확인
	var encryptor Encryptor
	
	encryptor = NewAESEncryptor("test")
	assert.NotNil(t, encryptor)
	
	encryptor = NewNoOpEncryptor()
	assert.NotNil(t, encryptor)
}

func BenchmarkAESEncryption(b *testing.B) {
	encryptor := NewAESEncryptor("benchmark-key")
	
	testData := []struct {
		name string
		data string
	}{
		{"Small", "small test data"},
		{"Medium", string(make([]byte, 1024))},   // 1KB
		{"Large", string(make([]byte, 10240))},   // 10KB
	}
	
	for _, td := range testData {
		b.Run("Encrypt_"+td.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				encryptor.Encrypt(td.data)
			}
		})
		
		// 복호화 벤치마크를 위해 먼저 암호화
		encrypted, _ := encryptor.Encrypt(td.data)
		b.Run("Decrypt_"+td.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				encryptor.Decrypt(string(encrypted))
			}
		})
	}
}

func BenchmarkEncryptorComparison(b *testing.B) {
	aesEncryptor := NewAESEncryptor("benchmark-key")
	noOpEncryptor := NewNoOpEncryptor()
	testData := "benchmark test data for comparison"
	
	b.Run("AES Encrypt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			aesEncryptor.Encrypt(testData)
		}
	})
	
	b.Run("NoOp Encrypt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			noOpEncryptor.Encrypt(testData)
		}
	})
}
