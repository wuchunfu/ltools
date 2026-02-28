package bookmark

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// BookmarkService 暴露给前端的服务接口
type BookmarkService struct {
	plugin *BookmarkPlugin
	app    *application.App
}

// NewBookmarkService 创建服务
func NewBookmarkService(app *application.App, plugin *BookmarkPlugin) *BookmarkService {
	return &BookmarkService{
		plugin: plugin,
		app:    app,
	}
}

// Search 搜索书签
func (s *BookmarkService) Search(query string) ([]SearchResult, error) {
	return s.plugin.Search(query), nil
}

// Sync 同步书签
func (s *BookmarkService) Sync() error {
	return s.plugin.Sync()
}

// GetCacheStatus 获取缓存状态
func (s *BookmarkService) GetCacheStatus() (map[string]interface{}, error) {
	return s.plugin.GetCacheStatus(), nil
}

// OpenURL 在浏览器中打开 URL
func (s *BookmarkService) OpenURL(url string) error {
	// 使用系统默认浏览器打开
	return s.app.Browser.OpenURL(url)
}

// ExportHTML 导出为 HTML 格式（弹出保存对话框）
func (s *BookmarkService) ExportHTML() (string, error) {
	if s.app == nil {
		return "", fmt.Errorf("application not initialized")
	}

	// 生成默认文件名
	defaultFilename := fmt.Sprintf("bookmarks_%s.html", time.Now().Format("2006-01-02"))

	// 弹出保存对话框
	path, err := s.app.Dialog.SaveFile().
		SetMessage("导出书签为 HTML").
		SetFilename(defaultFilename).
		AddFilter("HTML Files", "*.html").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()

	if err != nil {
		return "", fmt.Errorf("failed to show save dialog: %w", err)
	}

	if path == "" {
		// 用户取消
		return "", nil
	}

	// 执行导出
	exporter := NewExporter(s.plugin)
	if err := exporter.ExportHTML(path); err != nil {
		return "", fmt.Errorf("failed to export: %w", err)
	}

	return path, nil
}

// ExportJSON 导出为 JSON 格式（弹出保存对话框）
func (s *BookmarkService) ExportJSON() (string, error) {
	if s.app == nil {
		return "", fmt.Errorf("application not initialized")
	}

	// 生成默认文件名
	defaultFilename := fmt.Sprintf("bookmarks_%s.json", time.Now().Format("2006-01-02"))

	// 弹出保存对话框
	path, err := s.app.Dialog.SaveFile().
		SetMessage("导出书签为 JSON").
		SetFilename(defaultFilename).
		AddFilter("JSON Files", "*.json").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()

	if err != nil {
		return "", fmt.Errorf("failed to show save dialog: %w", err)
	}

	if path == "" {
		// 用户取消
		return "", nil
	}

	// 执行导出
	exporter := NewExporter(s.plugin)
	if err := exporter.ExportJSON(path); err != nil {
		return "", fmt.Errorf("failed to export: %w", err)
	}

	return path, nil
}
