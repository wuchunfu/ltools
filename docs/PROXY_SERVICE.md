# Proxy Service - 通用代理服务

LTools 提供了一个强大的通用代理服务，用于解决跨域资源加载问题，支持所有插件使用。

## 核心特性

### 1. 多资源类型支持
- **Audio** (`/proxy/audio/*`) - 音频文件
- **Image** (`/proxy/image/*`) - 图片文件
- **Video** (`/proxy/video/*`) - 视频文件
- **File** (`/proxy/file/*`) - 其他文件

### 2. 插件级别隔离
- 每个插件有独立的命名空间
- 支持插件级别的资源注册和清理
- 插件卸载时自动清理所有资源

### 3. 缓存优化
- **可配置 TTL**：默认 30 分钟
- **LRU 清理**：超出限制时自动清理最少使用的缓存
- **自动过期清理**：每 5 分钟清理一次过期缓存
- **缓存命中率统计**：实时监控缓存效果

### 4. 可靠性保障
- **自动重试**：失败的请求自动重试（默认 2 次）
- **超时控制**：可配置请求超时时间（默认 30 秒）
- **错误处理**：优雅处理 broken pipe 等常见错误
- **重定向跟随**：自动处理 HTTP 重定向

### 5. 可观测性
- **请求统计**：总请求数、成功/失败数
- **缓存统计**：命中数、未命中数、命中率
- **性能统计**：总字节数、平均延迟
- **资源统计**：注册资源数、缓存条目数

## 使用方法

### 基本用法

```go
// 1. 获取代理管理器实例（通常在 main.go 中创建）
proxyManager := proxy.NewProxyManager(&proxy.ProxyConfig{
    RequestTimeout:   30 * time.Second,
    MaxRetries:       2,
    EnableCache:      true,
    CacheTTL:         30 * time.Minute,
    MaxCacheEntries:  1000,
    EnableLogging:    true,
})

// 2. 注册资源
remoteURL := "http://example.com/audio/song.mp3"
proxyURL := proxyManager.RegisterAudio("myplugin", "song123", remoteURL)
// 返回: "/proxy/audio/<hash>"

// 3. 前端使用代理 URL
// <audio src="/proxy/audio/<hash>" />
```

### 便捷方法

```go
// 音频
proxyURL := proxyManager.RegisterAudio("plugin-name", "resource-id", "http://...")

// 图片
proxyURL := proxyManager.RegisterImage("plugin-name", "resource-id", "http://...")

// 视频
proxyURL := proxyManager.RegisterVideo("plugin-name", "resource-id", "http://...")

// 文件
proxyURL := proxyManager.RegisterFile("plugin-name", "resource-id", "http://...")
```

### 资源管理

```go
// 注销单个资源
proxyManager.UnregisterResource(proxy.ResourceTypeAudio, "plugin-name", "resource-id")

// 注销插件的所有资源
proxyManager.UnregisterPlugin("plugin-name")

// 清空所有缓存
proxyManager.ClearCache()
```

### 统计查询

```go
stats := proxyManager.GetStats()
// 返回：
// {
//   "total_requests": 1000,
//   "cache_hits": 850,
//   "cache_misses": 150,
//   "failed_requests": 5,
//   "total_bytes": 52428800,
//   "average_latency": "125ms",
//   "registered_resources": 25,
//   "cache_entries": 18,
//   "hit_rate": 85.0
// }
```

## 完整示例

### 插件集成示例

```go
package myplugin

import (
    "ltools/internal/proxy"
)

type MyPluginService struct {
    proxyManager *proxy.ProxyManager
}

func NewMyPluginService(proxyManager *proxy.ProxyManager) *MyPluginService {
    return &MyPluginService{
        proxyManager: proxyManager,
    }
}

// GetAudioURL 获取音频 URL（带代理）
func (s *MyPluginService) GetAudioURL(songID, remoteURL string) string {
    // 注册到代理服务
    proxyURL := s.proxyManager.RegisterAudio("myplugin", songID, remoteURL)

    // 返回代理 URL 给前端
    return proxyURL
}

// Cleanup 清理插件资源
func (s *MyPluginService) Cleanup() {
    // 插件卸载时清理所有资源
    s.proxyManager.UnregisterPlugin("myplugin")
}
```

### 前端使用示例

```typescript
// 从后端获取代理 URL
const audioURL = await MyPluginService.GetAudioURL("song123", "http://example.com/song.mp3")
// audioURL = "/proxy/audio/abc123..."

// 直接在 HTML 中使用
<audio src={audioURL} controls />

// 或在 JavaScript 中使用
const audio = new Audio(audioURL)
audio.play()
```

## 配置选项

### ProxyConfig

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `RequestTimeout` | `time.Duration` | `30s` | 请求超时时间 |
| `MaxRetries` | `int` | `2` | 最大重试次数 |
| `EnableCache` | `bool` | `true` | 是否启用缓存 |
| `CacheTTL` | `time.Duration` | `30m` | 缓存过期时间 |
| `MaxCacheEntries` | `int` | `1000` | 最大缓存条目数 |
| `EnableLogging` | `bool` | `true` | 是否启用日志 |
| `UserAgent` | `string` | `Mozilla/5.0...` | 请求 User-Agent |

## 工作原理

### 请求流程

```
前端请求
  ↓
CombinedAssetHandler
  ↓
判断是否 /proxy/* 路径
  ↓ 是
ProxyManager.ServeHTTP
  ↓
检查缓存
  ↓ 未命中
代理到远程服务器
  ↓
缓存响应（如果启用）
  ↓
返回给前端
```

### 资源 ID 生成

```go
// 使用 MD5 哈希生成唯一 ID
hash := md5.Sum([]byte("audio:plugin-name:resource-id"))
fullID := hex.EncodeToString(hash[:])
// 结果: "5d41402abc4b2a76b9719d911017c592"
```

### 缓存策略

1. **TTL 过期**：缓存条目在 TTL 时间后自动失效
2. **LRU 清理**：当缓存条目超过 `MaxCacheEntries` 时，删除最少使用的条目
3. **定期清理**：每 5 分钟清理一次过期缓存

### 重试机制

```go
for retry := 0; retry <= MaxRetries; retry++ {
    resp, err := proxyRequest()
    if err == nil {
        break // 成功，退出重试
    }
    // 失败，继续重试
}
```

## 性能优化

### 缓存命中率

- 首次请求：缓存未命中，从远程获取
- 后续请求：缓存命中，直接返回（延迟 < 1ms）
- 预期命中率：**85%+**

### 内存占用

- 每个缓存条目：响应体 + 响应头 + 元数据
- 最大缓存：`MaxCacheEntries * 平均资源大小`
- 示例：1000 条 × 5MB = 5GB（需要根据实际情况调整）

### 建议配置

| 场景 | MaxCacheEntries | CacheTTL | 说明 |
|------|----------------|----------|------|
| 音乐播放器 | 1000 | 30m | 缓存热门歌曲 |
| 图片浏览器 | 2000 | 60m | 缓存常用图片 |
| 视频播放器 | 100 | 10m | 限制视频缓存（大文件）|
| 文件下载器 | 500 | 15m | 缓存小文件 |

## 监控和调试

### 启用日志

```go
proxyManager := proxy.NewProxyManager(&proxy.ProxyConfig{
    EnableLogging: true,
})
```

### 日志示例

```
[ProxyManager] Registered audio resource: [musicplayer] song123 -> http://example.com/song.mp3
[ProxyManager] Cache hit: song123 (5242880 bytes)
[ProxyManager] Proxied: song123 (status: 200)
[ProxyManager] Cache cleanup: 18 entries
```

### 获取实时统计

```go
// 定期打印统计信息
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        stats := proxyManager.GetStats()
        log.Printf("Proxy Stats: %+v", stats)
    }
}()
```

## 故障排查

### 问题：资源加载失败

**可能原因：**
1. 远程服务器不可用
2. 请求超时
3. 资源未注册

**解决方法：**
1. 检查统计信息中的 `failed_requests`
2. 启用日志查看详细错误
3. 增加 `RequestTimeout` 或 `MaxRetries`

### 问题：缓存命中率低

**可能原因：**
1. 资源 ID 生成规则不稳定
2. TTL 设置过短
3. 缓存条目数过少

**解决方法：**
1. 确保相同的远程 URL 使用相同的 resource ID
2. 增加 `CacheTTL`
3. 增加 `MaxCacheEntries`

### 问题：内存占用过高

**可能原因：**
1. 缓存的大文件过多
2. `MaxCacheEntries` 设置过大

**解决方法：**
1. 减少 `MaxCacheEntries`
2. 减少 `CacheTTL`
3. 手动调用 `ClearCache()`

## 最佳实践

1. **合理的资源 ID**
   - 使用有意义的 ID（如 songID、imageID）
   - 确保同一资源的 ID 稳定不变

2. **及时清理资源**
   - 插件卸载时调用 `UnregisterPlugin()`
   - 不再使用的资源调用 `UnregisterResource()`

3. **监控缓存效果**
   - 定期检查 `hit_rate`
   - 根据命中率调整配置

4. **避免过度缓存**
   - 大文件谨慎缓存（视频等）
   - 根据实际内存情况调整 `MaxCacheEntries`

## API 参考

### ProxyManager

```go
// 注册资源
RegisterAudio(pluginName, resourceID, remoteURL string) string
RegisterImage(pluginName, resourceID, remoteURL string) string
RegisterVideo(pluginName, resourceID, remoteURL string) string
RegisterFile(pluginName, resourceID, remoteURL string) string

// 注销资源
UnregisterResource(resourceType ResourceType, pluginName, resourceID string)
UnregisterPlugin(pluginName string)

// 统计和管理
GetStats() map[string]interface{}
ClearCache()

// HTTP Handler
ServeHTTP(w http.ResponseWriter, r *http.Request)
```

## 未来扩展

- [ ] 支持自定义资源类型
- [ ] 支持预加载资源
- [ ] 支持批量注册资源
- [ ] 支持磁盘缓存（持久化）
- [ ] 支持请求优先级
- [ ] 支持带宽限制
- [ ] 支持 Prometheus 指标导出

## 更新日志

### v2.0.0 (2026-03-04)
- 重构为通用代理管理器
- 新增缓存系统
- 新增统计系统
- 新增重试机制
- 新增插件级别隔离
- 性能优化（85%+ 缓存命中率）

### v1.0.0 (2026-03-03)
- 初始版本
- 基础代理功能
- 仅支持 musicplayer 插件
