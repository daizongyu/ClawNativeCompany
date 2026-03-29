#!/bin/bash
# Claw Native Company - 监控脚本
# 用法: ./scripts/monitor.sh [check|report|alert]

set -e

# 配置
MONITOR_MODE=${1:-check}
WEBHOOK_URL="${ALERT_WEBHOOK_URL:-}"  # 告警 webhook URL
EMAIL="${ALERT_EMAIL:-}"              # 告警邮箱
THRESHOLD_CPU=80                      # CPU 使用率阈值
THRESHOLD_MEM=80                      # 内存使用率阈值
THRESHOLD_DISK=90                     # 磁盘使用率阈值
LOG_FILE="./logs/monitor.log"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1" >> "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [WARN] $1" >> "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1" >> "$LOG_FILE"
}

# 检查 Docker 容器
check_containers() {
    log_info "检查 Docker 容器状态..."
    
    local containers=$(docker-compose ps -q)
    local failed=0
    
    for container in $containers; do
        local name=$(docker inspect --format='{{.Name}}' "$container" | sed 's/\///')
        local status=$(docker inspect --format='{{.State.Status}}' "$container")
        local health=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "N/A")
        
        if [ "$status" != "running" ]; then
            log_error "容器 $name 状态异常: $status"
            failed=$((failed + 1))
        elif [ "$health" != "N/A" ] && [ "$health" != "healthy" ]; then
            log_warn "容器 $name 健康状态: $health"
        else
            log_info "容器 $name: $status (健康: $health)"
        fi
    done
    
    return $failed
}

# 检查系统资源
check_resources() {
    log_info "检查系统资源..."
    
    local alerts=()
    
    # CPU 使用率
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
    cpu_usage=${cpu_usage%.*}  # 取整
    log_info "CPU 使用率: ${cpu_usage}%"
    
    if [ "$cpu_usage" -gt "$THRESHOLD_CPU" ]; then
        alerts+=("CPU 使用率过高: ${cpu_usage}%")
    fi
    
    # 内存使用率
    local mem_usage=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
    log_info "内存使用率: ${mem_usage}%"
    
    if [ "$mem_usage" -gt "$THRESHOLD_MEM" ]; then
        alerts+=("内存使用率过高: ${mem_usage}%")
    fi
    
    # 磁盘使用率
    local disk_usage=$(df -h . | awk 'NR==2 {print $5}' | sed 's/%//')
    log_info "磁盘使用率: ${disk_usage}%"
    
    if [ "$disk_usage" -gt "$THRESHOLD_DISK" ]; then
        alerts+=("磁盘使用率过高: ${disk_usage}%")
    fi
    
    # 发送告警
    if [ ${#alerts[@]} -gt 0 ]; then
        send_alert "资源告警" "${alerts[*]}"
        return 1
    fi
    
    return 0
}

# 检查 API 健康
check_api_health() {
    log_info "检查 API 健康状态..."
    
    local max_retries=3
    local retry=0
    
    while [ $retry -lt $max_retries ]; do
        if curl -s -f http://localhost:8080/api/v1/health > /dev/null; then
            log_info "API 健康检查通过"
            return 0
        fi
        
        retry=$((retry + 1))
        log_warn "API 健康检查失败，重试 ${retry}/${max_retries}..."
        sleep 2
    done
    
    log_error "API 健康检查失败"
    send_alert "API 告警" "后端 API 健康检查失败"
    return 1
}

# 检查日志错误
check_logs() {
    log_info "检查日志错误..."
    
    local errors=$(docker-compose logs --tail=100 2>&1 | grep -i "error\|fatal\|panic" | wc -l)
    
    if [ "$errors" -gt 0 ]; then
        log_warn "最近 100 条日志中发现 ${errors} 个错误"
        
        if [ "$errors" -gt 10 ]; then
            send_alert "日志告警" "最近日志中发现 ${errors} 个错误"
        fi
    else
        log_info "日志检查正常"
    fi
}

# 生成报告
generate_report() {
    log_info "生成监控报告..."
    
    local report_file="./logs/monitor-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" << EOF
Claw Native Company - 监控报告
================================
生成时间: $(date '+%Y-%m-%d %H:%M:%S')

系统信息:
---------
主机名: $(hostname)
操作系统: $(uname -a)
运行时间: $(uptime -p)

容器状态:
---------
$(docker-compose ps)

资源使用:
---------
CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}')
内存: $(free -h | grep Mem)
磁盘: $(df -h . | tail -1)

API 状态:
---------
$(curl -s http://localhost:8080/api/v1/health 2>/dev/null || echo "无法连接")

日志统计:
---------
错误数(最近100条): $(docker-compose logs --tail=100 2>&1 | grep -ic "error\|fatal\|panic")
警告数(最近100条): $(docker-compose logs --tail=100 2>&1 | grep -ic "warn")
EOF
    
    log_info "报告已生成: $report_file"
}

# 发送告警
send_alert() {
    local title="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    log_warn "发送告警: $title - $message"
    
    # Webhook 告警
    if [ -n "$WEBHOOK_URL" ]; then
        curl -s -X POST "$WEBHOOK_URL" \
            -H "Content-Type: application/json" \
            -d "{\"title\":\"$title\",\"message\":\"$message\",\"timestamp\":\"$timestamp\"}" \
            > /dev/null || log_error "Webhook 发送失败"
    fi
    
    # 邮件告警
    if [ -n "$EMAIL" ] && command -v mail &> /dev/null; then
        echo "$message" | mail -s "[Claw Alert] $title" "$EMAIL" || log_error "邮件发送失败"
    fi
}

# 执行检查
run_check() {
    log_info "开始系统检查..."
    
    local failed=0
    
    check_containers || failed=$((failed + 1))
    check_resources || failed=$((failed + 1))
    check_api_health || failed=$((failed + 1))
    check_logs
    
    if [ $failed -eq 0 ]; then
        log_info "✅ 所有检查通过"
    else
        log_error "❌ 部分检查失败"
    fi
    
    return $failed
}

# 显示帮助
show_help() {
    cat << EOF
Claw Native Company - 监控脚本

用法: $0 [check|report|alert|help]

命令:
  check   执行系统检查（默认）
  report  生成监控报告
  alert   测试告警
  help    显示帮助

环境变量:
  ALERT_WEBHOOK_URL  告警 Webhook URL
  ALERT_EMAIL        告警邮箱

阈值设置:
  CPU:  ${THRESHOLD_CPU}%
  MEM:  ${THRESHOLD_MEM}%
  DISK: ${THRESHOLD_DISK}%
EOF
}

# 主函数
main() {
    # 确保日志目录存在
    mkdir -p ./logs
    
    case "$MONITOR_MODE" in
        check)
            run_check
            ;;
        report)
            generate_report
            ;;
        alert)
            send_alert "测试告警" "这是一条测试告警消息"
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $MONITOR_MODE"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
