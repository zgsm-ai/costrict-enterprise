#!/bin/bash
set -e
set -u
set -o pipefail 2>/dev/null || true

# -------------------------- Initialize Configuration --------------------------
SCRIPT_NAME=$(basename "$0")
LOG_FILE="${SCRIPT_NAME%.*}.log"
exec > >(tee -a "$LOG_FILE") 2>&1

# -------------------------- Constants Definition --------------------------

# Get the current directory
declare -r BASE_DIR=$(pwd)

# -------------------------- Function Definitions --------------------------
docker-compose() {
    # Check if docker has compose subcommand
    if docker compose version >/dev/null 2>&1; then
        command docker compose "$@"
    else
        command docker-compose "$@"
    fi
}

log() {
    local level=$1
    local message=$2
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S")
    echo -e "[${timestamp}] [${level}] ${message}"
}

check_core_dependencies() {
    local missing_deps=()
    local base_path="${BASE_DIR}"
    
    # 检查依赖项目录/文件
    local deps=(
        ".env"
        "scripts/gen-env-file.sh"
        "docker-compose.yml"
        "costrict-admin-backend"
        "apisix"
        "portal"
        "portal/data/costrict-admin"
        "portal/data/costrict-admin/index.html"
    )
    
    for dep in "${deps[@]}"; do
        local dep_path="${base_path}/${dep}"
        if [[ ! -e "$dep_path" ]]; then
            missing_deps+=("$dep")
            log "WARN" "缺失依赖: ${dep}"
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log "ERROR" "核心依赖检查失败，以下依赖项不存在:"
        for dep in "${missing_deps[@]}"; do
            log "ERROR" "  - ${dep}"
        done
        return 1
    fi
    
    log "INFO" "核心依赖检查通过"
    return 0
}

prepare() {
    log "INFO" "检查核心依赖..."
    if ! check_core_dependencies; then
        log "ERROR" "核心依赖检查失败，请确保所有依赖目录和文件都存在"
        return 1
    fi
    
    log "INFO" "生成环境变量文件 .env..."
    if ! bash scripts/gen-env-file.sh; then
        log "ERROR" "生成环境变量文件 .env失败"
        return 1
    fi
    return 0
}

start_docker_services() {
    if ! prepare; then
        return 1
    fi
    
    log "INFO" "启动Docker Compose服务..."
    if ! docker-compose  -f docker-compose.yml up -d; then
        log "ERROR" "Docker Compose服务启动失败"
        return 1
    fi
    . .env
    log "INFO" "Docker Compose服务启动完成"
    log "INFO" "系统启动完成"
    log "INFO" "请登录到诸葛神码后端管理页面 [http://${COSTRICT_HOST}:${PORT_APISIX_ENTRY}/costrict-admin/] (默认账号: admin, 密码: admin)"
    return 0
}

start_docker_service() {
    local service=$1

    if ! prepare; then
        return 1
    fi
    
    # 启动指定服务
    if ! docker-compose -f docker-compose.yml up -d "$service"; then
        log "ERROR" "启动服务 $service 失败"
        return 1
    fi
    return 0
}

restart_docker_service() {
    local service=$1

    log "INFO" "开始重启服务: $service"
    if ! prepare; then
        return 1
    fi
    # 重启指定服务
    if ! docker-compose -f docker-compose.yml restart "$service"; then
        log "ERROR" "重启服务 $service 失败"
        return 1
    fi
    log "INFO" "服务 $service 重启完成"
    return 0
}

stop_docker_services() {
    log "INFO" "停止Docker Compose服务..."
    if ! docker-compose  -f docker-compose.yml down; then
        log "ERROR" "Docker Compose服务停止失败"
        return 1
    fi
    log "INFO" "Docker Compose服务已停止"
    return 0
}

stop_docker_service() {
    local service=$1

    log "INFO" "开始停止服务: $service"
    
    # 停止指定服务
    if ! docker-compose -f docker-compose.yml stop "$service"; then
        log "ERROR" "停止服务 $service 失败"
        return 1
    fi
    
    log "INFO" "服务 $service 已停止"
    return 0
}

check_docker_status() {
    log "INFO" "检查Docker Compose服务状态..."
    # 获取docker-compose ps的输出
    local ps_output
    ps_output=$(docker-compose  -f docker-compose.yml ps 2>&1)
    local ps_exit_code=$?
    
    # 如果命令执行失败，说明服务未运行
    if [[ $ps_exit_code -ne 0 ]]; then
        log "INFO" "Docker Compose服务未运行"
        return 1
    fi
    
    # 逐行处理统计容器状态
    local healthy_count=0 unhealthy_count=0 starting_count=0 no_status_count=0
    local line_count=0
    
    # 逐行处理容器信息，跳过第一行（表头）
    while IFS= read -r line; do
        ((line_count++))
        # 跳过第一行表头
        [[ $line_count -eq 1 ]] && continue
        
        # 跳过空行
        [[ -z "$line" ]] && continue
        
        #根据状态标识分类统计
        if [[ "$line" == *"(health: starting)"* ]]; then
            ((starting_count++))
        elif [[ "$line" == *"(healthy)"* ]]; then
            ((healthy_count++))
        elif [[ "$line" == *"(unhealthy)"* ]]; then
            ((unhealthy_count++))
        else
            # 无状态标识的容器
            ((no_status_count++))
        fi
    done <<< "$ps_output"
    
    # 计算总容器数
    local total_count=$((healthy_count + unhealthy_count + starting_count + no_status_count))
    
    # 显示四类统计结果
    log "INFO" "容器状态统计: (总计 ${total_count} 个)"
    log "INFO" "  - 健康(healthy): ${healthy_count} 个"
    log "INFO" "  - 不健康(unhealthy): ${unhealthy_count} 个"
    log "INFO" "  - 健康检查中(health: starting): ${starting_count} 个"
    log "INFO" "  - 无状态标识(no status): ${no_status_count} 个"
    
    # 如果有不健康的容器，显示详细信息
    if [[ $unhealthy_count -gt 0 ]]; then
        # 显示不健康的容器
        log "WARN" "不健康容器详情:"
        echo "$ps_output" | while IFS= read -r line; do
            if [[ "$line" == *"(unhealthy)"* ]]; then
                local container_name
                container_name=$(echo "$line" | awk '{print $1}')
                log "WARN" "  - ${container_name}"
            fi
        done
    fi
    
    # 如果有正在健康检查的容器，显示详细信息
    if [[ $starting_count -gt 0 ]]; then
        # 显示健康检查中的容器
        log "INFO" "健康检查中容器详情:"
        echo "$ps_output" | while IFS= read -r line; do
            if [[ "$line" == *"(health: starting)"* ]]; then
                local container_name
                container_name=$(echo "$line" | awk '{print $1}')
                log "INFO" "  - ${container_name}"
            fi
        done
    fi
    
    # 如果有容器在运行（排除退出的），则认为服务正在运行
    local running_count=$((healthy_count + unhealthy_count + starting_count + no_status_count))
    if [[ $running_count -gt 0 ]]; then
        log "INFO" "Docker Compose服务正在运行"
        return 0
    else
        log "INFO" "Docker Compose服务未运行"
        return 1
    fi
}

# -------------------------- Command Functions --------------------------
cmd_start() {
    local service="${1:-}"
    
    if [[ -z "$service" ]]; then
        log "INFO" "开始启动系统..."
        
        # 启动Docker Compose服务
        if ! start_docker_services; then
            return 1
        fi
        
        return 0
    else
        log "INFO" "开始启动服务: $service"
        
        # 启动Docker Compose服务
        if ! start_docker_service ${service}; then
            return 1
        fi
        
        log "INFO" "服务 $service 启动完成"
        return 0
    fi
}

cmd_stop() {
    local service="${1:-}"

    if [[ -z "$service" ]]; then
        stop_docker_services
    else
        stop_docker_service ${service}
    fi
}

cmd_restart() {
    local service="${1:-}"
    
    if [[ -z "$service" ]]; then
        log "INFO" "开始重启系统..."
        stop_docker_services
        sleep 3
        start_docker_services
    else
        restart_docker_service $service
    fi
}

cmd_status() {
    log "INFO" "系统状态："
    log "INFO" "=========="
    
    # 检查Docker Compose服务状态
    if check_docker_status; then
        log "INFO" "Docker Compose服务: 运行中"
    else
        log "INFO" "Docker Compose服务: 未运行"
    fi
    
    log "INFO" "=========="
    return 0
}

cmd_details() {
    local service="${1:-}"
    
    if [[ -z "$service" ]]; then
        # 显示所有服务的详细信息
        docker-compose -f docker-compose.yml ps
    else
        # 显示指定服务的详细信息
        log "INFO" "查看服务 $service 的详细信息..."
        docker-compose -f docker-compose.yml ps "$service"
    fi
}

# -------------------------- Usage --------------------------
show_usage() {
    cat << EOF
用法: $0 [命令] [服务名]

命令:
    start [服务名]     启动系统或指定服务（默认）
    stop [服务名]      停止系统或指定服务
    restart [服务名]   重启系统或指定服务
    detail [服务名]    查看容器详细信息或指定服务详情
    status             查看系统状态
    -h, --help         显示帮助信息

首次运行前需先执行 init.sh 进行系统初始化。

示例:
    $0 start           # 启动所有服务
    $0 start apisix    # 仅启动 apisix 服务
    $0 stop            # 停止所有服务
    $0 stop postgres   # 仅停止 postgres 服务
    $0 restart         # 重启所有服务
    $0 restart etcd    # 仅重启 etcd 服务
    $0 status          # 查看系统状态
    $0 detail          # 查看所有容器详细信息
    $0 detail redis    # 查看 redis 服务详细信息

EOF
}

# -------------------------- Main Logic --------------------------
main() {
    local action="${1:-start}"
    
    log "INFO" "脚本启动，日志文件: $LOG_FILE"
    
    case "$action" in
        start)
            cmd_start "${2:-}"
            ;;
        stop)
            cmd_stop "${2:-}"
            ;;
        restart)
            cmd_restart "${2:-}"
            ;;
        detail)
            cmd_details "${2:-}"
            ;;
        status)
            cmd_status
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            log "ERROR" "未知参数: $action"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"