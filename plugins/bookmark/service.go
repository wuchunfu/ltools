package bookmark

import (
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
	// Wails v3 提供了 application.Browser.OpenURL 方法
	return s.app.Browser.OpenURL(url)
}

// ExportHTML 导出为 HTML 格式（Phase 2 实现）
func (s *BookmarkService) ExportHTML(outputPath string) error {
	return nil
}

// ExportJSON 导出为 JSON 格式（Phase 2 实现）
func (s *BookmarkService) ExportJSON(outputPath string) error {
	return nil
}
