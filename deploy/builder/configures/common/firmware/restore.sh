#!/bin/bash

################################################################################
# 恢复costrict后端系统
#
# 功能：将backup.sh备份的costrict后端系统恢复到工作环境
#
# 使用方法：
#   ./restore.sh [OPTIONS]
#   或
#   bash restore.sh --input /path/to/backup --backend /path/to/backend --data /path/to/data
#
# 支持的选项：
#   --input       备份数据来源目录，该目录由backup.sh脚本生成
#                 必须包含backup、costrict、images三个子目录
#                 此参数为必需参数，无默认值
#   --backend     backend所在目录，即启动costrict backend的docker compose环境目录
#                 默认值: /usr/local/costrict
#   --data        costrict的数据目录，即从cloud下载安装组件的位置，
#                 也是costrict-admin保存数据的位置
#                 默认值: /root/.costrict
#
# 恢复步骤（与backup.sh的步骤相反）：
#   步骤1. 环境检查：调用check.sh确认必要的组件和环境都正常
#   步骤2. 加载Docker镜像：调用$backend/scripts/load-images.sh加载镜像
#   步骤3. 恢复backend目录：将$input/backend下的内容拷贝到$backend目录
#   步骤4. 恢复data目录：将$input/data下的内容拷贝到$data目录
#   步骤5. 注册服务：source加载$backend/init.sh，运行register_services注册costrict-admin服务
#   步骤6. 启动服务：运行$backend/run.sh start，启动docker compose服务
#
# 备份目录结构要求：
#   $input/
#   ├── backend/      # 后端docker compose环境备份
#   ├── data/         # costrict数据目录备份
#   └── images/       # Docker镜像备份（tar文件）
#
# 示例：
#   # 恢复到默认路径
#   ./restore.sh --input /mnt/backup/costrict_backup_2024
#
#   # 恢复到指定路径
#   ./restore.sh --input /backup/latest --backend /opt/costrict --data /data/costrict
#
#   # 恢复到不同的服务器
#   ./restore.sh --input /tmp/backup --backend /usr/local/costrict_new --data /data/costrict_new
#
# 注意事项：
#   - 本脚本需要root权限执行，因为涉及服务管理和系统目录操作
#   - 确保输入目录是由backup.sh生成的完整备份
#   - 恢复前请确保目标目录（backend和data）为空，否则恢复将失败
#   - 建议在恢复前先备份当前环境，以防需要回滚
#   - 恢复后需要验证服务是否正常启动
#   - 版本兼容性：备份和恢复的costrict版本应该一致，否则可能出现兼容性问题
#   - Docker镜像加载可能需要较长时间，取决于镜像大小和数量
#   - 如果硬件环境（CPU架构等）与备份源不一致，镜像可能无法正常加载
#   - 恢复过程中的错误会导致脚本退出，不会部分恢复
#
# 依赖脚本：
#   - check.sh: 用于检查环境是否满足恢复要求
#   - $backend/scripts/load-images.sh: 用于加载Docker镜像
#   - $backend/init.sh: 用于注册costrict-admin服务
#   - $backend/run.sh: 用于启动docker compose服务
#
# 返回值：
#   0 - 恢复成功
#   非0 - 恢复失败（具体错误信息会输出到stderr）
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
declare -r DEFAULT_INPUT_DIR=""

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
    --input <path>     备份数据来源目录（必需），由backup.sh脚本生成
    --backend <path>   backend所在目录，默认为 /usr/local/costrict
    --data <path>      costrict数据目录，默认为 /root/.costrict
    -h, --help         显示帮助信息

示例:
    $0 --input /mnt/backup/costrict_2024           # 恢复到默认路径
    $0 --input /backup/latest --backend /opt/...   # 指定所有路径

说明:
    本脚本用于恢复由backup.sh脚本备份的costrict后端系统。
    恢复前请确保目标目录为空，否则会报错退出。

EOF
}

parse_arguments() {
    # 默认值
    BACKEND_DIR="$DEFAULT_BACKEND_DIR"
    DATA_DIR="$DEFAULT_DATA_DIR"
    INPUT_DIR="$DEFAULT_INPUT_DIR"
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --input)
                INPUT_DIR="$2"
                shift 2
                ;;
            --backend)
                BACKEND_DIR="$2"
                shift 2
                ;;
            --data)
                DATA_DIR="$2"
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
    if [[ -n "$INPUT_DIR" ]]; then
        INPUT_DIR="$(cd "$INPUT_DIR" 2>/dev/null && echo "$INPUT_DIR" || echo "$INPUT_DIR")"
    fi
    
    log "INFO" "输入目录: $INPUT_DIR"
    log "INFO" "Backend目录: $BACKEND_DIR"
    log "INFO" "Costrict数据目录: $DATA_DIR"
}

check_root_permission() {
    if [ "$(id -u)" -ne 0 ]; then
        log "ERROR" "本脚本需要root权限执行"
        return 1
    fi
    log "INFO" "Root权限检查通过"
    return 0
}

check_input_directory() {
    local missing_items=()
    
    # 检查输入目录结构
    if [[ ! -d "$INPUT_DIR" ]]; then
        missing_items+=("输入目录不存在: $INPUT_DIR")
    fi
    
    if [[ ! -d "$INPUT_DIR/backend" ]]; then
        missing_items+=("缺少: \$INPUT_DIR/backend")
    fi
    
    if [[ ! -d "$INPUT_DIR/data" ]]; then
        missing_items+=("缺少: \$INPUT_DIR/data")
    fi
    
    if [[ ! -d "$INPUT_DIR/images" ]]; then
        missing_items+=("缺少: \$INPUT_DIR/images")
    fi
    
    if [[ ${#missing_items[@]} -gt 0 ]]; then
        log "ERROR" "输入目录结构检查失败:"
        for item in "${missing_items[@]}"; do
            log "ERROR" "  - ${item}"
        done
        return 1
    fi
    
    log "INFO" "输入目录结构检查通过"
    return 0
}

check_prerequisites() {
    local missing_items=()
    local input_backend_dir="$INPUT_DIR/backend"
    
    # 检查backend目录
    if [[ ! -d "$input_backend_dir" ]]; then
        missing_items+=("Backend目录: $input_backend_dir")
    fi
    
    # 检查必要脚本
    if [[ ! -f "$input_backend_dir/run.sh" ]]; then
        missing_items+=("脚本: $input_backend_dir/run.sh")
    fi
    
    if [[ ! -f "$input_backend_dir/init.sh" ]]; then
        missing_items+=("脚本: $input_backend_dir/init.sh")
    fi
    
    if [[ ! -f "$input_backend_dir/scripts/load-images.sh" ]]; then
        missing_items+=("脚本: $input_backend_dir/scripts/load-images.sh")
    fi
    
    if [[ ! -f "$input_backend_dir/check.sh" ]]; then
        missing_items+=("脚本: $input_backend_dir/check.sh")
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

check_docker_service() {
    log "INFO" "检查Docker服务..."
    
    # 检查Docker是否安装
    if ! command -v docker >/dev/null 2>&1; then
        log "ERROR" "Docker未安装"
        return 1
    fi
    
    # 检查Docker服务是否运行
    if ! docker info >/dev/null 2>&1; then
        log "WARN" "Docker服务未运行，正在启动..."
        systemctl start docker
        if ! docker info >/dev/null 2>&1; then
            log "ERROR" "Docker服务启动失败"
            return 1
        fi
    fi
    
    log "INFO" "Docker服务状态正常"
    return 0
}

check_docker_compose() {
    log "INFO" "检查Docker Compose..."
    
    if ! docker-compose version >/dev/null 2>&1; then
        log "ERROR" "Docker Compose未安装或不可用"
        return 1
    fi
    
    log "INFO" "Docker Compose检查通过"
    return 0
}

load_docker_images() {
    log "INFO" "加载Docker镜像..."
    log "INFO" "镜像来源: $INPUT_DIR/images"
    
    cd "$BACKEND_DIR" || return 1
    
    if [[ ! -d "$INPUT_DIR/images" ]] || [[ -z "$(ls -A $INPUT_DIR/images 2>/dev/null)" ]]; then
        log "WARN" "镜像目录为空或不存在，跳过镜像加载"
        return 0
    fi
    
    if bash scripts/load-images.sh "$INPUT_DIR/images"; then
        log "INFO" "Docker镜像加载完成"
        return 0
    else
        log "ERROR" "Docker镜像加载失败"
        return 1
    fi
}

restore_backend_directory() {
    log "INFO" "恢复backend目录..."
    log "INFO" "备份源: $INPUT_DIR/backend"
    log "INFO" "目标目录: $BACKEND_DIR"
    
    # 检查目标目录是否为空
    if [[ -d "$BACKEND_DIR" ]]; then
        local file_count=$(find "$BACKEND_DIR" -mindepth 1 ! -path "$BACKEND_DIR/scripts" ! -path "$BACKEND_DIR/scripts/*" | wc -l)
        if [[ "$file_count" -ne 0 ]]; then
            log "ERROR" "目标目录不为空，请手动清理：$BACKEND_DIR"
            return 1
        fi
    fi
    
    # 确保目标目录存在
    mkdir -p "$BACKEND_DIR"
    
    # 复制备份内容
    if cp -rp "$INPUT_DIR/backend"/* "$BACKEND_DIR/"; then
        log "INFO" "backend目录恢复完成"
        return 0
    else
        log "ERROR" "backend目录恢复失败"
        return 1
    fi
}

restore_data_directory() {
    log "INFO" "恢复costrict数据目录..."
    log "INFO" "备份源: $INPUT_DIR/data"
    log "INFO" "目标目录: $DATA_DIR"
    
    # 检查目标目录是否为空
    if [[ -d "$DATA_DIR" ]]; then
        local file_count=$(find "$DATA_DIR" -mindepth 1 | wc -l)
        if [[ "$file_count" -ne 0 ]]; then
            log "ERROR" "目标目录不为空，请手动清理：$DATA_DIR"
            return 1
        fi
    fi
    
    # 确保目标目录存在
    mkdir -p "$DATA_DIR"
    
    # 复制备份内容
    if cp -rp "$INPUT_DIR/data"/* "$DATA_DIR/"; then
        log "INFO" "costrict数据目录恢复完成"
        return 0
    else
        log "ERROR" "costrict数据目录恢复失败"
        return 1
    fi
}

check_environment() {
    log "INFO" "执行环境检查..."
    cd "$BACKEND_DIR" || return 1
    if ! bash check.sh; then
        log "ERROR" "环境检查失败"
        return 1
    fi
    log "INFO" "环境检查完成"   
    return 0
}


register_services() {
    local initd_dir="${BACKEND_DIR}/init.d"
    
    # 检查 init.d 目录是否存在
    if [[ ! -d "$initd_dir" ]]; then
        log "INFO" "未找到 init.d 目录，跳过服务注册"
        return 0
    fi
    
    log "INFO" "开始注册系统服务..."
    
    # 遍历 init.d 目录下的所有脚本
    for service_script in "$initd_dir"/*; do
        if [[ ! -f "$service_script" ]]; then
            continue
        fi
        
        local service_name=$(basename "$service_script")
        local sys_initd_path="/etc/init.d/${service_name}"
        
        log "INFO" "注册服务: $service_name"
        
        # 复制到系统目录
        if [[ -e "$sys_initd_path" ]]; then
            log "INFO" "服务已存在，删除旧文件"
            sudo rm -f "$sys_initd_path"
        fi
        
        # 拷贝文件
        if sudo cp "$service_script" "$sys_initd_path"; then
            # 设置执行权限
            sudo chmod +x "$sys_initd_path"
            
            # 注册为系统服务
            if command -v chkconfig >/dev/null 2>&1; then
                # 使用 chkconfig (CentOS/RHEL)
                sudo chkconfig --add "$service_name" 2>/dev/null || true
                sudo chkconfig "$service_name" on 2>/dev/null || true
                log "INFO" "服务 $service_name 已注册为开机自启 (chkconfig)"
            elif command -v update-rc.d >/dev/null 2>&1; then
                # 使用 update-rc.d (Debian/Ubuntu)
                sudo update-rc.d "$service_name" defaults 2>/dev/null || true
                log "INFO" "服务 $service_name 已注册为开机自启 (update-rc.d)"
            else
                log "WARN" "未找到 chkconfig 或 update-rc.d，无法自动注册开机自启"
            fi
            
            log "INFO" "服务 $service_name 注册成功"
        else
            log "ERROR" "服务 $service_name 注册失败"
        fi
    done
    
    return 0
}

start_docker_compose_services() {
    log "INFO" "启动docker compose服务..."
    cd "$BACKEND_DIR" || return 1
    
    if bash run.sh start; then
        log "INFO" "docker compose服务已启动"
        return 0
    else
        log "ERROR" "docker compose服务启动失败"
        return 1
    fi
}

start_costrict_admin_service() {
    log "INFO" "启动costrict-admin服务..."
    
    if command -v service >/dev/null 2>&1; then
        if service costrict-admin start; then
            log "INFO" "costrict-admin服务已启动"
        else
            log "WARN" "costrict-admin服务启动失败或未安装"
        fi
    elif systemctl is-enabled costrict-admin &> /dev/null; then
        if systemctl start costrict-admin; then
            log "INFO" "costrict-admin服务已启动"
        else
            log "WARN" "costrict-admin服务启动失败"
        fi
    else
        log "WARN" "costrict-admin服务未安装，跳过启动"
    fi
    
    return 0
}

show_restore_summary() {
    log "INFO" "======================================"
    log "INFO" "恢复完成！"
    log "INFO" "======================================"
    log "INFO" "恢复来源: $INPUT_DIR"
    log "INFO" "Backend目录: $BACKEND_DIR"
    log "INFO" "Costrict数据目录: $DATA_DIR"
    log "INFO" ""
    log "INFO" "后续操作:"
    log "INFO" "  1. 检查服务状态: cd $BACKEND_DIR; bash run.sh status"
    log "INFO" "  2. 检查costrict-admin服务: service costrict-admin status"
    log "INFO" "  3. 查看日志: cd $BACKEND_DIR; bash run.sh logs"
    log "INFO" "======================================"
    log "INFO" ""
    log "INFO" "注意事项:"
    log "INFO" "  - 请验证所有服务是否正常启动"
    log "INFO" "  - 如果服务启动失败，请检查日志文件"
    log "INFO" "  - 如需回滚，请使用之前创建的备份目录"
    log "INFO" "======================================"
}

# -------------------------- Main Execution --------------------------
main() {
    log "INFO" "======================================"
    log "INFO" "开始恢复 costrict 后端系统"
    log "INFO" "======================================"
    
    # 解析参数
    parse_arguments "$@"
    
    # 检查必需参数
    if [[ -z "$INPUT_DIR" ]]; then
        log "ERROR" "缺少必需参数: --input"
        show_usage
        exit 1
    fi
    
    # 检查root权限
    if ! check_root_permission; then
        exit 1
    fi
    
    # 检查输入目录结构
    if ! check_input_directory; then
        exit 1
    fi
    
    # 检查前置条件
    if ! check_prerequisites; then
        exit 1
    fi
    
    check_docker_service
    check_docker_compose
    
    # 执行恢复
    load_docker_images
    restore_backend_directory
    restore_data_directory

    # 检查安装后的环境和依赖
    if ! check_environment; then
        exit 1
    fi

    if ! register_services; then
        log "WARN" "注册系统服务发现问题，但继续执行"
    fi

    start_docker_compose_services
    start_costrict_admin_service
    
    # 显示汇总信息
    show_restore_summary
    
    log "INFO" "恢复流程全部完成"
}

main "$@"
exit 0
