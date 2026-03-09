# macOS DMG 安装包

本文档说明 LTools 的 DMG 安装包创建和自定义。

## 概述

从 v0.1.4 开始，LTools 为 macOS 提供 DMG 安装包，提供更专业的安装体验。

### DMG 特点

- **拖拽安装** - 将应用拖拽到 Applications 文件夹
- **自定义背景** - 品牌化的安装界面
- **图标布局** - 应用和 Applications 文件夹位置固定
- **压缩优化** - zlib 压缩，节省约 80% 空间

## 下载

从 [Releases](https://github.com/lian-yang/ltools/releases) 页面下载 DMG 文件：

```
ltools-{VERSION}-darwin-arm64.dmg  # Apple Silicon (M1/M2/M3)
```

## 安装步骤

### 1. 下载并打开 DMG

```bash
# 下载（以 v0.1.4 为例）
curl -LO https://github.com/lian-yang/ltools/releases/download/v0.1.4/ltools-v0.1.4-darwin-arm64.dmg

# 打开 DMG
open ltools-v0.1.4-darwin-arm64.dmg
```

### 2. 拖拽安装

在打开的窗口中，将 **LTools** 图标拖拽到 **Applications** 文件夹。

![DMG 安装界面](../resource/images/dmg-installer.png)

### 3. 移除隔离属性

首次打开可能需要移除 macOS 隔离属性：

```bash
sudo xattr -rd com.apple.quarantine /Applications/LTools.app
```

### 4. 启动应用

```bash
open /Applications/LTools.app
```

## 本地构建 DMG

### 使用 Task 命令

```bash
# 创建 DMG（推荐）
wails3 task darwin:create:dmg

# 完整签名流程（需要签名证书）
wails3 task darwin:package:dmg
```

### 手动创建

```bash
# 1. 构建 .app bundle
wails3 task darwin:package

# 2. 安装 create-dmg
brew install create-dmg

# 3. 创建 DMG
create-dmg \
  --volname "LTools Installer" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "LTools.app" 200 185 \
  --app-drop-link 600 185 \
  --background build/darwin/dmg/background.png \
  bin/ltools.dmg \
  bin/ltools.app
```

## 自定义 DMG 背景

### 背景文件位置

```
build/darwin/dmg/
├── background.png      # DMG 窗口背景 (800x400)
├── background@2x.png   # Retina 版本 (1600x800, 可选)
└── README.md           # 设计规范
```

### 设计规范

| 属性 | 规格 |
|------|------|
| 尺寸 | 800 x 400 像素 |
| 格式 | PNG (支持透明) |
| 应用图标位置 | 左侧 (200, 185) |
| Applications 位置 | 右侧 (600, 185) |

### 修改背景

1. 替换 `build/darwin/dmg/background.png`
2. 重新运行 `wails3 task darwin:create:dmg`

## GitHub Actions 集成

### 自动化流程

GitHub Actions 在发布时自动创建 DMG：

1. **安装 create-dmg** - 使用 Homebrew（带缓存）
2. **创建 DMG** - 使用自定义背景（如果存在）
3. **上传 Release** - DMG 文件自动上传

### 缓存优化

create-dmg 工具会被缓存以加速构建：

```yaml
- name: Cache Homebrew create-dmg
  uses: actions/cache@v4
  with:
    path: |
      ~/Library/Caches/Homebrew/create-dmg--*
      /opt/homebrew/Cellar/create-dmg
    key: ${{ runner.os }}-create-dmg-v1
```

首次构建约 60 秒，后续构建约 1 秒。

## 故障排查

### DMG 无法打开

```bash
# 验证 DMG 文件
hdiutil imageinfo ltools.dmg

# 手动挂载
hdiutil attach ltools.dmg
```

### 背景图不显示

1. 确认 `build/darwin/dmg/background.png` 存在
2. 检查文件尺寸是否为 800x400
3. 重新创建 DMG

### 图标位置错误

修改 `--icon` 和 `--app-drop-link` 参数调整位置：

```bash
create-dmg \
  --icon "LTools.app" 200 185 \     # 应用图标位置 (x, y)
  --app-drop-link 600 185 \          # Applications 链接位置 (x, y)
  ...
```

## 相关文件

| 文件 | 用途 |
|------|------|
| `build/darwin/Taskfile.yml` | DMG 创建任务定义 |
| `build/darwin/dmg/background.png` | 自定义背景图 |
| `.github/workflows/release.yml` | CI/CD 集成 |
| `docs/DMG_INSTALLER.md` | 本文档 |

## 参考资料

- [create-dmg GitHub](https://github.com/create-dmg/create-dmg)
- [Apple DMG 文档](https://developer.apple.com/library/archive/documentation/DeveloperTools/Conceptual/PackageMakerUserGuide/PackageMakerUserGuide.pdf)
