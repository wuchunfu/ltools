package plugins

import (
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/plugins/applauncher/apps"
)

// AppLauncherService 接口定义（避免循环依赖）
type AppLauncherService interface {
	Search(query string) ([]*apps.AppInfo, error)
	LaunchApp(appID string) error
}

// SearchWindowService manages the global search window (Spotlight/Alfred-like)
type SearchWindowService struct {
	app                   *application.App
	pluginService         *PluginService
	shortcutService       *ShortcutService
	appLauncherService    AppLauncherService
	searchWindow          *application.WebviewWindow
	mainWindow            *application.WebviewWindow // 主窗口引用
	isVisible             bool
	lastPosition          *windowPosition
	searchHotkeyPluginID  string // Plugin ID for the search hotkey
	mu                    sync.RWMutex
}

// windowPosition stores the window position
type windowPosition struct {
	X int
	Y int
}

// SearchResult represents a plugin search result
type SearchResult struct {
	PluginID      string   `json:"pluginId"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Icon          string   `json:"icon"`
	MatchedFields []string `json:"matchedFields"` // Fields that matched the search query
	Type          string   `json:"type"`          // "plugin" or "app"
	AppID         string   `json:"appId,omitempty"` // For apps
}

// NewSearchWindowService creates a new search window service
func NewSearchWindowService(app *application.App, pluginService *PluginService, shortcutService *ShortcutService) *SearchWindowService {
	return &SearchWindowService{
		app:             app,
		pluginService:   pluginService,
		shortcutService: shortcutService,
		isVisible:       false,
	}
}

// SetAppLauncherService 设置应用启动器服务
func (s *SearchWindowService) SetAppLauncherService(service AppLauncherService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appLauncherService = service
}

// SetMainWindow 设置主窗口引用
func (s *SearchWindowService) SetMainWindow(window *application.WebviewWindow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mainWindow = window
}

// ServiceStartup is called when the application starts
func (s *SearchWindowService) ServiceStartup(app *application.App) error {
	app.Logger.Info("[SearchWindowService] Starting up...")

	// Register the search hotkey (Cmd+Space / Ctrl+Space)
	// We'll use a special plugin ID for the search window itself
	s.searchHotkeyPluginID = "search.window.builtin"

	// Note: The actual hotkey registration will be done by ShortcutService
	// We just need to make sure we listen for the search window events

	return nil
}

// ServiceShutdown is called when the application shuts down
func (s *SearchWindowService) ServiceShutdown(app *application.App) error {
	app.Logger.Info("[SearchWindowService] Shutting down...")

	// Hide and clean up the search window if it exists
	if s.searchWindow != nil {
		s.searchWindow.Hide()
	}

	return nil
}

// Show displays the search window and brings it to front
func (s *SearchWindowService) Show() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.app.Logger.Info("[SearchWindowService] Showing search window...")

	// Create window if it doesn't exist
	if s.searchWindow == nil {
		if err := s.createWindow(); err != nil {
			s.app.Logger.Error(fmt.Sprintf("[SearchWindowService] Failed to create search window: %v", err))
			return err
		}
	}

	// Restore last position if saved, otherwise use default centered position
	if s.lastPosition != nil {
		s.searchWindow.SetPosition(s.lastPosition.X, s.lastPosition.Y)
	}

	// Show and focus the window
	s.searchWindow.Show()
	s.searchWindow.SetAlwaysOnTop(true)

	// Emit event to frontend to focus the search input
	s.app.Event.Emit("search:opened", "")

	s.isVisible = true
	s.app.Logger.Info("[SearchWindowService] Search window is now visible")

	return nil
}

// Hide hides the search window
func (s *SearchWindowService) Hide() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.app.Logger.Info("[SearchWindowService] Hiding search window...")

	if s.searchWindow == nil {
		return fmt.Errorf("search window not created")
	}

	// Save current position
	x, y := s.searchWindow.Position()
	s.lastPosition = &windowPosition{X: x, Y: y}

	s.searchWindow.Hide()
	s.isVisible = false

	// Emit event to frontend
	s.app.Event.Emit("search:closed", "")

	s.app.Logger.Info("[SearchWindowService] Search window hidden")
	return nil
}

// Toggle toggles the search window visibility
func (s *SearchWindowService) Toggle() error {
	s.mu.RLock()
	isVisible := s.isVisible
	s.mu.RUnlock()

	if isVisible {
		return s.Hide()
	}
	return s.Show()
}

// Search performs a plugin search based on the query
func (s *SearchWindowService) Search(query string) ([]*SearchResult, error) {
	s.app.Logger.Info(fmt.Sprintf("[SearchWindowService] Searching for: %s", query))

	results := make([]*SearchResult, 0)

	// 1. 搜索插件
	plugins := s.pluginService.List()
	for _, plugin := range plugins {
		// Skip disabled plugins
		if plugin.State != PluginStateEnabled {
			continue
		}

		// Check if plugin matches the query
		matchedFields := s.matchPlugin(plugin, query)
		if len(matchedFields) > 0 {
			results = append(results, &SearchResult{
				PluginID:      plugin.ID,
				Name:          plugin.Name,
				Description:   plugin.Description,
				Icon:          s.getPluginIcon(plugin.ID),
				MatchedFields: matchedFields,
				Type:          "plugin",
			})
		}
	}

	// 注意：应用搜索由前端直接调用 AppLauncherService.Search() 处理
	// 这样可以避免 Wails 绑定警告，并保持更清晰的职责分离

	s.app.Logger.Info(fmt.Sprintf("[SearchWindowService] Found %d plugin results", len(results)))
	return results, nil
}

// OpenPlugin opens a plugin by sending a shortcut event
func (s *SearchWindowService) OpenPlugin(pluginID string) error {
	s.app.Logger.Info(fmt.Sprintf("[SearchWindowService] Opening plugin: %s", pluginID))

	// Show the main window before opening the plugin
	s.mu.RLock()
	mainWindow := s.mainWindow
	s.mu.RUnlock()

	if mainWindow != nil {
		s.app.Logger.Info("[SearchWindowService] Showing main window")
		mainWindow.Show()
	}

	// Emit the shortcut triggered event to navigate to the plugin
	s.app.Event.Emit("shortcut:triggered", pluginID)

	// Hide the search window after opening
	return s.Hide()
}

// OpenApp 打开应用
func (s *SearchWindowService) OpenApp(appID string) error {
	s.app.Logger.Info(fmt.Sprintf("[SearchWindowService] Opening app: %s", appID))

	s.mu.RLock()
	appLauncher := s.appLauncherService
	s.mu.RUnlock()

	if appLauncher == nil {
		return fmt.Errorf("app launcher service not available")
	}

	if err := appLauncher.LaunchApp(appID); err != nil {
		return err
	}

	// Hide the search window after opening
	return s.Hide()
}

// OpenItem 根据类型打开插件或应用（统一接口）
func (s *SearchWindowService) OpenItem(resultType, id string) error {
	switch resultType {
	case "plugin":
		return s.OpenPlugin(id)
	case "app":
		return s.OpenApp(id)
	default:
		return fmt.Errorf("unknown result type: %s", resultType)
	}
}

// matchPlugin checks if a plugin matches the search query
func (s *SearchWindowService) matchPlugin(plugin *PluginMetadata, query string) []string {
	if query == "" {
		return []string{} // Empty query matches nothing by default
	}

	matchedFields := make([]string, 0)

	// Simple case-insensitive matching
	queryLower := toLower(query)

	// Check name
	if contains(toLower(plugin.Name), queryLower) {
		matchedFields = append(matchedFields, "name")
	}

	// Check description
	if contains(toLower(plugin.Description), queryLower) {
		matchedFields = append(matchedFields, "description")
	}

	// Check keywords
	for _, keyword := range plugin.Keywords {
		if contains(toLower(keyword), queryLower) {
			matchedFields = append(matchedFields, "keyword")
			break
		}
	}

	// Check author
	if contains(toLower(plugin.Author), queryLower) {
		matchedFields = append(matchedFields, "author")
	}

	return matchedFields
}

// getPluginIcon returns the icon for a plugin
func (s *SearchWindowService) getPluginIcon(pluginID string) string {
	// This could be enhanced to return actual icon data
	// For now, return emoji based on plugin ID
	iconMap := map[string]string{
		"datetime.builtin":      "🕐",
		"calculator.builtin":    "🔢",
		"clipboard.builtin":     "📋",
		"sysinfo.builtin":       "💻",
		"jsoneditor.builtin":    "📝",
		"processmanager.builtin": "⚙️",
	}

	if icon, ok := iconMap[pluginID]; ok {
		return icon
	}
	return "🧩"
}

// createWindow creates the search window
func (s *SearchWindowService) createWindow() error {
	s.app.Logger.Info("[SearchWindowService] Creating search window...")

	s.searchWindow = s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "LTools Search",
		Width:           600,
		Height:          500,
		MinWidth:        500,
		MinHeight:       400,
		Frameless:       true,
		AlwaysOnTop:     true,
		BackgroundType:  application.BackgroundTypeTransparent,
		URL:             "/search",
		InitialPosition: application.WindowCentered,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 0,
			Backdrop:               application.MacBackdropTranslucent,
		},
		DisableResize: false,
	})

	s.app.Logger.Info("[SearchWindowService] Search window created")
	return nil
}
