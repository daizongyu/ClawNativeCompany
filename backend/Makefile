# Claw Backend Makefile

.PHONY: build build-linux test clean run dev lint migrate

# 变量
BINARY_NAME=server
BUILD_DIR=bin
MAIN_FILE=cmd/server/main.go
DOCKER_IMAGE=claw-backend

# 默认目标
all: build

# 构建本地二进制
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建 Linux 二进制（用于 Docker）
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux"

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./internal/... -count=1

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# 运行开发服务器
dev:
	@echo "Starting development server..."
	go run $(MAIN_FILE)

# 运行生产服务器（需要先构建）
run: build
	@echo "Starting server..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# 代码格式化
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet..."; \
		go vet ./...; \
	fi

# 依赖检查
deps:
	@echo "Checking dependencies..."
	go mod tidy
	go mod verify

# 构建 Docker 镜像
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .

# 运行 Docker 容器
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		--name claw-backend \
		-p 8080:8080 \
		-v $(PWD)/data:/data \
		$(DOCKER_IMAGE):latest

# 停止 Docker 容器
docker-stop:
	@echo "Stopping Docker container..."
	docker stop claw-backend || true
	docker rm claw-backend || true

# 数据库迁移（简化版，实际应该使用 migrate 工具）
migrate:
	@echo "Running database migrations..."
	@echo "Note: Migrations are auto-run on server start"

# 生成 Swagger 文档（如果需要）
swagger:
	@echo "Generating Swagger docs..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g $(MAIN_FILE); \
	else \
		echo "swag not installed. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-linux    - Build Linux binary for Docker"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev            - Run development server"
	@echo "  run            - Build and run production server"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  deps           - Check and tidy dependencies"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-stop    - Stop Docker container"
	@echo "  migrate        - Run database migrations"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  help           - Show this help message"
