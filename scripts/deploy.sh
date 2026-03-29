#!/bin/bash
# Claw Native Company - 部署脚本
# 用法: ./scripts/deploy.sh [production|staging]

set -e

# 配置
DEPLOY_ENV=${1:-production}
PROJECT_NAME="claw-native-company"
BACKUP_DIR="/opt/backups/${PROJECT_NAME}"
DEPLOY_DIR="/opt/apps/${PROJECT_NAME}"
DOCKER_COMPOSE_FILE="docker-compose.yml"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
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
        log_error "Docker 未安装"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    log_info "依赖检查通过"
}

# 创建目录
setup_directories() {
    log_info "创建目录..."
    mkdir -p "${BACKUP_DIR}"
    mkdir -p "${DEPLOY_DIR}"
    mkdir -p "${DEPLOY_DIR}/data"
    log_info "目录创建完成"
}

# 备份数据
backup_data() {
    log_info "备份数据..."
    
    if [ -d "${DEPLOY_DIR}/data" ]; then
        BACKUP_FILE="${BACKUP_DIR}/backup-$(date +%Y%m%d-%H%M%S).tar.gz"
        tar -czf "${BACKUP_FILE}" -C "${DEPLOY_DIR}" data/ 2>/dev/null || true
        log_info "数据已备份到: ${BACKUP_FILE}"
    else
        log_warn "没有找到数据目录，跳过备份"
    fi
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    docker-compose -f "${DOCKER_COMPOSE_FILE}" build --no-cache
    log_info "镜像构建完成"
}

# 部署应用
deploy_app() {
    log_info "部署应用到 ${DEPLOY_ENV} 环境..."
    
    # 停止旧容器
    log_info "停止旧容器..."
    docker-compose -f "${DOCKER_COMPOSE_FILE}" down || true
    
    # 启动新容器
    log_info "启动新容器..."
    docker-compose -f "${DOCKER_COMPOSE_FILE}" up -d
    
    log_info "部署完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待服务启动
    sleep 5
    
    # 检查后端服务
    for i in {1..10}; do
        if curl -s http://localhost:8080/api/v1/health > /dev/null; then
            log_info "后端服务健康检查通过"
            break
        fi
        
        if [ $i -eq 10 ]; then
            log_error "后端服务健康检查失败"
            return 1
        fi
        
        log_warn "等待后端服务启动... (${i}/10)"
        sleep 3
    done
    
    # 检查前端服务
    for i in {1..10}; do
        if curl -s http://localhost:80 > /dev/null; then
            log_info "前端服务健康检查通过"
            break
        fi
        
        if [ $i -eq 10 ]; then
            log_error "前端服务健康检查失败"
            return 1
        fi
        
        log_warn "等待前端服务启动... (${i}/10)"
        sleep 3
    done
    
    log_info "所有服务健康检查通过"
}

# 清理旧备份
cleanup_old_backups() {
    log_info "清理旧备份..."
    
    # 保留最近 7 天的备份
    find "${BACKUP_DIR}" -name "backup-*.tar.gz" -mtime +7 -delete 2>/dev/null || true
    
    log_info "旧备份清理完成"
}

# 显示部署信息
show_info() {
    log_info "========================================"
    log_info "部署完成！"
    log_info "========================================"
    log_info "环境: ${DEPLOY_ENV}"
    log_info "前端地址: http://localhost"
    log_info "后端地址: http://localhost:8080"
    log_info "========================================"
    log_info "常用命令:"
    log_info "  查看日志: docker-compose logs -f"
    log_info "  停止服务: docker-compose down"
    log_info "  重启服务: docker-compose restart"
    log_info "========================================"
}

# 主函数
main() {
    log_info "开始部署 Claw Native Company..."
    log_info "环境: ${DEPLOY_ENV}"
    
    check_dependencies
    setup_directories
    backup_data
    build_images
    deploy_app
    health_check
    cleanup_old_backups
    show_info
    
    log_info "部署成功！"
}

# 执行主函数
main "$@"
