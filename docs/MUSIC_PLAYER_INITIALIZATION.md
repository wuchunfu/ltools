# 音乐播放器插件初始化流程

## 概述

音乐播放器插件 (`musicplayer.builtin`) 是一个基于 LX Music 的随机音乐播放器，支持网易云、QQ 音乐、酷狗等多个平台。本文档详细说明该插件在**开发环境**和**构建环境**中的初始化流程。

---

## 一、开发环境 (Development Mode)

### 1. 前端初始化流程

#### 1.1 启动方式
```bash
# 方式 1: 使用 Wails 开发模式（推荐）
task dev
# 或
wails3 dev -config ./build/config.yml -port 9245

# 方式 2: 仅启动前端开发服务器
cd frontend && npm run dev
```

#### 1.2 前端加载流程
```
1. Vite Dev Server 启动 (http://localhost:9245)
   ↓
2. React 应用加载 (main.tsx)
   ↓
3. React Router 初始化
   ↓
4. 用户访问音乐播放器页面
   ├─ 路由: /plugins/musicplayer.builtin
   ├─ 组件: MusicPlayerPage.tsx
   └─ Widget: MusicPlayerWidget.tsx
   ↓
5. 前端调用后端 Service (通过 Wails 绑定)
   ├─ musicplayer.Search(keyword, page)
   ├─ musicplayer.GetRandomSongs(count)
   ├─ musicplayer.GetSongURLWithMetadata(song, quality)
   └─ musicplayer.GetLyric(song)
```

#### 1.3 关键特性
- **热重载**: 前端代码修改实时更新
- **Source Map**: 完整的调试信息
- **开发工具**: React DevTools、浏览器调试器可用
- **API 绑定**: `frontend/bindings/` 中的自动生成绑定

---

### 2. 后端初始化流程

#### 2.1 服务启动序列
```
main.go: main()
   ↓
1. 创建插件实例
   musicPlayerPlugin := musicplayer.NewMusicPlayerPlugin()
   ↓
2. 注册插件到 PluginManager
   pluginManager.Register(musicPlayerPlugin)
   ↓
3. 创建 ServiceLX 实例
   musicPlayerServiceLX, err := musicplayer.NewServiceLX(
       musicPlayerPlugin,
       app,
       musicPlayerProxyAdapter
   )
   ↓
4. 注册 Service 到 App
   app.RegisterService(application.NewService(musicPlayerServiceLX))
```

#### 2.2 ServiceLX 初始化细节
```go
// service_lx.go: NewServiceLX()
func NewServiceLX(...) (*ServiceLX, error) {
    // 1. 初始化配置管理器
    configManager, err := NewConfigManager()

    // 2. 初始化窗口管理器
    windowManager := NewWindowManager(plugin, app)

    // 3. 初始化缓存管理器
    cacheManager, err := NewCacheManager(&CacheConfig{
        MemoryCapacity: 1000,
        DefaultTTL:     30 * time.Minute,
    })

    // 4. 初始化源管理器
    sourceManager := NewSourceManager()

    // 5. 初始化进程管理器（核心）
    processManager, err := NewProcessManager(&ProcessManagerConfig{
        NodePath: "node", // 使用系统 Node.js
    })

    // 6. 启动 LX Music 服务进程
    if err := service.startLXService(); err != nil {
        return nil, err
    }

    return service, nil
}
```

#### 2.3 进程管理器初始化（开发环境）
```go
// process_manager.go: NewProcessManager()
func NewProcessManager(config *ProcessManagerConfig) (*ProcessManager, error) {
    // 1. 查找 Node.js 可执行文件
    actualNodePath, err := findNodeExecutable()

    // 2. 查找服务目录（开发环境）
    possiblePaths := []struct {
        path string
        desc string
    }{
        // 开发环境路径优先级：
        // 6. CWD/lx-music-service/dist (bundle 模式)
        {filepath.Join(cwd, "lx-music-service", "dist"),
         "Development (CWD/lx-music-service/dist)"},

        // 7. CWD/lx-music-service (完整模式)
        {filepath.Join(cwd, "lx-music-service"),
         "Development (CWD/lx-music-service)"},

        // 其他备选路径...
    }

    // 3. 检查服务目录是否存在
    if _, err := os.Stat(config.ServiceDir); os.IsNotExist(err) {
        // 尝试提取嵌入的服务（bundle 模式）
        extractedDir, extractErr := extractEmbeddedService()
        if extractErr == nil {
            config.ServiceDir = extractedDir
        }
    }

    return &ProcessManager{...}, nil
}
```

#### 2.4 Node.js 进程启动（开发环境）
```go
// process_manager.go: Start()
func (pm *ProcessManager) Start() error {
    // 1. 确定启动文件
    bundlePath := filepath.Join(pm.serviceDir, "server.bundle.js")
    serverPath := "server.js" // 默认

    if _, err := os.Stat(bundlePath); err == nil {
        // Bundle 模式（优化后）
        serverPath = "server.bundle.js"
    } else {
        // 完整模式（开发环境）
        serverPath = "server.js" // 或 server-simple.js
    }

    // 2. 启动 Node.js 进程
    pm.cmd = exec.Command(pm.nodePath, serverPath)
    pm.cmd.Dir = pm.serviceDir

    // 3. 设置标准输入输出管道
    pm.stdin, _ = pm.cmd.StdinPipe()
    pm.stdout, _ = pm.cmd.StdoutPipe()
    pm.stderr, _ = pm.cmd.StderrPipe()

    // 4. 启动进程
    pm.cmd.Start()
}
```

#### 2.5 开发环境特点
- **直接运行**: 使用源代码目录中的 `lx-music-service/`
- **完整依赖**: 包含完整的 `node_modules/` (约 15MB)
- **实时调试**: Node.js 进程日志输出到 stderr
- **热重载**: 修改 JS 代码需要重启 Go 应用

---

## 二、构建环境 (Build/Production Mode)

### 1. 构建流程

#### 1.1 前端构建
```bash
task build
# 或
task darwin:build  # macOS
```

构建步骤：
```
1. 安装前端依赖
   task install:frontend:deps
   cd frontend && npm install

2. 生成 Go 绑定
   task generate:bindings
   wails3 generate bindings -clean=true -ts

3. 构建前端资源
   cd frontend && npm run build
   ↓
   输出: frontend/dist/
   - index.html
   - assets/*.js
   - assets/*.css
```

#### 1.2 后端构建
```bash
# macOS
task darwin:build

# 构建步骤:
1. Go mod tidy
2. 构建前端 (task build:frontend)
3. 构建 Go 二进制文件
   go build -tags production -trimpath -ldflags="-w -s"
   ↓
   输出: bin/ltools (18MB)
```

#### 1.3 LX Music Service 构建（Bundle 优化）
```bash
task common:build:lx-music-service
```

构建步骤：
```
1. 安装依赖
   cd lx-music-service && npm install

2. 打包 Bundle
   npm run build
   ↓
   esbuild 配置 (build.js):
   - entryPoints: ['./server-simple.js']
   - bundle: true
   - platform: node
   - minify: true
   ↓
   输出: lx-music-service/dist/
   - server.bundle.js (13KB) ✅ 优化后
   - sources/元力kw1.1.0.js (41KB)
   - sources/开心汽水_0.1.5_鸿蒙.js (34KB)
   ↓
   总大小: 96KB (原 15MB)
```

#### 1.4 嵌入到 Go 二进制
```go
// plugins/musicplayer/embed.go

//go:embed all:lx-music-service/dist
//go:embed lx-music-service/sources/*
var lxMusicServiceFS embed.FS
```

嵌入内容：
- `lx-music-service/dist/server.bundle.js` (13KB)
- `lx-music-service/dist/sources/*.js` (75KB)
- `lx-music-service/sources/*.js` (75KB，备份)

嵌入后二进制大小：
- 原始: 18MB
- 嵌入后: 20MB (+176KB)

---

### 2. 打包流程 (macOS .app Bundle)

#### 2.1 创建 .app 包结构
```bash
task darwin:package
```

目录结构：
```
bin/ltools.app/
├── Contents/
│   ├── MacOS/
│   │   └── ltools (20MB, 含嵌入的服务)
│   ├── Resources/
│   │   ├── icons.icns
│   │   └── lx-music-service/
│   │       └── dist/
│   │           ├── server.bundle.js (13KB)
│   │           └── sources/ (75KB)
│   └── Info.plist
```

#### 2.2 复制 LX Music Service
```yaml
# build/darwin/Taskfile.yml
copy:music-service:
  deps:
    - task: common:build:lx-music-service
  cmds:
    - mkdir -p "bin/ltools.app/Contents/Resources/lx-music-service/dist"
    - cp -r lx-music-service/dist/* "{{.BIN_DIR}}/{{.APP_NAME}}.app/Contents/Resources/lx-music-service/dist/"
```

---

### 3. 运行时初始化流程

#### 3.1 应用启动
```
1. 用户双击 ltools.app
   ↓
2. macOS 加载 ltools 二进制
   ↓
3. main() 函数执行
   ↓
4. 初始化音乐播放器服务
   musicplayer.NewServiceLX()
```

#### 3.2 服务目录查找（生产环境）
```go
// process_manager.go: NewProcessManager()
possiblePaths := []struct {
    path string
    desc string
}{
    // 1. .app bundle (bundle 模式，优先)
    {filepath.Join(appDir, "..", "Resources", "lx-music-service", "dist"),
     "macOS .app bundle (bundle mode)"},

    // 2. .app bundle (完整模式)
    {filepath.Join(appDir, "..", "Resources", "lx-music-service"),
     "macOS .app bundle (full)"},

    // 3. 可执行文件同目录 (bundle 模式)
    {filepath.Join(appDir, "lx-music-service", "dist"),
     "Same directory as executable (bundle mode)"},

    // 其他备选路径...
}
```

#### 3.3 嵌入服务提取（首次运行）
```go
// embed.go: extractEmbeddedService()
func extractEmbeddedService() (string, error) {
    // 1. 确定目标目录
    configDir, _ := os.UserConfigDir()
    targetDir := filepath.Join(configDir, "ltools", "lx-music-service")
    // macOS: ~/Library/Application Support/ltools/lx-music-service

    // 2. 检查是否已提取
    bundlePath := filepath.Join(targetDir, "server.bundle.js")
    if _, err := os.Stat(bundlePath); err == nil {
        return targetDir, nil // 已存在，直接使用
    }

    // 3. 首次运行，提取服务
    log.Printf("Extracting embedded lx-music-service to: %s", targetDir)
    os.MkdirAll(targetDir, 0755)

    // 4. 从 embed.FS 提取文件
    extractDir("lx-music-service/dist", targetDir)

    return targetDir, nil
}
```

提取后的目录结构：
```
~/Library/Application Support/ltools/lx-music-service/
├── server.bundle.js (13KB)
└── sources/
    ├── 元力kw1.1.0.js (41KB)
    └── 开心汽水_0.1.5_鸿蒙.js (34KB)
```

#### 3.4 Node.js 进程启动（生产环境）
```go
// process_manager.go: Start()
func (pm *ProcessManager) Start() error {
    // 1. 确定启动文件
    // 优先使用 bundle 模式
    serverPath := "server.bundle.js"

    // 2. 启动 Node.js 进程
    pm.cmd = exec.Command(pm.nodePath, serverPath)
    pm.cmd.Dir = pm.serviceDir

    // 3. 启动进程
    pm.cmd.Start()

    // 4. 等待初始化完成
    time.Sleep(500 * time.Millisecond)
}
```

---

## 三、对比总结

| 维度 | 开发环境 | 构建环境 |
|------|----------|----------|
| **前端资源** | Vite Dev Server (localhost:9245) | 静态文件 (frontend/dist/) |
| **热重载** | ✅ 支持 | ❌ 不支持 |
| **后端二进制** | 动态编译 | 静态编译 + 优化 |
| **LX Service 位置** | 项目根目录 `lx-music-service/` | .app bundle + 用户配置目录 |
| **依赖方式** | 完整 `node_modules/` (15MB) | Bundle 单文件 (13KB) |
| **首次运行** | 直接使用源码 | 提取嵌入的服务 |
| **调试工具** | ✅ 完整 (DevTools, 日志) | ⚠️ 有限 (日志) |
| **总体积** | ~500MB (含依赖) | ~20MB (.app bundle) |

---

## 四、关键优化点

### 4.1 Bundle 优化
- **工具**: esbuild
- **效果**: 15MB → 13KB (99.9% 减少)
- **原理**: 打包所有依赖到单文件，移除 `node_modules/`

### 4.2 嵌入优化
- **工具**: Go embed
- **内容**: 仅嵌入 `dist/` 和 `sources/`
- **大小**: 176KB (原 15MB+)

### 4.3 提取策略
- **时机**: 首次运行时提取
- **位置**: 用户配置目录
- **缓存**: 后续运行直接使用提取的文件

---

## 五、故障排查

### 5.1 开发环境问题

**问题**: 找不到 `lx-music-service` 目录
```
Error: lx-music-service directory not found
```

**解决方案**:
```bash
# 1. 确认在项目根目录运行
pwd  # 应该显示 .../ltools

# 2. 检查目录是否存在
ls -la lx-music-service/

# 3. 安装依赖
cd lx-music-service && npm install
```

---

### 5.2 构建环境问题

**问题**: Bundle 未生成
```
Error: lx-music-service/dist/server.bundle.js not found
```

**解决方案**:
```bash
# 1. 构建 bundle
task common:build:lx-music-service

# 2. 验证输出
ls -lh lx-music-service/dist/server.bundle.js
# 应该显示: -rw-r--r-- 1 user staff 13K ... server.bundle.js

# 3. 重新打包应用
task darwin:package
```

---

**问题**: 嵌入提取失败
```
Error: failed to extract dist: ...
```

**调试步骤**:
```bash
# 1. 检查用户配置目录
ls -la ~/Library/Application\ Support/ltools/lx-music-service/

# 2. 删除已提取的文件（强制重新提取）
rm -rf ~/Library/Application\ Support/ltools/lx-music-service/

# 3. 重新运行应用
open bin/ltools.app

# 4. 查看日志（应该显示提取过程）
# [ProcessManager] Extracting embedded lx-music-service to: ...
# [ProcessManager] Extracted: server.bundle.js (7774 bytes)
```

---

**问题**: Node.js 版本过旧
```
[ERROR] Node.js version 10.15.3 is too old. Please upgrade to Node.js >= 16.0.0
```

**原因**: 系统安装了多个 Node.js 版本，应用选中了旧版本。

**解决方案**:
```bash
# 1. 检查已安装的 Node.js 版本
ls ~/.nvm/versions/node/

# 2. 确保有 >= 16.0.0 的版本
nvm install --lts

# 3. 验证应用能找到正确版本
# 日志应该显示:
# [ProcessManager] Found 7 nvm versions (sorted by version)
# [ProcessManager] ✅ Found Node.js at candidate #0: .../v22.21.1/bin/node
```

---

**问题**: Node.js 进程启动失败
```
Error: failed to start LX process: ...
```

**调试步骤**:
```bash
# 1. 检查 Node.js 是否安装
node --version  # 应该 >= 16.0.0

# 2. 手动测试 bundle
node ~/Library/Application\ Support/ltools/lx-music-service/server.bundle.js

# 3. 检查进程日志
# 应用日志中应该显示:
# [ProcessManager] Server path: server.bundle.js
# [ProcessManager] Working directory: /Users/.../lx-music-service
```

---

## 六、性能指标

### 6.1 启动时间
- **开发环境**: ~2 秒（包含热重载检查）
- **构建环境**: ~0.5 秒（直接启动）

### 6.2 内存占用
- **开发环境**: ~150MB (Go + Node.js + Vite)
- **构建环境**: ~100MB (Go + Node.js)

### 6.3 磁盘占用
- **开发环境**: ~15MB (lx-music-service/ 含 node_modules)
- **构建环境**: ~1.8MB (.app bundle 中的 lx-music-service/dist/)
- **嵌入后二进制**: 增加 ~1.8MB

---

## 七、未来优化方向

### 7.1 短期优化
- [ ] 添加 Bundle 校验机制（防止损坏）
- [ ] 支持自动更新 LX Music Service
- [ ] 优化提取速度（并行提取）

### 7.2 长期优化
- [ ] 探索 Node.js SEA (Single Executable Application)
- [ ] 支持 WebAssembly 替代 Node.js
- [ ] 实现纯 Go 实现的音乐服务

---

## 八、参考资料

### 8.1 内部文档
- [插件系统架构](../design/plugin-system.md)
- [设计系统](../DESIGN_SYSTEM.md)
- [窗口行为](../WAILS_WINDOW_BEHAVIOR.md)

### 8.2 外部资源
- [Wails v3 文档](https://v3.wails.io/)
- [esbuild 文档](https://esbuild.github.io/)
- [Go embed 文档](https://pkg.go.dev/embed)

### 8.3 相关代码
- `plugins/musicplayer/` - 音乐播放器插件实现
- `lx-music-service/` - Node.js 音乐服务
- `build/Taskfile.yml` - 构建任务配置
- `build/darwin/Taskfile.yml` - macOS 打包配置

---

**文档版本**: 1.1
**最后更新**: 2026-03-06
**作者**: LTools Team

---

## 附录 A：环境变量与双击启动

### A.1 问题描述

在 macOS 上，通过不同方式启动应用会导致不同的环境变量：

| 启动方式 | PATH 环境变量 | 工作目录 |
|---------|-------------|---------|
| `open bin/ltools.app` | 继承 shell 的 PATH（包含 nvm） | 当前目录 |
| 双击 `ltools.app` | 最小化 PATH（`/usr/bin:/bin`） | `/`（根目录） |

### A.2 解决方案

#### 问题 1: 剪贴板插件日志路径

**症状**：双击启动时，剪贴板插件初始化失败：
```
[Clipboard Plugin] Failed to create logs directory: mkdir logs: read-only file system
```

**原因**：工作目录为 `/`，无法创建相对路径 `logs/` 目录。

**解决方案**：使用 `os.UserConfigDir()` 获取用户配置目录：
```go
// plugins/clipboard/clipboard.go
configDir, _ := os.UserConfigDir()
logsDir := filepath.Join(configDir, "ltools", "logs")
// macOS: ~/Library/Application Support/ltools/logs
```

#### 问题 2: Node.js 版本查找

**症状**：双击启动时，音乐播放器服务无法启动：
```
[ERROR] Node.js version 10.15.3 is too old. Please upgrade to Node.js >= 16.0.0
```

**原因**：PATH 中没有 node，回退到查找 nvm 目录时，选中了旧版本。

**解决方案**：对 nvm 版本按版本号排序，优先使用最新版本：
```go
// plugins/musicplayer/process_manager.go
sort.Slice(matches, func(i, j int) bool {
    vi := extractVersionFromPath(matches[i])
    vj := extractVersionFromPath(matches[j])
    return compareVersions(vi, vj) > 0  // 降序排列
})
```

### A.3 验证修复

```bash
# 模拟双击启动（最小化环境）
env -i HOME="$HOME" USER="$USER" PATH="/usr/bin:/bin:/usr/sbin:/sbin" \
    ./bin/ltools.app/Contents/MacOS/ltools

# 检查日志中的 Node.js 版本
# 应该显示:
# [ProcessManager] Found 7 nvm versions (sorted by version)
# [ProcessManager]   Version 1: /Users/.../.nvm/versions/node/v22.21.1/bin/node
# [ProcessManager] ✅ Found Node.js at candidate #0: .../v22.21.1/bin/node
```

---

## 附录 B：打包结构

### B.1 最终目录结构

```
plugins/musicplayer/lx-music-service/
└── dist/                          # 嵌入到 Go 二进制
    ├── server.bundle.js           # 打包后的服务代码 (~1.7MB)
    └── sources/                   # 音源插件文件
        ├── 元力kw1.1.0.js
        └── 开心汽水_0.1.5_鸿蒙.js
```

### B.2 构建流程

```
lx-music-service/
├── build.js           # esbuild 打包配置
├── server.js          # 入口文件
├── sources/           # 原始音源插件
└── dist/              # 构建输出
    ├── server.bundle.js   # esbuild 生成
    └── sources/           # build.js 复制
```

**注意**：sources 只由 `build.js` 复制一次到 `dist/sources/`，Taskfile 不再单独复制，避免重复。
