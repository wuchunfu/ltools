package applauncher

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/plugins/applauncher/apps"
)

// AppLauncherService 暴露给前端的接口
type AppLauncherService struct {
	app    *application.App
	plugin *AppLauncherPlugin
}

// NewAppLauncherService 创建服务
func NewAppLauncherService(app *application.App, plugin *AppLauncherPlugin) *AppLauncherService {
	return &AppLauncherService{
		app:    app,
		plugin: plugin,
	}
}

// Search 搜索应用
func (s *AppLauncherService) Search(query string) ([]*apps.AppInfo, error) {
	s.app.Logger.Info("[AppLauncherService] Searching for:", query)
	return s.plugin.Search(query)
}

// LaunchApp 启动应用
func (s *AppLauncherService) LaunchApp(appID string) error {
	s.app.Logger.Info("[AppLauncherService] Launching app:", appID)
	return s.plugin.LaunchApp(appID)
}

// RefreshCache 刷新应用列表
func (s *AppLauncherService) RefreshCache() error {
	s.app.Logger.Info("[AppLauncherService] Refreshing cache...")
	return s.plugin.RefreshCache()
}

// GetCacheStatus 获取缓存状态
func (s *AppLauncherService) GetCacheStatus() map[string]interface{} {
	return s.plugin.GetCacheStatus()
}
