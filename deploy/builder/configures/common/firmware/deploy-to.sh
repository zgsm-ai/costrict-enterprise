#!/bin/bash
set -e
set -u
set -o pipefail 2>/dev/null || true

# -------------------------- Initialize Configuration --------------------------
SCRIPT_NAME=$(basename "$0")
LOG_FILE="${SCRIPT_NAME%.*}.log"
exec > >(tee -a "$LOG_FILE") 2>&1

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
--from <path>     部署源目录，默认为当前目录
--to <path>       部署目标目录，默认为 /usr/local/costrict
--manifest <file> 部署清单文件，默认为 MANIFEST
--clean           部署前清理目标目录，先删除目标目录下的所有文件和目录
-h, --help        显示帮助信息

MANIFEST 指令:
@@ <path>         设置目标相对路径（后续文件部署到此路径下）
@clean            清空当前目标相对路径下的所有内容

示例:
    $0                                            # 使用默认设置部署
    $0 --from /tmp/source                         # 指定源目录部署
    $0 --to /opt                                  # 部署到 /opt/costrict
    $0 --manifest custom.MANIFEST                 # 使用自定义清单文件
    $0 --from /tmp/source --to /opt/costrict --manifest custom.MANIFEST  # 完整示例
    $0 --to /opt/costrict --clean                 # 清理后再部署

MANIFEST 文件示例：
    # 设置目标相对路径为 config
    @@ config
    # 清空 config 目录下的所有内容
    @clean
    # 部署文件到 config 目录
    config.yaml
    app.conf

EOF
}

parse_arguments() {
    # 默认值
    INSTALL_FROM="$(pwd)"
    INSTALL_TO="/usr/local/costrict"
    MANIFEST_FILE="MANIFEST"
    CLEAN_TARGET=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --from)
                INSTALL_FROM="$2"
                shift 2
                ;;
            --to)
                INSTALL_TO="$2"
                shift 2
                ;;
            --manifest)
                MANIFEST_FILE="$2"
                shift 2
                ;;
            --clean)
                CLEAN_TARGET=true
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
    INSTALL_FROM="$(cd "$INSTALL_FROM" && pwd)"
    
    # 安全检查：确保 INSTALL_TO 不是危险目录
    check_dangerous_dirs "$INSTALL_TO"
    
    # 处理 MANIFEST_FILE 路径
    [[ "$MANIFEST_FILE" != /* ]] && MANIFEST_FILE="${INSTALL_FROM}/${MANIFEST_FILE}"
    
    log "INFO" "部署源: $INSTALL_FROM"
    log "INFO" "部署目标: $INSTALL_TO"
    log "INFO" "部署清单: $MANIFEST_FILE"
    if [[ "$CLEAN_TARGET" == "true" ]]; then
        log "INFO" "清理模式: 是"
    fi
}

check_dangerous_dirs() {
    local target_dir="$1"
    
    # 安全检查：确保参数非空
    if [[ -z "$target_dir" ]]; then
        log "ERROR" "安全错误: 目标目录参数为空，拒绝检查操作"
        exit 1
    fi
    
    # 安全检查：确保不是系统关键目录
    local dangerous_dirs=("/" "/usr" "/usr/local" "/bin" "/sbin" "/etc" "/home" "/root")
    for dangerous_dir in "${dangerous_dirs[@]}"; do
        if [[ "$(cd "$target_dir" && pwd)" == "$(cd "$dangerous_dir" && pwd)" ]]; then
            log "ERROR" "安全警告: 禁止部署到系统关键目录: $dangerous_dir"
            exit 1
        fi
    done
}

clean_target_directory() {
    local target_dir="$1"
    
    # 检查是否为危险目录，如果是则直接退出程序
    check_dangerous_dirs "$target_dir"
    
    log "INFO" "开始清理目标目录: $target_dir"
    
    # 检查目标目录是否存在
    if [[ ! -d "$target_dir" ]]; then
        log "INFO" "目标目录不存在，无需清理: $target_dir"
        return 0
    fi
    
    # 安全保护：使用路径变量保护语法，防止意外展开
    # 删除目标目录下的所有文件和目录
    log "INFO" "正在删除目标目录下的所有内容: $target_dir/*"
    if sudo rm -rf "${target_dir:?}"/* 2>/dev/null; then
        # 如果删除了所有内容，再次删除可能残留的隐藏文件
        sudo find "${target_dir:?}" -mindepth 1 -delete 2>/dev/null || true
        log "INFO" "目标目录清理完成"
        return 0
    else
        log "ERROR" "目标目录清理失败: $target_dir"
        return 1
    fi
}

validate_manifest() {
    if [[ ! -f "$MANIFEST_FILE" ]]; then
        log "ERROR" "部署清单文件不存在: $MANIFEST_FILE"
        return 1
    fi
    
    log "INFO" "找到部署清单文件: $MANIFEST_FILE"
    return 0
}

install_files() {
    local target_base="${INSTALL_TO}"
    local installed_count=0
    local failed_count=0
    
    log "INFO" "开始部署文件"
    
    # 创建目标根目录
    if ! sudo mkdir -p "$target_base"; then
        log "ERROR" "创建目标目录失败: $target_base"
        return 1
    fi
    
    local target_rel_path=""
    
    # 读取部署清单并逐行处理
    while read -r line; do
        # 处理 @@ 指令（设置目标相对路径）
        if [[ "$line" =~ ^[[:space:]]*@@[[:space:]]*(.*) ]]; then
            local directive_value="${BASH_REMATCH[1]}"
            # 规范化值（去除首尾空格）
            directive_value=$(echo "$directive_value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
            
            if [[ "$directive_value" == "~" ]] || [[ -z "$directive_value" ]]; then
                # @@~ 或 @@ 表示相对路径为空，即 target_base 本身
                target_rel_path=""
                log "INFO" "设置目标相对路径为空（target_base 本身）"
            else
                target_rel_path="/$directive_value"
                log "INFO" "设置目标相对路径为: $directive_value"
            fi
            continue
        fi
        
        # 处理 @clean 指令（清空当前目标路径下的内容）
        if [[ "$line" =~ ^[[:space:]]*@clean[[:space:]]*$ ]]; then
            local clean_path="${target_base}${target_rel_path}"
            log "INFO" "清理目录: $clean_path"
            
            # 检查目录是否存在
            if [[ ! -d "$clean_path" ]]; then
                log "INFO" "目录不存在，跳过清理: $clean_path"
            else
                # 删除目录下的所有内容（保留目录本身）
                if sudo rm -rf "${clean_path:?}"/* 2>/dev/null || sudo find "${clean_path}" -mindepth 1 -delete 2>/dev/null; then
                    log "INFO" "目录清理完成: $clean_path"
                else
                    log "ERROR" "目录清理失败: $clean_path"
                    failed_count=$((failed_count + 1))
                fi
            fi
            continue
        fi
        
        # 解析源文件和目标文件
        local source_rel=""
        local target_rel=""
        
        # 处理 "source -> target" 格式
        local arrow_pattern='^(.+)[[:space:]]*->[[:space:]]*(.*)'
        if [[ "$line" =~ $arrow_pattern ]]; then
            source_rel="${BASH_REMATCH[1]}"
            target_rel="${BASH_REMATCH[2]}"
        else
            # 单行只有一个源文件
            source_rel="$line"
        fi
        
        # 跳过空行和注释
        [[ -z "$source_rel" || "$source_rel" =~ ^[[:space:]]*# ]] && continue
        
        # 规范化路径（去除开头/结尾的空格）
        source_rel=$(echo "$source_rel" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        target_rel=$(echo "$target_rel" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        
        # 跳过空行
        [[ -z "$source_rel" ]] && continue
        
        # 构建源完整路径
        local source_full="${INSTALL_FROM}/${source_rel}"
        
        # 检查源文件是否存在
        if [[ ! -f "$source_full" && ! -d "$source_full" ]]; then
            log "ERROR" "源文件不存在: $source_full"
            failed_count=$((failed_count + 1))
            continue
        fi
        
        # 确定目标路径
        local target_full=""
        
        if [[ -z "$target_rel" ]]; then
            # 情况3: 目标路径未指定，和源文件同名
            target_full="${target_base}${target_rel_path}/${source_rel}"
        elif [[ "$target_rel" = /* ]]; then
            # 情况2: 目标路径是绝对路径，直接使用
            target_full="$target_rel"
        else
            # 情况1: 目标路径是相对路径，放到 ${target_base}${target_rel_path} 下
            target_full="${target_base}${target_rel_path}/${target_rel}"
        fi
        
        # 创建目标目录
        local target_dir=$(dirname "$target_full")
        if ! sudo mkdir -p "$target_dir"; then
            log "ERROR" "创建目标目录失败: $target_dir"
            failed_count=$((failed_count + 1))
            continue
        fi
        
        # 复制文件
        log "INFO" "复制: $source_rel -> $target_full"
        
        # 如果目标和源都是目录，使用 cp -r source/. target 合并内容
        if [[ -d "$source_full" && -d "$target_full" ]]; then
            sudo cp -r "${source_full}/." "$target_full"
        else
            sudo cp -r "$source_full" "$target_full"
        fi
        
        # 如果是文件，设置执行权限
        if [[ -f "$source_full" && -x "$source_full" ]]; then
            sudo chmod +x "$target_full"
        fi
        
        installed_count=$((installed_count + 1))
    done < "$MANIFEST_FILE"
    
    log "INFO" "文件部署完成"
    log "INFO" "成功: $installed_count 个文件/目录"
    if [[ $failed_count -gt 0 ]]; then
        log "WARN" "失败: $failed_count 个文件/目录"
    fi
    
    return 0
}

# -------------------------- Main Logic --------------------------
main() {
    log "INFO" "部署脚本启动，日志文件: $LOG_FILE"
    
    # 解析命令行参数
    parse_arguments "$@"
    
    # 清理目标目录（如果指定了 --clean 选项）
    if [[ "$CLEAN_TARGET" == "true" ]]; then
        if ! clean_target_directory "$INSTALL_TO"; then
            log "ERROR" "目标目录清理失败，部署中止"
            exit 1
        fi
    fi
    
    # 验证部署清单
    if ! validate_manifest; then
        log "ERROR" "部署清单验证失败"
        exit 1
    fi
    
    # 复制文件
    if ! install_files; then
        log "ERROR" "文件部署失败"
        exit 1
    fi
    
    log "INFO" "部署完成！"
    log "INFO" "部署位置: ${INSTALL_TO}"
    log "INFO" "后续步骤："
    log "INFO" "  1. 完成数据初始化: cd ${INSTALL_TO} && bash init.sh"
    log "INFO" "  2. 启动服务: cd ${INSTALL_TO} && bash run.sh"
}

main "$@"