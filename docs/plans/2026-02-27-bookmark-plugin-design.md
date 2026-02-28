# 浏览器书签搜索插件设计文档

**创建日期**: 2026-02-27
**状态**: 设计完成，待实施

## 概述

浏览器书签搜索插件提供跨浏览器的书签搜索、管理和导出功能。支持 Chrome、Safari、Firefox 三大浏览器，采用智能匹配算法，提供混合式 UI（搜索窗口 + 独立页面）。

## 核心功能

### 1. 浏览器支持
- **Chrome**: JSON 格式书签文件解析
- **Safari**: 二进制 plist 格式解析（macOS 专用）
- **Firefox**: SQLite 数据库解析
- **扩展性**: 预留 Edge、Brave 等浏览器接口

### 2. 搜索功能
- **匹配字段**: 标题、URL、文件夹名称
- **匹配方式**:
  - 标题前缀匹配（优先级最高）
  - 标题包含匹配
  - URL 包含匹配
  - 文件夹匹配
  - 拼音匹配（中文标题）
- **排序策略**: 匹配分数 > 标题长度 > 添加时间

### 3. 缓存和同步
- **缓存有效期**: 7 天（可配置）
- **同步策略**:
  - 首次加载自动缓存所有浏览器书签
  - 启动时检查缓存是否过期
  - 提供一键手动同步按钮
- **状态显示**: 上次同步时间、总书签数、各浏览器统计

### 4. 导出备份
- **HTML 格式**: 标准 Netscape Bookmark File Format，可直接导入浏览器
- **JSON 格式**: 包含完整元数据（标题、URL、文件夹、来源、时间）
- **导出内容**: 所有浏览器书签的合并结果

### 5. UI 交互
- **搜索窗口**:
  - 触发方式: 输入 "书签" 或 "BM" + 回车
  - 显示: 标题、URL、浏览器图标、文件夹路径
  - 快捷键: Enter 打开、Cmd+Enter 新标签页、Cmd+C 复制
- **独立页面**:
  - 搜索栏 + 浏览器过滤
  - 书签列表（分组显示）
  - 同步和导出操作区
  - 缓存状态统计

## 技术架构

### 后端结构

```
plugins/bookmark/
├── plugin.go              # 插件主文件
├── service.go             # 前端服务接口
├── browser/               # 浏览器书签解析器
│   ├── interface.go       # 统一接口定义
│   ├── chrome.go          # Chrome 书签解析
│   ├── safari.go          # Safari 书签解析
│   ├── firefox.go         # Firefox 书签解析
│   └── factory.go         # 浏览器工厂（预留扩展）
├── cache.go               # 缓存管理
├── search.go              # 搜索引擎
└── export.go              # 导出功能
```

### 前端结构

```
frontend/src/
├── components/
│   └── BookmarkWidget.tsx       # 搜索窗口小部件
├── pages/
│   └── BookmarkPage.tsx         # 独立管理页面
└── hooks/
    └── useBookmarks.ts          # 书签 Hook
```

### 核心数据结构

```go
type Bookmark struct {
    ID           string    `json:"id"`           // 唯一标识（URL hash）
    Title        string    `json:"title"`        // 书签标题
    URL          string    `json:"url"`          // 书签 URL
    Folder       string    `json:"folder"`       // 所属文件夹路径
    Browser      string    `json:"browser"`      // 来源浏览器
    AddedAt      time.Time `json:"added_at"`     // 添加时间
    PinyinTitle  string    `json:"pinyin_title"` // 标题拼音（用于搜索）
}

type SearchResult struct {
    Bookmark   Bookmark `json:"bookmark"`
    Score      int      `json:"score"`      // 匹配分数（用于排序）
    MatchType  string   `json:"match_type"` // "prefix" | "contains" | "pinyin"
}

type CacheData struct {
    Bookmarks    []Bookmark     `json:"bookmarks"`
    LastSync     time.Time      `json:"last_sync"`
    BrowserStats map[string]int `json:"browser_stats"` // 各浏览器书签数
    Version      int            `json:"version"`
}

type CacheStatus struct {
    Available    bool           `json:"available"`
    LastSync     string         `json:"last_sync"`
    TotalCount   int            `json:"total_count"`
    BrowserStats map[string]int `json:"browser_stats"`
    IsExpired    bool           `json:"is_expired"`
}
```

## 浏览器书签解析

### 统一接口

```go
type BookmarkParser interface {
    Name() string                       // 浏览器名称
    Parse() ([]Bookmark, error)         // 解析书签
    IsAvailable() bool                  // 检查浏览器是否可用
    GetBookmarksPath() (string, error)  // 获取书签文件路径
}
```

### Chrome 实现
- **书签文件**: `~/Library/Application Support/Google/Chrome/Default/Bookmarks`
- **格式**: JSON
- **解析方式**: 读取 JSON 结构，递归遍历书签栏和其他文件夹

### Safari 实现
- **书签文件**: `~/Library/Safari/Bookmarks.plist`
- **格式**: 二进制 plist（macOS 特有）
- **解析方式**: 使用 `howett.net/plist` 库解析
- **平台限制**: 仅 macOS 可用

### Firefox 实现
- **书签数据库**: `~/Library/Application Support/Firefox/Profiles/*.default/places.sqlite`
- **格式**: SQLite 数据库
- **解析方式**: 查询 `moz_bookmarks` 和 `moz_places` 表

### 平台兼容性

```go
func NewParsers() []BookmarkParser {
    parsers := []BookmarkParser{
        &ChromeParser{},   // 跨平台
    }

    if runtime.GOOS == "darwin" {
        parsers = append(parsers, &SafariParser{})
    }

    parsers = append(parsers, &FirefoxParser{}) // 跨平台

    return parsers
}
```

## 搜索引擎

### 匹配算法

```go
func (e *SearchEngine) Search(query string) []SearchResult {
    // 1. 预处理查询（转小写、去除空格）
    // 2. 多维度匹配
    //    - 标题前缀匹配（分数：100）
    //    - 标题包含匹配（分数：80）
    //    - URL 包含匹配（分数：60）
    //    - 文件夹匹配（分数：40）
    //    - 拼音匹配（分数：50）
    // 3. 按分数排序
    // 4. 返回前 50 个结果
}
```

### 分数计算

- **标题前缀匹配**: 100 分 + (100 - 标题长度) / 10
- **标题完全匹配**: 95 分
- **标题包含匹配**: 80 分 + (100 - 匹配位置) / 10
- **拼音前缀匹配**: 70 分
- **URL 包含匹配**: 60 分
- **文件夹匹配**: 40 分

## 缓存管理

### 缓存文件

- **位置**: `~/.ltools/bookmark_cache.json`
- **格式**: JSON
- **有效期**: 7 天（可配置）

### 缓存流程

1. **首次加载**: 解析所有浏览器书签，保存到缓存
2. **后续加载**: 从缓存读取，检查是否过期
3. **手动同步**: 重新解析所有浏览器，更新缓存
4. **状态显示**: 显示上次同步时间、总书签数、各浏览器数量

### 缓存操作

```go
func (c *Cache) Load() (*CacheData, error)
func (c *Cache) Save(data *CacheData) error
func (c *Cache) IsExpired() bool  // 检查是否过期
func (c *Cache) Clear() error     // 清除缓存
```

## 导出功能

### HTML 格式

生成标准的 Netscape Bookmark File Format：

```html
<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><H3>书签栏</H3>
    <DL><p>
        <DT><A HREF="https://example.com">Example</A>
        <DD>Example Description
    </DL><p>
</DL><p>
```

### JSON 格式

```json
{
  "exported_at": "2026-02-27T18:30:00Z",
  "total_count": 235,
  "browsers": ["chrome", "safari", "firefox"],
  "bookmarks": [
    {
      "id": "abc123",
      "title": "Example",
      "url": "https://example.com",
      "folder": "书签栏/工作",
      "browser": "chrome",
      "added_at": "2025-01-15T10:00:00Z"
    }
  ]
}
```

## 服务接口

```go
type BookmarkService struct {
    plugin *BookmarkPlugin
    app    *application.App
}

// 暴露给前端的方法
func (s *BookmarkService) Search(query string) ([]SearchResult, error)
func (s *BookmarkService) Sync() error
func (s *BookmarkService) GetCacheStatus() (*CacheStatus, error)
func (s *BookmarkService) ExportHTML(outputPath string) error
func (s *BookmarkService) ExportJSON(outputPath string) error
func (s *BookmarkService) OpenURL(url string) error
```

## 事件系统

```go
// 在 main.go init() 中注册
application.RegisterEvent[string]("bookmark:sync-started")
application.RegisterEvent[string]("bookmark:sync-completed")
application.RegisterEvent[string]("bookmark:sync-error")
application.RegisterEvent[string]("bookmark:exported")
```

## 实施计划

### Phase 1 - 核心功能（MVP）

**目标**: 实现基本可用的书签搜索功能

1. ✅ 插件框架搭建
   - 创建插件结构
   - 注册到 main.go
   - 生成 TypeScript 绑定

2. ✅ Chrome 书签解析
   - 实现 Chrome JSON 解析
   - 递归遍历书签树
   - 提取完整元数据

3. ✅ 基础搜索功能
   - 实现标题和 URL 匹配
   - 基础排序算法
   - 返回搜索结果

4. ✅ 缓存机制
   - 实现缓存读写
   - 过期检查
   - 手动同步

5. ✅ 搜索窗口小部件
   - 实现触发关键字
   - 显示搜索结果
   - 打开 URL 功能

### Phase 2 - 完善功能

**目标**: 支持多浏览器并提供完整管理界面

1. ⏳ Safari 和 Firefox 解析
   - Safari plist 解析（macOS）
   - Firefox SQLite 查询
   - 多浏览器合并

2. ⏳ 拼音搜索
   - 集成 go-pinyin 库
   - 生成拼音索引
   - 拼音匹配算法

3. ⏳ 独立管理页面
   - 完整的书签列表
   - 浏览器过滤
   - 状态显示

4. ⏳ 导出功能
   - HTML 导出
   - JSON 导出
   - 文件保存对话框

### Phase 3 - 优化增强

**目标**: 性能优化和用户体验提升

1. ⏳ 高级排序算法
   - 访问频率统计（如果可用）
   - 智能推荐
   - 搜索历史

2. ⏳ 性能优化
   - 大量书签的加载优化
   - 搜索响应速度
   - 内存占用优化

3. ⏳ 错误处理
   - 浏览器不可用提示
   - 书签文件损坏处理
   - 权限问题处理

## 依赖库

### Go 依赖

```go
// go.mod 新增
require (
    github.com/mozillazg/go-pinyin v0.20.0  // 拼音转换
    howett.net/plist v1.0.1                  // plist 解析（Safari）
    github.com/mattn/go-sqlite3 v1.14.22     // SQLite（Firefox）
)
```

### 前端依赖

无需新增依赖，使用项目现有的 React、TypeScript、TailwindCSS。

## 配置项

```go
type BookmarkConfig struct {
    CacheExpiryDays int      `json:"cache_expiry_days"` // 缓存有效期（天）
    MaxResults      int      `json:"max_results"`       // 最大搜索结果数
    EnablePinyin    bool     `json:"enable_pinyin"`     // 启用拼音搜索
    TriggerKeywords []string `json:"trigger_keywords"`  // 触发关键字
}

// 默认配置
var DefaultConfig = BookmarkConfig{
    CacheExpiryDays: 7,
    MaxResults:      50,
    EnablePinyin:     true,
    TriggerKeywords: []string{"书签", "bookmark", "bm"},
}
```

## 测试策略

### 单元测试

- 浏览器解析器测试（使用示例书签文件）
- 搜索算法测试（各种匹配场景）
- 缓存读写测试
- 导出格式验证测试

### 集成测试

- 多浏览器书签合并测试
- 搜索到打开 URL 的完整流程
- 缓存过期和同步流程

### 手动测试

- 各浏览器书签导入验证
- 搜索准确性验证
- UI 交互验证

## 已知限制

1. **Safari 限制**: 仅 macOS 可用，Windows/Linux 用户无法使用
2. **Firefox 配置文件**: 需要自动检测默认配置文件，多配置文件情况可能不准确
3. **权限要求**: 需要文件系统读取权限访问浏览器书签文件
4. **缓存一致性**: 浏览器书签变更后需要手动同步，无法实时更新

## 未来扩展

1. **Edge 和 Brave 支持**: 添加基于 Chromium 的浏览器支持
2. **浏览器配置文件**: 支持选择不同的浏览器配置文件
3. **书签去重**: 自动检测和合并重复书签
4. **标签系统**: 为书签添加自定义标签
5. **云同步**: 支持导出到云存储服务
6. **实时监控**: 使用 fsnotify 监听书签文件变化
