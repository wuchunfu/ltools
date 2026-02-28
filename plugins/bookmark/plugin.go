package bookmark

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
	"ltools/plugins/bookmark/browser"
)

const (
	PluginID      = "bookmark.builtin"
	PluginName    = "书签搜索"
	PluginVersion = "1.0.0"
)

// BookmarkPlugin is a browser bookmark search plugin
type BookmarkPlugin struct {
	*plugins.BasePlugin
	app     *application.App
	dataDir string
	config  *BookmarkConfig
	cache   *Cache
	search  *SearchEngine
	parsers []browser.BookmarkParser
}

// NewBookmarkPlugin creates a new bookmark plugin instance
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
			plugins.PermissionFileSystem, // Read bookmark files
		},
		Keywords:   []string{"书签", "bookmark", "bm", "浏览器"},
		ShowInMenu: plugins.BoolPtr(false), // Triggered via search window
		HasPage:    plugins.BoolPtr(true),  // Has standalone management page
	}

	return &BookmarkPlugin{
		BasePlugin: plugins.NewBasePlugin(metadata),
		config: &BookmarkConfig{
			CacheExpiryDays: 7,
			MaxResults:      50,
			EnablePinyin:    false, // Phase 2 implementation
			TriggerKeywords: []string{"书签", "bookmark", "bm"},
		},
	}
}

// Init initializes the plugin
func (p *BookmarkPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// SetDataDir sets the data directory
func (p *BookmarkPlugin) SetDataDir(dataDir string) error {
	p.dataDir = dataDir

	// Initialize cache
	cache, err := NewCache(dataDir)
	if err != nil {
		p.app.Logger.Error("Failed to create cache: " + err.Error())
		return err
	}
	p.cache = cache

	// Initialize search engine
	p.search = NewSearchEngine(p.config)

	// Initialize parsers
	p.initParsers()

	// Load bookmarks asynchronously in background
	go p.loadBookmarks()

	return nil
}

// ServiceStartup is called when the service starts
func (p *BookmarkPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown is called when the service shuts down
func (p *BookmarkPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// loadBookmarks loads bookmarks (from cache or browser)
func (p *BookmarkPlugin) loadBookmarks() {
	// Try to load from cache
	_, err := p.cache.Load()
	if err == nil && !p.cache.IsExpired() {
		p.app.Logger.Info("Loaded bookmarks from cache")
		return
	}

	// Cache expired or missing, resync
	p.app.Logger.Info("Cache expired or missing, syncing...")
	if err := p.Sync(); err != nil {
		p.app.Logger.Error("Failed to sync bookmarks: " + err.Error())
	}
}

// Sync synchronizes bookmarks from all browsers
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

	// Update search engine
	p.search.SetBookmarks(allBookmarks)

	// Save to cache
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

// initParsers initializes browser parsers
func (p *BookmarkPlugin) initParsers() {
	p.parsers = []browser.BookmarkParser{
		browser.NewChromeParser(),
		browser.NewSafariParser(),
		browser.NewFirefoxParser(),
	}
}

// Search searches bookmarks matching the query
func (p *BookmarkPlugin) Search(query string) []SearchResult {
	if p.search == nil {
		return []SearchResult{}
	}
	return p.search.Search(query)
}

// GetCacheStatus returns cache status information
func (p *BookmarkPlugin) GetCacheStatus() map[string]interface{} {
	if p.cache == nil {
		return map[string]interface{}{
			"available": false,
		}
	}
	return p.cache.Status()
}
