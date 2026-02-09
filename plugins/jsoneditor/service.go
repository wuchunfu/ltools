package jsoneditor

import (
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// JSONEditorService exposes JSON Editor functionality to the frontend
type JSONEditorService struct {
	plugin *JSONEditorPlugin
	app    *application.App
}

func NewJSONEditorService(plugin *JSONEditorPlugin, app *application.App) *JSONEditorService {
	return &JSONEditorService{
		plugin: plugin,
		app:    app,
	}
}

// SetApp sets the application instance for file operations
func (s *JSONEditorService) SetApp(app *application.App) {
	s.app = app
}

// FormatJSON 格式化 JSON
func (s *JSONEditorService) FormatJSON(jsonStr string) (string, error) {
	return s.plugin.FormatJSON(jsonStr)
}

// MinifyJSON 压缩 JSON
func (s *JSONEditorService) MinifyJSON(jsonStr string) (string, error) {
	return s.plugin.MinifyJSON(jsonStr)
}

// ValidateJSON 验证 JSON
func (s *JSONEditorService) ValidateJSON(jsonStr string) bool {
	return s.plugin.ValidateJSON(jsonStr)
}

// GetJSONError 获取错误信息
func (s *JSONEditorService) GetJSONError(jsonStr string) string {
	return s.plugin.GetJSONError(jsonStr)
}

// ParseJSONToMap 解析为 Map
func (s *JSONEditorService) ParseJSONToMap(jsonStr string) (map[string]interface{}, error) {
	return s.plugin.ParseJSONToMap(jsonStr)
}

// ParseJSONToArray 解析为 Array
func (s *JSONEditorService) ParseJSONToArray(jsonStr string) ([]interface{}, error) {
	return s.plugin.ParseJSONToArray(jsonStr)
}

// GetJSONType 获取类型
func (s *JSONEditorService) GetJSONType(jsonStr string) string {
	return s.plugin.GetJSONType(jsonStr)
}

// GetJSONStats 获取统计信息
func (s *JSONEditorService) GetJSONStats(jsonStr string) JSONStats {
	return s.plugin.GetJSONStats(jsonStr)
}

// ReadFile 读取文件内容
func (s *JSONEditorService) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile 写入文件内容
func (s *JSONEditorService) WriteFile(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

// ImportFileResult 导入文件结果
type ImportFileResult struct {
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
}

// ImportFile 打开文件选择对话框并读取文件内容
// 使用后端对话框 API 以获得更好的文件过滤器支持
func (s *JSONEditorService) ImportFile() (*ImportFileResult, error) {
	if s.app == nil {
		return nil, os.ErrInvalid
	}

	path, err := s.app.Dialog.OpenFile().
		SetTitle("选择 JSON 文件").
		AddFilter("JSON Files", "*.json").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()

	if err != nil || path == "" {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &ImportFileResult{
		FilePath: path,
		Content:  string(data),
	}, nil
}

// ExportFile 打开保存文件对话框并写入内容
// 使用后端对话框 API 以获得更好的文件过滤器支持
func (s *JSONEditorService) ExportFile(content string, defaultFilename string) (filePath string, err error) {
	if s.app == nil {
		return "", os.ErrInvalid
	}

	if defaultFilename == "" {
		defaultFilename = "data.json"
	}

	path, err := s.app.Dialog.SaveFile().
		SetMessage("保存 JSON 文件").
		SetFilename(defaultFilename).
		AddFilter("JSON Files", "*.json").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()

	if err != nil || path == "" {
		return "", err
	}

	err = os.WriteFile(path, []byte(content), 0644)
	return path, err
}
