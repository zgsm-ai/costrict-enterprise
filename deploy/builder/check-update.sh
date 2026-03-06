#!/bin/bash

# check-update.sh - 检测packages.json中包的版本和内容变化，更新包版本
#
# 功能：
# - 遍历packages.json中的builds数组
# - 计算包path所指目录的CHECKSUM和文件数
# - 比较当前版本和checksum与latest.json中的记录
# - 根据变化情况输出提示或更新latest.json
# --update: 当checksum变化时自动更新版本号(递增包的patch版本号)

set -e

# 默认参数值
UPDATE_VERSION=false
VERBOSE=false
PACKAGES=""

# 配置文件路径
PACKAGES_JSON="packages.json"
LATEST_JSON="latest.json"

# 打印帮助信息的函数
print_usage() {
    echo "Usage: check-update.sh [-u|--update] [-p|--packages PACKAGE1,PACKAGE2,...] [-h|--help] [-v|--verbose]"
    echo "Options:"
    echo "  -u, --update          Update package version when checksum changes"
    echo "  -p, --packages        Only check specified packages (comma-separated list)"
    echo "  -v, --verbose         Show checksum calculation details for each file"
    echo "  -h, --help            Show this help message"
}

log() {
    local level=$1
    local message=$2
    echo -e "[${level}] ${message}" >&2
}

prompt() {
    local message=$1
    echo -e "${message}" >&2
}

# 从文件列表计算CHECKSUM（通用函数）
# 输入：通过参数传递的文件列表（可变参数）
# 输出格式：第一行checksum，第二行文件数
calculate_checksum_from_file_list() {
    # 如果没有传参，返回空
    if [ $# -eq 0 ]; then
        prompt "no any files"
        echo ""
        echo "0"
        return
    fi
    
    # 统计文件数
    local file_count=$#
    
    # 计算checksum
    local sha256_output=$(printf "%s\n" "$@" | xargs sha256sum 2>/dev/null | sort)
    local checksum=$(echo "$sha256_output" | sha256sum | awk '{print $1}')
    
    # 如果开启了verbose模式，输出每个文件的checksum
    if [ "$VERBOSE" = true ]; then
        prompt "$sha256_output"
    fi
    
    # 输出两行：checksum, file_count
    echo "$checksum"
    echo "$file_count"
}

# 计算目录的CHECKSUM（用于exec、zip类型）
# 输出格式：第一行checksum，第二行文件数
calculate_directory_checksum() {
    local dir="$1"
    
    if [ ! -d "$dir" ]; then
        prompt "$dir is not a directory"
        echo ""
        echo "0"
        return
    fi
    
    # 使用find获取文件列表并缓存在数组中
    local file_list=()
    while IFS= read -r -d '' file; do
        file_list+=("$file")
    done < <(find "$dir" -type f -print0 2>/dev/null)
    
    # 调用通用函数计算checksum
    calculate_checksum_from_file_list "${file_list[@]}"
}

# 计算目录中Go相关文件的CHECKSUM（用于exec类型）
# 输出格式：第一行checksum，第二行文件数
calculate_go_directory_checksum() {
    local dir="$1"
    
    if [ ! -d "$dir" ]; then
        prompt "$dir is not a directory"
        echo ""
        echo "0"
        return
    fi
    
    # 使用find获取Go文件列表并缓存在数组中
    local file_list=()
    while IFS= read -r -d '' file; do
        file_list+=("$file")
    done < <(find "$dir" -type f \( -name "*.go" -o -name "*.mod" \) ! -path "*/.git/*" -print0 2>/dev/null)
    
    # 调用通用函数计算checksum
    calculate_checksum_from_file_list "${file_list[@]}"
}

# 计算conf类型包的CHECKSUM（跨所有平台）
# 输出格式：第一行checksum，第二行文件数
calculate_conf_package_checksum() {
    local path="$1"
    local target="$2"
    
    local file_list=()
    
    # 定义平台列表
    local platforms=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64" "windows/arm64" "common/common")
    
    # 收集所有平台的文件
    for platform in "${platforms[@]}"; do
        local file_path="$path/$platform/$target"
        if [ -f "$file_path" ]; then
            file_list+=("$file_path")
        fi
    done
    
    # 添加 common 目录的文件（如果存在）
    local common_file="$path/common/$target"
    if [ -f "$common_file" ]; then
        file_list+=("$common_file")
    fi

    # 调用通用函数计算checksum
    calculate_checksum_from_file_list "${file_list[@]}"
}

# 计算zip类型包的CHECKSUM（跨所有平台）
# 输出格式：第一行checksum，第二行文件数
calculate_zip_package_checksum() {
    local path="$1"
    local package_name="$2"
    
    local file_list=()
    
    # 定义平台列表
    local platforms=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64" "windows/arm64" "common/common")
    
    # 收集所有平台的文件
    for platform in "${platforms[@]}"; do
        local dir_path="$path/$platform/$package_name"
        if [ -d "$dir_path" ]; then
            while IFS= read -r -d '' file; do
                file_list+=("$file")
            done < <(find "$dir_path" -type f -print0 2>/dev/null)
        fi
    done
    
    # 添加 common 目录中的包目录（如果存在）
    local common_dir="$path/common/$package_name"
    if [ -d "$common_dir" ]; then
        while IFS= read -r -d '' file; do
            file_list+=("$file")
        done < <(find "$common_dir" -type f -print0 2>/dev/null)
    fi
    
    # 调用通用函数计算checksum
    calculate_checksum_from_file_list "${file_list[@]}"
}

# 递增包的patch版本号
increment_patch_version() {
    local package_name="$1"
    local current_version="$2"

    # 自动递增 patch 版本号
    local MAJOR=$(echo "$current_version" | cut -d'.' -f1)
    local MINOR=$(echo "$current_version" | cut -d'.' -f2)
    local PATCH=$(echo "$current_version" | cut -d'.' -f3)
    local NEW_PATCH=$((PATCH + 1))
    local NEW_VERSION="$MAJOR.$MINOR.$NEW_PATCH"

    # 使用 jq 更新packages.json中的版本号
    jq "(.builds[] | select(.name == \"$package_name\") | .version) |= \"$NEW_VERSION\"" "$PACKAGES_JSON" > "$PACKAGES_JSON.tmp"

    if [ $? -ne 0 ]; then
        rm -f "$PACKAGES_JSON.tmp"
        log "ERROR" "Failed to update package version in packages.json"
        exit 1
    fi
    mv "$PACKAGES_JSON.tmp" "$PACKAGES_JSON"
    
    echo "$NEW_VERSION"
}

# 主函数
main() {
    prompt "=============================================="
    prompt "Checking package updates from $PACKAGES_JSON"
    prompt "=============================================="
    prompt ""
    
    # 使用jq解析JSON
    local packages_json=$(cat "$PACKAGES_JSON")
    local package_count=$(echo "$packages_json" | jq '.builds | length')
    
    prompt "Found $package_count packages"
    prompt ""
    
    # 遍历每个包
    local i
    local modified_packages=()
    
    # 如果指定了packages选项，则将逗号分隔的字符串转换为数组
    local target_packages=()
    if [ -n "$PACKAGES" ]; then
        # 将逗号分隔的字符串转换为数组
        IFS=',' read -ra target_packages <<< "$PACKAGES"
        log "INFO" "Checking only specified packages: ${target_packages[*]}"
    fi
    
    for ((i=0; i<package_count; i++)); do
        local package_name=$(echo "$packages_json" | jq -r ".builds[$i].name")
        
        # 如果指定了packages选项，检查当前包是否在目标列表中
        if [ ${#target_packages[@]} -gt 0 ]; then
            local found=false
            for target in "${target_packages[@]}"; do
                if [ "$package_name" = "$target" ]; then
                    found=true
                    break
                fi
            done
            if [ "$found" = false ]; then
                continue
            fi
        fi
        local package_version=$(echo "$packages_json" | jq -r ".builds[$i].version")
        local package_path=$(echo "$packages_json" | jq -r ".builds[$i].path")
        local package_type=$(echo "$packages_json" | jq -r ".builds[$i].type")
        local package_target=$(echo "$packages_json" | jq -r ".builds[$i].target")
        
        prompt "Processing package: $package_name"
        
        # 检查version字段是否存在
        if [ -z "$package_version" ] || [ "$package_version" = "null" ]; then
            log "WARN" "No version found for package '$package_name', skipping..."
            prompt ""
            continue
        fi
        
        # 检查path字段是否存在
        if [ -z "$package_path" ] || [ "$package_path" = "null" ]; then
            log "WARN" "No path found for package '$package_name', skipping..."
            prompt ""
            continue
        fi
        
        # 根据包类型计算CHECKSUM和文件数
        local new_checksum=""
        local new_file_count=0
        
        if [ "$package_type" = "conf" ]; then
            # conf类型：path、target跨所有平台查找文件
            if [ -z "$package_target" ] || [ "$package_target" = "null" ]; then
                log "WARN" "No target found for conf package '$package_name', skipping..."
                prompt ""
                continue
            fi
            local result=$(calculate_conf_package_checksum "$package_path" "$package_target")
            new_checksum=$(echo "$result" | head -n1)
            new_file_count=$(echo "$result" | tail -n1)
        elif [ "$package_type" = "zip" ]; then
            # zip类型：跨所有平台查找目录
            local result=$(calculate_zip_package_checksum "$package_path" "$package_name")
            new_checksum=$(echo "$result" | head -n1)
            new_file_count=$(echo "$result" | tail -n1)
        elif [ "$package_type" = "exec" ]; then
            # exec类型：只扫描Go相关文件（.go和.mod文件）
            local full_path="$package_path"
            local result=$(calculate_go_directory_checksum "$full_path")
            new_checksum=$(echo "$result" | head -n1)
            new_file_count=$(echo "$result" | tail -n1)
        else
            # 其他类型：扫描目录下所有文件
            local full_path="$package_path"
            local result=$(calculate_directory_checksum "$full_path")
            new_checksum=$(echo "$result" | head -n1)
            new_file_count=$(echo "$result" | tail -n1)
        fi
        
        if [ -z "$new_checksum" ]; then
            log "ERROR" "Failed to calculate checksum for package '$package_name' at path '$full_path'"
            prompt ""
            continue
        fi
        
        # 从latest.json中读取之前的版本和checksum
        local old_version=$(jq -r ".\"$package_name\".version // \"null\"" "$LATEST_JSON")
        local old_checksum=$(jq -r ".\"$package_name\".checksum // \"null\"" "$LATEST_JSON")
        
        prompt "Package '$package_name': files=$new_file_count"
        
        # 比较
        if [ "$old_version" = "null" ]; then
            # 首次记录
            log "INFO" "First time recording package '$package_name': version=$package_version"
            jq ".\"$package_name\" = {\"version\": \"$package_version\", \"checksum\": \"$new_checksum\", \"file_count\": $new_file_count}" "$LATEST_JSON" > "$LATEST_JSON.tmp"
            mv "$LATEST_JSON.tmp" "$LATEST_JSON"
            modified_packages+=("$package_name")
        elif [ "$package_version" = "$old_version" ]; then
            # 版本号未变
            if [ "$new_checksum" = "$old_checksum" ]; then
                # 版本号和CHECKSUM都没变
                log "INFO" "No changes for package '$package_name'"
            else
                # 版本号未变但CHECKSUM变了
                if [ "$UPDATE_VERSION" = true ]; then
                    # 启用自动版本递增
                    local new_version=$(increment_patch_version "$package_name" "$package_version")
                    log "MODIFIED" "Update version for '$package_name': $package_version -> $new_version"
                    
                    # 更新latest.json
                    jq ".\"$package_name\" = {\"version\": \"$new_version\", \"checksum\": \"$new_checksum\", \"file_count\": $new_file_count}" "$LATEST_JSON" > "$LATEST_JSON.tmp"
                    mv "$LATEST_JSON.tmp" "$LATEST_JSON"
                    modified_packages+=("$package_name")
                else
                    # 未启用自动版本递增，仅记录
                    log "MODIFIED" "Module '$package_name' has been modified (version=$package_version, checksum changed)"
                    modified_packages+=("$package_name")
                fi
            fi
        else
            # 版本号变了
            if [ "$new_checksum" = "$old_checksum" ]; then
                # 版本号变了但CHECKSUM未变（这是不正常的，但也记录）
                log "WARN" "Module '$package_name' version changed but checksum didn't: $old_version -> $package_version"
            else
                # 版本号和CHECKSUM都变了
                log "INFO" "Module '$package_name' version updated: $old_version -> $package_version"
            fi
            
            # 更新latest.json
            jq ".\"$package_name\" = {\"version\": \"$package_version\", \"checksum\": \"$new_checksum\", \"file_count\": $new_file_count}" "$LATEST_JSON" > "$LATEST_JSON.tmp"
            mv "$LATEST_JSON.tmp" "$LATEST_JSON"
            modified_packages+=("$package_name")
        fi
        
        prompt ""
    done
    
    prompt "=============================================="
    if [ ${#modified_packages[@]} -gt 0 ]; then
        prompt "Check completed. $LATEST_JSON has been updated."
    else
        prompt "Check completed. No updates found."
    fi
    prompt "=============================================="
    
    # 输出所有发生变化的包名到标准输出（以逗号分隔）
    if [ ${#modified_packages[@]} -gt 0 ]; then
        IFS=',' echo "${modified_packages[*]}"
    fi
}

# Parse command line options
args=$(getopt -o uhp:v --long help,update,packages:,verbose -n 'check-update.sh' -- "$@")
[ $? -ne 0 ] && print_usage && exit 1

eval set -- "$args"

while true; do
    case "$1" in
        -u|--update) UPDATE_VERSION=true; shift;;
        -p|--packages) PACKAGES="$2"; shift 2;;
        -v|--verbose) VERBOSE=true; shift;;
        -h|--help) print_usage; exit 0;;
        --) shift; break;;
        *) print_usage; exit 1;;
    esac
done

# 检查packages.json是否存在
if [ ! -f "$PACKAGES_JSON" ]; then
    log "ERROR" "$PACKAGES_JSON not found!"
    exit 1
fi

# 初始化或读取latest.json
if [ ! -f "$LATEST_JSON" ]; then
    log "INFO" "$LATEST_JSON not found, creating new one..."
    echo "{}" > "$LATEST_JSON"
fi

main
