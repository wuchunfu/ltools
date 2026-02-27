package vault

import (
	"ltools/internal/plugins"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// VaultPlugin 密码保险库插件
type VaultPlugin struct {
	*plugins.BasePlugin
	app     *application.App
	storage *Storage
}

// NewVaultPlugin 创建新的保险库插件
func NewVaultPlugin(dataDir string) *VaultPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          "vault.builtin",
		Name:        "密码保险库",
		Version:     "1.0.0",
		Author:      "LTools",
		Description: "安全存储和管理密码、登录凭证等敏感信息",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{
			plugins.PermissionFileSystem,
		},
		Keywords:    []string{"密码", "password", "保险库", "vault", "1password", "管理"},
		ShowInMenu: plugins.BoolPtr(true),
		HasPage:    plugins.BoolPtr(true),
	}

	return &VaultPlugin{
		BasePlugin: plugins.NewBasePlugin(metadata),
		storage:    NewStorage(dataDir),
	}
}

// Metadata 返回插件元数据
func (p *VaultPlugin) Metadata() *plugins.PluginMetadata {
	return p.BasePlugin.Metadata()
}

// Init 初始化插件
func (p *VaultPlugin) Init(app *application.App) error {
	p.app = app
	return nil
}

// ServiceStartup 服务启动
func (p *VaultPlugin) ServiceStartup(app *application.App) error {
	return nil
}

// ServiceShutdown 服务关闭
func (p *VaultPlugin) ServiceShutdown(app *application.App) error {
	// 锁定保险库，清除内存中的敏感数据
	if p.storage != nil {
		p.storage.Clear()
	}
	return nil
}

// Enabled 返回插件是否启用
func (p *VaultPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled 设置插件启用状态
func (p *VaultPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}
