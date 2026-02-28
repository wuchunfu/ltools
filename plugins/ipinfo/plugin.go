package ipinfo

import (
	"ltools/internal/plugins"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Plugin IP信息插件
type Plugin struct {
	*plugins.BasePlugin
	app     *application.App
	service *Service
}

// NewPlugin 创建新的IP信息插件
func NewPlugin() *Plugin {
	metadata := &plugins.PluginMetadata{
		ID:          "ipinfo.builtin",
		Name:        "IP信息",
		Version:     "1.0.0",
		Author:      "LTools",
		Description: "获取当前公网IP地址和地理位置信息",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{
			plugins.PermissionNetwork,
		},
		Keywords:   []string{"IP", "地址", "位置", "网络", "公网"},
		ShowInMenu: plugins.BoolPtr(true),
		HasPage:    plugins.BoolPtr(true),
	}

	return &Plugin{
		BasePlugin: plugins.NewBasePlugin(metadata),
	}
}

// Metadata 返回插件元数据
func (p *Plugin) Metadata() *plugins.PluginMetadata {
	return p.BasePlugin.Metadata()
}

// Init 初始化插件
func (p *Plugin) Init(app *application.App) error {
	p.app = app
	return nil
}

// ServiceStartup 服务启动
func (p *Plugin) ServiceStartup(app *application.App) error {
	p.service = NewService(p, app)
	return nil
}

// ServiceShutdown 服务关闭
func (p *Plugin) ServiceShutdown(app *application.App) error {
	return nil
}

// Enabled 返回插件是否启用
func (p *Plugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled 设置插件启用状态
func (p *Plugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// GetService 获取服务实例
func (p *Plugin) GetService() *Service {
	return p.service
}
