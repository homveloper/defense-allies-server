package timesquare

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"defense-allies-server/configs"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 연결 설정 Value Object
type RedisConfig struct {
	DefaultURL string
	Topics     map[string]string
}

// Validate Redis 설정 유효성 검사
func (r *RedisConfig) Validate() error {
	if r.DefaultURL == "" {
		return fmt.Errorf("default Redis URL is required")
	}
	return nil
}

// GetURL 토픽별 Redis URL 반환
func (r *RedisConfig) GetURL(topic string) string {
	if url, exists := r.Topics[topic]; exists {
		return url
	}
	return r.DefaultURL
}

// JWTConfig JWT 설정 Value Object
type JWTConfig struct {
	PublicKeyPEM string
}

// Validate JWT 설정 유효성 검사
func (j *JWTConfig) Validate() error {
	if j.PublicKeyPEM == "" {
		return fmt.Errorf("JWT public key is required")
	}

	// 공개키 파싱 테스트
	if _, err := j.ParsePublicKey(); err != nil {
		return fmt.Errorf("invalid JWT public key: %w", err)
	}

	return nil
}

// ParsePublicKey JWT 공개키를 파싱
func (j *JWTConfig) ParsePublicKey() (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(j.PublicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

// GuardianConfig Guardian 연결 설정 Value Object
type GuardianConfig struct {
	URL string
}

// Validate Guardian 설정 유효성 검사
func (g *GuardianConfig) Validate() error {
	if g.URL == "" {
		return fmt.Errorf("Guardian URL is required")
	}
	return nil
}

// Config TimeSquare 앱 설정 (Value Object 조합)
type Config struct {
	Redis    RedisConfig
	JWT      JWTConfig
	Guardian GuardianConfig
	Enabled  bool
}

// NewConfigFromFile 설정 파일에서 TimeSquare Config 생성
func NewConfigFromFile(configPath string) (*Config, error) {
	// 환경변수로 설정 파일 경로 오버라이드
	if configPath == "" {
		configPath = "configs/config.json"
	}

	// 전역 설정 로드
	globalConfig, err := configs.LoadConfigFromPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// TimeSquare 설정 추출
	tsConfig := &Config{
		Redis: RedisConfig{
			DefaultURL: globalConfig.GetRedisURL("default"),
			Topics:     globalConfig.Redis.Topics,
		},
		JWT: JWTConfig{
			PublicKeyPEM: globalConfig.TimeSquare.JWTPublicKey,
		},
		Guardian: GuardianConfig{
			URL: globalConfig.TimeSquare.GuardianURL,
		},
		Enabled: globalConfig.TimeSquare.Enabled,
	}

	return tsConfig, nil
}

// Validate TimeSquare 설정 유효성 검사
func (c *Config) Validate() error {
	if !c.Enabled {
		return fmt.Errorf("TimeSquare app is disabled")
	}

	if err := c.JWT.Validate(); err != nil {
		return fmt.Errorf("JWT config validation failed: %w", err)
	}

	if err := c.Guardian.Validate(); err != nil {
		return fmt.Errorf("Guardian config validation failed: %w", err)
	}

	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("Redis config validation failed: %w", err)
	}

	return nil
}

// ParsePublicKey JWT 공개키를 파싱 (위임)
func (c *Config) ParsePublicKey() (*rsa.PublicKey, error) {
	return c.JWT.ParsePublicKey()
}

// CreateRedisClient 기본 Redis 클라이언트를 생성합니다
func (c *Config) CreateRedisClient() (*redis.Client, error) {
	return c.CreateRedisClientForTopic("default")
}

// CreateRedisClientForTopic 특정 토픽용 Redis 클라이언트를 생성합니다
func (c *Config) CreateRedisClientForTopic(topic string) (*redis.Client, error) {
	url := c.Redis.GetURL(topic)

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL for topic %s: %w", topic, err)
	}

	client := redis.NewClient(opt)
	return client, nil
}

// GetGuardianURL Guardian 서버 URL 반환 (위임)
func (c *Config) GetGuardianURL() string {
	return c.Guardian.URL
}
