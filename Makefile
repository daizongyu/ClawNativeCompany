# Claw Native Company - Makefile
# 统一构建前后端

.PHONY: all build build-backend build-frontend dev clean test lint docker

# 默认目标
all: build

# 构建全部
build: build-backend build-frontend
	@echo "✅ Build completed"

# 构建后端
build-backend:
	@echo "🔨 Building backend..."
	cd backend && $(MAKE) build-linux

# 构建前端
build-frontend:
	@echo "🔨 Building frontend..."
	cd frontend && npm install && npm run build

# 开发模式（同时启动前后端）
dev:
	@echo "🚀 Starting development servers..."
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:3000"
	@trap 'kill %1 %2' EXIT; \
	cd backend && go run cmd/server/main.go & \
	cd frontend && npm run dev & \
	wait

# 启动后端
dev-backend:
	@echo "🚀 Starting backend server..."
	cd backend && go run cmd/server/main.go

# 启动前端
dev-frontend:
	@echo "🚀 Starting frontend server..."
	cd frontend && npm run dev

# 运行测试
test:
	@echo "🧪 Running tests..."
	cd backend && $(MAKE) test
	cd frontend && npm run lint

# 清理构建产物
clean:
	@echo "🧹 Cleaning..."
	cd backend && $(MAKE) clean
	cd frontend && rm -rf dist node_modules
	rm -rf data/*.db

# 代码检查
lint:
	@echo "🔍 Running lint..."
	cd backend && go vet ./...
	cd frontend && npm run lint

# Docker 构建
docker:
	@echo "🐳 Building Docker images..."
	docker-compose build

# Docker 启动
docker-up:
	@echo "🐳 Starting Docker containers..."
	docker-compose up -d

# Docker 停止
docker-down:
	@echo "🛑 Stopping Docker containers..."
	docker-compose down

# 查看日志
logs:
	docker-compose logs -f

# 数据库迁移
migrate:
	@echo "🗄️ Running database migrations..."
	cd backend && go run cmd/server/main.go migrate

# 初始化项目（首次使用）
init:
	@echo "🚀 Initializing project..."
	cd frontend && npm install
	cd backend && go mod download
	@echo "✅ Initialization completed"
	@echo "Run 'make dev' to start development"

# 帮助
help:
	@echo "Available targets:"
	@echo "  make build          - Build backend and frontend"
	@echo "  make build-backend  - Build backend only"
	@echo "  make build-frontend - Build frontend only"
	@echo "  make dev            - Start development servers"
	@echo "  make dev-backend    - Start backend only"
	@echo "  make dev-frontend   - Start frontend only"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run lint checks"
	@echo "  make docker         - Build Docker images"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make logs           - View Docker logs"
	@echo "  make migrate        - Run database migrations"
	@echo "  make init           - Initialize project"
