#!/bin/bash
# Claw Native Company - 数据备份脚本
# 用法: ./scripts/backup.sh [full|db|config]

set -e

# 配置
BACKUP_TYPE=${1:-full}
BACKUP_DIR="/opt/backups/claw-native-company"
DATA_DIR="./data"
CONFIG_FILE="./backend/config.yaml"
RETENTION_DAYS=30

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

# 创建备份目录
setup_backup_dir() {
    local date_str=$(date +%Y%m%d)
    local time_str=$(date +%H%M%S)
    BACKUP_PATH="${BACKUP_DIR}/${date_str}"
    BACKUP_NAME="claw-backup-${date_str}-${time_str}"
    
    mkdir -p "${BACKUP_PATH}"
    log_info "备份目录: ${BACKUP_PATH}"
}

# 备份数据库
backup_database() {
    log_info "备份数据库..."
    
    if [ -d "${DATA_DIR}" ]; then
        tar -czf "${BACKUP_PATH}/${BACKUP_NAME}-db.tar.gz" -C "$(dirname ${DATA_DIR})" $(basename ${DATA_DIR})
        log_info "数据库备份完成: ${BACKUP_PATH}/${BACKUP_NAME}-db.tar.gz"
    else
        log_warn "数据目录不存在: ${DATA_DIR}"
    fi
}

# 备份配置
backup_config() {
    log_info "备份配置文件..."
    
    if [ -f "${CONFIG_FILE}" ]; then
        cp "${CONFIG_FILE}" "${BACKUP_PATH}/${BACKUP_NAME}-config.yaml"
        log_info "配置备份完成: ${BACKUP_PATH}/${BACKUP_NAME}-config.yaml"
    else
        log_warn "配置文件不存在: ${CONFIG_FILE}"
    fi
    
    # 备份 docker-compose.yml
    if [ -f "docker-compose.yml" ]; then
        cp "docker-compose.yml" "${BACKUP_PATH}/${BACKUP_NAME}-docker-compose.yml"
        log_info "Docker Compose 配置备份完成"
    fi
}

# 备份日志
backup_logs() {
    log_info "备份日志..."
    
    if [ -d "./logs" ]; then
        tar -czf "${BACKUP_PATH}/${BACKUP_NAME}-logs.tar.gz" logs/ 2>/dev/null || true
        log_info "日志备份完成"
    fi
}

# 创建备份清单
create_manifest() {
    log_info "创建备份清单..."
    
    cat > "${BACKUP_PATH}/${BACKUP_NAME}-manifest.txt" << EOF
Claw Native Company Backup Manifest
====================================
备份时间: $(date '+%Y-%m-%d %H:%M:%S')
备份类型: ${BACKUP_TYPE}
主机名: $(hostname)
用户: $(whoami)

备份文件:
$(ls -lh "${BACKUP_PATH}")

磁盘使用情况:
$(df -h .)

数据目录大小:
$(du -sh ${DATA_DIR} 2>/dev/null || echo "N/A")
EOF
    
    log_info "备份清单创建完成"
}

# 上传到远程存储 (可选)
upload_to_remote() {
    log_info "检查远程存储配置..."
    
    # 如果配置了 S3 或其他远程存储，可以在这里实现上传逻辑
    # 例如: aws s3 cp "${BACKUP_PATH}" s3://your-bucket/backups/
    
    if [ -n "${S3_BUCKET}" ]; then
        log_info "上传到 S3: ${S3_BUCKET}"
        # aws s3 sync "${BACKUP_PATH}" "s3://${S3_BUCKET}/backups/$(basename ${BACKUP_PATH})/"
    fi
}

# 清理旧备份
cleanup_old_backups() {
    log_info "清理 ${RETENTION_DAYS} 天前的备份..."
    
    find "${BACKUP_DIR}" -type d -mtime +${RETENTION_DAYS} -exec rm -rf {} + 2>/dev/null || true
    
    local remaining=$(find "${BACKUP_DIR}" -type d -name "20*" | wc -l)
    log_info "剩余备份数量: ${remaining}"
}

# 显示备份信息
show_backup_info() {
    log_info "========================================"
    log_info "备份完成！"
    log_info "========================================"
    log_info "备份类型: ${BACKUP_TYPE}"
    log_info "备份路径: ${BACKUP_PATH}"
    log_info "备份大小: $(du -sh ${BACKUP_PATH} | cut -f1)"
    log_info "========================================"
    log_info "备份文件列表:"
    ls -lh "${BACKUP_PATH}"
    log_info "========================================"
}

# 全量备份
full_backup() {
    log_info "执行全量备份..."
    backup_database
    backup_config
    backup_logs
    create_manifest
}

# 仅数据库备份
db_backup() {
    log_info "执行数据库备份..."
    backup_database
    create_manifest
}

# 仅配置备份
config_backup() {
    log_info "执行配置备份..."
    backup_config
    create_manifest
}

# 主函数
main() {
    log_info "开始备份 Claw Native Company..."
    log_info "备份类型: ${BACKUP_TYPE}"
    
    setup_backup_dir
    
    case "${BACKUP_TYPE}" in
        full)
            full_backup
            ;;
        db)
            db_backup
            ;;
        config)
            config_backup
            ;;
        *)
            log_error "未知的备份类型: ${BACKUP_TYPE}"
            log_info "用法: $0 [full|db|config]"
            exit 1
            ;;
    esac
    
    upload_to_remote
    cleanup_old_backups
    show_backup_info
    
    log_info "备份成功！"
}

# 执行主函数
main "$@"
