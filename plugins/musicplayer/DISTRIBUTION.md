# 音乐播放器插件 - 生产环境构建指南

## 概述

音乐播放器插件依赖 Node.js 服务来提供音源支持。本文档说明如何在生产环境中自动化构建和分发。

## 架构说明

```
ltools/
├── bin/                      # 编译后的二进制文件
│   └── ltools               # 主程序
├── lx-music-service/         # Node.js 音源服务（需要随应用分发）
│   ├── server.js            # JSON-RPC 服务入口
│   ├── node_modules/        # 生产环境依赖
│   ├── sources/             # 音源插件
│   └── package.json
└── plugins/musicplayer/      # Go 插件实现
    ├── process_manager.go   # Node.js 进程管理
    ├── service_lx.go        # 前端服务接口
    └── lx_client.go         # JSON-RPC 客户端
```

## 系统要求

### 用户端要求
- **Node.js**: 版本 ≥ 16.0.0（推荐 LTS 版本）
- 下载地址: https://nodejs.org/

### 开发端要求
- Go 1.25+
- Node.js 16+ (用于构建)
- Task (taskfile) 构建工具

## 依赖管理

### 不提交到 Git 的文件

以下目录不会提交到版本控制：
- `frontend/node_modules/` - 前端依赖
- `lx-music-service/node_modules/` - Node 服务依赖
- `bin/` - 编译产物

**原因：**
- `node_modules` 文件数量多、体积大（~15MB）
- 可通过 `package.json` 和 `package-lock.json` 重建
- 不同平台的二进制文件可能不同

### 自动依赖安装

构建任务会自动检测并安装依赖：

```bash
# 首次构建会自动安装所有依赖
task build

# 手动安装 Node 服务依赖
task common:install:lx-music-service:deps
```

**Task 会自动跳过已安装的依赖：**
- 如果 `node_modules` 存在且 `package.json` 未变化，会跳过安装
- 这使得后续构建更快

## 生产环境构建

### 一键构建

```bash
# 完整的生产环境构建（包含依赖安装）
task package:production
```

这个命令会：
1. 安装 `lx-music-service` 的生产环境依赖
2. 构建应用程序
3. 显示分发说明

**注意：** `lx-music-service/node_modules` 不会提交到 git，首次构建时会自动安装。

### 首次克隆项目后

```bash
# 克隆仓库
git clone <repo-url>
cd ltools

# 首次构建（会自动安装所有依赖）
task build

# 或者只安装 Node 服务依赖
task common:install:lx-music-service:deps
```

## 分发打包

### macOS (.app)

```bash
# 构建 macOS 应用包
task darwin:package

# 分发结构
LTools.app/
├── Contents/
│   ├── MacOS/
│   │   └── ltools           # 主程序
│   └── Resources/
│       └── lx-music-service/ # Node 服务（需要添加）
```

**创建 DMG 前：**
```bash
# 确保 Node 服务包含在应用包中
cp -r lx-music-service LTools.app/Contents/Resources/
```

### Windows (.exe)

```bash
# 构建 Windows 可执行文件
task windows:build

# 分发结构
ltools/
├── ltools.exe               # 主程序
└── lx-music-service/        # Node 服务
```

**创建安装包：**
使用 NSIS 或 Inno Setup，确保包含 `lx-music-service/` 目录

### Linux (.AppImage / .deb)

```bash
# 构建 Linux 二进制
task linux:build

# 分发结构
ltools/
├── ltools                   # 主程序
└── lx-music-service/        # Node 服务
```

**创建 .deb 包：**
在 `debian/control` 中添加依赖：
```
Depends: nodejs (>= 16.0.0)
```

## 首次运行检测

应用启动时会自动检测：

1. **Node.js 安装检测**
   - 如果未安装：显示友好的错误提示和下载链接
   - 如果版本过低：提示升级到 Node.js 16+

2. **服务目录检测**
   - 开发环境：从项目根目录查找
   - 生产环境：从可执行文件旁边查找

3. **健康检查**
   - 启动后 5 秒内完成健康检查
   - 失败时记录详细错误日志

## 错误处理示例

### Node.js 未安装

```
错误: Node.js check failed: node.js not found at 'node': exec: "node": executable file not found in $PATH

The music player plugin requires Node.js to run the music source service.

Please install Node.js:
  • Download from: https://nodejs.org/
  • Recommended: LTS version (Long Term Support)
  • Minimum required: Node.js 16.0.0 or later

After installation, restart the application.
```

### Node.js 版本过低

```
错误: Node.js version 14.17.0 is too old.
Please upgrade to Node.js >= 16.0.0 (recommended: LTS version from https://nodejs.org)
```

### 服务目录缺失

```
错误: lx-music-service directory not found: /path/to/lx-music-service

This directory is required for music player functionality.
Expected location: next to the application executable or in the project root during development.

If you're a user:
  Please ensure the application was installed correctly with all bundled files.

If you're a developer:
  Run 'task package:production' to prepare all dependencies before building
```

## 优化建议

### 减小分发体积

```bash
# 只安装生产依赖
cd lx-music-service
npm install --production
npm prune --production

# 清理不必要的文件
rm -rf node_modules/.cache
rm -rf node_modules/.package-lock.json
```

### 嵌入式 Node.js（未来优化）

考虑使用以下方案消除用户端 Node.js 依赖：
- Node.js SEA (Single Executable Application)
- pkg 打包工具
- 嵌入平台特定的 Node 二进制

## CI/CD 集成示例

### GitHub Actions

```yaml
name: Build Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        os: [macos-latest, windows-latest, ubuntu-latest]

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Install Task
        run: sh -c "$(curl --location https://taskfile.dev/install.sh)"

      - name: Build Production
        run: ~/bin/task package:production

      - name: Package Distribution
        run: |
          mkdir -p dist
          cp -r bin/* dist/
          cp -r lx-music-service dist/

      - name: Upload Artifact
        uses: actions/upload-artifact@v3
        with:
          name: ltools-${{ matrix.os }}
          path: dist/
```

## 故障排查

### 开发环境

```bash
# 检查 Node 服务是否正常
cd lx-music-service
node server.js
# 应该启动并等待 stdin 输入

# 手动测试健康检查
echo '{"method":"health","id":1}' | node server.js
```

### 生产环境

```bash
# 检查服务目录位置
# 应该在可执行文件旁边
ls -l /path/to/ltools
ls -l /path/to/lx-music-service/

# 测试 Node.js 版本
node --version  # 应该 >= v16.0.0

# 查看应用日志
# 日志中会显示：
# [ProcessManager] Node.js version: vX.X.X
# [ProcessManager] Process started (PID: XXXX)
```

## 相关文档

- [插件系统架构](../../docs/design/plugin-system.md)
- [音乐播放器 API 文档](./README.md)
- [LX Music 项目](https://github.com/lyswhut/lx-music-mobile)
