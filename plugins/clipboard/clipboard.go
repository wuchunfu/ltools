package clipboard

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

var debugLog *os.File

func init() {
	// Initialize debug log file in project directory
	f, err := os.OpenFile("/Users/yanglian/code/ltools/clipboard-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		debugLog = f
		log.SetOutput(f)
		fmt.Println("[Clipboard Plugin] Debug log initialized: /Users/yanglian/code/ltools/clipboard-debug.log")
		debugWrite("Debug log initialized")
	} else {
		fmt.Println("[Clipboard Plugin] Failed to initialize debug log:", err)
	}
}

// debugWrite writes a message to the debug log file
func debugWrite(format string, args ...interface{}) {
	if debugLog != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		msg := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
		debugLog.WriteString(msg)
	}
}

const (
	PluginID      = "clipboard.builtin"
	PluginName    = "剪贴板管理器"
	PluginVersion = "1.0.0"
)

// ClipboardItem represents an item in the clipboard history
type ClipboardItem struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // text, image, etc.
}

// ClipboardPlugin provides clipboard management functionality
type ClipboardPlugin struct {
	*plugins.BasePlugin
	app            *application.App
	history        []ClipboardItem
	maxHistory     int
	lastClipboard  string        // Track last clipboard content to detect changes
	stopMonitoring chan struct{} // Channel to stop monitoring
	ready          bool          // Flag to indicate when app is ready for clipboard operations
}

// NewClipboardPlugin creates a new clipboard plugin
func NewClipboardPlugin() *ClipboardPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "剪贴板历史记录管理器，自动保存复制的内容",
		Icon:        "clipboard",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{
			plugins.PermissionClipboard,
		},
		Keywords: []string{"剪贴板", "复制", "粘贴", "clipboard", "copy", "paste"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &ClipboardPlugin{
		BasePlugin:      base,
		history:         make([]ClipboardItem, 0),
		maxHistory:      100, // Keep last 100 items
		stopMonitoring:  make(chan struct{}),
	}
}

// Init initializes the plugin
func (p *ClipboardPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceStartup is called when the application starts
func (p *ClipboardPlugin) ServiceStartup(app *application.App) error {
	debugWrite("[Clipboard Plugin] ServiceStartup called\n")
	if err := p.BasePlugin.ServiceStartup(app); err != nil {
		debugWrite("[Clipboard Plugin] ServiceStartup BasePlugin error: %v\n", err)
		return err
	}
	p.app = app
	p.ready = false // Will be set to true after delay

	// Request clipboard permission
	p.emitEvent("permission:requested", "clipboard")

	// Start clipboard monitoring with a delay to ensure app is fully initialized
	debugWrite("[Clipboard Plugin] Starting clipboard monitoring...\n")
	fmt.Println("[Clipboard Plugin] Starting clipboard monitoring...")
	go func() {
		// Wait for app to be fully ready
		time.Sleep(2 * time.Second)
		p.ready = true
		debugWrite("[Clipboard Plugin] App is ready, starting monitoring loop\n")
		p.monitorClipboard()
	}()

	return nil
}

// ServiceShutdown is called when the application shuts down
func (p *ClipboardPlugin) ServiceShutdown(app *application.App) error {
	// Stop monitoring
	close(p.stopMonitoring)
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *ClipboardPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *ClipboardPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// Helper method to emit events
func (p *ClipboardPlugin) emitEvent(eventName, data string) {
	if p.app != nil {
		p.app.Event.Emit("clipboard:"+eventName, data)
	}
}

// monitorClipboard monitors the clipboard for changes
func (p *ClipboardPlugin) monitorClipboard() {
	defer func() {
		if r := recover(); r != nil {
			debugWrite("[Clipboard Plugin] Panic recovered in monitorClipboard: %v\n", r)
			fmt.Printf("[Clipboard Plugin] Panic recovered: %v\n", r)
		}
	}()

	ticker := time.NewTicker(500 * time.Millisecond) // Check every 500ms
	defer ticker.Stop()

	fmt.Println("[Clipboard Plugin] Monitor started, checking every 500ms")

	for {
		select {
		case <-p.stopMonitoring:
			fmt.Println("[Clipboard Plugin] Monitor stopped")
			return
		case <-ticker.C:
			// Wait until app is ready
			if !p.ready {
				debugWrite("[Clipboard Plugin] App not ready yet, skipping clipboard check\n")
				continue
			}

			// Check both plugin enabled state and metadata state
			enabled := p.Enabled()
			metadataState := p.Metadata().State
			stateEnabled := plugins.PluginStateEnabled
			shouldMonitor := enabled || metadataState == stateEnabled

			// Debug logging
			debugWrite("[Clipboard Plugin] Debug: enabled=%v, state=%q, stateEnabled=%q, shouldMonitor=%v\n",
				enabled, metadataState, stateEnabled, shouldMonitor)

			if shouldMonitor {
				currentClipboard := p.getSystemClipboard()
				if currentClipboard != "" && currentClipboard != p.lastClipboard {
					// Clipboard has changed
					debugWrite("[Clipboard Plugin] Clipboard changed! New content length: %d\n", len(currentClipboard))
					p.lastClipboard = currentClipboard
					p.AddToHistory(currentClipboard, "text")
				}
			}
		}
	}
}

// getSystemClipboard gets the current system clipboard content using Wails native API
func (p *ClipboardPlugin) getSystemClipboard() string {
	// Defensive checks
	if p.app == nil {
		debugWrite("[Clipboard Plugin] Error: p.app is nil\n")
		return ""
	}

	if !p.ready {
		debugWrite("[Clipboard Plugin] Error: App not ready for clipboard operations\n")
		return ""
	}

	// Use Wails native clipboard API - cross-platform and reliable
	// This must be called with panic recovery as it may fail during app initialization
	defer func() {
		if r := recover(); r != nil {
			debugWrite("[Clipboard Plugin] Panic recovered in getSystemClipboard: %v\n", r)
			fmt.Printf("[Clipboard Plugin] Panic recovered in getSystemClipboard: %v\n", r)
		}
	}()

	text, ok := p.app.Clipboard.Text()
	if !ok {
		debugWrite("[Clipboard Plugin] Failed to read clipboard using Wails API\n")
		return ""
	}

	// Debug: Log clipboard content (truncated)
	if len(text) > 0 {
		debugWrite("[Clipboard Plugin] Read clipboard: %q (%d chars)\n",
			truncateString(text, 50), len(text))
		fmt.Printf("[Clipboard Plugin] Read clipboard: %q (%d chars)\n",
			truncateString(text, 50), len(text))
	}

	return text
}

// Helper function to truncate string for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetCurrentClipboard returns the current system clipboard content
func (p *ClipboardPlugin) GetCurrentClipboard() string {
	return p.getSystemClipboard()
}

// GetHistory returns the clipboard history
func (p *ClipboardPlugin) GetHistory() []ClipboardItem {
	return p.history
}

// AddToHistory adds an item to the clipboard history
func (p *ClipboardPlugin) AddToHistory(content, itemType string) {
	debugWrite("[Clipboard Plugin] AddToHistory called with content length: %d, type: %s\n", len(content), itemType)

	item := ClipboardItem{
		Content:   content,
		Timestamp: time.Now(),
		Type:      itemType,
	}

	// Add to history
	p.history = append([]ClipboardItem{item}, p.history...)

	// Trim if exceeds max history
	if len(p.history) > p.maxHistory {
		p.history = p.history[:p.maxHistory]
	}

	// Emit event
	debugWrite("[Clipboard Plugin] Added to history: %q (total: %d items)\n", truncateString(content, 50), len(p.history))
	p.emitEvent("new", content)
	p.emitEvent("count", fmt.Sprintf("%d", len(p.history)))
}

// ClearHistory clears all clipboard history
func (p *ClipboardPlugin) ClearHistory() {
	debugWrite("[Clipboard Plugin] ClearHistory called, current history size: %d\n", len(p.history))
	p.history = make([]ClipboardItem, 0)
	// Update lastClipboard to prevent the monitor from immediately re-adding the current clipboard content
	p.lastClipboard = p.getSystemClipboard()
	debugWrite("[Clipboard Plugin] History cleared, new size: %d, lastClipboard updated\n", len(p.history))
	p.emitEvent("cleared", "")
	p.emitEvent("count", "0")
}

// GetLastItem returns the most recent clipboard item
func (p *ClipboardPlugin) GetLastItem() *ClipboardItem {
	if len(p.history) == 0 {
		return nil
	}
	return &p.history[0]
}

// SearchHistory searches clipboard history by content
func (p *ClipboardPlugin) SearchHistory(query string) []ClipboardItem {
	var results []ClipboardItem

	for _, item := range p.history {
		if contains(item.Content, query) {
			results = append(results, item)
		}
	}

	return results
}

// SetMaxHistory sets the maximum number of history items to keep
func (p *ClipboardPlugin) SetMaxHistory(max int) {
	p.maxHistory = max

	// Trim existing history if needed
	if len(p.history) > max {
		p.history = p.history[:max]
		p.emitEvent("count", fmt.Sprintf("%d", len(p.history)))
	}
}

// GetMaxHistory returns the maximum number of history items
func (p *ClipboardPlugin) GetMaxHistory() int {
	return p.maxHistory
}

// DeleteItem removes an item from history by index
func (p *ClipboardPlugin) DeleteItem(index int) error {
	if index < 0 || index >= len(p.history) {
		return fmt.Errorf("index out of bounds")
	}

	p.history = append(p.history[:index], p.history[index+1:]...)
	p.emitEvent("deleted", fmt.Sprintf("%d", index))
	p.emitEvent("count", fmt.Sprintf("%d", len(p.history)))
	return nil
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 len(s) > len(substr) && (s[0:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
