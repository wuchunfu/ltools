package screenshot

import (
	"log"
	"path/filepath"

	"ltools/internal/plugins"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	PluginID      = "screenshot.builtin"
	PluginName    = "截图工具"
	PluginVersion = "1.0.0"
)

// ScreenshotPlugin provides screenshot functionality
type ScreenshotPlugin struct {
	*plugins.BasePlugin
	app          *application.App
	editorWindow *application.WebviewWindow
	currentImage []byte
	storage      *Storage
	clipboard    *Clipboard
	tempDir      string
}

// NewScreenshotPlugin creates a new screenshot plugin instance
func NewScreenshotPlugin() *ScreenshotPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "微信风格截图工具，支持屏幕捕获、标注编辑和保存复制",
		Icon:        "screenshot",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{
			plugins.PermissionClipboard,
			plugins.PermissionFileSystem,
		},
		Keywords: []string{"截图", "屏幕", "捕获", "screenshot", "screen", "capture"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &ScreenshotPlugin{
		BasePlugin: base,
	}
}

// Init initializes the plugin
func (p *ScreenshotPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app

	// Get user data directory for temp files
	userDataDir, err := GetDefaultSavePath()
	if err != nil {
		log.Printf("[Screenshot] Failed to get default save path: %v", err)
		// Fallback to temp directory
		p.tempDir = filepath.Join(userDataDir, "ltools", "screenshots")
	} else {
		p.tempDir = userDataDir
	}

	// Initialize storage
	p.storage = NewStorage(p.tempDir)

	// Initialize clipboard
	p.clipboard = NewClipboard(app)

	log.Printf("[Screenshot] Plugin initialized with temp dir: %s", p.tempDir)
	return nil
}

// ServiceStartup is called when the application starts
func (p *ScreenshotPlugin) ServiceStartup(app *application.App) error {
	log.Printf("[Screenshot] Service startup")
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (p *ScreenshotPlugin) ServiceShutdown(app *application.App) error {
	// Clean up editor window if open
	if p.editorWindow != nil {
		p.editorWindow.Close()
		p.editorWindow = nil
	}

	log.Printf("[Screenshot] Service shutdown")
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *ScreenshotPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *ScreenshotPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// Helper method to emit events
func (p *ScreenshotPlugin) emitEvent(eventName, data string) {
	if p.app != nil {
		p.app.Event.Emit("screenshot:"+eventName, data)
	}
}

// SetCurrentImage stores the current captured image
func (p *ScreenshotPlugin) SetCurrentImage(imgData []byte) {
	p.currentImage = imgData
}

// GetCurrentImage returns the current captured image
func (p *ScreenshotPlugin) GetCurrentImage() []byte {
	return p.currentImage
}

// GetStorage returns the storage instance
func (p *ScreenshotPlugin) GetStorage() *Storage {
	return p.storage
}

// GetClipboard returns the clipboard instance
func (p *ScreenshotPlugin) GetClipboard() *Clipboard {
	return p.clipboard
}

// SetEditorWindow sets the editor window reference
func (p *ScreenshotPlugin) SetEditorWindow(window *application.WebviewWindow) {
	p.editorWindow = window
}

// GetEditorWindow returns the editor window reference
func (p *ScreenshotPlugin) GetEditorWindow() *application.WebviewWindow {
	return p.editorWindow
}

// OpenEditorWindow opens the screenshot editor window
func (p *ScreenshotPlugin) OpenEditorWindow(imgData string) (*application.WebviewWindow, error) {
	// Close existing editor window if open
	if p.editorWindow != nil {
		p.editorWindow.Close()
		p.editorWindow = nil
	}

	// Get screen dimensions for full-screen window
	displays := p.GetDisplays()
	var width, height int
	if len(displays) > 0 {
		// Use primary display dimensions
		width = displays[0].Width
		height = displays[0].Height
	} else {
		// Fallback dimensions
		width = 1920
		height = 1080
	}

	// Create editor window
	editorWindow := p.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Screenshot Editor",
		Width:  width,
		Height: height,
		Mac: application.MacWindow{
			TitleBar:                application.MacTitleBarHiddenInset,
			InvisibleTitleBarHeight: 0,
			Backdrop:                application.MacBackdropTranslucent,
		},
		BackgroundColour: application.NewRGB(0, 0, 0),
		URL:              "/screenshot-editor",
		AlwaysOnTop:      true,
	})

	p.editorWindow = editorWindow
	p.emitEvent("started", "editor opened")

	return editorWindow, nil
}

// CloseEditorWindow closes the screenshot editor window
func (p *ScreenshotPlugin) CloseEditorWindow() {
	if p.editorWindow != nil {
		p.editorWindow.Close()
		p.editorWindow = nil
		p.emitEvent("cancelled", "editor closed")
	}
}
