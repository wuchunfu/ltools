package proxy

// MusicPlayerAdapter 适配器，让 ProxyManager 兼容 musicplayer 的 ProxyHandler 接口
type MusicPlayerAdapter struct {
	manager  *ProxyManager
	pluginName string
}

// NewMusicPlayerAdapter 创建适配器
func NewMusicPlayerAdapter(manager *ProxyManager) *MusicPlayerAdapter {
	return &MusicPlayerAdapter{
		manager:    manager,
		pluginName: "musicplayer",
	}
}

// RegisterAudioURL 注册音频 URL（实现 musicplayer.ProxyHandler 接口）
// 返回代理 URL
func (a *MusicPlayerAdapter) RegisterAudioURL(resourceID, remoteURL string) string {
	return a.manager.RegisterAudio(a.pluginName, resourceID, remoteURL)
}

// RegisterImageURL 注册图片 URL（实现 musicplayer.ProxyHandler 接口）
// 返回代理 URL
func (a *MusicPlayerAdapter) RegisterImageURL(resourceID, remoteURL string) string {
	return a.manager.RegisterImage(a.pluginName, resourceID, remoteURL)
}
