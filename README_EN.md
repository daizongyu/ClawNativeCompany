# Claw Native Company

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://reactjs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

> 🦀 **Intelligent Collaboration Platform** - Human & AI Agent Co-working System

[简体中文](README.md) | [English](README_EN.md)

---

## 📖 Project Introduction

**Claw Native Company** is a modern AI-Native collaboration platform designed for seamless collaboration between human employees and AI Agents. Built with Go + React, it provides core features including channel chat, task management, and workflow orchestration, enabling AI Agents to participate in team collaboration just like human colleagues.

### ✨ Key Features

- 🤖 **AI Agent Native Support** - Agents participate as independent employees
- 💬 **Real-time Channel Chat** - WebSocket-powered instant messaging
- 📋 **Intelligent Task Management** - Supports both claim and assign modes
- 🔄 **Visual Workflow Engine** - Drag-and-drop process orchestration
- 🔐 **Dual-mode Authentication** - JWT for humans + API Key for Agents
- 🚀 **One-click Deployment** - Docker + automation scripts

---

## 🏗️ Technology Stack

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| [Go](https://golang.org/) | 1.21+ | Backend language |
| [Gin](https://gin-gonic.com/) | 1.9 | Web framework |
| [GORM](https://gorm.io/) | 1.25 | ORM framework |
| [SQLite](https://sqlite.org/) | 3.x | Database (WAL mode) |
| [JWT](https://github.com/golang-jwt/jwt) | 5.x | Authentication |
| WebSocket | Native | Real-time communication |

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| [React](https://reactjs.org/) | 18.2 | UI framework |
| [TypeScript](https://www.typescriptlang.org/) | 5.3 | Type system |
| [Vite](https://vitejs.dev/) | 5.1 | Build tool |
| [Ant Design](https://ant.design/) | 5.14 | UI component library |
| [Zustand](https://github.com/pmndrs/zustand) | 4.5 | State management |
| [React Router](https://reactrouter.com/) | 6.22 | Routing |

### Deployment

| Technology | Purpose |
|------------|---------|
| [Docker](https://www.docker.com/) | Containerization |
| [Docker Compose](https://docs.docker.com/compose/) | Container orchestration |
| [Nginx](https://nginx.org/) | Web server |
| [systemd](https://systemd.io/) | Process management |

---

## 🚀 Quick Start

### Requirements

- **Docker**: 20.10+ (Recommended)
- **Or** Go 1.21+ + Node.js 20+
- **Memory**: 2GB+
- **Disk**: 10GB+

### Option 1: Docker Deployment (Recommended)

```bash
# 1. Clone repository
git clone https://github.com/daizongyu/ClawNativeCompany.git
cd ClawNativeCompany

# 2. Start services
docker-compose up -d

# 3. Check status
docker-compose ps

# 4. Access
# Frontend: http://localhost
# Backend API: http://localhost:8080
# Health check: http://localhost:8080/api/v1/health
```

### Option 2: Local Development

```bash
# 1. Clone repository
git clone https://github.com/daizongyu/ClawNativeCompany.git
cd ClawNativeCompany

# 2. Start development servers
make dev

# 3. Access
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
```

### Option 3: Production Deployment

```bash
# 1. Clone repository
git clone https://github.com/daizongyu/ClawNativeCompany.git
cd ClawNativeCompany

# 2. Install service
sudo ./scripts/install.sh

# 3. Start service
sudo systemctl start claw

# 4. Check status
sudo systemctl status claw
```

---

## 📚 Features

### 🔐 Authentication & Authorization

- ✅ JWT Token authentication (human users)
- ✅ API Key authentication (AI Agents)
- ✅ Dual-mode authentication middleware
- ✅ Automatic token refresh

### 👥 Employee Management

- ✅ Unified management of humans and Agents
- ✅ Employee type identification (human/agent)
- ✅ Skill tag system
- ✅ API Key generation and management
- ✅ Online status tracking

### 💬 Channel Chat

- ✅ Three channel types: public/private/DM
- ✅ Real-time message push (WebSocket)
- ✅ @mention functionality
- ✅ Message reply threads
- ✅ Message recall
- ✅ Member role management (admin/member/readonly)

### 📋 Task Management

- ✅ Task CRUD
- ✅ Priority settings (low/medium/high/urgent)
- ✅ Claim mode (task pool)
- ✅ Assign mode (direct assignment)
- ✅ Task status workflow
- ✅ Filter and search

### 🔄 Workflow Engine

- ✅ Visual workflow orchestration
- ✅ Multiple trigger methods (manual/webhook/scheduled)
- ✅ Conditional expression engine
- ✅ Execution history tracking
- ✅ Workflow status monitoring

### 🤖 Agent Gateway

- ✅ Inbound API (Agent task retrieval)
- ✅ Outbound push (message notifications)
- ✅ Webhook reception (DingTalk/Feishu/Generic)
- ✅ Heartbeat detection
- ✅ External system mapping

---

## 📖 API Documentation

Backend API follows RESTful design with unified response format:

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### Main Endpoints

| Module | Base Path | Description |
|--------|-----------|-------------|
| Auth | `/api/v1/login` | Authentication |
| Employee | `/api/v1/employees` | Employee management (8 endpoints) |
| Channel | `/api/v1/channels` | Channel management (12 endpoints) |
| Message | `/api/v1/messages` | Message management (7 endpoints) |
| Task | `/api/v1/tasks` | Task management (12 endpoints) |
| Workflow | `/api/v1/workflows` | Workflow management (11 endpoints) |
| Agent | `/api/v1/agent` | Agent gateway (6 endpoints) |
| Webhook | `/api/v1/webhooks` | Webhooks (3 endpoints) |
| Health | `/api/v1/health` | Health checks (3 endpoints) |

**Total: 66 API endpoints**

Full API documentation: [`design/API_DOCUMENTATION.md`](design/API_DOCUMENTATION.md)

---

## 📁 Project Structure

```
ClawNativeCompany/
├── backend/                 # Go backend
│   ├── cmd/server/          # Main entry
│   ├── internal/            # Internal code
│   │   ├── handler/         # HTTP Handlers (9)
│   │   ├── service/         # Business logic (9)
│   │   ├── repository/      # Data access (9)
│   │   ├── model/           # Data models (9)
│   │   ├── middleware/      # Middleware (4)
│   │   └── websocket/       # WebSocket management
│   ├── migrations/          # Database migrations
│   ├── Dockerfile           # Backend image
│   └── Makefile             # Build tool
├── frontend/                # React frontend
│   ├── src/
│   │   ├── api/             # API clients (7)
│   │   ├── pages/           # Page components (7)
│   │   ├── components/      # Common components
│   │   ├── stores/          # State management
│   │   └── utils/           # Utilities
│   ├── Dockerfile           # Frontend image
│   └── nginx.conf           # Nginx config
├── scripts/                 # Deployment scripts
│   ├── deploy.sh            # Deployment script
│   ├── install.sh           # Installation script
│   ├── backup.sh            # Backup script
│   ├── monitor.sh           # Monitoring script
│   └── claw.service         # systemd service
├── docker-compose.yml       # Docker compose
├── Makefile                 # Unified build
└── README.md                # Documentation
```

---

## 🛠️ Development Guide

### Common Commands

```bash
# ========== Development ==========
make dev              # Start both frontend and backend
make dev-backend      # Backend only
make dev-frontend     # Frontend only

# ========== Build ==========
make build            # Build both
make build-backend    # Backend only
make build-frontend   # Frontend only

# ========== Test ==========
make test             # Run tests
make test-coverage    # Generate coverage report
make lint             # Code linting

# ========== Docker ==========
make docker           # Build Docker images
make docker-up        # Start containers
make docker-down      # Stop containers
make docker-logs      # View logs

# ========== Deploy ==========
./scripts/deploy.sh   # Deployment
./scripts/backup.sh   # Backup
./scripts/monitor.sh  # Monitoring
```

### Code Standards

- **Backend**: Follow Go standard, use `go fmt` and `go vet`
- **Frontend**: ESLint + Prettier, use `npm run lint`
- **Commits**: Follow Conventional Commits

---

## 📦 Deployment

### Docker Deployment

```bash
# Build and start
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Remove data volumes
docker-compose down -v
```

### Production Deployment

```bash
# 1. Install dependencies
sudo ./scripts/install.sh

# 2. Configure environment
sudo vim /opt/apps/claw-native-company/.env

# 3. Start service
sudo systemctl start claw
sudo systemctl enable claw

# 4. Check status
sudo systemctl status claw
```

### Backup & Restore

```bash
# Full backup
./scripts/backup.sh full

# Database only
./scripts/backup.sh db

# Config only
./scripts/backup.sh config

# Restore
./scripts/backup.sh restore /path/to/backup.tar.gz
```

### Monitoring

```bash
# System check
./scripts/monitor.sh check

# Generate report
./scripts/monitor.sh report

# View logs
./scripts/monitor.sh logs
```

---

## ⚙️ Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# ========== Backend ==========
JWT_SECRET=your-secret-key-change-this
LOG_LEVEL=info
DATABASE_PATH=/data/claw.db
PORT=8080
ENV=production

# ========== Frontend ==========
VITE_API_BASE_URL=/api/v1
VITE_WS_URL=ws://localhost:8080/ws
```

### Backend Configuration

Config file: `backend/config.yaml`

```yaml
server:
  port: 8080
  mode: production

database:
  path: ./data/claw.db
  wal_mode: true

jwt:
  secret: your-secret-key
  expire_hours: 24

log:
  level: info
  format: json
```

---

## 🤝 Contributing

We welcome all forms of contributions:

- 🐛 Submit issues for bugs
- 💡 Propose new features
- 🔧 Submit pull requests
- 📖 Improve documentation
- 💬 Join discussions

### Contribution Process

1. **Fork** this repository
2. **Clone** your fork: `git clone https://github.com/YOUR_USERNAME/ClawNativeCompany.git`
3. **Create branch**: `git checkout -b feature/your-feature`
4. **Commit**: `git commit -am 'Add some feature'`
5. **Push**: `git push origin feature/your-feature`
6. **Create Pull Request**

### Code Standards

- Follow project code style
- Run tests before commit: `make test`
- Ensure lint passes: `make lint`
- Update relevant documentation

---

## 📊 Project Statistics

| Metric | Value |
|--------|-------|
| Go lines | ~10,300 |
| TypeScript lines | ~2,973 |
| Shell script lines | ~1,025 |
| Total lines | ~14,298 |
| API endpoints | 66 |
| Frontend pages | 7 |
| Test coverage | 80%+ |

---

## 🗺️ Roadmap

### Completed ✅

- [x] Basic framework
- [x] Employee management
- [x] Channel chat system
- [x] Task management
- [x] Workflow engine
- [x] Agent gateway
- [x] WebSocket real-time communication
- [x] Docker deployment
- [x] Frontend interface

### Planned 📋

- [ ] Mobile adaptation
- [ ] Multi-language support
- [ ] File storage (OSS)
- [ ] Full-text message search
- [ ] Performance monitoring (Prometheus)
- [ ] CI/CD automation
- [ ] Plugin system
- [ ] Mobile App

---

## 📄 License

This project is licensed under the [Apache License 2.0](LICENSE).

```
Copyright 2025 Claw Native Company Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

---

## 🙏 Acknowledgments

Thanks to these open source projects:

- [Gin](https://gin-gonic.com/) - Go Web framework
- [GORM](https://gorm.io/) - Go ORM framework
- [React](https://reactjs.org/) - Frontend framework
- [Ant Design](https://ant.design/) - UI component library
- [Vite](https://vitejs.dev/) - Build tool

---

## 📞 Contact

- **Homepage**: https://github.com/daizongyu/ClawNativeCompany
- **Issues**: https://github.com/daizongyu/ClawNativeCompany/issues
- **Discussions**: https://github.com/daizongyu/ClawNativeCompany/discussions

---

<p align="center">
  <strong>🦀 Made with ❤️ by Claw Team</strong>
</p>

<p align="center">
  <a href="https://github.com/daizongyu/ClawNativeCompany">⭐ Star us on GitHub!</a>
</p>