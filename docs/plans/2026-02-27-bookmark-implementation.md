# 浏览器书签搜索插件实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现一个支持 Chrome、Safari、Firefox 的浏览器书签搜索插件，提供智能搜索、缓存管理和导出功能。

**Architecture:** 采用插件化架构，后端使用 Go 实现浏览器书签解析器、搜索引擎和缓存管理，前端使用 React 实现搜索窗口小部件和独立管理页面。

**Tech Stack:** Go 1.25+, Wails v3, React 18.2, TypeScript 5.2, SQLite (Firefox), plist (Safari)

---

## Task 1: 创建插件目录结构和基础文件

**Files:**
- Create: `plugins/bookmark/plugin.go`
- Create: `plugins/bookmark/types.go`

**Step 1: 创建插件目录**

```bash
mkdir -p plugins/bookmark/browser
```

**Step 2: 创建类型定义文件**

Create `plugins/bookmark/types.go`:

```go
package bookmark

import "time"

// Bookmark 表示一个书签
type Bookmark struct {
	ID          string    `json:"id"`           // 唯一标识（URL hash）
	Title       string    `json:"title"`        // 书签标题
	URL         string    `json:"url"`          // 书签 URL
	Folder      string    `json:"folder"`       // 所属文件夹路径
	Browser     string    `json:"browser"`      // 来源浏览器
	AddedAt     time.Time `json:"added_at"`     // 添加时间
	PinyinTitle string    `json:"pinyin_title"` // 标题拼音（用于搜索）
}

// SearchResult 搜索结果
type SearchResult struct {
	Bookmark  Bookmark `json:"bookmark"`
	Score     int      `json:"score"`      // 匹配分数（用于排序）
	MatchType string   `json:"match_type"` // "prefix" | "contains" | "pinyin"
}

// CacheData 缓存数据
type CacheData struct {
	Bookmarks    []Bookmark     `json:"bookmarks"`
	LastSync     time.Time      `json:"last_sync"`
	BrowserStats map[string]int `json:"browser_stats"` // 各浏览器书签数
	Version      int            `json:"version"`
}

// CacheStatus 缓存状态
type CacheStatus struct {
	Available    bool           `json:"available"`
	LastSync     string         `json:"last_sync"`
	TotalCount   int            `json:"total_count"`
	BrowserStats map[string]int `json:"browser_stats"`
	IsExpired    bool           `json:"is_expired"`
}

// BookmarkConfig 插件配置
type BookmarkConfig struct {
	CacheExpiryDays int      `json:"cache_expiry_days"` // 缓存有效期（天）
	MaxResults      int      `json:"max_results"`       // 最大搜索结果数
	EnablePinyin    bool     `json:"enable_pinyin"`     // 启用拼音搜索
	TriggerKeywords []string `json:"trigger_keywords"`  // 触发关键字
}
```

**Step 3: 创建插件主文件**

Create `plugins/bookmark/plugin.go`:

```go
package bookmark

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "bookmark.builtin"
	PluginName    = "书签搜索"
	PluginVersion = "1.0.0"
)

// BookmarkPlugin 浏览器书签搜索插件
type BookmarkPlugin struct {
	*plugins.BasePlugin
	app     *application.App
	dataDir string
	config  *BookmarkConfig
	cache   *Cache
}

// NewBookmarkPlugin 创建插件实例
func NewBookmarkPlugin() *BookmarkPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "搜索 Chrome、Safari、Firefox 浏览器书签",
		Icon:        "bookmark",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{
			plugins.PermissionFileSystem, // 读取书签文件
		},
		Keywords:   []string{"书签", "bookmark", "bm", "浏览器"},
		ShowInMenu: plugins.BoolPtr(false), // 通过搜索窗口触发
		HasPage:    plugins.BoolPtr(true),  // 有独立管理页面
	}

	return &BookmarkPlugin{
		BasePlugin: plugins.NewBasePlugin(metadata),
		config: &BookmarkConfig{
			CacheExpiryDays: 7,
			MaxResults:      50,
			EnablePinyin:    false, // Phase 2 实现
			TriggerKeywords: []string{"书签", "bookmark", "bm"},
		},
	}
}

// Init 初始化插件
func (p *BookmarkPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// SetDataDir 设置数据目录
func (p *BookmarkPlugin) SetDataDir(dataDir string) error {
	p.dataDir = dataDir

	// 初始化缓存
	cache, err := NewCache(dataDir)
	if err != nil {
		p.app.Logger.Error("Failed to create cache: " + err.Error())
		return err
	}
	p.cache = cache

	// 后台异步加载书签
	go p.loadBookmarks()

	return nil
}

// ServiceStartup 服务启动时调用
func (p *BookmarkPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown 服务关闭时调用
func (p *BookmarkPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// loadBookmarks 加载书签（从缓存或浏览器）
func (p *BookmarkPlugin) loadBookmarks() {
	// 尝试从缓存加载
	cacheData, err := p.cache.Load()
	if err == nil && !p.cache.IsExpired() {
		p.app.Logger.Info("Loaded bookmarks from cache")
		return
	}

	// 缓存过期或不存在，重新同步
	p.app.Logger.Info("Cache expired or missing, syncing...")
	if err := p.Sync(); err != nil {
		p.app.Logger.Error("Failed to sync bookmarks: " + err.Error())
	}
}

// Sync 同步所有浏览器书签
func (p *BookmarkPlugin) Sync() error {
	// 将在后续任务中实现
	return nil
}
```

**Step 4: 提交基础结构**

```bash
git add plugins/bookmark/
git commit -m "feat(bookmark): 创建插件基础结构和类型定义

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 实现浏览器书签解析器接口

**Files:**
- Create: `plugins/bookmark/browser/interface.go`

**Step 1: 创建解析器接口**

Create `plugins/bookmark/browser/interface.go`:

```go
package browser

import "ltools/plugins/bookmark"

// BookmarkParser 浏览器书签解析器接口
type BookmarkParser interface {
	// Name 返回浏览器名称
	Name() string

	// Parse 解析书签
	Parse() ([]bookmark.Bookmark, error)

	// IsAvailable 检查浏览器是否可用（书签文件是否存在）
	IsAvailable() bool

	// GetBookmarksPath 获取书签文件路径
	GetBookmarksPath() (string, error)
}
```

**Step 2: 提交接口定义**

```bash
git add plugins/bookmark/browser/interface.go
git commit -m "feat(bookmark): 定义浏览器书签解析器接口

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 实现 Chrome 书签解析器

**Files:**
- Create: `plugins/bookmark/browser/chrome.go`

**Step 1: 创建 Chrome 解析器**

Create `plugins/bookmark/browser/chrome.go`:

```go
package browser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"ltools/plugins/bookmark"
)

// ChromeParser Chrome 书签解析器
type ChromeParser struct {
	bookmarksPath string
}

// NewChromeParser 创建 Chrome 解析器
func NewChromeParser() *ChromeParser {
	return &ChromeParser{}
}

// Name 返回浏览器名称
func (p *ChromeParser) Name() string {
	return "Chrome"
}

// GetBookmarksPath 获取书签文件路径
func (p *ChromeParser) GetBookmarksPath() (string, error) {
	if p.bookmarksPath != "" {
		return p.bookmarksPath, nil
	}

	var basePath string
	switch runtime.GOOS {
	case "darwin":
		basePath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Google", "Chrome", "Default")
	case "windows":
		basePath = filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Default")
	case "linux":
		basePath = filepath.Join(os.Getenv("HOME"), ".config", "google-chrome", "Default")
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	p.bookmarksPath = filepath.Join(basePath, "Bookmarks")
	return p.bookmarksPath, nil
}

// IsAvailable 检查 Chrome 是否可用
func (p *ChromeParser) IsAvailable() bool {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// chromeBookmark Chrome 书签 JSON 结构
type chromeBookmark struct {
	Checksum string `json:"checksum"`
	Roots    struct {
		BookmarkBar chromeBookmarkNode `json:"bookmark_bar"`
		Other       chromeBookmarkNode `json:"other"`
		Synced      chromeBookmarkNode `json:"synced"`
	} `json:"roots"`
	Version int `json:"version"`
}

// chromeBookmarkNode Chrome 书签节点
type chromeBookmarkNode struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Type         string                `json:"type"` // "url" or "folder"
	URL          string                `json:"url,omitempty"`
	DateAdded    string                `json:"date_added,omitempty"`
	Children     []chromeBookmarkNode  `json:"children,omitempty"`
}

// Parse 解析 Chrome 书签
func (p *ChromeParser) Parse() ([]bookmark.Bookmark, error) {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	var chromeData chromeBookmark
	if err := json.Unmarshal(data, &chromeData); err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks JSON: %w", err)
	}

	var bookmarks []bookmark.Bookmark

	// 解析书签栏
	bookmarks = append(bookmarks, p.parseNode(&chromeData.Roots.BookmarkBar, "书签栏")...)

	// 解析其他书签
	bookmarks = append(bookmarks, p.parseNode(&chromeData.Roots.Other, "其他书签")...)

	// 解析同步书签
	bookmarks = append(bookmarks, p.parseNode(&chromeData.Roots.Synced, "移动设备书签")...)

	return bookmarks, nil
}

// parseNode 递归解析书签节点
func (p *ChromeParser) parseNode(node *chromeBookmarkNode, folderPath string) []bookmark.Bookmark {
	var bookmarks []bookmark.Bookmark

	if node.Type == "url" && node.URL != "" {
		// 解析 Chrome 时间格式（WebKit timestamp: microseconds since 1601-01-01）
		addedAt := time.Now()
		if node.DateAdded != "" {
			if timestamp, err := parseChromeTimestamp(node.DateAdded); err == nil {
				addedAt = timestamp
			}
		}

		bookmarks = append(bookmarks, bookmark.Bookmark{
			ID:      generateID(node.URL),
			Title:   node.Name,
			URL:     node.URL,
			Folder:  folderPath,
			Browser: "chrome",
			AddedAt: addedAt,
		})
	} else if node.Type == "folder" && len(node.Children) > 0 {
		// 递归处理子节点
		for i := range node.Children {
			newPath := folderPath + "/" + node.Name
			bookmarks = append(bookmarks, p.parseNode(&node.Children[i], newPath)...)
		}
	}

	return bookmarks
}

// parseChromeTimestamp 解析 Chrome 时间戳
func parseChromeTimestamp(timestamp string) (time.Time, error) {
	// Chrome 使用 WebKit timestamp: microseconds since 1601-01-01 00:00:00 UTC
	var microseconds int64
	if _, err := fmt.Sscanf(timestamp, "%d", &microseconds); err != nil {
		return time.Time{}, err
	}

	// 转换为 Unix timestamp
	seconds := microseconds / 1000000
	nanos := (microseconds % 1000000) * 1000

	// WebKit epoch: 1601-01-01
	webkitEpoch := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	return webkitEpoch.Add(time.Duration(seconds)*time.Second + time.Duration(nanos)*time.Nanosecond), nil
}

// generateID 生成书签 ID（URL hash）
func generateID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:16]
}
```

**Step 2: 提交 Chrome 解析器**

```bash
git add plugins/bookmark/browser/chrome.go
git commit -m "feat(bookmark): 实现 Chrome 书签解析器

- 支持 macOS/Windows/Linux 跨平台
- 递归解析书签树结构
- 解析 Chrome WebKit 时间戳格式

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 实现缓存管理

**Files:**
- Create: `plugins/bookmark/cache.go`

**Step 1: 创建缓存管理器**

Create `plugins/bookmark/cache.go`:

```go
package bookmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheFileName = "bookmark_cache.json"
	cacheVersion  = 1
)

// Cache 缓存管理器
type Cache struct {
	dataDir      string
	cacheFile    string
	expiryDays   int
}

// NewCache 创建缓存管理器
func NewCache(dataDir string) (*Cache, error) {
	cacheDir := filepath.Join(dataDir, "bookmark")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		dataDir:    dataDir,
		cacheFile:  filepath.Join(cacheDir, cacheFileName),
		expiryDays: 7, // 默认 7 天
	}, nil
}

// Load 加载缓存
func (c *Cache) Load() (*CacheData, error) {
	data, err := os.ReadFile(c.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 缓存不存在
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &cacheData, nil
}

// Save 保存缓存
func (c *Cache) Save(data *CacheData) error {
	data.Version = cacheVersion
	data.LastSync = time.Now()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(c.cacheFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// IsExpired 检查缓存是否过期
func (c *Cache) IsExpired() bool {
	data, err := c.Load()
	if err != nil || data == nil {
		return true
	}

	expiryTime := data.LastSync.AddDate(0, 0, c.expiryDays)
	return time.Now().After(expiryTime)
}

// Clear 清除缓存
func (c *Cache) Clear() error {
	if err := os.Remove(c.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	return nil
}

// Status 获取缓存状态
func (c *Cache) Status() map[string]interface{} {
	data, err := c.Load()
	if err != nil || data == nil {
		return map[string]interface{}{
			"available": false,
		}
	}

	return map[string]interface{}{
		"available":     true,
		"last_sync":     data.LastSync.Format("2006-01-02 15:04:05"),
		"total_count":   len(data.Bookmarks),
		"browser_stats": data.BrowserStats,
		"is_expired":    c.IsExpired(),
	}
}
```

**Step 2: 提交缓存管理**

```bash
git add plugins/bookmark/cache.go
git commit -m "feat(bookmark): 实现书签缓存管理

- JSON 格式缓存到 ~/.ltools/bookmark/
- 7 天有效期检查
- 缓存状态查询接口

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 实现搜索引擎

**Files:**
- Create: `plugins/bookmark/search.go`

**Step 1: 创建搜索引擎**

Create `plugins/bookmark/search.go`:

```go
package bookmark

import (
	"sort"
	"strings"
)

// SearchEngine 搜索引擎
type SearchEngine struct {
	bookmarks []Bookmark
	config    *BookmarkConfig
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(config *BookmarkConfig) *SearchEngine {
	return &SearchEngine{
		config: config,
	}
}

// SetBookmarks 设置书签数据
func (e *SearchEngine) SetBookmarks(bookmarks []Bookmark) {
	e.bookmarks = bookmarks
}

// Search 搜索书签
func (e *SearchEngine) Search(query string) []SearchResult {
	if query == "" || len(e.bookmarks) == 0 {
		return []SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []SearchResult

	for i := range e.bookmarks {
		bm := &e.bookmarks[i]
		score, matchType := e.calculateScore(bm, query)

		if score > 0 {
			results = append(results, SearchResult{
				Bookmark:  *bm,
				Score:     score,
				MatchType: matchType,
			})
		}
	}

	// 按分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 限制结果数量
	if len(results) > e.config.MaxResults {
		results = results[:e.config.MaxResults]
	}

	return results
}

// calculateScore 计算匹配分数
func (e *SearchEngine) calculateScore(bm *Bookmark, query string) (int, string) {
	titleLower := strings.ToLower(bm.Title)
	urlLower := strings.ToLower(bm.URL)
	folderLower := strings.ToLower(bm.Folder)

	// 1. 标题前缀匹配（最高优先级）
	if strings.HasPrefix(titleLower, query) {
		score := 100 + (100-len(bm.Title))/10
		return score, "prefix"
	}

	// 2. 标题完全匹配
	if titleLower == query {
		return 95, "exact"
	}

	// 3. 标题包含匹配
	if idx := strings.Index(titleLower, query); idx >= 0 {
		score := 80 + (100-idx)/10
		return score, "contains"
	}

	// 4. URL 包含匹配
	if strings.Contains(urlLower, query) {
		return 60, "url"
	}

	// 5. 文件夹匹配
	if strings.Contains(folderLower, query) {
		return 40, "folder"
	}

	return 0, ""
}
```

**Step 2: 提交搜索引擎**

```bash
git add plugins/bookmark/search.go
git commit -m "feat(bookmark): 实现书签搜索引擎

- 多维度匹配（标题前缀/包含/URL/文件夹）
- 分数计算和排序
- 可配置最大结果数

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 完善插件主文件（集成解析器和搜索）

**Files:**
- Modify: `plugins/bookmark/plugin.go`

**Step 1: 添加解析器集成**

Update `plugins/bookmark/plugin.go`, add imports and update Sync method:

```go
// 在文件顶部添加 imports
import (
	"fmt"
	// ... 现有 imports
	"ltools/plugins/bookmark/browser"
)

// 在 BookmarkPlugin 结构体中添加字段
type BookmarkPlugin struct {
	*plugins.BasePlugin
	app      *application.App
	dataDir  string
	config   *BookmarkConfig
	cache    *Cache
	search   *SearchEngine
	parsers  []browser.BookmarkParser
}

// 更新 SetDataDir 方法
func (p *BookmarkPlugin) SetDataDir(dataDir string) error {
	p.dataDir = dataDir

	// 初始化缓存
	cache, err := NewCache(dataDir)
	if err != nil {
		p.app.Logger.Error("Failed to create cache: " + err.Error())
		return err
	}
	p.cache = cache

	// 初始化搜索引擎
	p.search = NewSearchEngine(p.config)

	// 初始化浏览器解析器
	p.initParsers()

	// 后台异步加载书签
	go p.loadBookmarks()

	return nil
}

// initParsers 初始化浏览器解析器
func (p *BookmarkPlugin) initParsers() {
	p.parsers = []browser.BookmarkParser{
		browser.NewChromeParser(),
		// Safari 和 Firefox 将在 Phase 2 添加
	}
}

// Sync 同步所有浏览器书签
func (p *BookmarkPlugin) Sync() error {
	p.app.Logger.Info("[Bookmark] Starting sync...")

	var allBookmarks []Bookmark
	browserStats := make(map[string]int)

	for _, parser := range p.parsers {
		if !parser.IsAvailable() {
			p.app.Logger.Info(fmt.Sprintf("[Bookmark] %s not available, skipping", parser.Name()))
			continue
		}

		bookmarks, err := parser.Parse()
		if err != nil {
			p.app.Logger.Error(fmt.Sprintf("[Bookmark] Failed to parse %s: %v", parser.Name(), err))
			continue
		}

		p.app.Logger.Info(fmt.Sprintf("[Bookmark] Parsed %d bookmarks from %s", len(bookmarks), parser.Name()))
		allBookmarks = append(allBookmarks, bookmarks...)
		browserStats[parser.Name()] = len(bookmarks)
	}

	// 更新搜索引擎
	p.search.SetBookmarks(allBookmarks)

	// 保存到缓存
	cacheData := &CacheData{
		Bookmarks:    allBookmarks,
		BrowserStats: browserStats,
	}

	if err := p.cache.Save(cacheData); err != nil {
		p.app.Logger.Error("[Bookmark] Failed to save cache: " + err.Error())
		return err
	}

	p.app.Logger.Info(fmt.Sprintf("[Bookmark] Sync completed, total %d bookmarks", len(allBookmarks)))
	return nil
}

// Search 搜索书签
func (p *BookmarkPlugin) Search(query string) []SearchResult {
	if p.search == nil {
		return []SearchResult{}
	}
	return p.search.Search(query)
}

// GetCacheStatus 获取缓存状态
func (p *BookmarkPlugin) GetCacheStatus() map[string]interface{} {
	if p.cache == nil {
		return map[string]interface{}{
			"available": false,
		}
	}
	return p.cache.Status()
}
```

**Step 2: 提交集成更新**

```bash
git add plugins/bookmark/plugin.go
git commit -m "feat(bookmark): 集成解析器和搜索引擎

- 初始化浏览器解析器列表
- 实现完整同步流程
- 集成搜索引擎

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: 创建前端服务接口

**Files:**
- Create: `plugins/bookmark/service.go`

**Step 1: 创建服务接口**

Create `plugins/bookmark/service.go`:

```go
package bookmark

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// BookmarkService 暴露给前端的服务接口
type BookmarkService struct {
	plugin *BookmarkPlugin
	app    *application.App
}

// NewBookmarkService 创建服务
func NewBookmarkService(app *application.App, plugin *BookmarkPlugin) *BookmarkService {
	return &BookmarkService{
		plugin: plugin,
		app:    app,
	}
}

// Search 搜索书签
func (s *BookmarkService) Search(query string) ([]SearchResult, error) {
	return s.plugin.Search(query), nil
}

// Sync 同步书签
func (s *BookmarkService) Sync() error {
	return s.plugin.Sync()
}

// GetCacheStatus 获取缓存状态
func (s *BookmarkService) GetCacheStatus() (map[string]interface{}, error) {
	return s.plugin.GetCacheStatus(), nil
}

// OpenURL 在浏览器中打开 URL
func (s *BookmarkService) OpenURL(url string) error {
	// 使用系统默认浏览器打开
	// Wails v3 提供了 application.Browser.OpenURL 方法
	return s.app.Browser.OpenURL(url)
}

// ExportHTML 导出为 HTML 格式（Phase 2 实现）
func (s *BookmarkService) ExportHTML(outputPath string) error {
	return nil
}

// ExportJSON 导出为 JSON 格式（Phase 2 实现）
func (s *BookmarkService) ExportJSON(outputPath string) error {
	return nil
}
```

**Step 2: 提交服务接口**

```bash
git add plugins/bookmark/service.go
git commit -m "feat(bookmark): 创建前端服务接口

- Search: 搜索书签
- Sync: 手动同步
- GetCacheStatus: 获取缓存状态
- OpenURL: 在浏览器打开链接

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: 注册插件到 main.go

**Files:**
- Modify: `main.go`

**Step 1: 导入插件包**

在 `main.go` 的 import 部分添加：

```go
import (
	// ... 现有 imports
	"ltools/plugins/bookmark"
)
```

**Step 2: 注册事件**

在 `main.go` 的 `init()` 函数中添加事件注册：

```go
func init() {
	// ... 现有事件注册

	// Register custom events for the bookmark plugin
	application.RegisterEvent[string]("bookmark:sync-started")
	application.RegisterEvent[string]("bookmark:sync-completed")
	application.RegisterEvent[string]("bookmark:sync-error")
	application.RegisterEvent[string]("bookmark:exported")
}
```

**Step 3: 注册插件和服务**

在 `main()` 函数中找到插件注册部分，添加：

```go
// 在插件管理器注册部分添加
bookmarkPlugin := bookmark.NewBookmarkPlugin()
pluginManager.Register(bookmarkPlugin)

// 在服务注册部分添加
bookmarkService := bookmark.NewBookmarkService(app, bookmarkPlugin)
app.RegisterService(application.NewService(bookmarkService))
```

**Step 4: 提交注册**

```bash
git add main.go
git commit -m "feat(bookmark): 注册书签插件到 main.go

- 注册书签相关事件
- 注册插件到插件管理器
- 注册前端服务接口

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: 生成 TypeScript 绑定

**Files:**
- Generated: `frontend/bindings/bookmark/`

**Step 1: 生成绑定**

```bash
task common:generate:bindings
```

**Step 2: 验证绑定文件**

检查是否生成了 `frontend/bindings/bookmark/` 目录和相关文件。

**Step 3: 提交绑定文件**

```bash
git add frontend/bindings/bookmark/
git commit -m "feat(bookmark): 生成 TypeScript 绑定

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 10: 创建前端书签 Hook

**Files:**
- Create: `frontend/src/hooks/useBookmarks.ts`

**Step 1: 创建 Hook**

Create `frontend/src/hooks/useBookmarks.ts`:

```typescript
import { useState, useCallback } from 'react';
import { BookmarkService } from '../bindings/bookmark/BookmarkService';

export interface Bookmark {
  id: string;
  title: string;
  url: string;
  folder: string;
  browser: string;
  added_at: string;
}

export interface SearchResult {
  bookmark: Bookmark;
  score: number;
  match_type: string;
}

export interface CacheStatus {
  available: boolean;
  last_sync: string;
  total_count: number;
  browser_stats: Record<string, number>;
  is_expired: boolean;
}

export function useBookmarks() {
  const [searching, setSearching] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const search = useCallback(async (query: string): Promise<SearchResult[]> => {
    if (!query.trim()) {
      return [];
    }

    setSearching(true);
    setError(null);

    try {
      const results = await BookmarkService.Search(query);
      return results || [];
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '搜索失败';
      setError(errorMsg);
      return [];
    } finally {
      setSearching(false);
    }
  }, []);

  const sync = useCallback(async (): Promise<boolean> => {
    setSyncing(true);
    setError(null);

    try {
      await BookmarkService.Sync();
      return true;
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '同步失败';
      setError(errorMsg);
      return false;
    } finally {
      setSyncing(false);
    }
  }, []);

  const getCacheStatus = useCallback(async (): Promise<CacheStatus | null> => {
    try {
      const status = await BookmarkService.GetCacheStatus();
      return status as CacheStatus;
    } catch (err) {
      return null;
    }
  }, []);

  const openURL = useCallback(async (url: string): Promise<boolean> => {
    try {
      await BookmarkService.OpenURL(url);
      return true;
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '打开链接失败';
      setError(errorMsg);
      return false;
    }
  }, []);

  return {
    search,
    sync,
    getCacheStatus,
    openURL,
    searching,
    syncing,
    error,
  };
}
```

**Step 2: 提交 Hook**

```bash
git add frontend/src/hooks/useBookmarks.ts
git commit -m "feat(bookmark): 创建书签管理 Hook

- search: 搜索书签
- sync: 同步书签
- getCacheStatus: 获取缓存状态
- openURL: 打开链接

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 11: 创建书签小部件组件

**Files:**
- Create: `frontend/src/components/BookmarkWidget.tsx`

**Step 1: 创建小部件**

Create `frontend/src/components/BookmarkWidget.tsx`:

```typescript
import React, { useState, useEffect, useRef } from 'react';
import { useBookmarks, SearchResult } from '../hooks/useBookmarks';
import { Icon } from './Icon';

interface BookmarkWidgetProps {
  query: string;
  onSelect?: () => void;
}

export const BookmarkWidget: React.FC<BookmarkWidgetProps> = ({ query, onSelect }) => {
  const { search, openURL } = useBookmarks();
  const [results, setResults] = useState<SearchResult[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [searching, setSearching] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // 搜索书签
  useEffect(() => {
    const searchBookmarks = async () => {
      if (!query.trim()) {
        setResults([]);
        return;
      }

      setSearching(true);
      const searchResults = await search(query);
      setResults(searchResults);
      setSelectedIndex(0);
      setSearching(false);
    };

    const debounce = setTimeout(searchBookmarks, 200);
    return () => clearTimeout(debounce);
  }, [query, search]);

  // 处理键盘事件
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (results.length === 0) return;

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex((prev) => (prev + 1) % results.length);
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex((prev) => (prev - 1 + results.length) % results.length);
          break;
        case 'Enter':
          e.preventDefault();
          handleSelect(results[selectedIndex]);
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [results, selectedIndex]);

  // 处理选择书签
  const handleSelect = async (result: SearchResult) => {
    await openURL(result.bookmark.url);
    onSelect?.();
  };

  // 获取浏览器图标
  const getBrowserIcon = (browser: string) => {
    const icons: Record<string, string> = {
      chrome: '🌐',
      safari: '🧭',
      firefox: '🦊',
    };
    return icons[browser] || '🔖';
  };

  if (!query.trim()) {
    return (
      <div className="p-4 text-center text-gray-400">
        输入关键词搜索浏览器书签
      </div>
    );
  }

  if (searching) {
    return (
      <div className="p-4 text-center text-gray-400">
        <Icon name="loading" className="inline animate-spin mr-2" />
        搜索中...
      </div>
    );
  }

  if (results.length === 0) {
    return (
      <div className="p-4 text-center text-gray-400">
        未找到匹配的书签
      </div>
    );
  }

  return (
    <div ref={containerRef} className="max-h-96 overflow-y-auto">
      {results.map((result, index) => (
        <div
          key={result.bookmark.id}
          className={`px-4 py-3 cursor-pointer transition-colors ${
            index === selectedIndex
              ? 'bg-purple-500/20 border-l-2 border-purple-500'
              : 'hover:bg-white/5'
          }`}
          onClick={() => handleSelect(result)}
          onMouseEnter={() => setSelectedIndex(index)}
        >
          <div className="flex items-start gap-3">
            <span className="text-xl flex-shrink-0">
              {getBrowserIcon(result.bookmark.browser)}
            </span>
            <div className="flex-1 min-w-0">
              <div className="text-white font-medium truncate">
                {result.bookmark.title}
              </div>
              <div className="text-gray-400 text-sm truncate">
                {result.bookmark.url}
              </div>
              {result.bookmark.folder && (
                <div className="text-gray-500 text-xs mt-1 truncate">
                  📁 {result.bookmark.folder}
                </div>
              )}
            </div>
            <div className="text-xs text-gray-500 flex-shrink-0">
              {result.match_type}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};
```

**Step 2: 提交小部件**

```bash
git add frontend/src/components/BookmarkWidget.tsx
git commit -m "feat(bookmark): 创建书签搜索小部件

- 实时搜索（200ms 防抖）
- 键盘导航（上下箭头 + 回车）
- 显示浏览器图标和文件夹路径
- 高亮选中项

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 12: 测试和验证

**Step 1: 构建后端**

```bash
go build
```

Expected: 构建成功，无错误

**Step 2: 生成绑定（如果尚未生成）**

```bash
task common:generate:bindings
```

**Step 3: 构建前端**

```bash
cd frontend && npm run build
```

Expected: 构建成功，无错误

**Step 4: 运行应用测试**

```bash
task dev
```

手动测试：
1. 启动应用
2. 在搜索窗口输入 "书签" 或 "bm"
3. 输入搜索关键词
4. 验证搜索结果
5. 点击书签验证是否在浏览器中打开

**Step 5: 提交最终测试**

```bash
git add -A
git commit -m "test(bookmark): 验证基础功能

- 后端构建成功
- 前端构建成功
- Chrome 书签解析正常
- 搜索功能正常
- 打开链接功能正常

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Phase 1 完成清单

- [x] 插件基础结构和类型定义
- [x] 浏览器解析器接口
- [x] Chrome 书签解析器
- [x] 缓存管理
- [x] 搜索引擎
- [x] 前端服务接口
- [x] 注册到 main.go
- [x] 生成 TypeScript 绑定
- [x] 前端 Hook
- [x] 书签小部件组件
- [x] 测试验证

---

## Phase 2 任务（待实施）

1. **Safari 解析器** - 使用 plist 库解析 Safari 书签
2. **Firefox 解析器** - 使用 SQLite 查询 Firefox 书签
3. **拼音搜索** - 集成 go-pinyin 库
4. **独立管理页面** - 完整的书签管理界面
5. **导出功能** - HTML 和 JSON 导出

---

## Phase 3 任务（待实施）

1. **高级排序** - 访问频率、智能推荐
2. **性能优化** - 大量书签加载优化
3. **错误处理** - 浏览器不可用、权限问题
