#!/bin/bash

# gen-secret.sh - 从指定源文件中提取内容，并把其中的转换标记替换成随机生成的密钥，输出到目标文件
# 
# 使用方法:
#   ./gen-secret.sh -i <源文件> -o <目标文件>
#
# 示例:
#   ./gen-secret.sh -i sample.env -o output.env

# 生成 base64 编码的随机密钥
gen_base64() {
    local size=$1
    local result=""
    local chars='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
    local chars_length=${#chars}
    
    # 生成 base64 随机字符串
    for ((i=0; i<size; i++)); do
        local index=$((RANDOM % chars_length))
        result="${result}${chars:index:1}"
    done
    
    echo "$result"
}

# 生成十六进制格式的随机密钥
gen_hex() {
    local size=$1
    local result=""
    local chars='0123456789abcdef'
    local chars_length=${#chars}
    
    # 生成十六进制随机字符串
    for ((i=0; i<size; i++)); do
        local index=$((RANDOM % chars_length))
        result="${result}${chars:index:1}"
    done
    
    echo "$result"
}

# 生成密码（包含字母、数字和特殊符号）
gen_password() {
    local size=$1
    local password=""
    local chars='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+-=[]{}|;:,.<>?'
    local chars_length=${#chars}
    
    # 生成随机密码
    for ((i=0; i<size; i++)); do
        local index=$((RANDOM % chars_length))
        password="${password}${chars:index:1}"
    done
    
    echo "$password"
}

# 替换文件中的密钥标记
replace_secrets() {
    local source_file=$1
    local target_file=$2
    
    if [[ ! -f "$source_file" ]]; then
        echo "[ERROR] 源文件不存在: $source_file"
        return 1
    fi
    
    echo "[INFO] 开始处理源文件: $source_file"
    
    # 清空目标文件（如果存在）
    > "$target_file"
    
    # 逐行处理文件
    local line_num=0
    local replaced_count=0
    
    while IFS= read -r line || [[ -n "$line" ]]; do
        local new_line="$line"
        
        # 查找行中的所有标记（不去重）- 更灵活的正则表达式
        while [[ "$new_line" =~ \{\{([a-zA-Z0-9_]+)[[:space:]]+([0-9]+)\}\} ]]; do
            local match="${BASH_REMATCH[0]}"
            local func="${BASH_REMATCH[1]}"
            local size="${BASH_REMATCH[2]}"
            
            # 根据函数类型生成对应的密钥
            local secret=""
            local func_lower=$(echo "$func" | tr '[:upper:]' '[:lower:]')
            case "$func_lower" in
                base64)
                    secret=$(gen_base64 "$size")
                    ;;
                hex)
                    secret=$(gen_hex "$size")
                    ;;
                password)
                    secret=$(gen_password "$size")
                    ;;
                *)
                    echo "[ERROR] 未知的密钥生成函数: $func"
                    return 1
                    ;;
            esac
            
            # 验证密钥生成是否成功
            if [[ -z "$secret" ]]; then
                echo "[ERROR] 生成密钥失败: $func $size"
                return 1
            fi
            
            echo "[INFO] 第 $((line_num + 1)) 行: 替换标记 $match 为 ${func} 密钥 (${#secret} 字符)"
            
            # 替换该行中的第一个匹配标记
            new_line="${new_line//$match/$secret}"
            replaced_count=$((replaced_count + 1))
        done
        
        # 写入处理后的行
        echo "$new_line" >> "$target_file"
        line_num=$((line_num + 1))
        
    done < "$source_file"
    
    if [[ $replaced_count -eq 0 ]]; then
        echo "[INFO] 文件中未找到密钥标记"
    else
        echo "[INFO] 已替换 $replaced_count 个密钥标记"
    fi
    
    echo "[INFO] 密钥替换完成，结果已写入: $target_file"
    return 0
}

# 显示使用说明
usage() {
    echo "Usage: gen-secret.sh [options]"
    echo "Generate random secrets from a template file"
    echo ""
    echo "Options:"
    echo "  -i <SOURCE_FILE>  Input source file (required)"
    echo "  -o <TARGET_FILE>  Output target file (required)"
    echo "  -f                Force overwrite existing target file"
    echo "  -h                Show this help message"
    echo ""
    echo "Description:"
    echo "  This script reads content from the source file, replaces secret generation"
    echo "  markers with randomly generated secrets, and writes the result to the target file."
    echo ""
    echo "Supported marker formats:"
    echo "  {{base64 <size>}}   Generate a base64-encoded random string"
    echo "  {{hex <size>}}      Generate a hexadecimal random string"
    echo "  {{password <size>}} Generate a password with letters, numbers, and special characters"
    echo ""
    echo "Examples:"
    echo "  # Generate secrets from template"
    echo "  gen-secret.sh -i sample.env -o output.env"
    echo ""
    echo "  # Force overwrite existing file"
    echo "  gen-secret.sh -i sample.env -o output.env -f"
    echo ""
    echo "  Example markers in source file:"
    echo '  PASSWORD_APISIX_DASHBOARD="{{password 12}}"'
    echo '  APIKEY_APISIX_ADMIN="{{base64 32}}"'
    echo '  APIKEY_APISIX_VIEWER="{{hex 32}}"'
}

# 默认值
SOURCE_FILE=""
TARGET_FILE=""
FORCE=false

# 使用 getopts 解析参数
while getopts "i:o:fh" opt
do
    case $opt in
        i)
            SOURCE_FILE="$OPTARG"
            ;;
        o)
            TARGET_FILE="$OPTARG"
            ;;
        f)
            FORCE=true
            ;;
        h)
            usage
            exit 0
            ;;
        ?)
            usage
            exit 1
            ;;
    esac
done

# 检查必需参数
if [[ -z "$SOURCE_FILE" || -z "$TARGET_FILE" ]]; then
    echo "[ERROR] 缺少必需参数"
    usage
    exit 1
fi

# 检查源文件
if [[ ! -f "$SOURCE_FILE" ]]; then
    echo "[ERROR] 源文件不存在: $SOURCE_FILE"
    exit 1
fi

# 检查目标文件是否存在，如果已存在且未启用 force 选项则退出
if [[ -f "$TARGET_FILE" && "$FORCE" != "true" ]]; then
    echo "[ERROR] 目标文件已存在: $TARGET_FILE"
    echo "[INFO] 使用 -f 选项强制覆盖"
    exit 1
fi

# 执行密钥替换
if replace_secrets "$SOURCE_FILE" "$TARGET_FILE"; then
    echo "[INFO] 成功生成密钥文件: $TARGET_FILE"
    exit 0
else
    echo "[ERROR] 生成密钥文件失败"
    exit 1
fi
