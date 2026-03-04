package proxy

import (
	"net/http"
)

// Service 代理服务接口（供插件使用）
type Service interface {
	// RegisterAudio 注册音频资源
	RegisterAudio(pluginName, resourceID, remoteURL string) string

	// RegisterImage 注册图片资源
	RegisterImage(pluginName, resourceID, remoteURL string) string

	// RegisterVideo 注册视频资源
	RegisterVideo(pluginName, resourceID, remoteURL string) string

	// RegisterFile 注册文件资源
	RegisterFile(pluginName, resourceID, remoteURL string) string

	// UnregisterResource 注销资源
	UnregisterResource(resourceType ResourceType, pluginName, resourceID string)

	// UnregisterPlugin 注销插件的所有资源
	UnregisterPlugin(pluginName string)

	// GetStats 获取统计信息
	GetStats() map[string]interface{}

	// ClearCache 清空缓存
	ClearCache()

	// GetHandler 获取 HTTP Handler
	GetHandler() http.Handler
}

// 确保 ProxyManager 实现了 Service 接口
var _ Service = (*ProxyManager)(nil)

// GetHandler 获取 HTTP Handler
func (pm *ProxyManager) GetHandler() http.Handler {
	return pm
}
