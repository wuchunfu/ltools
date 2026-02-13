package hosts

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "hosts.builtin"
	PluginName    = "Hosts 管理器"
	PluginVersion = "1.0.0"
)

// HostsPlugin provides hosts file management functionality
type HostsPlugin struct {
	*plugins.BasePlugin
	app    *application.App
	config *HostsConfig
}

// NewHostsPlugin creates a new hosts plugin
func NewHostsPlugin() *HostsPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "Hosts 文件管理器，支持场景化切换和自动备份",
		Icon:        "network",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{
			plugins.PermissionFileSystem,
		},
		Keywords:   []string{"hosts", "域名", "domain", "switch"},
		// ShowInMenu 和 HasPage 将由 NewBasePlugin 设置为默认值 true
	}

	return &HostsPlugin{
		BasePlugin: plugins.NewBasePlugin(metadata),
	}
}

// Init initializes the plugin
func (p *HostsPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app

	// Initialize default config (will be loaded from disk in ServiceStartup)
	p.config = &HostsConfig{
		Version:          configVersion,
		Scenarios:        []Scenario{},
		Backups:          []Backup{},
		CurrentScenario:  "",
	}

	return nil
}

// ServiceStartup is called when the application starts
func (p *HostsPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (p *HostsPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Helper method to emit events
func (p *HostsPlugin) emitEvent(eventName, data string) {
	if p.app != nil {
		p.app.Event.Emit("hosts:"+eventName, data)
	}
}

// SetDataDir sets the data directory for the plugin
func (p *HostsPlugin) SetDataDir(dataDir string) error {
	// Load config from data directory
	if err := p.LoadConfig(dataDir); err != nil {
		return err
	}
	return nil
}
