package password

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "password.builtin"
	PluginName    = "密码生成器"
	PluginVersion = "1.0.0"
)

// PasswordPlugin provides password generation functionality
// Note: The actual password generation is done on the frontend using crypto.getRandomValues()
// This plugin exists for plugin system integration and management purposes
type PasswordPlugin struct {
	*plugins.BasePlugin
	app *application.App
}

// NewPasswordPlugin creates a new password plugin instance
func NewPasswordPlugin() *PasswordPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "生成安全的随机密码，支持自定义长度、字符类型和强度检测",
		Icon:        "key",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:     []string{"密码", "生成器", "随机", "password", "generator", "random"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &PasswordPlugin{
		BasePlugin: base,
	}
}

// Init initializes plugin
func (p *PasswordPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceStartup is called when application starts
func (p *PasswordPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown is called when application shuts down
func (p *PasswordPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if plugin is enabled
func (p *PasswordPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *PasswordPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}
