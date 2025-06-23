package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config 전체 애플리케이션 설정
type Config struct {
	Server     ServerConfig     `json:"server"`
	Redis      RedisConfig      `json:"redis"`
	TimeSquare TimeSquareConfig `json:"timesquare"`
	Guardian   GuardianConfig   `json:"guardian"`
	Logging    LoggingConfig    `json:"logging"`
}

// ServerConfig 공용 서버 설정
type ServerConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	GracefulTimeout int    `json:"graceful_timeout"`
}

// RedisConfig Redis 토픽별 URL 설정
type RedisConfig struct {
	Default string            `json:"default"`
	Topics  map[string]string `json:"topics"`
}

// LoggingConfig 로깅 설정
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// TimeSquareConfig TimeSquare 앱 설정
type TimeSquareConfig struct {
	Enabled      bool   `json:"enabled"`
	GuardianURL  string `json:"guardian_url"`
	JWTPublicKey string `json:"jwt_public_key"`
}

// GuardianConfig Guardian 앱 설정
type GuardianConfig struct {
	Enabled       bool   `json:"enabled"`
	MongoDBURI    string `json:"mongodb_uri"`
	JWTPrivateKey string `json:"jwt_private_key"`
	JWTPublicKey  string `json:"jwt_public_key"`
}

// LoadConfig JSON 파일과 환경변수에서 설정을 로드합니다
func LoadConfig() (*Config, error) {
	return LoadConfigFromPath("")
}

// LoadConfigFromPath 지정된 경로에서 설정을 로드합니다
func LoadConfigFromPath(configPath string) (*Config, error) {
	// 1. JSON 파일에서 기본 설정 로드
	config, err := loadFromJSON(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from JSON: %w", err)
	}

	// 2. 환경변수로 오버라이드
	overrideWithEnv(config)

	return config, nil
}

// loadFromJSON JSON 파일에서 설정 로드
func loadFromJSON(configPath string) (*Config, error) {
	// 설정 파일 경로 결정
	if configPath == "" {
		configPath = getEnv("CONFIG_PATH", "configs/config.json")
	}
	
	// 상대 경로를 절대 경로로 변환
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		configPath = filepath.Join(wd, configPath)
	}

	// 파일 읽기
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// JSON 파싱
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// overrideWithEnv 환경변수로 설정 오버라이드
func overrideWithEnv(config *Config) {
	// Server 설정
	if host := getEnv("SERVER_HOST", ""); host != "" {
		config.Server.Host = host
	}
	if port := getEnvAsInt("SERVER_PORT", 0); port != 0 {
		config.Server.Port = port
	}
	if timeout := getEnvAsInt("SERVER_GRACEFUL_TIMEOUT", 0); timeout != 0 {
		config.Server.GracefulTimeout = timeout
	}

	// Redis 설정
	if defaultRedis := getEnv("REDIS_DEFAULT", ""); defaultRedis != "" {
		config.Redis.Default = defaultRedis
	}
	
	// Redis 토픽별 URL 오버라이드
	for topic := range config.Redis.Topics {
		envKey := fmt.Sprintf("REDIS_%s", topic)
		if url := getEnv(envKey, ""); url != "" {
			config.Redis.Topics[topic] = url
		}
	}

	// TimeSquare 설정
	if guardianURL := getEnv("TIMESQUARE_GUARDIAN_URL", ""); guardianURL != "" {
		config.TimeSquare.GuardianURL = guardianURL
	}
	if jwtPublicKey := getEnv("TIMESQUARE_JWT_PUBLIC_KEY", ""); jwtPublicKey != "" {
		config.TimeSquare.JWTPublicKey = jwtPublicKey
	}
	if enabled := getEnv("TIMESQUARE_ENABLED", ""); enabled != "" {
		config.TimeSquare.Enabled = enabled == "true"
	}

	// Guardian 설정
	if mongoURI := getEnv("GUARDIAN_MONGODB_URI", ""); mongoURI != "" {
		config.Guardian.MongoDBURI = mongoURI
	}
	if privateKey := getEnv("GUARDIAN_JWT_PRIVATE_KEY", ""); privateKey != "" {
		config.Guardian.JWTPrivateKey = privateKey
	}
	if publicKey := getEnv("GUARDIAN_JWT_PUBLIC_KEY", ""); publicKey != "" {
		config.Guardian.JWTPublicKey = publicKey
	}
	if enabled := getEnv("GUARDIAN_ENABLED", ""); enabled != "" {
		config.Guardian.Enabled = enabled == "true"
	}

	// Logging 설정
	if level := getEnv("LOG_LEVEL", ""); level != "" {
		config.Logging.Level = level
	}
	if format := getEnv("LOG_FORMAT", ""); format != "" {
		config.Logging.Format = format
	}
}


// GetRedisURL 토픽별 Redis URL 반환
func (c *Config) GetRedisURL(topic string) string {
	if url, exists := c.Redis.Topics[topic]; exists {
		return url
	}
	return c.Redis.Default
}

// getEnv 환경변수를 가져오거나 기본값을 반환합니다
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 환경변수를 정수로 가져오거나 기본값을 반환합니다
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
