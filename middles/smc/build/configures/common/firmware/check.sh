#!/bin/bash
set -e
set -u
set -o pipefail 2>/dev/null || true

# -------------------------- Initialize Configuration --------------------------
SCRIPT_NAME=$(basename "$0")
LOG_FILE="${SCRIPT_NAME%.*}.log"
exec > >(tee -a "$LOG_FILE") 2>&1

# -------------------------- Constants Definition --------------------------
declare -r BASE_DIR=$(pwd)
declare -r SERVER_IP=$(hostname -I | awk '{ print $1 }')
declare -r SERVER_IP

# -------------------------- Function Definitions --------------------------
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
    --dir <path>       检查目录，默认为当前目录 (pwd)
    -h, --help        显示帮助信息

示例:
    $0                                            # 检查当前目录
    $0 --dir /usr/local/costrict                  # 检查指定目录

EOF
}

validate_install_environment() {
    local target_base="${INSTALL_DIR:-$(pwd)}"
    local validation_errors=0
    local validation_warnings=0
    
    log "INFO" "验证安装环境..."
    log "INFO" "检查目标目录: $target_base"
    
    # ========== 1. 检查必要的系统命令 ==========
    log "INFO" "检查系统命令依赖..."
    
    declare -A required_commands=(
        ["bash"]="4.0"
        ["docker"]="19.0"
        ["curl"]="7.0"
        ["awk"]="1.0"
        ["sed"]="4.0"
        ["grep"]="2.0"
        ["hostname"]="1.0"
        ["date"]="1.0"
        ["getopt"]="1.0"
        ["find"]="4.0"
        ["mkdir"]="1.0"
        ["tail"]="1.0"
    )
    
    for cmd in "${!required_commands[@]}"; do
        local min_version="${required_commands[$cmd]}"
        
        if ! command -v "$cmd" >/dev/null 2>&1; then
            log "ERROR" "缺少必要命令: $cmd (最低要求版本: $min_version)"
            validation_errors=$((validation_errors + 1))
        else
            # 获取实际版本
            local actual_version=$(eval "$cmd" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "0.0")
            local major=$(echo "$actual_version" | cut -d'.' -f1)
            local minor=$(echo "$actual_version" | cut -d'.' -f2)
            
            local required_major=$(echo "$min_version" | cut -d'.' -f1)
            local required_minor=$(echo "$min_version" | cut -d'.' -f2)
            
            # 简单版本比较
            if [[ "$major" -lt "$required_major" ]] || \
               [[ "$major" -eq "$required_major" && "$minor" -lt "$required_minor" ]]; then
                log "WARN" "$cmd 版本过低: $actual_version (要求: >= $min_version)"
                validation_warnings=$((validation_warnings + 1))
            else
                log "INFO" "$cmd 版本检查通过: $actual_version (要求: >= $min_version)"
            fi
        fi
    done
    
    # ========== 2. 检查 Docker Compose ==========
    log "INFO" "检查 Docker Compose..."
    if docker compose version >/dev/null 2>&1; then
        compose_version=$(docker compose version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        log "INFO" "Docker Compose (插件) 版本: $compose_version"
    elif command -v docker-compose >/dev/null 2>&1; then
        compose_version=$(docker-compose --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        log "INFO" "Docker Compose (独立版) 版本: $compose_version"
    else
        log "ERROR" "未找到 Docker Compose (插件或独立版)"
        validation_errors=$((validation_errors + 1))
    fi
    
    # 检查 Docker 服务状态
    if command -v docker >/dev/null 2>&1; then
        if docker info >/dev/null 2>&1; then
            log "INFO" "Docker 服务运行正常"
        else
            log "WARN" "Docker 服务未运行或无权限访问"
            validation_warnings=$((validation_warnings + 1))
        fi
    fi
    
    # ========== 3. 检查必要的服务目录 ==========
    log "INFO" "检查必要的服务目录..."
    local required_service_dirs=(
        "${target_base}/init.d"
        "${target_base}/scripts"
        "${target_base}/apisix"
        "${target_base}/casdoor"
        "${target_base}/chat-rag"
        "${target_base}/code-completion"
        "${target_base}/codebase-embedder"
        "${target_base}/codebase-querier"
        "${target_base}/codereview"
        "${target_base}/costrict-admin-backend"
        "${target_base}/cotun"
        "${target_base}/credit-manager"
        "${target_base}/etcd"
        "${target_base}/monitoring"
        "${target_base}/oneapi"
        "${target_base}/portal"
        "${target_base}/portal/data/costrict-admin"
        "${target_base}/postgres"
        "${target_base}/quota-manager"
        "${target_base}/redis"
        "${target_base}/weaviate"
    )
    
    for dir in "${required_service_dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            log "WARN" "服务目录不存在: $dir"
            validation_warnings=$((validation_warnings + 1))
        fi
    done
    
    # ========== 4. 检查核心配置文件 ==========
    log "INFO" "检查核心配置文件..."
    local required_configs=(
        "configure.sh"
        "tpl-resolve.sh"
        "docker-download-images.sh"
        "docker-compose.yml"
        ".env"
        "utils.sh"
    )
    
    for config in "${required_configs[@]}"; do
        local config_path="${target_base}/${config}"
        if [[ ! -f "$config_path" ]]; then
            log "ERROR" "核心配置文件不存在: $config_path"
            validation_errors=$((validation_errors + 1))
        else
            # 检查文件是否可读
            if [[ ! -r "$config_path" ]]; then
                log "ERROR" "配置文件不可读: $config_path"
                validation_errors=$((validation_errors + 1))
            else
                log "INFO" "配置文件存在: $config"
            fi
        fi
    done
    
    # ========== 5. 检查 scripts 子脚本 ==========
    log "INFO" "检查 scripts 子脚本..."
    local required_scripts=(
        "scripts/gen-env-file.sh"
        "scripts/verify-images.sh"
        "scripts/pull-images.sh"
    )
    
    for script in "${required_scripts[@]}"; do
        local script_path="${target_base}/${script}"
        if [[ ! -f "$script_path" ]]; then
            log "ERROR" "脚本文件不存在: $script_path"
            validation_errors=$((validation_errors + 1))
        elif [[ ! -x "$script_path" ]]; then
            log "WARN" "脚本文件无执行权限: $script_path"
            validation_warnings=$((validation_warnings + 1))
        else
            log "INFO" "脚本文件存在: $script"
        fi
    done
    
    # ========== 6. 检查 APISIX 配置脚本 ==========
    log "INFO" "检查 APISIX 配置脚本..."
    local apisix_scripts=(
        "apisix-casdoor.sh"
        "apisix-chatrag.sh"
        "apisix-codereview.sh"
        "apisix-completion-v2.sh"
        "apisix-costrict-apps.sh"
        "apisix-costrict-admin.sh"
        "apisix-cotun.sh"
        "apisix-credit-manager.sh"
        "apisix-embedder.sh"
        "apisix-grafana.sh"
        "apisix-issue.sh"
        "apisix-oidc-auth.sh"
    )
    
    for script in "${apisix_scripts[@]}"; do
        local script_path="${target_base}/${script}"
        if [[ ! -f "$script_path" ]]; then
            log "WARN" "APISIX配置脚本不存在: $script_path"
            validation_warnings=$((validation_warnings + 1))
        fi
    done
    
    # ========== 7. 检查镜像环境文件 ==========
    log "INFO" "检查镜像配置..."
    local image_env_count=$(find "${target_base}" -name "image.env" -type f 2>/dev/null | wc -l)
    if [[ "$image_env_count" -eq 0 ]]; then
        log "WARN" "未找到任何 image.env 文件"
        validation_warnings=$((validation_warnings + 1))
    else
        log "INFO" "找到 $image_env_count 个 image.env 文件"
    fi
    
    # ========== 8. 验证核心脚本语法 ==========
    log "INFO" "验证核心脚本语法..."
    local syntax_check_scripts=(
        "${target_base}/configure.sh"
        "${target_base}/tpl-resolve.sh"
        "${target_base}/utils.sh"
    )
    
    for script in "${syntax_check_scripts[@]}"; do
        if [[ -f "$script" ]]; then
            if bash -n "$script" 2>/dev/null; then
                log "INFO" "$(basename "$script") 语法检查通过"
            else
                log "WARN" "$(basename "$script") 语法检查失败"
                validation_warnings=$((validation_warnings + 1))
            fi
        fi
    done
    
    # ========== 9. 检查CPU核心数 ==========
    log "INFO" "检查CPU核心数..."
    if command -v nproc >/dev/null 2>&1; then
        local cpu_cores=$(nproc)
        local min_cpu_cores=2
        if [[ "$cpu_cores" -lt "$min_cpu_cores" ]]; then
            log "WARN" "CPU核心数不足: ${cpu_cores}C (要求: >= ${min_cpu_cores}C)"
            validation_warnings=$((validation_warnings + 1))
        else
            log "INFO" "CPU核心数: ${cpu_cores}C (要求: >= ${min_cpu_cores}C)"
        fi
    else
        log "WARN" "无法检测CPU核心数，缺少nproc命令"
        validation_warnings=$((validation_warnings + 1))
    fi
    
    # ========== 10. 检查内存大小 ==========
    log "INFO" "检查内存大小..."
    if command -v free >/dev/null 2>&1; then
        local total_mem_mb=$(free -m | awk '/^Mem:/ {print $2}')
        local total_mem_gb=$((total_mem_mb / 1024))
        local min_mem_gb=2
        if [[ "$total_mem_gb" -lt "$min_mem_gb" ]]; then
            log "WARN" "内存不足: ${total_mem_gb}GB (要求: >= ${min_mem_gb}GB)"
            validation_warnings=$((validation_warnings + 1))
        else
            log "INFO" "内存大小: ${total_mem_gb}GB (要求: >= ${min_mem_gb}GB)"
        fi
    else
        log "WARN" "无法检测内存大小，缺少free命令"
        validation_warnings=$((validation_warnings + 1))
    fi
    
    # ========== 11. 检查磁盘空间 ==========
    log "INFO" "检查磁盘空间..."
    if command -v df >/dev/null 2>&1; then
        local min_space_gb=10
        local available_space=$(df -BG "$target_base" 2>/dev/null | tail -1 | awk '{print $4}' | tr -d 'G')
        if [[ -n "$available_space" ]] && [[ "$available_space" -lt "$min_space_gb" ]]; then
            log "WARN" "可用磁盘空间不足 ${min_space_gb}GB: ${available_space}GB"
            validation_warnings=$((validation_warnings + 1))
        else
            log "INFO" "可用磁盘空间: ${available_space}GB (要求: >= ${min_space_gb}GB)"
        fi
    fi
    
    # ========== 12. 验证报告 ==========
    log "INFO" "安装环境验证完成"
    echo ""
    log "INFO" "======================================"
    log "INFO" "验证报告"
    log "INFO" "======================================"
    log "INFO" "错误: $validation_errors"
    log "INFO" "警告: $validation_warnings"
    log "INFO" "======================================"
    
    if [[ $validation_errors -gt 0 ]]; then
        log "ERROR" "发现 $validation_errors 个错误，无法继续安装"
        return 1
    fi
    
    if [[ $validation_warnings -gt 0 ]]; then
        log "WARN" "发现 $validation_warnings 个警告，安装可以继续但可能存在问题"
    fi
    
    return 0
}

# -------------------------- Main Logic --------------------------
main() {
    log "INFO" "环境检查脚本启动，日志文件: $LOG_FILE"
    
    # 解析命令行参数
    INSTALL_DIR="$(pwd)"
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dir)
                INSTALL_DIR="$2"
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
    INSTALL_DIR="$(cd "$INSTALL_DIR" 2>/dev/null && echo "$INSTALL_DIR" || echo "$INSTALL_DIR")"
    
    # 执行验证
    if validate_install_environment; then
        log "INFO" "环境检查通过"
        exit 0
    else
        log "ERROR" "环境检查失败"
        exit 1
    fi
}

main "$@"