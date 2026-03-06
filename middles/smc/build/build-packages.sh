#!/bin/bash

# 确保smc命令在PATH中
export PATH="$PATH:/root/.costrict/bin"

#
# 包管理系统的目录结构：
#
#-/-+-<package>/-+-<os>/-+-<arch>/-+-<ver>/-+-package.json: 对包数据文件进行签名保护
#   |            |       |         |        +-<package-data-file>
#   |            |       |         +-platform.json: 某个平台支持哪些版本
#   |            |       +-amd64-...
#   |            +-windows-...
#   |            +-platforms.json: 某个包支持哪些平台(OS&芯片架构)
#   +-packages.json: 系统有哪些包可以下载
#

usage() {
    echo "Usage: build-packages.sh [-p PACKAGE] [--packages PACKAGES] [--type TYPE] [--key KEY_FILE] [--clean] [--build] [--pack] [--index] [--upload] [--upload-packages] [--all]"
    echo "Options:"
    echo "  -p, --package        Package name (optional, if not specified, will process all packages)"
    echo "  --packages <list>    Package list (space-separated, e.g., \"pkg1 pkg2 pkg3\")"
    echo "  --type <type>        Package type filter (e.g., exec, conf, zip)"
    echo "  --key <key>          Private key file (default: costrict-private.pem)"
    echo "  --clean              Need clean first"
    echo "  --build              Need build packages"
    echo "  --pack               Need pack packages"
    echo "  --index              Need index packages"
    echo "  --upload             Need upload packages"
    echo "  --def                Execute default steps (build, pack, index)"
    echo "  --upload-packages    Need upload packages.json"
    echo "  --upload-to <env>    Upload package to <env>, env: def, all, prod, test, qianliu"
    echo "  -h, --help           Help information"
    exit 1
}

enable_upload() {
    case "$1" in
        def) NEED_UPLOAD=true; UPLOAD_PROD=true; UPLOAD_TEST=true; UPLOAD_QIANLIU=false;;
        all) NEED_UPLOAD=true; UPLOAD_PROD=true; UPLOAD_TEST=true; UPLOAD_QIANLIU=true;;
        prod) NEED_UPLOAD=true; UPLOAD_PROD=true;;
        test) NEED_UPLOAD=true; UPLOAD_TEST=true;;
        qianliu) NEED_UPLOAD=true; UPLOAD_QIANLIU=true;;
        *) usage;;
    esac
}
# 默认私钥文件
KEY_FILE="costrict-private.pem"

# 默认参数值
NEED_CLEAN=false
NEED_BUILD=false
NEED_PACK=false
NEED_INDEX=false
NEED_UPLOAD=false
NEED_UPLOAD_PACKAGES=false
UPLOAD_PROD=false
UPLOAD_TEST=false
UPLOAD_QIANLIU=false
PACKAGE_TYPE=""
PACKAGES=""

# Parse command line options
args=$(getopt -o hp:K: --long help,package:,packages:,kind:,type:,key:,clean,build,pack,index,def,upload,upload-packages,upload-to: -n 'build-packages.sh' -- "$@")
[ $? -ne 0 ] && usage

eval set -- "$args"

while true; do
    case "$1" in
        -p|--package) package="$2"; shift 2;;
        --packages) PACKAGES="$2"; shift 2;;
        --type) PACKAGE_TYPE="$2"; shift 2;;
        --key) KEY_FILE="$2"; shift 2;;
        --clean) NEED_CLEAN=true; shift;;
        --build) NEED_BUILD=true; shift;;
        --pack) NEED_PACK=true; shift;;
        --index) NEED_INDEX=true; shift;;
        --def) NEED_BUILD=true; NEED_PACK=true; NEED_INDEX=true; shift;;
        --upload) enable_upload "def"; shift;;
        --upload-packages) NEED_UPLOAD_PACKAGES=true; shift;;
        --upload-to) enable_upload "$2"; shift 2;;
        -h|--help) usage; exit 0;;
        --) shift; break;;
        *) usage;;
    esac
done

# Function to build a package for multiple platforms
build_app() {
    local package_name="$1"
    local version="$2"
    local source_dir="$3"
    local platforms_json="$4"

    # 获取当前路径的绝对路径
    local current_dir=$(pwd)   
    # 解析platforms数组
    local platform_count=$(echo "$platforms_json" | jq 'length')

    echo "Starting multi-platform build for package: $package_name, version: $version"
    echo ""
    echo "Source directory: $source_dir"
    echo "Building for $platform_count platform(s): $platforms_json"
    # 遍历每个平台
    local i
    for ((i=0; i<platform_count; i++)); do
        local os=$(echo "$platforms_json" | jq -r ".[$i].os")
        local arch=$(echo "$platforms_json" | jq -r ".[$i].arch")
        
        echo "==== Building $package_name for $os/$arch ===="
        
        # 创建输出目录
        local output_dir="$current_dir/packages/$package_name/$os/$arch/$version"
        mkdir -p "$output_dir"
        
        # 设置输出文件名
        local output_file="$package_name"
        if [ "$os" = "windows" ]; then
            output_file="$output_file.exe"
        fi
        
        # 完整输出路径
        local output_target="$output_dir/$output_file"
        
        echo "Output target: $output_target"
        
        # 到目标路径执行build.py
        (cd "$source_dir" && python ./build.py --software "$version" --os "$os" --arch "$arch" --output "$output_target")
        if [ $? -ne 0 ]; then
            echo "Build failed for $package_name on $os/$arch"
            exit 1
        fi
        echo ""
    done
    
    echo "All builds completed successfully for package: $package_name"
}

# Function to build configuration package directories
build_conf() {
    local package_name="$1"
    local version="$2"
    local source_dir="$3"
    local platforms_json="$4"
    
    # 获取当前路径的绝对路径
    local current_dir=$(pwd)
    # 解析platforms数组
    local platform_count=$(echo "$platforms_json" | jq 'length')
    # Get target from packages.json
    local target=$(jq -r ".builds[] | select(.name == \"${package_name}\") | .target" packages.json)
    if [ -z "$target" ] || [ "$target" = "null" ]; then
        echo "Error: 'target' not found for package '${package_name}' in packages.json!"
        exit 1
    fi
    
    echo "Starting configuration build for package: $package_name, version: $version"
    echo ""
    echo "Source directory: $source_dir"
    echo "Building for $platform_count platform(s): $platforms_json"
    
    # 遍历每个平台
    local i
    for ((i=0; i<platform_count; i++)); do
        local os=$(echo "$platforms_json" | jq -r ".[$i].os")
        local arch=$(echo "$platforms_json" | jq -r ".[$i].arch")
        
        echo "==== Building $package_name for $os/$arch ===="
        
        # 创建输出目录
        local output_dir="$current_dir/packages/$package_name/$os/$arch/$version"
        mkdir -p "$output_dir"
        
        # 源文件路径
        local source_file="$source_dir/$os/$arch/$target"
        # 目标文件路径
        local target_file="$output_dir/$target"
        
        # 检查源文件是否存在
        if [ ! -f "$source_file" ]; then
            if [ -f "$source_dir/common/$target" ]; then
                source_file="$source_dir/common/$target"
            else
                echo "Warning: Source file $source_file does not exist, skipping..."
                continue
            fi
        fi
        
        echo "Source file: $source_file"
        echo "Target file: $target_file"
        
        # 复制文件
        cp "$source_file" "$target_file"
        if [ $? -ne 0 ]; then
            echo "Error: Failed to copy $source_file to $target_file"
            exit 1
        fi
        
        echo "Successfully copied $source_file to $target_file"
        echo ""
    done
    
    echo "All configuration builds completed successfully for package: $package_name"
}

# Function to build zip package directories
build_zip() {
    local package_name="$1"
    local version="$2"
    local source_dir="$3"
    local platforms_json="$4"
    
    # 获取当前路径的绝对路径
    local current_dir=$(pwd)
    # 解析platforms数组
    local platform_count=$(echo "$platforms_json" | jq 'length')

    echo "Starting zip package build for package: $package_name, version: $version"
    echo ""
    echo "Source directory: $source_dir"
    echo "Building for $platform_count platform(s): $platforms_json"
    
    # 遍历每个平台
    local i
    for ((i=0; i<platform_count; i++)); do
        local os=$(echo "$platforms_json" | jq -r ".[$i].os")
        local arch=$(echo "$platforms_json" | jq -r ".[$i].arch")
        
        echo "==== Building $package_name for $os/$arch ===="
        
        # 创建输出目录
        local output_dir="$current_dir/packages/$package_name/$os/$arch/$version"
        mkdir -p "$output_dir"
        
        # 默认zip文件名：包名.zip
        local zip_filename="${package_name}.zip"
        # 目标zip文件路径
        local target_zip="$output_dir/$zip_filename"
        
        # 平台特定源目录路径
        local platform_source_dir="$source_dir/$os/$arch/$package_name"
        
        # 检查平台特定目录是否存在，如果不存在则使用common目录
        if [ ! -d "$platform_source_dir" ]; then
            echo "Warning: Platform directory $platform_source_dir does not exist"
            if [ -d "$source_dir/common/$package_name" ]; then
                platform_source_dir="$source_dir/common/$package_name"
                echo "Using common directory: $platform_source_dir"
            else
                echo "Warning: Common directory $source_dir/common does not exist either, skipping..."
                continue
            fi
        fi
        
        echo "Source directory: $platform_source_dir"
        echo "Target zip file: $target_zip"
        
        # 将源目录打包成zip文件
        (cd "$platform_source_dir" && zip -r "$target_zip" .)
        if [ $? -ne 0 ]; then
            echo "Error: Failed to create zip package for $package_name on $os/$arch"
            exit 1
        fi
        
        echo "Successfully created zip package: $target_zip"
        echo ""
    done
    
    echo "All zip packages built successfully for package: $package_name"
}

build_package() {
    local package="$1"
    
    # 从packages.json中获取指定包的版本号和路径
    local package_version=$(jq -r ".builds[] | select(.name == \"${package}\") | .version" packages.json)
    local package_path=$(jq -r ".builds[] | select(.name == \"${package}\") | .path" packages.json)
    local package_type=$(jq -r ".builds[] | select(.name == \"${package}\") | .type" packages.json)
    local package_platforms=$(jq -r ".builds[] | select(.name == \"${package}\") | .platforms" packages.json)

    if [ -z "$package_path" ] || [ "$package_path" = "null" ]; then
        echo "Skipping build step for ${package}..."
        return
    fi

    # 计算source_dir
    local current_dir=$(pwd)
    local source_dir="$current_dir/$package_path"
    # 检查源路径是否存在
    if [ ! -d "$source_dir" ]; then
        echo "Error: Source directory $source_dir does not exist!"
        exit 1
    fi
    
    if [ -z "$package_version" ] || [ "$package_version" = "null" ]; then
        echo "Error: Version not found for package '${package}' in packages.json!"
        exit 1
    fi

    if [ -z "$package_type" ] || [ "$package_type" = "null" ]; then
        echo "Error: 'type' not found for package '${package}' in packages.json!"
        exit 1
    fi
    
    # 如果没有定义platforms，则使用common
    if [ -z "$package_platforms" ] || [ "$package_platforms" = "null" ] || [ "$package_platforms" = "" ]; then
        package_platforms='[{"os":"common","arch":"common"}]'
    fi
    
    echo "=============================================="
    echo "Building package: $package, version: $package_version, path: $package_path, type: $package_type"
    echo "=============================================="
    if [ "exec" == "$package_type" ]; then
        build_app "${package}" "${package_version}" "${source_dir}" "${package_platforms}"
    elif [ "zip" == "$package_type" ]; then
        build_zip "${package}" "${package_version}" "${source_dir}" "${package_platforms}"
    else
        build_conf "${package}" "${package_version}" "${source_dir}" "${package_platforms}"
    fi
}

# Function to build multiple packages
build_packages() {
    local version="$1"

    # 从package-versions.json读取包信息
    echo "Reading package information from packages.json..."
    
    # 使用jq解析JSON
    local packages_json=$(cat packages.json)
    local package_count=$(echo "$packages_json" | jq '.builds | length')
    
    echo "Found $package_count packages to build"
    echo ""

    # 遍历每个包
    local i
    for ((i=0; i<package_count; i++)); do
        local package_name=$(echo "$packages_json" | jq -r ".builds[$i].name")

        build_package "${package_name}"
        echo ""
    done
    
    echo "All packages built successfully!"
}

# Function to get package type from packages.json
get_package_type() {
    local package_name=$1

    local package_type=$(jq -r ".builds[] | select(.name == \"${package_name}\") | .type // empty" packages.json)
    if [ -z "$package_type" ] || [ "$package_type" = "null" ]; then
        echo "exec"
    else
        echo "$package_type"
    fi
}

# Function to get package description from packages.json
get_package_description() {
    local package_name=$1

    local package_description=$(jq -r ".builds[] | select(.name == \"${package_name}\") | .description // empty" packages.json)
    if [ -z "$package_description" ] || [ "$package_description" = "null" ]; then
        echo "No description information"
    else
        echo "$package_description"
    fi
}

get_package_filename() {
    local package_name=$1

    local package_filename=$(jq -r ".builds[] | select(.name == \"${package_name}\") | .filename // empty" packages.json)
    if [ -z "$package_filename" ] || [ "$package_filename" = "null" ]; then
        echo ""
    else
        echo "$package_filename"
    fi
}

pack_package() {
    local package=$1
    local os=$2
    local arch=$3
    local ver=$4
    local file=$5
    local type=$6
    local description="$7"
    local filename="$8"
    
    echo "smc package build ${package} -f ${file} -k ${KEY_FILE} --os ${os} --arch ${arch} --version ${ver} --type ${type} --filename ${filename} --description ${description}"
    smc package build ${package} -f ${file} -k ${KEY_FILE} --os ${os} --arch ${arch} --version ${ver} --type ${type} --filename "${filename}" --description "${description}"
}

pack_dir_packages() {
    local package_dir=$1
    
    # 提取路径信息，先去掉末尾多余的/，再去掉开头多余的./
    local clean_packages=${package_dir%/}
    local clean_packages=${clean_packages#./}
    
    # 去掉开头的 'packages/' 基础路径
    local clean_packages=${clean_packages#packages/}
    
    local path_parts=(${clean_packages//\// })
    
    # 检查路径是否包含足够的部分
    if [ ${#path_parts[@]} -ne 4 ]; then
        echo "Internal Error: invalid directory: ${package_dir}"
        return 0
    fi
    
    # 从路径第一节获取包名
    local pkg_name=${path_parts[0]}
    local os=${path_parts[1]}
    local arch=${path_parts[2]}
    local ver=${path_parts[3]}
    
    echo "Processing: ${pkg_name}/${os}/${arch}/${ver} ..."
    
    local pkg_type=$(get_package_type "${pkg_name}")
    local pkg_description=$(get_package_description "${pkg_name}")
    local pkg_filename=$(get_package_filename "${pkg_name}")
    
    # 查找目录中非package.json的文件
    local file
    for file in "${package_dir}"*; do
        [ -f "${file}" ] || continue
        [ "$(basename "${file}")" = "package.json" ] && continue
        pack_package "${pkg_name}" "${os}" "${arch}" "${ver}" "${file}" "${pkg_type}" "${pkg_description}" "${pkg_filename}"
    done
}

index_packages() {
    local dir=$1
    
    echo "smc package index -b ${dir}"
    smc package index -b "${dir}"
}

# Function to clean up old version directories for a package
cleanup_old_versions() {
    local package_name="$1"
        
    # 从package-versions.json中获取指定包的版本号
    local target_version=$(jq -r ".builds[] | select(.name == \"${package_name}\") | .version" packages.json)
    local target_path=$(jq -r ".builds[] | select(.name == \"${package_name}\") | .path" packages.json)
    
    if [ -z "$target_version" ] || [ "$target_version" = "null" ]; then
        echo "Skipping clean step for package '${package_name}'..."
       return 0
    fi
    if [ -z "$target_path" ] || [ "$target_path" = "null" ]; then
        echo "Skipping clean step for package '${package_name}'..."
       return 0
    fi
    
    echo "Cleaning up old versions for package: $package_name, keeping version: $target_version"
    
    # 检查包目录是否存在
    if [ ! -d "packages/${package_name}" ]; then
        echo "Warning: Package directory 'packages/${package_name}' not found, skipping clean."
        return 0
    fi
    
    # 遍历所有平台和架构目录
    local os_dir
    local arch_dir
    local version_dir
    for os_dir in "packages/${package_name}"/*/; do
        [ -d "${os_dir}" ] || continue
        
        local os=$(basename "${os_dir}")
        
        for arch_dir in "${os_dir}"*/; do
            [ -d "${arch_dir}" ] || continue
            
            local arch=$(basename "${arch_dir}")
            
            # 遍历所有版本目录
            for version_dir in "${arch_dir}"*/; do
                [ -d "${version_dir}" ] || continue
                
                local version=$(basename "${version_dir}")
                
                # 如果不是目标版本，则删除
                if [ "$version" != "$target_version" ]; then
                    echo "Removing old version: ${package_name}/${os}/${arch}/${version}"
                    rm -rf "${version_dir}"
                    if [ $? -eq 0 ]; then
                        echo "Successfully removed: ${version_dir}"
                    else
                        echo "Error: Failed to remove ${version_dir}"
                    fi
                else
                    echo "Keeping target version: ${package_name}/${os}/${arch}/${version}"
                fi
            done
        done
    done
    
    echo "Cleanup completed for package: $package_name"
}

# Function to clean up old versions for all packages
cleanup_all_old_versions() {
    # 从package-versions.json读取包信息
    echo "Reading package information from packages.json for clean..."
    
    # 使用jq解析JSON
    local packages_json=$(cat packages.json)
    local package_count=$(echo "$packages_json" | jq '.builds | length')
    
    echo "Found $package_count packages to clean"
    echo ""
    
    # 遍历每个包
    local i
    for ((i=0; i<package_count; i++)); do
        local package_name=$(echo "$packages_json" | jq -r ".builds[$i].name")
        
        echo "=============================================="
        echo "Cleaning up package: $package_name"
        echo "=============================================="
        cleanup_old_versions "$package_name"
        if [ $? -ne 0 ]; then
            echo "Cleanup failed for package: $package_name"
            exit 1
        fi
        echo ""
    done
    
    echo "All packages clean completed!"
}

upload_package() {
    local source_dir=$1
    local package=$2
    local ip=$3
    local port=$4
    local rootDir=$5

    local formalDir="${rootDir}/costrict"
    local uploadDir="${rootDir}/costrict-upload"

    local package_path="${source_dir}/${package}"

    echo rsync -avzP -e "ssh -p ${port}" ${package_path} "root@${ip}:${uploadDir}/"
    rsync -avzP -e "ssh -p ${port}" ${package_path} "root@${ip}:${uploadDir}/"

    ssh -p "${port}" "root@${ip}" <<EOF
        set -e
        echo "Transfer ${package} to formal directory..."
        if [ -d "${formalDir}/${package}" ]; then
            mv "${formalDir}/${package}" "${uploadDir}/${package}-tmp"
        fi
        mv "${uploadDir}/${package}" "${formalDir}/${package}"
        if [ -d "${uploadDir}/${package}-tmp" ]; then
            mv "${uploadDir}/${package}-tmp" "${uploadDir}/${package}"
        fi
EOF
}

upload_package_clouds() {
    local source_dir=$1
    local package=$2

    source ./.env

    if [ "$UPLOAD_TEST" = true ]; then
        echo "=============================================="
        echo "Upload package $package to ${test_host}..."
        echo "=============================================="
        upload_package "${source_dir}" "${package}" "${test_host}" "${test_port}" "${test_path}"
    fi

    if [ "$UPLOAD_PROD" = true ]; then
        echo "=============================================="
        echo "Upload package $package to ${prod_host}..."
        echo "=============================================="
        upload_package "${source_dir}" "${package}" "${prod_host}" "${prod_port}" "${prod_path}"
    fi

    if [ "$UPLOAD_QIANLIU" = true ]; then
        echo "=============================================="
        echo "Upload package $package to ${qianliu_host}..."
        echo "=============================================="
        upload_package "${source_dir}" "${package}" "${qianliu_host}" "${qianliu_port}" "${qianliu_path}"
    fi
}

process_package() {
    local package_name=$1
    # 处理指定包
    mkdir -p "packages/${package_name}"

    if [ "$NEED_CLEAN" = true ]; then
        echo "Cleaning up old versions for package: $package_name"
        cleanup_old_versions "$package_name"
    else
        echo "Skipping clean step for ${package_name}..."
    fi

    if [ "$NEED_BUILD" = true ]; then
        echo "Building target for ${package_name}..."
        build_package "${package_name}"
    else
        echo "Skipping build step for ${package_name}..."
    fi

    if [ "$NEED_PACK" = true ]; then
        echo "Building package.json for ${package_name}..."
        # 检查私钥文件是否存在
        if [ ! -f "${KEY_FILE}" ]; then
            echo "Error: Private key file '${KEY_FILE}' not found!"
            exit 1
        fi
        for package_dir in "packages/${package_name}"/*/*/*/; do
            [ -d "${package_dir}" ] || continue
            pack_dir_packages "${package_dir}"
        done
    else
        echo "Skipping package step for ${package_name}..."
    fi

    if [ "$NEED_INDEX" = true ]; then
        echo "Building index for ${package_name}..."
        index_packages "packages/${package_name}"
    else
        echo "Skipping index step for ${package_name}..."
    fi

    if [ "$NEED_UPLOAD" = true ]; then
        echo "Uploading package: $package_name"
        upload_package_clouds "packages" "${package_name}"
    fi
}

# Function to process multiple packages
process_packages() {
    local packages=$1
    
    # 解析包列表（支持空格分隔的包名）
    local package_list=()
    if [ -n "$packages" ]; then
        IFS=' ' read -ra package_list <<< "$packages"
    fi
    
    # 如果包列表为空，从packages.json读取所有包
    if [ ${#package_list[@]} -eq 0 ]; then
        echo "No packages specified, reading from packages.json..."
        local packages_json=$(cat packages.json)
        local package_count=$(echo "$packages_json" | jq '.builds | length')
        
        for ((i=0; i<package_count; i++)); do
            package_list+=("$(echo "$packages_json" | jq -r ".builds[$i].name")")
        done
    fi
    
    echo "Processing ${#package_list[@]} package(s): ${package_list[*]}"
    echo ""
    
    # 遍历每个包并处理
    for pkg in "${package_list[@]}"; do
        echo "=============================================="
        echo "Processing package: $pkg"
        echo "=============================================="
        process_package "$pkg"
        echo ""
    done
    
    echo "All packages processed successfully!"
}

# Function to process packages by type
process_type() {
    local target_type=$1
    
    if [ -z "$target_type" ]; then
        echo "Error: Target type is empty!"
        exit 1
    fi
    
    echo "Processing packages of type: $target_type"
    echo ""
    
    # 使用jq解析JSON
    local packages_json=$(cat packages.json)
    local package_count=$(echo "$packages_json" | jq '.builds | length')
    
    local processed_count=0
    local i
    for ((i=0; i<package_count; i++)); do
        local package_name=$(echo "$packages_json" | jq -r ".builds[$i].name")
        local package_type=$(echo "$packages_json" | jq -r ".builds[$i].type // empty")
        
        if [ "$package_type" = "$target_type" ]; then
            echo "=============================================="
            echo "Processing package: $package_name (type: $package_type)"
            echo "=============================================="
            process_package "$package_name"
            echo ""
            ((processed_count++))
        fi
    done
    
    echo "Processed $processed_count package(s) of type '$target_type'"
}

if [ "$NEED_CLEAN" = true ] || [ "$NEED_BUILD" = true ] || [ "$NEED_PACK" = true ]; then
    # 检查jq工具是否可用
    if ! command -v jq >/dev/null 2>&1; then
        echo "Error: jq command not found! Please install jq to parse JSON files."
        echo "Installation instructions:"
        echo "  Ubuntu/Debian: sudo apt-get install jq"
        echo "  CentOS/RHEL: sudo yum install jq"
        echo "  macOS: brew install jq"
        echo "  Windows: Download from https://stedolan.github.io/jq/download/"
        exit 1
    fi
    # 检查是否有模块定义文件packages.json
    if [ ! -f "packages.json" ]; then
        echo "Error: packages.json file not found!"
        exit 1
    fi
fi

if [ "$NEED_UPLOAD_PACKAGES" = true ]; then
    echo "Uploading packages.json..."
    upload_package_clouds "." "packages.json"
    exit 0
fi

if [ -n "$PACKAGE_TYPE" ]; then
    # 处理指定类型的包
    process_type "$PACKAGE_TYPE"
elif [ -n "$PACKAGES" ]; then
    # 处理指定的包列表
    process_packages "$PACKAGES"
elif [ -z "$package" ]; then
    # 处理所有包
    process_packages ""
else
    process_package "$package"
fi

echo "Build completed."
