#!/bin/bash

# Claw Native Company 部署脚本
# 用于自动化部署 Claw 系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_info "依赖检查通过"
}

# 创建必要目录
setup_directories() {
    log_info "创建必要目录..."
    
    mkdir -p data
    mkdir -p logs
    mkdir -p backups
    
    log_info "目录创建完成"
}

# 生成环境配置
generate_env() {
    if [ ! -f .env ]; then
        log_info "生成环境配置文件..."
        
        JWT_SECRET=$(openssl rand -base64 32)
        
        cat > .env << EOF
# Claw 环境配置
ENV=production
PORT=8080

# 数据库配置
DATABASE_PATH=/data/claw.db

# JWT 配置
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRE_HOURS=24

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json

# WebSocket 配置
WS_MAX_CONNECTIONS=10000

# 钉钉配置（可选）
# DINGTALK_APP_KEY=
# DINGTALK_APP_SECRET=

# 飞书配置（可选）
# FEISHU_APP_ID=
# FEISHU_APP_SECRET=
EOF
        
        log_info "环境配置文件已生成: .env"
        log_warn "请检查并修改 .env 文件中的配置"
    else
        log_info "环境配置文件已存在，跳过生成"
    fi
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    docker-compose build --no-cache
    
    log_info "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    docker-compose up -d
    
    log_info "服务启动完成"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    max_attempts=30
    attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8080/api/v1/health > /dev/null; then
            log_info "后端服务已就绪"
            return 0
        fi
        
        log_info "等待服务就绪... ($attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "服务启动超时"
    return 1
}

# 显示服务状态
show_status() {
    log_info "服务状态:"
    docker-compose ps
    
    echo ""
    log_info "健康检查:"
    curl -s http://localhost:8080/api/v1/health | jq . 2>/dev/null || curl -s http://localhost:8080/api/v1/health
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker-compose down
    log_info "服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    docker-compose restart
    wait_for_services
    show_status
}

# 查看日志
show_logs() {
    if [ -z "$1" ]; then
        docker-compose logs -f
    else
        docker-compose logs -f "$1"
    fi
}

# 备份数据
backup_data() {
    log_info "备份数据..."
    
    BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    if [ -f data/claw.db ]; then
        cp data/claw.db "$BACKUP_DIR/"
        log_info "数据已备份到: $BACKUP_DIR"
    else
        log_warn "数据库文件不存在，跳过备份"
    fi
}

# 更新部署
update_deployment() {
    log_info "更新部署..."
    
    # 备份数据
    backup_data
    
    # 拉取最新代码（如果是 git 仓库）
    if [ -d .git ]; then
        log_info "拉取最新代码..."
        git pull
    fi
    
    # 重新构建并启动
    docker-compose down
    build_images
    start_services
    wait_for_services
    show_status
    
    log_info "更新完成"
}

# 清理资源
cleanup() {
    log_warn "清理资源..."
    
    read -p "确定要停止服务并清理所有数据吗？此操作不可恢复！[y/N] " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose down -v
        rm -rf data/*
        log_info "资源已清理"
    else
        log_info "取消清理操作"
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
Claw Native Company 部署脚本

用法: ./deploy.sh [命令]

命令:
    install     首次安装部署
    start       启动服务
    stop        停止服务
    restart     重启服务
    update      更新部署
    status      查看服务状态
    logs        查看日志 [service]
    backup      备份数据
    cleanup     清理所有资源（危险！）
    help        显示帮助信息

示例:
    ./deploy.sh install     # 首次安装
    ./deploy.sh logs backend # 查看后端日志
    ./deploy.sh update      # 更新到最新版本
EOF
}

# 主函数
main() {
    case "${1:-install}" in
        install)
            check_dependencies
            setup_directories
            generate_env
            build_images
            start_services
            wait_for_services
            show_status
            log_info "部署完成！"
            log_info "前端访问: http://localhost"
            log_info "后端 API: http://localhost:8080/api/v1"
            ;;
        start)
            start_services
            wait_for_services
            show_status
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        update)
            update_deployment
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        backup)
            backup_data
            ;;
        cleanup)
            cleanup
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
