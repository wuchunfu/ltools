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

# 生成平台信息
generate_platform_info() {
  local platform=$1
  local ext=$2
  local file="$UPDATE_DIR/ltools-$VERSION-$platform.$ext"

  if [ ! -f "$file" ]; then
    return
  fi

  local size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null)
  local checksum=$(calculate_checksum "$file")

  cat <<EOF
    "$platform": {
      "url": "$BASE_URL/ltools-$VERSION-$platform.$ext",
      "size": $size,
      "checksum": "sha256:$checksum"
    }
EOF
}

# 检查所有平台包
echo "检查平台包..."
PACKAGES_FOUND=0

if check_package "darwin-arm64" "tar.gz"; then PACKAGES_FOUND=$((PACKAGES_FOUND + 1)); fi
if check_package "darwin-amd64" "tar.gz"; then PACKAGES_FOUND=$((PACKAGES_FOUND + 1)); fi
if check_package "windows-amd64-portable" "zip"; then PACKAGES_FOUND=$((PACKAGES_FOUND + 1)); fi
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

# 生成 update.json
OUTPUT_FILE="$UPDATE_DIR/update.json"

cat > "$OUTPUT_FILE" <<EOF
{
  "version": "$VERSION",
  "releaseDate": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "releaseNotes": "## 新功能\n- 自动更新机制\n\n## 改进\n- 性能优化\n\n## 修复\n- Bug 修复",
  "mandatory": false,
  "platforms": {
$(generate_platform_info "darwin-arm64" "tar.gz")$(generate_platform_info "darwin-amd64" "tar.gz")$(generate_platform_info "windows-amd64-portable" "zip")$(generate_platform_info "linux-amd64" "AppImage")
  }
}
EOF

# 移除多余的逗号（JSON 格式化）
sed -i '' 's/},$/}/g' "$OUTPUT_FILE" 2>/dev/null || sed -i 's/},$/}/g' "$OUTPUT_FILE"

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
