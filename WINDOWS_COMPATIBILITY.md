# Windows 兼容性说明

> 版本：v1.0
> 更新日期：2025-03-29

## 概述

本项目已针对 Windows 平台进行兼容性优化，使用纯 Go 实现的 SQLite 驱动，**无需安装 GCC** 即可在 Windows 上编译和运行。

## 技术方案

### 驱动替换

| 项目 | 原方案 | 新方案 |
|------|--------|--------|
| SQLite 驱动 | `gorm.io/driver/sqlite` | `github.com/glebarez/sqlite` |
| 实现方式 | CGO + C 库 | 纯 Go |
| Windows 依赖 | 需要 GCC | 无需 GCC |

### 纯 Go SQLite 优势

1. **零依赖**：Windows 用户无需安装 MinGW-w64
2. **跨平台**：一套代码支持 Windows/Linux/Mac
3. **向后兼容**：现有数据和功能完全兼容
4. **易于分发**：单二进制文件，无需额外 DLL

## 快速开始

### Windows 用户

```powershell
# 1. 克隆项目
git clone https://github.com/daizongyu/ClawNativeCompany.git
cd ClawNativeCompany

# 2. 进入后端目录
cd backend

# 3. 直接构建（无需 GCC！）
go build -o bin/server.exe ./cmd/server/main.go

# 4. 运行
.\bin\server.exe
```

### 使用 Makefile

```powershell
# 构建 Windows 版本
make build-windows

# 或从根目录
make build-backend-windows
```

## 验证

### 验证构建

```powershell
# 检查生成的可执行文件
ls bin/
# server.exe  <- Windows 可执行文件
```

### 验证运行

```powershell
# 运行服务
.\bin\server.exe

# 预期输出：
# {"level":"INFO","message":"服务启动中",...}
# {"level":"INFO","message":"数据库连接成功",...}
# {"level":"INFO","message":"HTTP server started",...}
```

## 常见问题

### Q: 为什么之前需要 GCC？

A: 之前的 SQLite 驱动 `gorm.io/driver/sqlite` 底层使用 `github.com/mattn/go-sqlite3`，这是一个 CGO 绑定，需要调用 C 代码，因此需要 GCC 编译器。

### Q: 纯 Go 版本性能如何？

A: `github.com/glebarez/sqlite` 基于 `modernc.org/sqlite`，性能与 CGO 版本相当，对于大多数应用场景完全满足需求。

### Q: 数据兼容性如何？

A: 完全兼容。两个驱动都使用标准的 SQLite 数据库文件格式，可以无缝切换，原有数据无需迁移。

### Q: 是否支持所有 SQLite 功能？

A: 支持绝大多数常用功能，包括：
- 标准 SQL 语法
- 事务
- 索引
- WAL 模式
- 外键约束

### Q: 如何切换回 CGO 版本？

A: 如需切换回 CGO 版本，修改 `backend/internal/database/sqlite.go`：

```go
// 改为
import "gorm.io/driver/sqlite"
```

然后更新 go.mod：

```bash
go get gorm.io/driver/sqlite
go mod edit -droprequire=github.com/glebarez/sqlite
go mod tidy
```

## 构建矩阵

| 平台 | 命令 | 输出 |
|------|------|------|
| Windows | `make build-windows` | `bin/server.exe` |
| Linux | `make build-linux` | `bin/server-linux` |
| macOS | `go build -o bin/server-darwin ./cmd/server/main.go` | `bin/server-darwin` |

## 相关文件

- `backend/internal/database/sqlite.go` - 数据库连接配置
- `backend/go.mod` - Go 依赖管理
- `backend/Makefile` - 构建脚本

## 参考

- [glebarez/sqlite](https://github.com/glebarez/sqlite) - 纯 Go SQLite 驱动
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - 底层实现
