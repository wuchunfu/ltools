# 自动更新机制实现指南

## 概述

LTools 已集成自动检测更新机制，支持：
- ✅ 启动时自动检查更新（延迟 5 秒）
- ✅ 手动检查更新
- ✅ 后台下载（带进度）
- ✅ 校验和验证（SHA256）
- ✅ 跨平台支持（macOS/Windows/Linux）
- ✅ 补丁更新（Delta Updates）
- ✅ 前端通知 UI

---

## 架构说明

```
internal/update/
├── service.go              # 更新服务主逻辑
├── example_manifest.go     # update.json 示例

scripts/
└── generate-update-manifest.sh  # 生成更新清单脚本

frontend/src/widgets/
└── UpdateNotificationWidget.tsx  # 前端更新通知组件

main.go
└── 集成更新服务（第 549-583 行）
```

---

## 工作流程

### 1. 应用启动时

```go
// main.go 第 708-721 行
go func() {
    time.Sleep(5 * time.Second)  // 延迟 5 秒
    if updateService.IsEnabled() {
        info, err := updateService.CheckForUpdate()
        if err == nil && info != nil {
            app.Event.Emit("update:available", info)
        }
    }
}()
```

### 2. 前端接收事件

```typescript
// UpdateNotificationWidget.tsx
Events.On('update:available', (ev: { data: UpdateInfo }) => {
    setUpdateInfo(ev.data);  // 显示更新通知
});
```

### 3. 用户下载更新

```typescript
const filePath = await UpdateService.DownloadUpdate(
    updateInfo.downloadUrl,
    updateInfo.checksum
);
```

### 4. 安装并重启

```typescript
await UpdateService.InstallUpdate(filePath);
```

---

## 配置

### 1. 更新服务器 URL

在 `main.go` 中配置：

```go
updateService := update.NewService(&update.ServiceConfig{
    CurrentVersion: "0.1.0",
    UpdateURL:      "https://updates.ltools.app/stable/",
    DataDir:        dataDir,
    Enabled:        true,
})
```

### 2. 禁用自动更新

```go
Enabled: false,  // 或从配置文件读取
```

### 3. 版本号管理

当前版本硬编码在 `main.go` 中。建议：

**方案 A: 编译时注入**
```bash
go build -ldflags="-X main.version=0.2.0"
```

**方案 B: 从 build/config.yml 读取**
```go
// TODO: 实现读取 YAML 配置
version := readVersionFromConfig("build/config.yml")
```

---

## 托管更新清单

### 更新清单结构

```json
{
  "version": "0.2.0",
  "releaseDate": "2025-03-04T12:00:00Z",
  "releaseNotes": "## 新功能\n- 自动更新机制",
  "mandatory": false,
  "platforms": {
    "darwin-arm64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-darwin-arm64.tar.gz",
      "size": 52428800,
      "checksum": "sha256:abc123..."
    },
    "windows-amd64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-windows-amd64.zip",
      "size": 56623104,
      "checksum": "sha256:def456..."
    }
  }
}
```

### 托管方案

#### 选项 1: GitHub Releases

```bash
# 1. 构建应用
task build

# 2. 生成清单
./scripts/generate-update-manifest.sh 0.2.0 ./updates https://github.com/yourname/ltools/releases/download/v0.2.0

# 3. 创建 GitHub Release 并上传文件
gh release create v0.2.0 ./updates/*
```

**更新 URL**: `https://github.com/yourname/ltools/releases/download/v0.2.0/`

#### 选项 2: Amazon S3 / Cloudflare R2

```bash
# 1. 生成清单
./scripts/generate-update-manifest.sh 0.2.0 ./updates https://your-bucket.s3.amazonaws.com/stable

# 2. 上传到 S3
aws s3 sync ./updates/ s3://your-bucket/stable/ --acl public-read

# 3. 配置 CORS
# 见下方 CORS 配置
```

**更新 URL**: `https://your-bucket.s3.amazonaws.com/stable/`

#### 选项 3: 自托管服务器

```bash
# 1. 生成清单
./scripts/generate-update-manifest.sh 0.2.0 ./updates https://updates.ltools.app/stable

# 2. 上传到服务器
scp -r ./updates/* user@server:/var/www/updates/

# 3. 配置 Nginx
# 见下方 Nginx 配置
```

**更新 URL**: `https://updates.ltools.app/stable/`

---

## 服务器配置

### CORS 配置（S3）

```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["*"],
      "AllowedMethods": ["GET", "HEAD"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3000
    }
  ]
}
```

### Nginx 配置

```nginx
server {
    listen 443 ssl;
    server_name updates.ltools.app;

    root /var/www/updates;

    location / {
        autoindex on;

        # CORS headers
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, HEAD";

        # Cache control
        add_header Cache-Control "public, max-age=3600";
    }
}
```

---

## 发布流程

### 完整发布示例

```bash
# 1. 更新版本号
# 编辑 build/config.yml，将 version 改为 0.2.0

# 2. 构建所有平台
task darwin:build
task windows:build
task linux:build

# 3. 打包应用
# macOS
task darwin:package
tar czf updates/ltools-0.2.0-darwin-arm64.tar.gz -C bin LTools.app

# Windows
cd bin && zip -r ../updates/ltools-0.2.0-windows-amd64.zip ltools.exe
cd ..

# Linux
tar czf updates/ltools-0.2.0-linux-amd64.tar.gz -C bin ltools

# 4. 生成更新清单
./scripts/generate-update-manifest.sh 0.2.0 ./updates https://updates.ltools.app/stable

# 5. 编辑更新说明
nano updates/update.json  # 更新 releaseNotes

# 6. 上传到服务器
scp -r updates/* user@server:/var/www/updates/stable/

# 7. 测试更新
# 启动旧版本应用，验证更新提示
```

---

## Delta Updates（补丁更新）

### 生成补丁

```bash
# 安装 bsdiff
brew install bsdiff  # macOS
sudo apt-get install bsdiff  # Ubuntu

# 生成补丁
bsdiff old-app new-app patch.bsdiff

# 放置补丁
mkdir -p updates/patches
mv patch.bsdiff updates/patches/0.1.0-to-0.2.0-darwin-arm64.patch
```

### 更新清单格式

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

---

## 前端集成

### 添加到主应用

在 `App.tsx` 或主布局中添加：

```typescript
import { UpdateNotification } from './widgets/UpdateNotificationWidget';

export function App() {
  return (
    <>
      {/* 你的应用内容 */}

      {/* 更新通知组件 */}
      <UpdateNotification />
    </>
  );
}
```

### 设置页面集成

```typescript
function SettingsPage() {
  const [autoUpdate, setAutoUpdate] = useState(true);

  const handleCheckUpdate = async () => {
    const info = await UpdateService.CheckForUpdate();
    if (!info) {
      toast.success('您已经在使用最新版本！');
    }
  };

  return (
    <div>
      <h2>更新设置</h2>

      <label>
        <input
          type="checkbox"
          checked={autoUpdate}
          onChange={(e) => setAutoUpdate(e.target.checked)}
        />
        自动检查更新
      </label>

      <button onClick={handleCheckUpdate}>
        立即检查更新
      </button>

      <div>
        当前版本: {currentVersion}
      </div>
    </div>
  );
}
```

---

## 签名验证（可选，增强安全性）

### 1. 生成密钥对

```bash
# 生成 Ed25519 私钥
openssl genpkey -algorithm Ed25519 -out private.pem

# 导出公钥
openssl pkey -in private.pem -pubout -out public.pem

# 获取 Base64 公钥（用于配置）
cat public.pem | base64
```

### 2. 签名更新清单

```bash
# 签名
openssl pkeyutl -sign -inkey private.pem -in update.json -out update.json.sig

# 上传时同时上传签名文件
scp update.json update.json.sig user@server:/var/www/updates/
```

### 3. 配置验证（TODO）

```go
// 未来版本支持
updater := update.NewService(&update.ServiceConfig{
    RequireSignature: true,
    PublicKey:        "MCowBQYDK2VwAyEA...",  // Base64 公钥
})
```

---

## 故障排查

### 问题 1: 未检测到更新

**检查项：**
1. ✅ 更新 URL 是否正确
2. ✅ update.json 是否可访问
3. ✅ 版本号是否比当前版本新
4. ✅ 平台标识是否匹配（如 `darwin-arm64`）
5. ✅ CORS 是否配置正确

```bash
# 测试更新清单
curl -v https://updates.ltools.app/stable/update.json

# 检查平台
go run -e 'fmt.Println(runtime.GOOS, runtime.GOARCH)'
```

### 问题 2: 下载失败

**检查项：**
1. ✅ 下载 URL 是否可访问
2. ✅ 文件大小是否正确
3. ✅ 校验和是否匹配
4. ✅ 网络连接是否正常

```bash
# 手动下载测试
curl -O https://updates.ltools.app/stable/ltools-0.2.0-darwin-arm64.tar.gz

# 验证校验和
shasum -a 256 ltools-0.2.0-darwin-arm64.tar.gz
```

### 问题 3: 安装失败

**当前状态：** 平台特定安装逻辑尚未实现

**TODO:**
- [ ] 实现 macOS 安装逻辑
- [ ] 实现 Windows 安装逻辑
- [ ] 实现 Linux 安装逻辑

---

## 最佳实践

### ✅ 推荐做法

- ✅ 使用 HTTPS 托管更新
- ✅ 验证 SHA256 校验和
- ✅ 显示更新说明给用户
- ✅ 允许用户跳过非强制更新
- ✅ 保留旧版本以便回滚
- ✅ 测试更新流程后再发布

### ❌ 避免的做法

- ❌ 强制立即重启（应提示用户）
- ❌ 跳过校验和验证
- ❌ 在用户工作时打断
- ❌ 立即删除旧版本
- ❌ 忽略更新失败错误

---

## CI/CD 集成示例

### GitHub Actions

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-and-release:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Build
        run: |
          task darwin:build
          task darwin:package
          tar czf ltools-${{ github.ref_name }}-darwin-arm64.tar.gz -C bin LTools.app

      - name: Generate Manifest
        run: |
          ./scripts/generate-update-manifest.sh \
            ${{ github.ref_name }} \
            ./updates \
            https://github.com/${{ github.repository }}/releases/download/${{ github.ref_name }}

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            ltools-${{ github.ref_name }}-darwin-arm64.tar.gz
            updates/update.json
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 相关文档

- [Wails v3 自动更新文档](https://v3alpha.wails.io/guides/distribution/auto-updates/)
- [音乐播放器分发指南](../plugins/musicplayer/DISTRIBUTION.md)
- [项目架构文档](../CLAUDE.md)

---

## 下一步

1. **测试更新流程**
   ```bash
   # 修改版本号为 0.0.1
   # 运行应用，验证更新检测
   ```

2. **实现平台安装逻辑**
   - 编辑 `internal/update/service.go`
   - 实现 `installMacOS()`, `installWindows()`, `installLinux()`

3. **配置生产环境更新服务器**
   - 选择托管方案（GitHub Releases / S3 / 自托管）
   - 配置 CORS 和 HTTPS

4. **添加签名验证**
   - 生成密钥对
   - 配置签名验证

5. **集成到设置页面**
   - 添加"检查更新"按钮
   - 显示当前版本
   - 自动更新开关
