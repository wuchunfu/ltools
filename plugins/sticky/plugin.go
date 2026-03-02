package sticky

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "sticky.builtin"
	PluginName    = "便利贴"
	PluginVersion = "1.0.0"
)

// StickyPlugin provides sticky note functionality
type StickyPlugin struct {
	*plugins.BasePlugin
	app           *application.App
	config        *StickyConfig
	dataDir       string
	windowManager *WindowManager
}

// NewStickyPlugin creates a new sticky plugin
func NewStickyPlugin() *StickyPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools",
		Description: "创建和管理多个便利贴窗口",
		Icon:        "sticky",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:    []string{"便利贴", "便签", "便条", "note", "sticky"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &StickyPlugin{
		BasePlugin: base,
		config:     &StickyConfig{},
	}
}

// Metadata returns the plugin metadata
func (p *StickyPlugin) Metadata() *plugins.PluginMetadata {
	return p.BasePlugin.Metadata()
}

// Init initializes the plugin
func (p *StickyPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceStartup is called when the application starts
func (p *StickyPlugin) ServiceStartup(app *application.App) error {
	if err := p.BasePlugin.ServiceStartup(app); err != nil {
		return err
	}
	p.app = app

	// Load config if dataDir is set
	if p.dataDir != "" {
		if err := p.LoadConfig(p.dataDir); err != nil {
			return fmt.Errorf("failed to load sticky config: %w", err)
		}
	}

	// Create window manager (but don't auto-restore windows)
	// Windows will only be created when user explicitly opens them
	p.windowManager = NewWindowManager(app)

	return nil
}

// ServiceShutdown is called when the application shuts down
func (p *StickyPlugin) ServiceShutdown(app *application.App) error {
	// Close all sticky windows
	if p.windowManager != nil {
		p.windowManager.CloseAll()
	}

	// Save config
	if p.dataDir != "" && p.config != nil {
		if err := p.SaveConfig(p.dataDir); err != nil {
			return fmt.Errorf("failed to save sticky config: %w", err)
		}
	}

	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *StickyPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *StickyPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// SetDataDir sets the data directory and loads config
func (p *StickyPlugin) SetDataDir(dataDir string) error {
	p.dataDir = dataDir

	// Load config
	if err := p.LoadConfig(dataDir); err != nil {
		return err
	}

	// Don't auto-restore windows - user must explicitly open them
	return nil
}

// emitEvent emits a sticky event
func (p *StickyPlugin) emitEvent(eventType string, data string) {
	if p.app != nil {
		p.app.Event.Emit("sticky:"+eventType, data)
	}
}

// generateID generates a unique ID for a sticky note
func generateID() string {
	return fmt.Sprintf("note-%d", time.Now().UnixNano())
}
