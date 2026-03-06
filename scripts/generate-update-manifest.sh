#!/bin/bash

# 生成更新清单脚本
# 用于生成 Wails v3 自动更新所需的 update.json 文件

set -e

# 配置
VERSION="${1:-}"
UPDATE_DIR="${2:-./updates}"
BASE_URL="${3:-https://updates.ltools.app/stable}"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version> [update-dir] [base-url]"
  echo "Example: $0 0.2.0 ./updates https://updates.ltools.app/stable"
  exit 1
fi

echo "================================"
echo "生成更新清单"
echo "================================"
echo "版本: $VERSION"
echo "目录: $UPDATE_DIR"
echo "URL: $BASE_URL"
echo ""

# 创建更新目录
mkdir -p "$UPDATE_DIR"
mkdir -p "$UPDATE_DIR/patches"

# 检查平台包是否存在
check_package() {
  local platform=$1
  local ext=$2
  local file="$UPDATE_DIR/ltools-$VERSION-$platform.$ext"

  if [ ! -f "$file" ]; then
    echo "警告: 未找到 $platform 包: $file"
    return 1
  fi

  echo "✓ 找到: $file ($(du -h "$file" | cut -f1))"
  return 0
}

# 计算文件 SHA256
calculate_checksum() {
  local file=$1

  if command -v shasum &> /dev/null; then
    shasum -a 256 "$file" | cut -d' ' -f1
  elif command -v sha256sum &> /dev/null; then
    sha256sum "$file" | cut -d' ' -f1
  else
    echo "错误: 未找到 shasum 或 sha256sum 命令"
    exit 1
  fi
}

# 检查所有平台包
echo "检查平台包..."
PACKAGES_FOUND=0

if check_package "darwin-arm64" "tar.gz"; then PACKAGES_FOUND=$((PACKAGES_FOUND + 1)); fi
if check_package "darwin-amd64" "tar.gz"; then PACKAGES_FOUND=$((PACKAGES_FOUND + 1)); fi
if check_package "windows-amd64-installer" "exe"; then PACKAGES_FOUND=$((PACKAGES_FOUND + 1)); fi
if check_package "linux-amd64" "AppImage"; then PACKAGES_FOUND=$((PACKAGES_FOUND + 1)); fi

if [ $PACKAGES_FOUND -eq 0 ]; then
  echo ""
  echo "错误: 未找到任何平台包"
  echo "请先运行构建命令:"
  echo "  task darwin:package"
  echo "  task windows:package"
  echo "  task linux:package"
  exit 1
fi

echo ""
echo "找到 $PACKAGES_FOUND 个平台包"
echo ""

# 收集所有平台信息
PLATFORM_INFOS=()

check_and_add_platform() {
  local platform=$1
  local ext=$2
  local file="$UPDATE_DIR/ltools-$VERSION-$platform.$ext"

  if [ -f "$file" ]; then
    local size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null)
    local checksum=$(calculate_checksum "$file")

    # 从平台标识中移除 "-installer" 后缀（Windows 特殊处理）
    # 这样 "windows-amd64-installer" 会映射到 "windows-amd64" 平台键
    local platform_key="${platform%-installer}"

    # 单行格式，稍后会被逗号分隔
    PLATFORM_INFOS+=("    \"$platform_key\": {\"url\": \"$BASE_URL/ltools-$VERSION-$platform.$ext\", \"size\": $size, \"checksum\": \"sha256:$checksum\"}")
  fi
}

check_and_add_platform "darwin-arm64" "tar.gz"
check_and_add_platform "darwin-amd64" "tar.gz"
check_and_add_platform "windows-amd64-installer" "exe"
check_and_add_platform "linux-amd64" "AppImage"

# 生成 update.json，用逗号连接平台信息
OUTPUT_FILE="$UPDATE_DIR/update.json"

# 移除版本号前缀的 'v'（如果有）
CLEAN_VERSION="${VERSION#v}"

# 生成 platforms 部分，用逗号和换行符连接
PLATFORMS_JSON=""
for i in "${!PLATFORM_INFOS[@]}"; do
  if [ $i -gt 0 ]; then
    PLATFORMS_JSON+=",\n"
  fi
  PLATFORMS_JSON+="${PLATFORM_INFOS[$i]}"
done

cat > "$OUTPUT_FILE" <<EOF
{
  "version": "$CLEAN_VERSION",
  "releaseDate": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "releaseNotes": "## 新功能\n- 自动更新机制\n\n## 改进\n- 性能优化\n\n## 修复\n- Bug 修复",
  "mandatory": false,
  "platforms": {
$(echo -e "$PLATFORMS_JSON")
  }
}
EOF

# 如果安装了 jq，美化 JSON
if command -v jq &> /dev/null; then
  jq '.' "$OUTPUT_FILE" > "$OUTPUT_FILE.tmp" && mv "$OUTPUT_FILE.tmp" "$OUTPUT_FILE"
fi

echo "================================"
echo "✓ 生成成功"
echo "================================"
echo "文件: $OUTPUT_FILE"
echo ""
echo "内容预览:"
echo "---"
cat "$OUTPUT_FILE"
echo ""
echo "---"
echo ""
echo "下一步:"
echo "1. 编辑 $OUTPUT_FILE 更新 releaseNotes"
echo "2. 上传到更新服务器:"
echo "   scp -r $UPDATE_DIR/* user@server:/var/www/updates/"
echo "3. 或使用 GitHub Releases / S3 等托管服务"
