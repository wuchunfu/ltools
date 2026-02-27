package markdown

import (
	"fmt"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "markdown.builtin"
	PluginName    = "Markdown 编辑器"
	PluginVersion = "1.0.0"
)

// MarkdownStats Markdown 统计信息
type MarkdownStats struct {
	Characters int    `json:"characters"` // 字符数
	Words      int    `json:"words"`      // 字数
	Lines      int    `json:"lines"`      // 行数
	ReadTime   string `json:"readTime"`   // 预计阅读时间
}

type MarkdownPlugin struct {
	*plugins.BasePlugin
	app *application.App
}

func NewMarkdownPlugin() *MarkdownPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools",
		Description: "支持实时预览的 Markdown 编辑与导出工具",
		Icon:        "document-text",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:    []string{"markdown", "md", "编辑器", "文档", "editor", "document"},
		HasPage:     plugins.BoolPtr(true),
		ShowInMenu:  plugins.BoolPtr(true),
	}

	base := plugins.NewBasePlugin(metadata)
	return &MarkdownPlugin{
		BasePlugin: base,
	}
}

// 实现 Plugin 接口方法
func (p *MarkdownPlugin) Metadata() *plugins.PluginMetadata {
	return p.BasePlugin.Metadata()
}

func (p *MarkdownPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

func (p *MarkdownPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

func (p *MarkdownPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

func (p *MarkdownPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

func (p *MarkdownPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// ========== Markdown 操作方法 ==========

// GetStats 获取 Markdown 统计信息
func (p *MarkdownPlugin) GetStats(content string) MarkdownStats {
	stats := MarkdownStats{
		Characters: len(content),
		Lines:      strings.Count(content, "\n") + 1,
	}

	// 简单的字数统计（按空白分隔）
	words := strings.Fields(content)
	stats.Words = len(words)

	// 预计阅读时间（假设每分钟 200 字）
	minutes := stats.Words / 200
	if minutes < 1 {
		stats.ReadTime = "< 1 分钟"
	} else {
		stats.ReadTime = fmt.Sprintf("%d 分钟", minutes)
	}

	return stats
}
