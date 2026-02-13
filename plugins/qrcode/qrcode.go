package qrcode

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "qrcode.builtin"
	PluginName    = "二维码生成器"
	PluginVersion = "1.0.0"
)

// QrcodePlugin provides QR code generation functionality
// Note: The actual QR code generation is done on the frontend using qrcode.react
// This plugin primarily serves as a metadata container for the plugin system
type QrcodePlugin struct {
	*plugins.BasePlugin
	app *application.App
}

// NewQrcodePlugin creates a new qrcode plugin instance
func NewQrcodePlugin() *QrcodePlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "快速生成二维码，支持一键复制到剪贴板或保存为文件",
		Icon:        "qrcode",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:    []string{"二维码", "QR", "qrcode", "二维码生成", "QR码"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &QrcodePlugin{
		BasePlugin: base,
	}
}

// Init initializes the plugin
func (p *QrcodePlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceStartup is called when the application starts
func (p *QrcodePlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (p *QrcodePlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *QrcodePlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *QrcodePlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// Helper method to emit events
func (p *QrcodePlugin) emitEvent(eventName, data string) {
	if p.app != nil {
		p.app.Event.Emit("qrcode:"+eventName, data)
	}
}

// Generated emits an event when a QR code is generated
func (p *QrcodePlugin) Generated(content string) {
	p.emitEvent("generated", content)
}

// Copied emits an event when a QR code is copied to clipboard
func (p *QrcodePlugin) Copied(content string) {
	p.emitEvent("copied", content)
}

// Saved emits an event when a QR code is saved to file
func (p *QrcodePlugin) Saved(filename string) {
	p.emitEvent("saved", filename)
}
