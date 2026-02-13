package clipboard

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
	clipboardpkg "ltools/internal/plugins/clipboard"
)

var debugLog *os.File

func init() {
	// Debug logging disabled - comment out to re-enable
	// Initialize debug log file in project directory
	// f, err := os.OpenFile("/Users/yanglian/code/ltools/clipboard-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// if err == nil {
	// 	debugLog = f
	// 	log.SetOutput(f)
	// 	fmt.Println("[Clipboard Plugin] Debug log initialized: /Users/yanglian/code/ltools/clipboard-debug.log")
	// 	debugWrite("Debug log initialized")
	// } else {
	// 	fmt.Println("[Clipboard Plugin] Failed to initialize debug log:", err)
	// }
	debugLog = nil // Ensure debug logging is disabled
}

// debugWrite writes a message to the debug log file
func debugWrite(format string, args ...interface{}) {
	// Debug logging disabled
	if debugLog != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		msg := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
		debugLog.WriteString(msg)
	}
	// Always return early - do nothing
	// return // Redundant return removed
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
	maxImageHistory int           // Maximum number of image items to keep
	lastClipboard  string        // Track last clipboard content to detect changes
	lastImageHash  string        // Hash of last image clipboard content
	stopMonitoring chan struct{} // Channel to stop monitoring
	ready          bool          // Flag to indicate when app is ready for clipboard operations
	imgClipboard   *clipboardpkg.ImageClipboard  // Image clipboard instance
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
		BasePlugin:       base,
		history:          make([]ClipboardItem, 0),
		maxHistory:       100, // Keep last 100 text items
		maxImageHistory:  20,  // Keep last 20 image items
		stopMonitoring:   make(chan struct{}),
	}
}

// Init initializes the plugin
func (p *ClipboardPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	p.imgClipboard = clipboardpkg.NewImageClipboard(app)
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
	// debugWrite("[Clipboard Plugin] Starting clipboard monitoring...\n")
	// fmt.Println("[Clipboard Plugin] Starting clipboard monitoring...")
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

// monitorClipboard monitors the clipboard for changes (both text and image)
func (p *ClipboardPlugin) monitorClipboard() {
	defer func() {
		if r := recover(); r != nil {
			// debugWrite("[Clipboard Plugin] Panic recovered in monitorClipboard: %v\n", r)
			// fmt.Printf("[Clipboard Plugin] Panic recovered: %v\n", r)
		}
	}()

	ticker := time.NewTicker(500 * time.Millisecond) // Check every 500ms
	defer ticker.Stop()

	// fmt.Println("[Clipboard Plugin] Monitor started, checking every 500ms")
	// debugWrite("[Clipboard Plugin] Monitor started with image support\n")

	for {
		select {
		case <-p.stopMonitoring:
			// fmt.Println("[Clipboard Plugin] Monitor stopped")
			return
		case <-ticker.C:
			// Wait until app is ready
			if !p.ready {
				// debugWrite("[Clipboard Plugin] App not ready yet, skipping clipboard check\n")
				continue
			}

			// Check both plugin enabled state and metadata state
			enabled := p.Enabled()
			metadataState := p.Metadata().State
			stateEnabled := plugins.PluginStateEnabled
			shouldMonitor := enabled || metadataState == stateEnabled

			// Debug logging
			// debugWrite("[Clipboard Plugin] Debug: enabled=%v, state=%q, shouldMonitor=%v\n",
			// 	enabled, metadataState, shouldMonitor)

			if !shouldMonitor {
				continue
			}

			// Monitor text clipboard
			p.monitorTextClipboard()

			// Monitor image clipboard
			p.monitorImageClipboard()
		}
	}
}

// monitorTextClipboard checks for text clipboard changes
func (p *ClipboardPlugin) monitorTextClipboard() {
	defer func() {
		if r := recover(); r != nil {
			debugWrite("[Clipboard Plugin] Panic in monitorTextClipboard: %v\n", r)
		}
	}()

	currentClipboard := p.getSystemClipboard()
	if currentClipboard != "" && currentClipboard != p.lastClipboard {
		// Check if this is an image we already tracked (avoid duplicates)
		if strings.HasPrefix(currentClipboard, "data:image") {
			// This is likely a base64 image that was just copied, skip
			p.lastClipboard = currentClipboard
			return
		}

		// Text clipboard has changed
		debugWrite("[Clipboard Plugin] Text clipboard changed! Length: %d\n", len(currentClipboard))
		p.lastClipboard = currentClipboard
		p.AddToHistory(currentClipboard, "text")
	}
}

// monitorImageClipboard checks for image clipboard changes
func (p *ClipboardPlugin) monitorImageClipboard() {
	defer func() {
		if r := recover(); r != nil {
			debugWrite("[Clipboard Plugin] Panic in monitorImageClipboard: %v\n", r)
		}
	}()

	if p.imgClipboard == nil {
		return
	}

	// Try to get image from clipboard
	imgData, err := p.imgClipboard.GetImage()
	if err != nil {
		// No image in clipboard is normal
		return
	}

	if len(imgData) == 0 {
		return
	}

	// Limit image size to 5MB
	if len(imgData) > 5*1024*1024 {
		debugWrite("[Clipboard Plugin] Image too large (%d bytes), skipping\n", len(imgData))
		return
	}

	// Calculate hash to detect changes
	imgHash := hashImage(imgData)
	if imgHash == p.lastImageHash {
		// Same image, skip
		return
	}

	// New image detected
	debugWrite("[Clipboard Plugin] Image clipboard changed! Size: %d bytes, Hash: %s\n", len(imgData), imgHash)
	p.lastImageHash = imgHash

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(imgData)
	dataURI := fmt.Sprintf("data:image/png;base64,%s", base64Data)

	// Add to history
	p.AddToHistory(dataURI, "image")
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
			// fmt.Printf("[Clipboard Plugin] Panic recovered in getSystemClipboard: %v\n", r)
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
		// fmt.Printf("[Clipboard Plugin] Read clipboard: %q (%d chars)\n",
		// 	truncateString(text, 50), len(text))
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

	// Skip if content is empty
	if content == "" {
		return
	}

	item := ClipboardItem{
		Content:   content,
		Timestamp: time.Now(),
		Type:      itemType,
	}

	// Add to beginning of history
	p.history = append([]ClipboardItem{item}, p.history...)

	// Apply limits based on type
	p.enforceHistoryLimits()

	// Emit appropriate event
	if itemType == "image" {
		debugWrite("[Clipboard Plugin] Added image to history (total: %d items)\n", len(p.history))
		p.emitEvent("image:new", content)
	} else {
		debugWrite("[Clipboard Plugin] Added text to history: %q (total: %d items)\n", truncateString(content, 50), len(p.history))
		p.emitEvent("new", content)
	}
	p.emitEvent("count", fmt.Sprintf("%d", len(p.history)))
}

// enforceHistoryLimits trims history to respect max limits for each type
func (p *ClipboardPlugin) enforceHistoryLimits() {
	// Separate items by type
	var textItems, imageItems []ClipboardItem
	for _, item := range p.history {
		if item.Type == "image" {
			imageItems = append(imageItems, item)
		} else {
			textItems = append(textItems, item)
		}
	}

	// Apply limits
	if len(textItems) > p.maxHistory {
		textItems = textItems[:p.maxHistory]
	}
	if len(imageItems) > p.maxImageHistory {
		imageItems = imageItems[:p.maxImageHistory]
	}

	// Rebuild history maintaining original order (interleaved)
	// We need to reconstruct in the same order they appeared
	var newHistory []ClipboardItem
	textIdx, imageIdx := 0, 0

	for _, item := range p.history {
		if item.Type == "image" && imageIdx < len(imageItems) {
			// Check if this image is in our limited list
			if imageItems[imageIdx].Timestamp.Equal(item.Timestamp) &&
			   imageItems[imageIdx].Content == item.Content {
				newHistory = append(newHistory, item)
				imageIdx++
			}
		} else if item.Type != "image" && textIdx < len(textItems) {
			// Check if this text is in our limited list
			if textItems[textIdx].Timestamp.Equal(item.Timestamp) &&
			   textItems[textIdx].Content == item.Content {
				newHistory = append(newHistory, item)
				textIdx++
			}
		}
	}

	p.history = newHistory
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

// GetImageFromClipboard gets image data from the clipboard
func (p *ClipboardPlugin) GetImageFromClipboard() string {
	if p.imgClipboard == nil {
		debugWrite("[Clipboard Plugin] Image clipboard not initialized\n")
		return ""
	}

	imgData, err := p.imgClipboard.GetImage()
	if err != nil {
		// No image in clipboard is not an error
		if _, ok := err.(*clipboardpkg.NoImageInClipboardError); !ok {
			debugWrite("[Clipboard Plugin] Failed to get image from clipboard: %v\n", err)
		}
		return ""
	}

	if len(imgData) == 0 {
		return ""
	}

	// Limit image size to 5MB
	if len(imgData) > 5*1024*1024 {
		debugWrite("[Clipboard Plugin] Image too large (%d bytes), skipping\n", len(imgData))
		return ""
	}

	// Return base64 encoded image
	base64Data := base64.StdEncoding.EncodeToString(imgData)
	return fmt.Sprintf("data:image/png;base64,%s", base64Data)
}

// CopyImageToClipboard copies a base64 encoded image to the clipboard
func (p *ClipboardPlugin) CopyImageToClipboard(base64Data string) error {
	if p.imgClipboard == nil {
		return fmt.Errorf("image clipboard not initialized")
	}

	// Remove data URI prefix if present
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	// Decode base64
	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Copy to clipboard
	if err := p.imgClipboard.SetImage(imgData); err != nil {
		return fmt.Errorf("failed to set image to clipboard: %w", err)
	}

	p.emitEvent("image:copied", "")
	return nil
}

// SaveImageToFile saves a base64 encoded image to a file using system dialog
func (p *ClipboardPlugin) SaveImageToFile(base64Data, defaultFilename string) (string, error) {
	if p.app == nil {
		return "", fmt.Errorf("app not initialized")
	}

	// Remove data URI prefix if present
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	// Decode base64
	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Use default filename if not provided
	if defaultFilename == "" {
		defaultFilename = fmt.Sprintf("clipboard_image_%d.png", time.Now().Unix())
	}

	// Open save file dialog using Wails v3 API
	filePath, err := p.app.Dialog.SaveFile().
		SetFilename(defaultFilename).
		AddFilter("PNG Images", "*.png").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()

	if err != nil {
		return "", fmt.Errorf("dialog cancelled or failed: %w", err)
	}

	if filePath == "" {
		return "", fmt.Errorf("no file selected")
	}

	// Write image data to file
	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	p.emitEvent("image:saved", filePath)
	return filePath, nil
}

// GetClipboardContentType returns the type of content in the clipboard (text, image, or unknown)
func (p *ClipboardPlugin) GetClipboardContentType() string {
	// Check for text first
	if text := p.getSystemClipboard(); text != "" {
		return "text"
	}

	// Check for image
	if imgData := p.GetImageFromClipboard(); imgData != "" {
		return "image"
	}

	return "unknown"
}

// hashImage calculates a SHA256 hash of image data for change detection
func hashImage(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for shorter hash
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
