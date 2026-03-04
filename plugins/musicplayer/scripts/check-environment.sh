#!/bin/bash

# 音乐播放器环境检查脚本
# 用于诊断 Node.js 和服务配置问题

set -e

echo "================================"
echo "音乐播放器环境检查工具"
echo "================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 Node.js
echo "1. 检查 Node.js 安装..."
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo -e "   ${GREEN}✓${NC} Node.js 已安装: $NODE_VERSION"

    # 检查版本是否 >= 16
    MAJOR_VERSION=$(echo $NODE_VERSION | cut -d'v' -f 2 | cut -d'.' -f 1)
    if [ "$MAJOR_VERSION" -ge 16 ]; then
        echo -e "   ${GREEN}✓${NC} 版本符合要求 (≥ 16.0.0)"
    else
        echo -e "   ${RED}✗${NC} 版本过低 (当前: $NODE_VERSION, 需要: ≥ 16.0.0)"
        echo -e "   ${YELLOW}建议:${NC} 请从 https://nodejs.org/ 安装 LTS 版本"
        exit 1
    fi
else
    echo -e "   ${RED}✗${NC} Node.js 未安装"
    echo -e "   ${YELLOW}建议:${NC} 请从 https://nodejs.org/ 下载并安装 Node.js"
    exit 1
fi

echo ""

# 检查 npm
echo "2. 检查 npm 安装..."
if command -v npm &> /dev/null; then
    NPM_VERSION=$(npm --version)
    echo -e "   ${GREEN}✓${NC} npm 已安装: $NPM_VERSION"
else
    echo -e "   ${RED}✗${NC} npm 未安装"
    echo -e "   ${YELLOW}建议:${NC} npm 通常随 Node.js 一起安装，请检查安装"
    exit 1
fi

echo ""

# 检查服务目录
echo "3. 检查 lx-music-service 目录..."

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

# 检查多个可能的位置
SERVICE_DIRS=(
    "$PROJECT_ROOT/lx-music-service"
    "$(dirname "$0")/lx-music-service"
    "./lx-music-service"
)

SERVICE_FOUND=false
for DIR in "${SERVICE_DIRS[@]}"; do
    if [ -d "$DIR" ]; then
        echo -e "   ${GREEN}✓${NC} 找到服务目录: $DIR"
        SERVICE_DIR="$DIR"
        SERVICE_FOUND=true
        break
    fi
done

if [ "$SERVICE_FOUND" = false ]; then
    echo -e "   ${RED}✗${NC} 未找到 lx-music-service 目录"
    echo -e "   ${YELLOW}建议:${NC} 运行 'task package:production' 来准备依赖"
    exit 1
fi

echo ""

# 检查服务文件
echo "4. 检查服务文件..."
if [ -f "$SERVICE_DIR/server.js" ]; then
    echo -e "   ${GREEN}✓${NC} server.js 存在"
else
    echo -e "   ${RED}✗${NC} server.js 不存在"
    exit 1
fi

if [ -f "$SERVICE_DIR/package.json" ]; then
    echo -e "   ${GREEN}✓${NC} package.json 存在"
else
    echo -e "   ${RED}✗${NC} package.json 不存在"
    exit 1
fi

echo ""

# 检查 node_modules
echo "5. 检查依赖安装..."
if [ -d "$SERVICE_DIR/node_modules" ]; then
    NODE_MODULES_SIZE=$(du -sh "$SERVICE_DIR/node_modules" | cut -f1)
    echo -e "   ${GREEN}✓${NC} node_modules 已安装 (大小: $NODE_MODULES_SIZE)"

    # 检查关键依赖
    REQUIRED_DEPS=("axios" "cheerio" "crypto-js")
    ALL_DEPS_OK=true

    for DEP in "${REQUIRED_DEPS[@]}"; do
        if [ -d "$SERVICE_DIR/node_modules/$DEP" ]; then
            echo -e "   ${GREEN}  ✓${NC} $DEP 已安装"
        else
            echo -e "   ${RED}  ✗${NC} $DEP 未安装"
            ALL_DEPS_OK=false
        fi
    done

    if [ "$ALL_DEPS_OK" = false ]; then
        echo -e "   ${YELLOW}建议:${NC} 运行 'cd lx-music-service && npm install --production'"
        exit 1
    fi
else
    echo -e "   ${RED}✗${NC} node_modules 不存在"
    echo -e "   ${YELLOW}建议:${NC} 运行 'cd lx-music-service && npm install --production'"
    exit 1
fi

echo ""

# 测试服务启动
echo "6. 测试服务启动..."
cd "$SERVICE_DIR"

# 发送健康检查请求
HEALTH_RESPONSE=$(echo '{"method":"health","id":1}' | timeout 5s node server.js 2>&1)

if echo "$HEALTH_RESPONSE" | grep -q '"result":"ok"'; then
    echo -e "   ${GREEN}✓${NC} 服务健康检查通过"
else
    echo -e "   ${RED}✗${NC} 服务健康检查失败"
    echo "   响应: $HEALTH_RESPONSE"
    exit 1
fi

echo ""
echo "================================"
echo -e "${GREEN}所有检查通过！${NC}"
echo "================================"
echo ""
echo "音乐播放器插件已准备就绪。"
echo ""
