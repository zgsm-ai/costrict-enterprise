#!/bin/bash

log() {
    local level=$1
    local message=$2
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S")
    echo -e "[${timestamp}] [${level}] ${message}"
}

get_image_envs() {
    local output_file="$1"
    
    # 切换到目标目录
    cd "${FROM_DIR}" || {
        log "ERROR" "无法切换到目录: ${FROM_DIR}"
        return 1
    }
    
    log "INFO" "开始收集镜像环境配置..."
    
    # 清空或创建输出文件
    > "$output_file"
    
    # 查找所有 image.env 文件并合并内容
    local found_files=0
    while IFS= read -r env_file; do
        found_files=$((found_files + 1))
        log "INFO" "处理镜像配置文件: $env_file"
        
        # 将非空、非注释行追加到输出文件
        grep -v '^#' "$env_file" | grep -v '^$' >> "$output_file"
    done < <(find . -name "image.env" -type f)
    
    if [[ $found_files -eq 0 ]]; then
        log "WARN" "未找到任何 image.env 文件"
    else
        log "INFO" "已收集 ${found_files} 个镜像环境配置文件"
        log "INFO" "镜像环境配置已合并到: $output_file"
    fi
    
    # 切换回原目录
    cd - >/dev/null
    return 0
}

merge_env() {
    local target="$1"
    local source="$2"
    
    # 检查 source 文件是否存在
    if [[ ! -f "$source" ]]; then
        log "ERROR" "源文件不存在: $source"
        return 1
    fi
    
    # 检查 target 文件是否存在，不存在则创建
    if [[ ! -f "$target" ]]; then
        touch "$target"
    fi
    
    log "INFO" "开始合并环境配置: $source -> $target"
    
    # 读取 source 文件的非注释行和非空行，追加到 target 文件
    local line_count=0
    local line
    while IFS= read -r line || [[ -n "$line" ]]; do
        # 跳过注释行（以 # 开头，包括前面有空格的情况）
        if [[ "$line" =~ ^[[:space:]]*# ]]; then
            continue
        fi
        # 跳过纯空格的行
        if [[ -z "${line// /}" ]]; then
            continue
        fi
        
        # 追加到 target 文件
        echo "$line" >> "$target"
        line_count=$((line_count + 1))
    done < "$source"
    
    log "INFO" "已追加 ${line_count} 行到: $target"
    return 0
}

create_dot_env() {
    # 清空 .env 文件（如果存在）
    > ${dotenv_file} 2>/dev/null || :  # 使用 : 确保命令总是成功
    # 调用 merge_env，将 ".images.env" 和 "costrict.env" 合并到 ".env" 文件中
    merge_env ${dotenv_file} ${envs_file}
    merge_env ${dotenv_file} ${costrict_env_file}
    
    # source 加载 costrict.env，该文件中可能定义了COSTRICT_HOST，COSTRICT_BASEURL
    if [ -f ${costrict_env_file} ]; then
        source ${costrict_env_file}
    fi
    
    # 判断 COSTRICT_HOST 是否已定义，若未定义则添加到 .env
    if [ -z "${COSTRICT_HOST:-}" ]; then
        COSTRICT_HOST="${SERVER_IP}"
        echo "COSTRICT_HOST=\"${SERVER_IP}\"" >> ${dotenv_file}
    fi
    
    # 判断 COSTRICT_BASEURL 是否已定义，若未定义则添加到 .env
    if [ -z "${COSTRICT_BASEURL:-}" ]; then
        echo "COSTRICT_BASEURL=\"http://${COSTRICT_HOST}:${PORT_APISIX_ENTRY}\"" >> ${dotenv_file}
    fi
    
    return 0
}

# 使用getopt解析参数
TEMP=$(getopt -o o:f: --long output:,from: -n "$0" -- "$@")
eval set -- "$TEMP"

# 默认值
OUTPUT_DIR="."
FROM_DIR=$(pwd)

# 解析参数
while true ; do
    case "$1" in
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -f|--from)
            FROM_DIR="$2"
            shift 2
            ;;
        --) shift ; break ;;
        *) echo "参数解析错误" >&2 ; exit 1 ;;
    esac
done

# 确保output_dir以'/'结尾
[[ "${FROM_DIR: -1}" != "/" ]] && FROM_DIR="${FROM_DIR}/"
[[ "${OUTPUT_DIR: -1}" != "/" ]] && OUTPUT_DIR="${OUTPUT_DIR}/"

envs_file="${OUTPUT_DIR}.images.env"
list_file="${OUTPUT_DIR}.images.list"
dotenv_file="${OUTDIR_DIR}.env"
costrict_env_file="${FROM_DIR}costrict.env"

mkdir -p ${OUTPUT_DIR}

log "INFO" "生成镜像环境变量文件: ${envs_file} ..."
get_image_envs ${envs_file} || {
    log "ERROR" "生成镜像环境变量文件 ${envs_file} 失败"
    exit 1
}

# 从.images.env提取镜像列表
log "INFO" "生成镜像列表文件: $list_file ..."
awk -F'=' '{print $2}' "${envs_file}" > "${list_file}"

log "INFO" "生成环境变量文件 ${dotenv_file} ..."
if ! create_dot_env; then
    log "ERROR" "生成环境变量文件 ${dotenv_file} 失败"
    return 1
fi
