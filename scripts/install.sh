#!/bin/bash
# Claw Native Company - 安装脚本
# 用法: sudo ./scripts/install.sh

set -e

# 配置
INSTALL_DIR="/opt/apps/claw-native-company"
SERVICE_NAME="claw"
USER="claw"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 sudo 运行此脚本"
        exit 1
    fi
}

# 检查系统要求
check_system() {
    log_info "检查系统要求..."
    
    # 检查操作系统
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        log_info "操作系统: $NAME $VERSION"
    fi
    
    # 检查内存
    local mem=$(free -m | awk 'NR==2{printf "%.0f", $2/1024}')
    log_info "内存: ${mem}GB"
    
    if [ "$mem" -lt 2 ]; then
        log_warn "建议至少 2GB 内存"
    fi
    
    # 检查磁盘空间
    local disk=$(df -BG . | awk 'NR==2{print $4}' | sed 's/G//')
    log_info "可用磁盘空间: ${disk}GB"
    
    if [ "$disk" -lt 10 ]; then
        log_warn "建议至少 10GB 可用磁盘空间"
    fi
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        log_info "请安装 Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        log_info "请安装 Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    log_info "Docker 版本: $(docker --version)"
    log_info "Docker Compose 版本: $(docker-compose --version)"
}

# 创建用户
create_user() {
    log_info "创建用户..."
    
    if ! id "$USER" &>/dev/null; then
        useradd -r -s /bin/false "$USER"
        log_info "用户 $USER 已创建"
    else
        log_info "用户 $USER 已存在"
    fi
}

# 创建目录
setup_directories() {
    log_info "创建安装目录..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/data"
    mkdir -p "$INSTALL_DIR/logs"
    mkdir -p "/opt/backups/claw-native-company"
    
    chown -R "$USER:$USER" "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR"
}

# 复制文件
copy_files() {
    log_info "复制文件..."
    
    # 复制项目文件
    cp -r backend "$INSTALL_DIR/"
    cp -r frontend "$INSTALL_DIR/"
    cp docker-compose.yml "$INSTALL_DIR/"
    cp -r scripts "$INSTALL_DIR/"
    
    chown -R "$USER:$USER" "$INSTALL_DIR"
    log_info "文件复制完成"
}

# 创建环境文件
create_env_file() {
    log_info "创建环境配置文件..."
    
    local env_file="$INSTALL_DIR/.env"
    
    # 生成随机 JWT 密钥
    local jwt_secret=$(openssl rand -base64 32)
    
    cat > "$env_file" << EOF
# Claw Native Company - 环境配置
COMPOSE_PROJECT_NAME=claw

# 后端配置
JWT_SECRET=${jwt_secret}
LOG_LEVEL=info

# 数据库配置
DATABASE_PATH=/data/claw.db

# 前端配置
NODE_ENV=production
EOF
    
    chmod 600 "$env_file"
    chown "$USER:$USER" "$env_file"
    
    log_info "环境配置文件已创建"
}

# 安装系统服务
install_service() {
    log_info "安装系统服务..."
    
    # 复制服务文件
    cp scripts/claw.service "/etc/systemd/system/${SERVICE_NAME}.service"
    
    # 更新服务文件中的路径
    sed -i "s|/opt/apps/claw-native-company|$INSTALL_DIR|g" "/etc/systemd/system/${SERVICE_NAME}.service"
    
    # 重新加载 systemd
    systemctl daemon-reload
    
    # 启用服务
    systemctl enable "$SERVICE_NAME"
    
    log_info "系统服务已安装"
}

# 配置防火墙
configure_firewall() {
    log_info "配置防火墙..."
    
    # 检查是否有防火墙
    if command -v ufw &> /dev/null; then
        ufw allow 80/tcp
        ufw allow 8080/tcp
        log_info "UFW 防火墙规则已添加"
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port=80/tcp
        firewall-cmd --permanent --add-port=8080/tcp
        firewall-cmd --reload
        log_info "Firewalld 防火墙规则已添加"
    else
        log_warn "未检测到支持的防火墙"
    fi
}

# 构建 Docker 镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    cd "$INSTALL_DIR"
    docker-compose build
    
    log_info "Docker 镜像构建完成"
}

# 启动服务
start_service() {
    log_info "启动服务..."
    
    cd "$INSTALL_DIR"
    docker-compose up -d
    
    # 等待服务启动
    sleep 5
    
    # 检查服务状态
    if docker-compose ps | grep -q "Up"; then
        log_info "服务已启动"
    else
        log_error "服务启动失败"
        docker-compose logs
        exit 1
    fi
}

# 显示安装信息
show_install_info() {
    log_info "========================================"
    log_info "安装完成！"
    log_info "========================================"
    log_info "安装目录: $INSTALL_DIR"
    log_info "数据目录: $INSTALL_DIR/data"
    log_info "日志目录: $INSTALL_DIR/logs"
    log_info "========================================"
    log_info "访问地址:"
    log_info "  前端: http://localhost"
    log_info "  后端: http://localhost:8080"
    log_info "========================================"
    log_info "常用命令:"
    log_info "  查看状态: systemctl status $SERVICE_NAME"
    log_info "  启动服务: systemctl start $SERVICE_NAME"
    log_info "  停止服务: systemctl stop $SERVICE_NAME"
    log_info "  重启服务: systemctl restart $SERVICE_NAME"
    log_info "  查看日志: docker-compose -f $INSTALL_DIR/docker-compose.yml logs -f"
    log_info "========================================"
    log_info "备份命令:"
    log_info "  $INSTALL_DIR/scripts/backup.sh full"
    log_info "========================================"
}

# 主函数
main() {
    log_info "开始安装 Claw Native Company..."
    
    check_root
    check_system
    check_dependencies
    create_user
    setup_directories
    copy_files
    create_env_file
    install_service
    configure_firewall
    build_images
    start_service
    show_install_info
    
    log_info "安装成功！"
}

# 执行主函数
main "$@"
