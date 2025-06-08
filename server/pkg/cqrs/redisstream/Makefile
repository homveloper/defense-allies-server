# Redis Stream EventBus Makefile
# 개발 및 테스트 편의성을 위한 Makefile

.PHONY: help test test-integration test-benchmark test-coverage redis-start redis-stop clean deps lint

# Default target
help: ## 사용 가능한 명령어 표시
	@echo "Redis Stream EventBus Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Dependencies
deps: ## Go 의존성 설치
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Redis management
redis-start: ## Redis 컨테이너 시작
	@echo "🚀 Starting Redis container..."
	docker run -d --name redis-eventbus -p 6379:6379 redis:7-alpine
	@echo "✅ Redis started on localhost:6379"

redis-stop: ## Redis 컨테이너 중지 및 제거
	@echo "🛑 Stopping Redis container..."
	-docker stop redis-eventbus
	-docker rm redis-eventbus
	@echo "✅ Redis stopped and removed"

redis-restart: redis-stop redis-start ## Redis 컨테이너 재시작

# Testing
test: ## 단위 테스트 실행
	@echo "🧪 Running unit tests..."
	go test -v -short ./pkg/cqrs/redisstream/...

test-integration: ## 통합 테스트 실행 (Redis 필요)
	@echo "🔗 Running integration tests..."
	go test -v -tags=integration ./pkg/cqrs/redisstream/...

test-benchmark: ## 벤치마크 테스트 실행
	@echo "📊 Running benchmark tests..."
	go test -bench=. -benchmem ./pkg/cqrs/redisstream/...

test-coverage: ## 테스트 커버리지 확인
	@echo "📈 Generating test coverage report..."
	go test -cover -coverprofile=coverage.out ./pkg/cqrs/redisstream/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-all: test test-integration test-benchmark ## 모든 테스트 실행

# Code quality
lint: ## 코드 품질 검사
	@echo "🔍 Running linters..."
	go vet ./pkg/cqrs/redisstream/...
	go fmt ./pkg/cqrs/redisstream/...
	@echo "✅ Linting complete"

# Examples
example-basic: ## 기본 예제 실행
	@echo "🎯 Running basic example..."
	cd pkg/cqrs/redisstream/example && go run main.go

example-advanced: ## 고급 기능 데모 실행
	@echo "🚀 Running advanced demo..."
	cd pkg/cqrs/redisstream/example && go run advanced_demo.go

# Development workflow
dev-setup: deps redis-start ## 개발 환경 초기 설정
	@echo "🎉 Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run tests: make test"
	@echo "  2. Run examples: make example-basic"
	@echo "  3. Check coverage: make test-coverage"

dev-test: redis-start test redis-stop ## Redis 시작 → 테스트 → Redis 중지

# CI/CD targets
ci-test: deps test test-integration ## CI용 테스트 (Redis 외부 실행 가정)

# Cleanup
clean: ## 생성된 파일 정리
	@echo "🧹 Cleaning up..."
	-rm -f coverage.out coverage.html
	-docker stop redis-eventbus
	-docker rm redis-eventbus
	@echo "✅ Cleanup complete"

# Performance monitoring
perf-profile: ## 성능 프로파일링 실행
	@echo "⚡ Running performance profiling..."
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=BenchmarkEventBusPublishing ./pkg/cqrs/redisstream/
	@echo "📊 Profile files generated: cpu.prof, mem.prof"
	@echo "View with: go tool pprof cpu.prof"

# Documentation
docs: ## 문서 생성
	@echo "📚 Generating documentation..."
	go doc -all ./pkg/cqrs/redisstream/ > docs.txt
	@echo "✅ Documentation generated: docs.txt"

# Check system requirements
check-requirements: ## 시스템 요구사항 확인
	@echo "🔧 Checking system requirements..."
	@echo -n "Go version: "
	@go version || echo "❌ Go not installed"
	@echo -n "Docker version: "
	@docker --version || echo "❌ Docker not installed"
	@echo -n "Redis connectivity: "
	@redis-cli ping 2>/dev/null || echo "⚠️  Redis not running (run 'make redis-start')"
	@echo "✅ Requirements check complete"

# Full development cycle
full-cycle: clean dev-setup test-all test-coverage perf-profile ## 전체 개발 사이클 실행
	@echo ""
	@echo "🎉 Full development cycle completed!"
	@echo "📊 Coverage report: coverage.html"
	@echo "⚡ Performance profiles: cpu.prof, mem.prof"

# Quick start for new developers
quickstart: ## 새로운 개발자를 위한 빠른 시작
	@echo "🚀 QuickStart Guide for Redis Stream EventBus"
	@echo "============================================="
	@echo ""
	@echo "1. Install dependencies:"
	@echo "   make deps"
	@echo ""
	@echo "2. Start Redis:"
	@echo "   make redis-start"
	@echo ""
	@echo "3. Run tests:"
	@echo "   make test"
	@echo ""
	@echo "4. Try examples:"
	@echo "   make example-basic"
	@echo "   make example-advanced"
	@echo ""
	@echo "5. Check coverage:"
	@echo "   make test-coverage"
	@echo ""
	@echo "For full development setup: make dev-setup"
