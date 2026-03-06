#!/bin/bash

################################################################################
# 备份costrict后端系统
#
# 功能：备份整个costrict后端系统，包括docker compose环境、工作目录和Docker镜像
#
# 使用方法：
#   ./backup.sh [OPTIONS]
#   或
#   bash backup.sh --backend /path/to/backend --data /path/to/costrict --output /path/to/output
#
# 支持的选项：
#   --backend     backend所在目录，即启动costrict backend的docker compose环境目录
#                 默认值: /usr/local/costrict
#   --data        costrict软件的数据目录，即从cloud下载安装组件的位置，
#                 也是costrict-admin保存数据的位置
#                 默认值: /root/.costrict
#   --output      备份输出目录，备份文件将保存在此目录下
#                 默认值: ./backup_YYYY-MM-DD_HH-MM-SS
#
# 备份步骤：
#   1. 调用 $backend/run.sh stop 停止docker compose服务
#   2. 调用 service costrict-admin stop 停止costrict-admin服务
#   3. 拷贝目录 $backend 的所有内容到 $output/backend 下
#   4. 拷贝目录 $data 的所有内容到 $output/data 下
#   5. 调用 $backend/scripts/save-images.sh 将Docker镜像备份到 $output/images 下
#
# 输出结构：
#   $output/
#   ├── backend/          # 后端docker compose环境备份
#   ├── data/             # costrict数据目录备份
#   └── images/           # Docker镜像备份（tar文件）
#
# 示例：
#   # 使用默认路径进行备份
#   ./backup.sh
#
#   # 指定输出目录
#   ./backup.sh --output /mnt/backup/costrict_backup_2024
#
#   # 指定所有路径
#   ./backup.sh --backend /opt/costrict --data /data/costrict --output /backup/latest
#
# 注意事项：
#   - 本脚本需要root权限执行，因为涉及服务停止和系统目录复制
#   - 备份前会停止相关服务，请确保在低峰期执行
#   - 确保输出目录有足够的磁盘空间（通常需要数GB到数十GB）
#   - 备份过程可能需要较长时间，取决于数据量和镜像大小
#   - 备份完成后可以通过restore.sh脚本将系统恢复到备份时的状态
#
# 依赖脚本：
#   - $backend/run.sh: 用于停止/启动docker compose服务
#   - $backend/scripts/save-images.sh: 用于备份Docker镜像
#
################################################################################

set -e
set -u
set -o pipefail 2>/dev/null || true

# -------------------------- Initialize Configuration --------------------------
SCRIPT_NAME=$(basename "$0")
LOG_FILE="${SCRIPT_NAME%.*}.log"
exec > >(tee -a "$LOG_FILE") 2>&1

# -------------------------- Constants Definition --------------------------
declare -r SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
declare -r DEFAULT_BACKEND_DIR="/usr/local/costrict"
declare -r DEFAULT_DATA_DIR="/root/.costrict"
declare -r DEFAULT_OUTPUT_DIR="./backup_$(date +%Y-%m-%d_%H-%M-%S)"

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

show_usage() {
    cat << EOF
用法: $0 [选项]

选项:
    --backend <path>   backend所在目录，默认为 /usr/local/costrict
    --data <path>      costrict数据目录，默认为 /root/.costrict
    --output <path>    输出目录，默认为 ./backup_YYYY-MM-DD_HH-MM-SS
    -h, --help         显示帮助信息

示例:
    $0                                          # 使用默认设置备份
    $0 --output /mnt/backup/costrict_2024        # 指定输出目录
    $0 --backend /opt/costrict --output /backup  # 指定所有路径

说明:
    本脚本用于备份costrict后端系统，包括：
    - docker compose环境
    - costrict数据目录
    - Docker镜像

    备份完成后可通过 restore.sh 脚本恢复。

EOF
}

parse_arguments() {
    # 默认值
    BACKEND_DIR="$DEFAULT_BACKEND_DIR"
    DATA_DIR="$DEFAULT_DATA_DIR"
    OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --backend)
                BACKEND_DIR="$2"
                shift 2
                ;;
            --data)
                DATA_DIR="$2"
                shift 2
                ;;
            --output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log "ERROR" "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 转换为绝对路径
    BACKEND_DIR="$(cd "$BACKEND_DIR" 2>/dev/null && echo "$BACKEND_DIR" || echo "$BACKEND_DIR")"
    DATA_DIR="$(cd "$DATA_DIR" 2>/dev/null && echo "$DATA_DIR" || echo "$DATA_DIR")"
    
    log "INFO" "Backend目录: $BACKEND_DIR"
    log "INFO" "数据目录: $DATA_DIR"
    log "INFO" "输出目录: $OUTPUT_DIR"
}

check_root_permission() {
    if [ "$(id -u)" -ne 0 ]; then
        log "ERROR" "本脚本需要root权限执行"
        return 1
    fi
    log "INFO" "Root权限检查通过"
    return 0
}

check_prerequisites() {
    local missing_items=()
    
    # 检查backend目录
    if [[ ! -d "$BACKEND_DIR" ]]; then
        missing_items+=("Backend目录: $BACKEND_DIR")
    fi
    
    # 检查数据目录
    if [[ ! -d "$DATA_DIR" ]]; then
        missing_items+=("数据目录: $DATA_DIR")
    fi
    
    # 检查必要脚本
    if [[ ! -f "$BACKEND_DIR/run.sh" ]]; then
        missing_items+=("脚本: $BACKEND_DIR/run.sh")
    fi
    
    if [[ ! -f "$BACKEND_DIR/scripts/save-images.sh" ]]; then
        missing_items+=("脚本: $BACKEND_DIR/scripts/save-images.sh")
    fi
    
    if [[ ${#missing_items[@]} -gt 0 ]]; then
        log "ERROR" "前置条件检查失败，以下项目缺失:"
        for item in "${missing_items[@]}"; do
            log "ERROR" "  - ${item}"
        done
        return 1
    fi
    
    log "INFO" "前置条件检查通过"
    return 0
}

stop_docker_compose_services() {
    log "INFO" "停止docker compose服务..."
    cd "$BACKEND_DIR" || return 1
    if bash run.sh stop; then
        log "INFO" "docker compose服务已停止"
        return 0
    else
        log "ERROR" "docker compose服务停止失败"
        return 1
    fi
}

stop_costrict_admin_service() {
    log "INFO" "停止costrict-admin服务..."
    if service costrict-admin status >/dev/null 2>&1; then
        if service costrict-admin stop; then
            log "INFO" "costrict-admin服务已停止"
        else
            log "WARN" "costrict-admin服务停止失败"
        fi
    else
        log "INFO" "costrict-admin服务未运行，跳过"
    fi
    return 0
}

create_output_directory() {
    log "INFO" "创建输出目录: $OUTPUT_DIR"
    if mkdir -p "$OUTPUT_DIR"; then
        log "INFO" "输出目录创建成功"
        return 0
    else
        log "ERROR" "输出目录创建失败"
        return 1
    fi
}

backup_backend_directory() {
    log "INFO" "备份backend目录..."
    log "INFO" "源目录: $BACKEND_DIR"
    log "INFO" "目标目录: $OUTPUT_DIR/backend"
    
    if cp -rp "$BACKEND_DIR" "$OUTPUT_DIR/backend"; then
        log "INFO" "backend目录备份完成"
        return 0
    else
        log "ERROR" "backend目录备份失败"
        return 1
    fi
}

backup_data_directory() {
    log "INFO" "备份costrict数据目录..."
    log "INFO" "源目录: $DATA_DIR"
    log "INFO" "目标目录: $OUTPUT_DIR/data"
    
    if cp -rp "$DATA_DIR" "$OUTPUT_DIR/data"; then
        log "INFO" "数据目录备份完成"
        return 0
    else
        log "ERROR" "数据目录备份失败"
        return 1
    fi
}

backup_docker_images() {
    log "INFO" "备份Docker镜像..."
    local images_dir="$OUTPUT_DIR/images"
    
    # 清理并创建images目录
    if [[ -d "$images_dir" ]]; then
        rm -rf "$images_dir"
    fi
    mkdir -p "$images_dir"
    
    cd "$BACKEND_DIR" || return 1
    if bash scripts/save-images.sh -f .images.list -s "$images_dir"; then
        log "INFO" "Docker镜像备份完成"
        return 0
    else
        log "ERROR" "Docker镜像备份失败"
        return 1
    fi
}

generate_manifest() {
    log "INFO" "生成备份清单..."
    cat > "$OUTPUT_DIR/MANIFEST.txt" <<EOF
Costrict后端系统备份清单
========================
备份时间: $(date)
备份服务器: $(hostname)
Backend目录: $BACKEND_DIR
数据目录: $DATA_DIR

目录结构:
  backend/  - 后端docker compose环境
  data/     - 数据工作目录
  images/   - Docker镜像备份（tar文件）

恢复方法:
  cd $OUTPUT_DIR/backend
  bash restore.sh --input $OUTPUT_DIR

注意事项:
  - 恢复前确保满足系统和硬件要求
  - 版本兼容性: 备份和恢复的costrict版本应保持一致
  - 需要root权限执行恢复
EOF
    log "INFO" "备份清单已生成: $OUTPUT_DIR/MANIFEST.txt"
}

show_backup_summary() {
    log "INFO" "======================================"
    log "INFO" "备份完成！"
    log "INFO" "======================================"
    log "INFO" "备份位置: $OUTPUT_DIR"
    log "INFO" ""
    log "INFO" "目录大小:"
    du -sh "$OUTPUT_DIR"/* 2>/dev/null | while read -r line; do
        log "INFO" "  $line"
    done
    log "INFO" ""
    log "INFO" "后续操作:"
    log "INFO" "  可以将 $OUTPUT_DIR 目录传输到备份服务器"
    log "INFO" "  或者使用以下命令恢复: cd $OUTPUT_DIR/backend; bash restore.sh --input=$OUTPUT_DIR"
    log "INFO" "======================================"
}

# -------------------------- Main Execution --------------------------
main() {
    log "INFO" "======================================"
    log "INFO" "开始备份 costrict 后端系统"
    log "INFO" "======================================"
    
    # 解析参数
    parse_arguments "$@"
    
    # 检查root权限
    if ! check_root_permission; then
        exit 1
    fi
    
    # 检查前置条件
    if ! check_prerequisites; then
        exit 1
    fi
    
    # 执行备份
    stop_docker_compose_services
    stop_costrict_admin_service
    create_output_directory
    backup_backend_directory
    backup_data_directory
    backup_docker_images
    generate_manifest
    
    # 显示汇总信息
    show_backup_summary
    
    log "INFO" "备份流程全部完成"
}

main "$@"
exit 0
