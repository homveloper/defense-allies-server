# Defense Allies Server Makefile

# 변수 설정
BINARY_NAME=defense-allies-server
MAIN_PATH=./main.go
BUILD_DIR=./build

# 기본 타겟
.PHONY: all build run test clean help

all: build

# 빌드
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# 실행
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

# 테스트
test:
	@echo "Running tests..."
	go test -v ./...

# 테스트 커버리지
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 린트
lint:
	@echo "Running linter..."
	golangci-lint run

# 포맷팅
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 의존성 정리
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# 개발 환경 설정
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 클린업
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 도움말
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  tidy         - Tidy dependencies"
	@echo "  dev-setup    - Setup development environment"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help message"
