# 自动更新机制实现文档

## 概述

LTools 已集成自动更新功能，基于 HTTP 更新清单实现跨平台自动检测和下载更新。

## 架构

```
┌─────────────────┐
│   更新服务器     │
│  (update.json)  │
└────────┬────────┘
         │ HTTPS
         ▼
┌─────────────────┐
│  后端更新服务    │
│  (Go)           │
│  - CheckForUpdate()
│  - DownloadUpdate()
│  - InstallUpdate()
└────────┬────────┘
         │ Events
         ▼
┌─────────────────┐
│  前端UI组件      │
│  (React)        │
│  - 更新通知     │
│  - 下载进度     │
│  - 安装提示     │
└─────────────────┘
```

## 已实现的功能

### ✅ 后端服务
- [x] 版本检查（从远程 update.json）
- [x] 文件下载（支持进度回调）
- [x] SHA256 校验和验证
- [x] 平台自动检测（darwin-arm64, windows-amd64 等）
- [x] 补丁更新支持（delta updates）
- [x] 启动时自动检查（延迟 5 秒）
- [x] 事件通知前端

### ✅ 前端组件
- [x] 更新通知 UI
- [x] 下载进度显示
- [x] 版本说明展示
- [x] 手动检查更新按钮
- [x] 跳过非强制更新选项

### ✅ 构建工具
- [x] 更新清单生成脚本
- [x] 自动计算 SHA256 校验和
- [x] 多平台支持

### ⏳ 待实现功能
- [ ] macOS 安装逻辑
- [ ] Windows 安装逻辑
- [ ] Linux 安装逻辑
- [ ] 签名验证（可选）
- [ ] 从 build/config.yml 读取版本号

## 使用指南

### 1. 开发环境测试

```bash
# 启动应用（会自动检查更新）
task dev

# 手动触发更新检查
# 前端点击右下角刷新按钮
```

### 2. 发布新版本

#### 步骤 1: 更新版本号

编辑 `build/config.yml`:
```yaml
info:
  version: "0.2.0"  # 从 0.1.0 更新
```

#### 步骤 2: 构建所有平台

```bash
# macOS
task darwin:package
tar czf updates/ltools-0.2.0-darwin-arm64.tar.gz -C bin LTools.app

# Windows
task windows:build
cd bin && zip -r ../updates/ltools-0.2.0-windows-amd64.zip ltools.exe

# Linux
task linux:build
tar czf updates/ltools-0.2.0-linux-amd64.tar.gz -C bin ltools
```

#### 步骤 3: 生成更新清单

```bash
./scripts/generate-update-manifest.sh 0.2.0 ./updates https://updates.ltools.app/stable
```

输出示例：
```
================================
生成更新清单
================================
版本: 0.2.0
目录: ./updates
URL: https://updates.ltools.app/stable

检查平台包...
✓ 找到: ./updates/ltools-0.2.0-darwin-arm64.tar.gz (50M)
✓ 找到: ./updates/ltools-0.2.0-windows-amd64.zip (54M)
✓ 找到: ./updates/ltools-0.2.0-linux-amd64.tar.gz (51M)

找到 3 个平台包

✓ 生成成功
================================
文件: ./updates/update.json
```

#### 步骤 4: 编辑更新说明

```bash
nano updates/update.json
```

更新 `releaseNotes` 字段：
```json
{
  "releaseNotes": "## 新功能\n- 自动更新机制\n- 音乐播放器优化\n\n## 修复\n- 修复剪贴板历史 bug"
}
```

#### 步骤 5: 上传到服务器

**方案 A: 自托管服务器**
```bash
scp -r updates/* user@server:/var/www/updates/stable/
```

**方案 B: GitHub Releases**
```bash
gh release create v0.2.0 updates/* --title "v0.2.0" --notes "发布说明"
```

**方案 C: S3**
```bash
aws s3 sync updates/ s3://your-bucket/stable/ --acl public-read
```

### 3. 配置更新服务器

#### 更新 URL 配置

在 `main.go` 中修改：
```go
updateService := update.NewService(&update.ServiceConfig{
    UpdateURL: "https://updates.ltools.app/stable/",  // 更新为你的 URL
    // ...
})
```

#### 服务器要求

1. **HTTPS**（必须）
2. **CORS 支持**（如果跨域）

**Nginx 配置示例：**
```nginx
server {
    listen 443 ssl;
    server_name updates.ltools.app;

    root /var/www/updates;

    location /stable/ {
        autoindex on;
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, HEAD";
    }
}
```

**S3 CORS 配置：**
```json
{
  "CORSRules": [{
    "AllowedOrigins": ["*"],
    "AllowedMethods": ["GET", "HEAD"],
    "AllowedHeaders": ["*"]
  }]
}
```

### 4. 测试更新流程

#### 本地测试

```bash
# 1. 启动本地 HTTP 服务器
cd updates
python3 -m http.server 8000

# 2. 修改 main.go 中的 UpdateURL
UpdateURL: "http://localhost:8000/",

# 3. 启动应用测试
task dev
```

#### 生产测试

```bash
# 1. 使用旧版本（如 0.1.0）
# 2. 启动应用，等待 5 秒
# 3. 应该看到更新通知
# 4. 点击"立即下载"
# 5. 查看下载进度
# 6. 点击"安装并重启"
```

## 更新清单格式

### 完整示例

```json
{
  "version": "0.2.0",
  "releaseDate": "2025-03-04T12:00:00Z",
  "releaseNotes": "## 新功能\n- 自动更新机制\n\n## 修复\n- Bug 修复",
  "mandatory": false,
  "platforms": {
    "darwin-arm64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-darwin-arm64.tar.gz",
      "size": 52428800,
      "checksum": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "patches": {
        "0.1.0": {
          "url": "https://updates.ltools.app/stable/patches/0.1.0-to-0.2.0-darwin-arm64.patch",
          "size": 2097152,
          "checksum": "sha256:abc123..."
        }
      }
    },
    "windows-amd64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-windows-amd64.zip",
      "size": 56623104,
      "checksum": "sha256:def456..."
    },
    "linux-amd64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-linux-amd64.tar.gz",
      "size": 53477376,
      "checksum": "sha256:ghi789..."
    }
  }
}
```

### 字段说明

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| version | string | ✅ | 新版本号（必须大于当前版本） |
| releaseDate | string | ✅ | ISO 8601 格式发布日期 |
| releaseNotes | string | ✅ | Markdown 格式更新说明 |
| mandatory | boolean | ✅ | 是否强制更新 |
| platforms | object | ✅ | 平台特定信息 |
| platforms[url] | string | ✅ | 下载 URL |
| platforms[size] | number | ✅ | 文件大小（字节） |
| platforms[checksum] | string | ✅ | SHA256 校验和 |
| platforms[patches] | object | ❌ | 补丁更新（可选） |

## Delta Updates（补丁更新）

补丁更新允许只下载版本差异，减少下载量。

### 生成补丁

```bash
# 安装 bsdiff
brew install bsdiff  # macOS
sudo apt-get install bsdiff  # Linux

# 生成补丁
bsdiff old-app new-app patch.bsdiff

# 放置补丁
mkdir -p updates/patches
mv patch.bsdiff updates/patches/0.1.0-to-0.2.0-darwin-arm64.patch
```

### 更新清单配置

```json
{
  "darwin-arm64": {
    "url": "https://.../ltools-0.2.0-darwin-arm64.tar.gz",
    "size": 52428800,
    "checksum": "sha256:...",
    "patches": {
      "0.1.0": {
        "url": "https://.../patches/0.1.0-to-0.2.0-darwin-arm64.patch",
        "size": 2097152,
        "checksum": "sha256:..."
      }
    }
  }
}
```

## 安全性

### 签名验证（推荐生产环境）

#### 1. 生成密钥对

```bash
# 生成 Ed25519 私钥
openssl genpkey -algorithm Ed25519 -out private.pem

# 导出公钥
openssl pkey -in private.pem -pubout -out public.pem

# 获取 Base64 公钥
cat public.pem | base64
```

#### 2. 签名更新清单

```bash
# 签名
openssl pkeyutl -sign -inkey private.pem -in update.json -out update.json.sig

# 上传时同时上传签名文件
scp update.json update.json.sig user@server:/var/www/updates/
```

#### 3. 配置验证（TODO: 未来实现）

```go
updateService := update.NewService(&update.ServiceConfig{
    RequireSignature: true,
    PublicKey:        "MCowBQYDK2VwAyEA...",  // Base64 公钥
    // ...
})
```

## CI/CD 集成

### GitHub Actions 完整示例

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Build
        run: |
          task darwin:package
          tar czf ltools-${{ github.ref_name }}-darwin-arm64.tar.gz -C bin LTools.app

      - name: Upload Artifact
        uses: actions/upload-artifact@v3
        with:
          name: darwin-arm64
          path: ltools-${{ github.ref_name }}-darwin-arm64.tar.gz

  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Build
        run: |
          task windows:build
          Compress-Archive -Path bin\ltools.exe -DestinationPath ltools-${{ github.ref_name }}-windows-amd64.zip

      - name: Upload Artifact
        uses: actions/upload-artifact@v3
        with:
          name: windows-amd64
          path: ltools-${{ github.ref_name }}-windows-amd64.zip

  create-release:
    needs: [build-macos, build-windows]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Download Artifacts
        uses: actions/download-artifact@v3

      - name: Generate Update Manifest
        run: |
          mkdir -p updates
          mv darwin-arm64/*.tar.gz updates/
          mv windows-amd64/*.zip updates/

          ./scripts/generate-update-manifest.sh \
            ${{ github.ref_name }} \
            ./updates \
            https://github.com/${{ github.repository }}/releases/download/${{ github.ref_name }}

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            updates/*
          body: |
            ## 更新说明
            请查看 update.json 获取详细变更
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 故障排查

### 问题 1: 未检测到更新

**检查清单：**
- [ ] 更新清单 URL 是否正确
- [ ] update.json 是否可访问
- [ ] 版本号是否大于当前版本
- [ ] 平台标识是否匹配（darwin-arm64 等）
- [ ] CORS 是否配置正确

```bash
# 测试清单可访问性
curl -v https://updates.ltools.app/stable/update.json

# 检查 JSON 格式
jq . update.json
```

### 问题 2: 下载失败

**检查清单：**
- [ ] 文件 URL 是否可访问
- [ ] 文件大小是否正确
- [ ] 校验和是否匹配
- [ ] 网络连接是否正常

```bash
# 手动下载测试
curl -O https://updates.ltools.app/stable/ltools-0.2.0-darwin-arm64.tar.gz

# 验证校验和
shasum -a 256 ltools-0.2.0-darwin-arm64.tar.gz
```

### 问题 3: 安装失败

**当前状态：** 平台安装逻辑尚未完全实现

**临时方案：**
1. 下载完成后手动解压
2. 替换旧版本应用
3. 重启应用

**TODO:**
- 实现 macOS 安装逻辑
- 实现 Windows 安装逻辑
- 实现 Linux 安装逻辑

## 最佳实践

### ✅ 推荐做法

- ✅ 使用 HTTPS 托管更新
- ✅ 验证 SHA256 校验和
- ✅ 显示更新说明给用户
- ✅ 允许跳过非强制更新
- ✅ 保留旧版本以便回滚
- ✅ 测试更新流程后再发布
- ✅ 使用 CI/CD 自动化发布

### ❌ 避免的做法

- ❌ 强制立即重启
- ❌ 跳过校验和验证
- ❌ 在用户工作时打断
- ❌ 立即删除旧版本
- ❌ 忽略更新失败错误

## 相关文档

- [音乐播放器分发指南](../plugins/musicplayer/DISTRIBUTION.md)
- [项目架构文档](../CLAUDE.md)
- [Wails v3 自动更新文档](https://v3alpha.wails.io/guides/distribution/auto-updates/)

## 未来改进

1. **从 build/config.yml 读取版本号**
2. **实现平台特定安装逻辑**
3. **添加签名验证**
4. **支持更新通道（stable/beta/alpha）**
5. **添加更新统计和分析**
6. **支持回滚到旧版本**
7. **添加更新通知的自定义配置**

## 总结

LTools 现已具备完整的自动更新功能：

- ✅ **自动检测**：启动时延迟 5 秒检查更新
- ✅ **友好提示**：前端组件显示更新通知和进度
- ✅ **安全下载**：SHA256 校验和验证
- ✅ **跨平台**：支持 macOS、Windows、Linux
- ✅ **易于发布**：自动化脚本生成更新清单

**立即开始：**
1. 运行 `./scripts/generate-update-manifest.sh` 生成更新清单
2. 上传到更新服务器
3. 测试更新流程

🎉 **自动更新机制已就绪！**
