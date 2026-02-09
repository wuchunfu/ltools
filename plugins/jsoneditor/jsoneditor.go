package jsoneditor

import (
	"encoding/json"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "jsoneditor.builtin"
	PluginName    = "JSON 编辑器"
	PluginVersion = "1.0.0"
)

// JSONStats JSON 统计信息
type JSONStats struct {
	Size     int    `json:"size"`     // 字符数
	Lines    int    `json:"lines"`    // 行数
	Type     string `json:"type"`     // 顶层类型
	IsValid  bool   `json:"isValid"`  // 是否有效
	ErrorMsg string `json:"errorMsg"` // 错误信息
}

type JSONEditorPlugin struct {
	*plugins.BasePlugin
	app *application.App
}

func NewJSONEditorPlugin() *JSONEditorPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools",
		Description: "JSON 格式化、验证和可视化编辑工具",
		Icon:        "code",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:    []string{"json", "编辑器", "格式化", "验证", "editor", "formatter"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &JSONEditorPlugin{
		BasePlugin: base,
	}
}

// 实现 Plugin 接口方法
func (p *JSONEditorPlugin) Metadata() *plugins.PluginMetadata {
	return p.BasePlugin.Metadata()
}

func (p *JSONEditorPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

func (p *JSONEditorPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

func (p *JSONEditorPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

func (p *JSONEditorPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

func (p *JSONEditorPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// ========== JSON 操作方法 ==========

// FormatJSON 格式化 JSON 字符串（2 空格缩进）
func (p *JSONEditorPlugin) FormatJSON(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	formatted, err := json.MarshalIndent(data, "", "  ")
	return string(formatted), err
}

// MinifyJSON 压缩 JSON 字符串
func (p *JSONEditorPlugin) MinifyJSON(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	minified, err := json.Marshal(data)
	return string(minified), err
}

// ValidateJSON 验证 JSON 是否有效
func (p *JSONEditorPlugin) ValidateJSON(jsonStr string) bool {
	var data interface{}
	return json.Unmarshal([]byte(jsonStr), &data) == nil
}

// GetJSONError 获取 JSON 解析错误的详细信息
func (p *JSONEditorPlugin) GetJSONError(jsonStr string) string {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err.Error()
	}
	return ""
}

// ParseJSONToMap 将 JSON 解析为 map（用于树形视图）
func (p *JSONEditorPlugin) ParseJSONToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// ParseJSONToArray 将 JSON 解析为 array（用于树形视图）
func (p *JSONEditorPlugin) ParseJSONToArray(jsonStr string) ([]interface{}, error) {
	var result []interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// GetJSONType 获取 JSON 值的类型
func (p *JSONEditorPlugin) GetJSONType(jsonStr string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "invalid"
	}

	switch data.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// GetJSONStats 获取 JSON 统计信息
func (p *JSONEditorPlugin) GetJSONStats(jsonStr string) JSONStats {
	stats := JSONStats{
		Size:  len(jsonStr),
		Lines: strings.Count(jsonStr, "\n") + 1,
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		stats.IsValid = false
		stats.ErrorMsg = err.Error()
		stats.Type = "invalid"
		return stats
	}

	stats.IsValid = true
	stats.Type = p.GetJSONType(jsonStr)
	return stats
}
