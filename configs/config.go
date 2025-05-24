package configs

import (
	"os"
	"strconv"
)

// Config 구조체는 애플리케이션 설정을 담습니다
type Config struct {
	Server ServerConfig
	Redis  RedisConfig
}

// ServerConfig 서버 관련 설정
type ServerConfig struct {
	Port string
	Host string
}

// RedisConfig Redis 관련 설정
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// LoadConfig 환경변수에서 설정을 로드합니다
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
	}
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
