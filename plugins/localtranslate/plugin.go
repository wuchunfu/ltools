package localtranslate

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "localtranslate.builtin"
	PluginName    = "AI翻译"
	PluginVersion = "1.0.0"
)

// LocalTranslatePlugin provides local translation functionality
type LocalTranslatePlugin struct {
	*plugins.BasePlugin
	app    *application.App
	config *Config
}

// NewLocalTranslatePlugin creates a new local translation plugin instance
func NewLocalTranslatePlugin() *LocalTranslatePlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "多供应商 AI 翻译插件，支持 Ollama 本地离线翻译和云端 API（OpenAI、Anthropic、DeepSeek），提供中英日韩等多语言互译",
		Icon:        "translate",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{plugins.PermissionFileSystem},
		Keywords:    []string{"翻译", "translate", "AI翻译", "离线翻译", "中英日韩", "Ollama", "OpenAI", "DeepSeek", "Claude"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &LocalTranslatePlugin{
		BasePlugin: base,
	}
}

// Init initializes the plugin
func (p *LocalTranslatePlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app

	// Load or create configuration
	config, err := LoadConfig()
	if err != nil {
		// If loading fails, use default config
		p.config = DefaultConfig()
	} else {
		p.config = config
	}

	return nil
}

// ServiceStartup is called when the application starts
func (p *LocalTranslatePlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (p *LocalTranslatePlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *LocalTranslatePlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *LocalTranslatePlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// emitEvent emits an event with the given name and data
func (p *LocalTranslatePlugin) emitEvent(eventName string, data any) {
	if p.app != nil {
		p.app.Event.Emit("localtranslate:"+eventName, data)
	}
}
