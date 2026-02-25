package kanban

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "kanban.builtin"
	PluginName    = "看板"
	PluginVersion = "1.0.0"
)

// KanbanPlugin provides kanban board functionality
type KanbanPlugin struct {
	*plugins.BasePlugin
	app      *application.App
	config   *KanbanConfig
	dataDir  string
}

// NewKanbanPlugin creates a new Kanban plugin
func NewKanbanPlugin() *KanbanPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools",
		Description: "项目管理看板，支持多看板、自定义列、标签和子任务",
		Icon:        "kanban",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:    []string{"看板", "kanban", "项目", "任务", "管理", "project", "task", "board"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &KanbanPlugin{
		BasePlugin: base,
		config:     &KanbanConfig{},
	}
}

// Metadata returns the plugin metadata
func (p *KanbanPlugin) Metadata() *plugins.PluginMetadata {
	return p.BasePlugin.Metadata()
}

// Init initializes the plugin
func (p *KanbanPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceStartup is called when the application starts
func (p *KanbanPlugin) ServiceStartup(app *application.App) error {
	if err := p.BasePlugin.ServiceStartup(app); err != nil {
		return err
	}
	return nil
}

// ServiceShutdown is called when the application shuts down
func (p *KanbanPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *KanbanPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *KanbanPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// SetDataDir sets the data directory and loads config
func (p *KanbanPlugin) SetDataDir(dataDir string) error {
	p.dataDir = dataDir
	return p.LoadConfig(dataDir)
}
