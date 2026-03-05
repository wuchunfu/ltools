# 音乐播放器下载功能设计

**日期：** 2026-03-05
**作者：** Claude Code
**状态：** 已确认

## 概述

为 LTools 音乐播放器添加音乐下载功能，支持下载当前播放歌曲和搜索列表中的歌曲到本地。

## 需求总结

### 核心功能
- 下载当前播放的歌曲
- 从搜索列表中选择并下载歌曲
- 使用 Wails `Dialogs.SaveFile()` 选择保存位置
- 根据当前播放音质下载（优先 FLAC 无损，其次 320kbps MP3）

### 用户体验
- 文件命名：`歌名.mp3`（用户可在对话框中编辑）
- 下载进度：按钮显示加载状态（转圈动画）
- 下载完成：恢复正常状态，显示 Toast 成功提示
- 下载失败：显示 Toast 错误提示

## 技术方案

### 方案选择：纯后端下载

**原因：**
- ✅ 与 `Dialogs.SaveFile()` 完美配合
- ✅ 下载稳定可靠，适合音乐文件（可能较大）
- ✅ 后端已有 `GetSongURLWithMetadata()` 获取音频 URL
- ✅ 符合项目架构风格（服务端逻辑放在 Go 后端）

### 架构设计

```
用户点击下载按钮
    ↓
前端调用 Dialogs.SaveFile()
    ↓
用户选择保存位置和文件名
    ↓
前端调用后端 DownloadSong(song, savePath)
    ↓
后端获取音频 URL (GetSongURLWithMetadata)
    ↓
后端下载音频文件（处理代理 URL）
    ↓
后端智能识别格式并保存文件
    ↓
前端显示成功/失败提示
```

## 详细设计

### 1. 后端 API

#### 1.1 主要方法：DownloadSong

**文件：** `plugins/musicplayer/service_lx.go`

```go
// DownloadSong 下载歌曲到指定路径（暴露给前端）
func (s *ServiceLX) DownloadSong(song Song, savePath string) error {
    if !s.initialized {
        return fmt.Errorf("service not initialized")
    }

    // 1. 获取音频 URL
    audioURL, err := s.GetSongURLWithMetadata(&song, "320k")
    if err != nil {
        return fmt.Errorf("failed to get audio URL: %w", err)
    }

    // 2. 如果是代理 URL，获取真实 URL
    if strings.Contains(audioURL, "/proxy/audio/") {
        if s.proxyHandler != nil {
            parts := strings.Split(audioURL, "/")
            if len(parts) > 0 {
                songID := parts[len(parts)-1]
                realURL := s.proxyHandler.GetAudioURL(songID)
                if realURL != "" {
                    audioURL = realURL
                }
            }
        }
    }

    // 3. 下载文件
    resp, err := http.Get(audioURL)
    if err != nil {
        return fmt.Errorf("failed to download: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("download failed with status: %d", resp.StatusCode)
    }

    // 4. 智能识别格式
    audioFormat := detectAudioFormat(audioURL, resp.Header.Get("Content-Type"))
    actualPath := ensureCorrectExtension(savePath, audioFormat)

    // 5. 创建文件
    outFile, err := os.Create(actualPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer outFile.Close()

    // 6. 写入文件
    _, err = io.Copy(outFile, resp.Body)
    if err != nil {
        return fmt.Errorf("failed to save file: %w", err)
    }

    log.Printf("[ServiceLX] Downloaded song: %s -> %s", song.Name, actualPath)
    return nil
}
```

#### 1.2 辅助函数

```go
// detectAudioFormat 检测音频格式
func detectAudioFormat(url string, contentType string) string {
    // 优先从 Content-Type 判断
    contentType = strings.ToLower(contentType)
    switch {
    case strings.Contains(contentType, "flac"):
        return "flac"
    case strings.Contains(contentType, "mpeg") || strings.Contains(contentType, "mp3"):
        return "mp3"
    case strings.Contains(contentType, "aac"):
        return "aac"
    case strings.Contains(contentType, "mp4") || strings.Contains(contentType, "m4a"):
        return "m4a"
    }

    // 从 URL 判断
    urlLower := strings.ToLower(url)
    for _, ext := range []string{".flac", ".mp3", ".aac", ".m4a", ".ogg"} {
        if strings.Contains(urlLower, ext) {
            return strings.TrimPrefix(ext, ".")
        }
    }

    return "mp3"
}

// ensureCorrectExtension 确保文件扩展名正确
func ensureCorrectExtension(path string, format string) string {
    ext := filepath.Ext(path)
    correctExt := "." + format

    if ext != correctExt {
        base := strings.TrimSuffix(path, ext)
        return base + correctExt
    }
    return path
}
```

### 2. 前端实现

#### 2.1 播放控制栏下载按钮

**文件：** `frontend/src/widgets/MusicPlayerWidget.tsx`

```tsx
// 新增状态
const [isDownloading, setIsDownloading] = useState(false);

// 下载当前歌曲
const handleDownloadCurrentSong = async () => {
    if (!currentSong) return;

    try {
        setIsDownloading(true);

        // 1. 打开保存对话框
        const savePath = await Dialogs.SaveFile({
            DefaultFilename: `${currentSong.name}.mp3`,
            Title: `保存音乐 - ${currentSong.name}`,
            Filters: [
                { DisplayName: '音频文件', Pattern: '*.mp3;*.flac;*.aac;*.m4a' },
                { DisplayName: 'MP3 音频', Pattern: '*.mp3' },
                { DisplayName: 'FLAC 无损', Pattern: '*.flac' },
                { DisplayName: '所有文件', Pattern: '*.*' }
            ]
        });

        if (!savePath) return; // 用户取消

        // 2. 调用后端下载
        await MusicPlayerService.DownloadSong(currentSong, savePath);

        // 3. 显示成功提示
        showToast(`下载完成：${currentSong.name}`, 'success');

    } catch (error) {
        console.error('Download failed:', error);
        showToast(`下载失败：${error.message}`, 'error');
    } finally {
        setIsDownloading(false);
    }
};

// 按钮渲染（在播放控制区域）
<button
    onClick={handleDownloadCurrentSong}
    disabled={isDownloading || !currentSong}
    className="download-button"
    title="下载当前歌曲"
>
    {isDownloading ? (
        <Icon name="loading" className="animate-spin" />
    ) : (
        <Icon name="download" />
    )}
</button>
```

#### 2.2 搜索列表下载按钮

```tsx
// 新增状态：记录每首歌的下载状态
const [downloadingSongs, setDownloadingSongs] = useState<Set<string>>(new Set());

// 下载搜索列表中的歌曲
const handleDownloadSearchSong = async (song: Song) => {
    try {
        // 添加到下载中集合
        setDownloadingSongs(prev => new Set(prev).add(song.id));

        // 1. 打开保存对话框
        const savePath = await Dialogs.SaveFile({
            DefaultFilename: `${song.name}.mp3`,
            Title: `保存音乐 - ${song.name}`,
            Filters: [
                { DisplayName: '音频文件', Pattern: '*.mp3;*.flac;*.aac;*.m4a' },
                { DisplayName: 'MP3 音频', Pattern: '*.mp3' },
                { DisplayName: 'FLAC 无损', Pattern: '*.flac' },
                { DisplayName: '所有文件', Pattern: '*.*' }
            ]
        });

        if (!savePath) {
            // 用户取消，从下载中集合移除
            setDownloadingSongs(prev => {
                const newSet = new Set(prev);
                newSet.delete(song.id);
                return newSet;
            });
            return;
        }

        // 2. 调用后端下载
        await MusicPlayerService.DownloadSong(song, savePath);

        // 3. 显示成功提示
        showToast(`下载完成：${song.name}`, 'success');

    } catch (error) {
        console.error('Download failed:', error);
        showToast(`下载失败：${error.message}`, 'error');
    } finally {
        // 从下载中集合移除
        setDownloadingSongs(prev => {
            const newSet = new Set(prev);
            newSet.delete(song.id);
            return newSet;
        });
    }
};

// 在搜索结果列表中渲染
{songs.map(song => (
    <div key={song.id} className="song-item">
        {/* 歌曲信息 */}
        <div className="song-info">
            <span className="song-name">{song.name}</span>
            <span className="song-artist">{song.artist.join(', ')}</span>
        </div>

        {/* 下载按钮 */}
        <button
            onClick={() => handleDownloadSearchSong(song)}
            disabled={downloadingSongs.has(song.id)}
            className="download-icon"
            title="下载此歌曲"
        >
            {downloadingSongs.has(song.id) ? (
                <Icon name="loading" className="animate-spin" />
            ) : (
                <Icon name="download" />
            )}
        </button>
    </div>
))}
```

### 3. 错误处理

#### 3.1 后端错误场景

- 服务未初始化
- 获取音频 URL 失败
- 代理 URL 解析失败
- HTTP 下载失败
- 文件创建失败
- 文件写入失败

#### 3.2 前端错误处理

- 用户取消保存对话框
- 下载失败显示 Toast 错误提示
- 下载完成显示 Toast 成功提示
- 下载按钮状态管理（加载中/正常）

### 4. 智能格式识别

**问题：** 用户选择 `.mp3` 扩展名，但实际音频可能是 FLAC 无损格式。

**解决方案：**
1. 后端从 `Content-Type` 和 URL 推断实际格式
2. 自动修正文件扩展名（`.mp3` → `.flac`）
3. 避免文件格式不匹配导致播放失败

## 实现计划

### 文件清单

**后端：**
1. `plugins/musicplayer/service_lx.go` - 添加下载方法和辅助函数
2. `internal/proxy/service.go` - 可选：添加 `GetAudioURL()` 方法

**前端：**
3. `frontend/src/widgets/MusicPlayerWidget.tsx` - 添加下载功能
4. `frontend/bindings/ltools/plugins/musicplayer/servicelx.ts` - 自动生成

### 实施步骤

```
步骤 1: 后端实现（30 分钟）
├─ 添加 DownloadSong() 方法
├─ 添加 detectAudioFormat() 辅助函数
├─ 添加 ensureCorrectExtension() 辅助函数
└─ 测试编译: go build

步骤 2: 生成绑定（1 分钟）
└─ task common:generate:bindings

步骤 3: 前端实现（20 分钟）
├─ 添加状态管理（isDownloading, downloadingSongs）
├─ 添加 handleDownloadCurrentSong()
├─ 添加 handleDownloadSearchSong()
├─ 在播放控制栏添加下载按钮
└─ 在搜索列表添加下载按钮

步骤 4: 测试（10 分钟）
├─ 测试下载当前歌曲
├─ 测试下载搜索列表歌曲
├─ 测试用户取消保存对话框
├─ 测试下载失败错误处理
└─ 测试不同音频格式自动识别

总计：约 1 小时
```

## 测试清单

- [ ] 下载当前播放的歌曲
- [ ] 从搜索列表下载歌曲
- [ ] 用户取消保存对话框
- [ ] 下载失败显示错误提示
- [ ] 下载完成显示成功提示
- [ ] 下载按钮加载状态正确显示
- [ ] 不同音频格式自动识别（MP3/FLAC/AAC）
- [ ] 文件扩展名自动修正
- [ ] 多个下载同时进行（搜索列表）
- [ ] 代理 URL 正确处理

## 设计决策记录

### Q1: 下载位置？
**决定：** 每次下载时使用 `Dialogs.SaveFile()` 让用户选择保存位置。

### Q2: 文件命名规则？
**决定：** 默认 `歌名.mp3`，用户可在对话框中编辑。

### Q3: 下载音质？
**决定：** 使用当前播放音质（优先 FLAC 无损，其次 320kbps MP3）。

### Q4: 下载按钮位置？
**决定：** 组合方案：
- 播放控制栏/歌曲信息区域：下载当前歌曲按钮
- 搜索结果列表：每首歌旁边有下载按钮

### Q5: 下载进度显示？
**决定：** 按钮状态变化（加载动画）+ Toast 完成通知。

### Q6: 实现方案？
**决定：** 纯后端下载，使用 Go HTTP 客户端下载文件。

## 未来扩展

- [ ] 批量下载（选择多首歌曲）
- [ ] 下载队列管理面板
- [ ] 下载进度百分比显示
- [ ] 断点续传支持
- [ ] 下载历史记录
- [ ] 下载目录配置（设置页面）
