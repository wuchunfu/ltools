#!/bin/bash

# 生产环境打包辅助脚本
# 用于准备分发所需的所有文件

set -e

echo "================================"
echo "LTools 生产环境打包工具"
echo "================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取项目根目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

cd "$PROJECT_ROOT"

# 检查是否安装了 Task
if ! command -v task &> /dev/null; then
    echo -e "${RED}错误:${NC} 未找到 task 命令"
    echo -e "${YELLOW}请先安装 Taskfile:${NC}"
    echo "  sh -c \"\$(curl --location https://taskfile.dev/install.sh)\" -- -d -b ~/.local/bin"
    exit 1
fi

echo "步骤 1/4: 安装 lx-music-service 依赖..."
echo ""
cd lx-music-service

if [ -d "node_modules" ]; then
    echo -e "${YELLOW}node_modules 已存在，跳过安装${NC}"
    echo "如需重新安装，请先删除 node_modules 目录"
else
    echo "安装生产环境依赖..."
    npm install --production
    npm prune --production
fi

cd "$PROJECT_ROOT"
echo -e "${GREEN}✓ 依赖安装完成${NC}"
echo ""

echo "步骤 2/4: 构建应用程序..."
echo ""
task build
echo -e "${GREEN}✓ 应用构建完成${NC}"
echo ""

echo "步骤 3/4: 准备分发目录..."
echo ""

DIST_DIR="$PROJECT_ROOT/dist"
mkdir -p "$DIST_DIR"

# 复制二进制文件
if [ -d "bin" ]; then
    cp -r bin/* "$DIST_DIR/"
    echo -e "${GREEN}✓${NC} 已复制二进制文件到 dist/"
else
    echo -e "${RED}✗${NC} bin/ 目录不存在，请先运行构建"
    exit 1
fi

# 复制 lx-music-service
if [ -d "lx-music-service" ]; then
    cp -r lx-music-service "$DIST_DIR/"
    echo -e "${GREEN}✓${NC} 已复制 lx-music-service 到 dist/"
else
    echo -e "${RED}✗${NC} lx-music-service/ 目录不存在"
    exit 1
fi

# 创建 README
cat > "$DIST_DIR/README.txt" << EOF
LTools 音乐播放器插件 - 分发包
================================

系统要求:
- Node.js >= 16.0.0 (推荐 LTS 版本)
- 下载地址: https://nodejs.org/

安装说明:
1. 安装 Node.js (如果尚未安装)
2. 确保 lx-music-service 目录与主程序在同一目录
3. 运行主程序

目录结构:
.
├── ltools (或 ltools.exe)    # 主程序
└── lx-music-service/          # Node.js 音源服务
    ├── server.js
    ├── node_modules/
    └── ...

故障排查:
如果音乐播放器无法使用，请运行:
  ./plugins/musicplayer/scripts/check-environment.sh

或手动检查:
  node --version  # 应该 >= v16.0.0
  ls lx-music-service/  # 应该包含 server.js 和 node_modules/

更多信息:
- 文档: plugins/musicplayer/DISTRIBUTION.md
- 问题反馈: https://github.com/your-repo/ltools/issues
EOF

echo -e "${GREEN}✓${NC} 已创建 README.txt"
echo ""

echo "步骤 4/4: 生成分发信息..."
echo ""

# 计算目录大小
TOTAL_SIZE=$(du -sh "$DIST_DIR" | cut -f1)
SERVICE_SIZE=$(du -sh "$DIST_DIR/lx-music-service" | cut -f1)

echo -e "${BLUE}分发目录准备完成！${NC}"
echo ""
echo "目录位置: $DIST_DIR"
echo "总大小: $TOTAL_SIZE"
echo "  - lx-music-service: $SERVICE_SIZE"
echo ""
echo "包含文件:"
ls -lh "$DIST_DIR"
echo ""

echo -e "${YELLOW}下一步:${NC}"
echo "1. 检查 dist/ 目录内容"
echo "2. 测试运行: cd dist && ./ltools"
echo "3. 创建安装包:"
echo "   - macOS: 创建 .dmg 或 .app"
echo "   - Windows: 使用 NSIS 或 Inno Setup"
echo "   - Linux: 创建 .deb, .rpm 或 .AppImage"
echo ""
echo "注意: 用户需要预先安装 Node.js >= 16.0.0"
echo ""
