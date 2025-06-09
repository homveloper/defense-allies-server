// encryption.go - 암호화 관련 인터페이스와 구현
package cqrsx

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor는 데이터 암호화/복호화를 담당하는 인터페이스입니다
type Encryptor interface {
	Encrypt(plaintext string) ([]byte, error)
	Decrypt(ciphertext string) (string, error)
}

// AESEncryptor는 AES-GCM을 사용한 암호화 구현체입니다
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor는 새로운 AES 암호화기를 생성합니다
func NewAESEncryptor(passphrase string) *AESEncryptor {
	// 패스프레이즈를 SHA256으로 해시하여 32바이트 키 생성
	hash := sha256.Sum256([]byte(passphrase))
	return &AESEncryptor{
		key: hash[:],
	}
}

// Encrypt는 평문을 암호화합니다
func (e *AESEncryptor) Encrypt(plaintext string) ([]byte, error) {
	if plaintext == "" {
		return nil, fmt.Errorf("plaintext cannot be empty")
	}

	// AES 블록 암호 생성
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// GCM 모드 생성
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// nonce 생성
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 암호화 수행
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Base64 인코딩하여 반환
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return []byte(encoded), nil
}

// Decrypt는 암호문을 복호화합니다
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", fmt.Errorf("ciphertext cannot be empty")
	}

	// Base64 디코딩
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// AES 블록 암호 생성
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// GCM 모드 생성
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// nonce 크기 확인
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// nonce와 암호문 분리
	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	// 복호화 수행
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// NoOpEncryptor는 암호화를 수행하지 않는 더미 구현체입니다 (테스트용)
type NoOpEncryptor struct{}

// NewNoOpEncryptor는 새로운 NoOp 암호화기를 생성합니다
func NewNoOpEncryptor() *NoOpEncryptor {
	return &NoOpEncryptor{}
}

// Encrypt는 데이터를 그대로 반환합니다
func (e *NoOpEncryptor) Encrypt(plaintext string) ([]byte, error) {
	return []byte(plaintext), nil
}

// Decrypt는 데이터를 그대로 반환합니다
func (e *NoOpEncryptor) Decrypt(ciphertext string) (string, error) {
	return ciphertext, nil
}
