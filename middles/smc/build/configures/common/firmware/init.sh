#!/bin/bash
set -e
set -u
set -o pipefail 2>/dev/null || true

# -------------------------- Initialize Configuration --------------------------
SCRIPT_NAME=$(basename "$0")
LOG_FILE="${SCRIPT_NAME%.*}.log"
exec > >(tee -a "$LOG_FILE") 2>&1

# -------------------------- Constants Definition --------------------------
# 标识文件：表示系统已经初始化成功
declare -r BASE_DIR=$(pwd)
declare -r INITIALIZED_FLAG="${BASE_DIR}/.system-initialized"
# Get the machine's IP
SERVER_IP=$(hostname -I | awk '{ print $1 }')
declare -r SERVER_IP

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
    --dir <path>       安装目录，默认为 /usr/local/costrict
    --force            强制重新初始化，跳过初始化标识检查
    -h, --help        显示帮助信息

示例:
    $0                                            # 使用默认设置安装
    $0 --dir /opt/costrict                        # 安装目录为 /opt/costrict
    $0 --force                                    # 强制重新初始化系统

EOF
}

parse_arguments() {
    # 默认值
    INSTALL_DIR="/usr/local/costrict"
    FORCE_REINIT=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --force)
                FORCE_REINIT=true
                shift
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
    
    log "INFO" "安装目录: $INSTALL_DIR"
    if [ "$FORCE_REINIT" = true ]; then
        log "INFO" "强制重新初始化模式：跳过初始化标识检查"
    fi
}

fix_permissions() {
    local target_base="${INSTALL_DIR}"
    
    # 需要修正权限的目录（相对于安装目录）
    declare -a dirs=(
        "etcd/data"
        "es/data"
        "portal/data"
        "oneapi/data"
        "postgres/initdb.d"
    )
    
    log "INFO" "开始修正目录权限..."
    
    for dir in "${dirs[@]}"; do
        local full_path="${target_base}/${dir}"
        
        # 自动创建目录（如果不存在）
        if [[ ! -d "$full_path" ]]; then
            log "INFO" "创建目录: $full_path"
            if ! sudo mkdir -p "$full_path"; then
                log "ERROR" "创建目录失败: $full_path"
                continue
            fi
        fi
        
        log "INFO" "处理目录: $full_path"
        
        # 修改所有权
        if sudo chown -R 1000:1000 "$full_path" 2>/dev/null; then
            log "INFO" "修改所有权成功: 1000:1000"
        else
            log "WARN" "修改所有权失败，可能当前用户权限不足或目录不存在"
        fi
        
        # 修改权限
        if sudo chmod -R 0775 "$full_path" 2>/dev/null; then
            log "INFO" "修改权限成功: 0775"
        else
            log "WARN" "修改权限失败"
        fi
    done
    
    # 设置脚本文件的执行权限
    log "INFO" "设置脚本文件执行权限..."
    find "${target_base}" -type f -name "*.sh" -exec sudo chmod +x {} \; 2>/dev/null || true
    
    log "INFO" "权限修正完成"
    return 0
}

process_template_files() {
    local target_base="${INSTALL_DIR}"
    
    # 切换到目标目录
    cd "$target_base" || {
        log "ERROR" "无法切换到安装目录: $target_base"
        return 1
    }
    
    log "INFO" "开始处理模板文件..."
    
    # 检查模板解析脚本是否存在
    if [[ ! -f "tpl-resolve.sh" ]]; then
        log "WARN" "模板解析脚本不存在: tpl-resolve.sh，跳过模板处理"
        cd - >/dev/null
        return 0
    fi
    
    # 检查配置文件是否存在
    if [[ ! -f "configure.sh" ]]; then
        log "WARN" "配置文件不存在: configure.sh，跳过模板处理"
        cd - >/dev/null
        return 0
    fi
    
    # 执行模板解析脚本
    log "INFO" "执行模板解析..."
    if bash tpl-resolve.sh; then
        log "INFO" "模板文件处理成功"
    else
        log "WARN" "模板文件处理失败，但继续安装"
    fi
    
    # 切换回原目录
    cd - >/dev/null
    return 0
}

register_services() {
    local target_base="${INSTALL_DIR}"
    local initd_dir="${target_base}/init.d"
    
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

download_docker_images() {
    local target_base="${INSTALL_DIR}"
    
    # 切换到目标目录
    cd "$target_base" || {
        log "ERROR" "无法切换到安装目录: $target_base"
        return 1
    }
    
    log "INFO" "开始下载Docker镜像..."
    
    # 检查镜像下载脚本是否存在
    if [[ ! -f "docker-download-images.sh" ]]; then
        log "ERROR" "镜像下载脚本不存在: docker-download-images.sh"
        cd - >/dev/null
        return 1
    fi
    
    # 检查Docker是否可用
    if ! command -v docker >/dev/null 2>&1; then
        log "WARN" "Docker未安装，跳过镜像下载"
        cd - >/dev/null
        return 0
    fi
    
    # 检查Docker服务是否运行
    if ! docker info >/dev/null 2>&1; then
        log "WARN" "Docker服务未运行，跳过镜像下载"
        cd - >/dev/null
        return 0
    fi
    
    # 执行镜像下载脚本
    log "INFO" "执行镜像下载..."
    if bash docker-download-images.sh; then
        log "INFO" "Docker镜像下载成功"
    else
        log "WARN" "Docker镜像下载失败，但继续安装"
    fi
    
    # 切换回原目录
    cd - >/dev/null
    return 0
}

wait_for_apisix_ready() {
    local max_attempts=30
    local wait_seconds=2
    local attempt=1
    
    source ./configure.sh
    
    log "INFO" "等待APISIX服务启动..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f http://$APISIX_ADDR/apisix/admin/routes -H "$AUTH" -H "$TYPE" -X GET > /dev/null 2>&1; then
            log "INFO" "APISIX服务已准备就绪，可以添加upstream配置"
            return 0
        fi
        log "INFO" "curl -s -f http://$APISIX_ADDR/apisix/admin/routes -H \"$AUTH\" -H \"$TYPE\" -X GET"
        log "INFO" "APISIX服务尚未就绪，等待${wait_seconds}秒后重试... (尝试 $attempt/$max_attempts)"
        sleep $wait_seconds
        attempt=$((attempt + 1))
    done
    
    log "ERROR" "APISIX服务在${max_attempts}次尝试后仍未准备就绪"
    return 1
}

configure_apisix_routes() {
    log "INFO" "配置APISIX路由..."
    
    # 等待APISIX服务就绪
    if ! wait_for_apisix_ready; then
        log "ERROR" "APISIX服务启动失败，无法继续配置"
        return 1
    fi
    
    # 配置APISIX路由
    local apisix_scripts=(
        #"apisix-ai-gateway.sh"
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
        #"apisix-quota-manager.sh"
    )
    
    for script in "${apisix_scripts[@]}"; do
        log "INFO" "执行APISIX配置: $script"
        if ! bash "$script"; then
            log "ERROR" "APISIX配置失败: $script"
            return 1
        fi
    done
    
    log "INFO" "APISIX路由配置完成"
    return 0
}

is_system_initialized() {
    # 检查系统初始化完成标记文件
    # 文件存在 = 系统已初始化，文件不存在 = 首次运行
    if [[ -f "$INITIALIZED_FLAG" ]]; then
        return 0  # 文件存在，系统已初始化
    fi
    return 1  # 文件不存在，首次运行
}

mark_system_initialized() {
    # 创建系统初始化完成标记文件
    touch "$INITIALIZED_FLAG"
    log "INFO" "已创建系统初始化完成标记文件: $INITIALIZED_FLAG"
}

# -------------------------- Main Logic --------------------------
main() {
    log "INFO" "安装脚本启动，日志文件: $LOG_FILE"
    
    # 解析命令行参数
    parse_arguments "$@"
    
    # 检查系统是否已初始化（非强制模式）
    if [ "$FORCE_REINIT" = false ] ; then
        if is_system_initialized; then
            log "WARN" "系统已初始化，无需重复执行初始化"
            log "INFO" "如需重新初始化系统，请使用 --force 选项"
            log "INFO" "可以使用以下命令强制重新初始化: bash init.sh --force"
            exit 0
        fi
    else
        log "INFO" "强制重新初始化模式，跳过初始化检查..."
    fi
    
    # 验证安装环境
    log "INFO" "执行环境检查脚本..."
    if ! bash check.sh --dir "${INSTALL_DIR}"; then
        log "WARN" "安装环境验证发现问题，但继续执行"
    fi

    # 注册系统服务
    if ! register_services; then
        log "WARN" "注册系统服务发现问题，但继续执行"
    fi
        
    # 修正目录权限
    if ! fix_permissions; then
        log "WARN" "权限修正失败，但继续执行"
    fi
    
    # 处理模板文件
    if ! process_template_files; then
        log "WARN" "模板文件处理失败，但继续执行"
    fi
    
    # 下载Docker镜像
    if ! download_docker_images; then
        log "WARN" "Docker镜像下载失败，请手动执行 ${INSTALL_DIR}/docker-download-images.sh"
    fi
    
    # 启动Docker Compose服务
    log "INFO" "启动Docker Compose服务..."
    cd "${target_base:-$INSTALL_DIR}" || {
        log "ERROR" "无法切换到安装目录"
        return 1
    }
    if ! docker-compose --env-file .env --env-file .images.env -f docker-compose.yml up -d; then
        log "ERROR" "Docker Compose服务启动失败"
        return 1
    fi
    log "INFO" "Docker Compose服务启动完成"
    
    # 配置APISIX路由
    if ! configure_apisix_routes; then
        log "WARN" "APISIX路由配置失败，但系统已启动"
    fi
    
    # 切换回原目录（如果在切换后）
    cd - >/dev/null || true
    
    # 标记系统已初始化完成
    mark_system_initialized
    
    log "INFO" "系统初始化完成！"
    log "INFO" "安装位置: ${INSTALL_DIR}"
    log "INFO" "后续步骤："
    log "INFO" "  1. 启动服务: cd ${INSTALL_DIR} && bash run.sh"
    log "INFO" "  2. 访问管理界面: http://${SERVER_IP}:39080/costrict-admin (默认账号: admin, 密码: admin)"
}

main "$@"