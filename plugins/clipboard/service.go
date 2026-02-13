package clipboard

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ClipboardService exposes Clipboard functionality to the frontend
type ClipboardService struct {
	app    *application.App
	plugin *ClipboardPlugin
}

// NewClipboardService creates a new clipboard service
func NewClipboardService(plugin *ClipboardPlugin, app *application.App) *ClipboardService {
	return &ClipboardService{
		app:    app,
		plugin: plugin,
	}
}

// ServiceStartup is called when the application starts
func (s *ClipboardService) ServiceStartup(app *application.App) error {
	return s.plugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (s *ClipboardService) ServiceShutdown(app *application.App) error {
	return s.plugin.ServiceShutdown(app)
}

// GetHistory returns the clipboard history
func (s *ClipboardService) GetHistory() []ClipboardItem {
	return s.plugin.GetHistory()
}

// AddToHistory adds an item to the clipboard history
func (s *ClipboardService) AddToHistory(content, itemType string) {
	s.plugin.AddToHistory(content, itemType)
}

// ClearHistory clears all clipboard history
func (s *ClipboardService) ClearHistory() {
	s.plugin.ClearHistory()
}

// GetLastItem returns the most recent clipboard item
func (s *ClipboardService) GetLastItem() *ClipboardItem {
	return s.plugin.GetLastItem()
}

// SearchHistory searches clipboard history by content
func (s *ClipboardService) SearchHistory(query string) []ClipboardItem {
	return s.plugin.SearchHistory(query)
}

// SetMaxHistory sets the maximum number of history items to keep
func (s *ClipboardService) SetMaxHistory(max int) {
	s.plugin.SetMaxHistory(max)
}

// GetMaxHistory returns the maximum number of history items
func (s *ClipboardService) GetMaxHistory() int {
	return s.plugin.GetMaxHistory()
}

// DeleteItem removes an item from history by index
func (s *ClipboardService) DeleteItem(index int) error {
	return s.plugin.DeleteItem(index)
}

// GetCurrentClipboard returns the current system clipboard content
func (s *ClipboardService) GetCurrentClipboard() string {
	return s.plugin.GetCurrentClipboard()
}

// GetImageFromClipboard returns the current image from clipboard as base64
func (s *ClipboardService) GetImageFromClipboard() string {
	return s.plugin.GetImageFromClipboard()
}

// CopyImageToClipboard copies a base64 encoded image to the clipboard
func (s *ClipboardService) CopyImageToClipboard(base64Data string) error {
	return s.plugin.CopyImageToClipboard(base64Data)
}

// SaveImageToFile saves a base64 encoded image to a file
func (s *ClipboardService) SaveImageToFile(base64Data, defaultFilename string) (string, error) {
	return s.plugin.SaveImageToFile(base64Data, defaultFilename)
}

// GetClipboardContentType returns the type of content in the clipboard
func (s *ClipboardService) GetClipboardContentType() string {
	return s.plugin.GetClipboardContentType()
}
