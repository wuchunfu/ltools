package apps

import "time"

// ResultType 搜索结果类型
type ResultType string

const (
	ResultTypePlugin ResultType = "plugin"
	ResultTypeApp    ResultType = "app"
)

// AppInfo 应用信息
type AppInfo struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	IconPath       string     `json:"iconPath"`
	IconData       string     `json:"iconData"`     // Base64 编码图标
	ExecutablePath string     `json:"executablePath"`
	Type           ResultType `json:"type"`
	// 平台特定字段
	DesktopFileID  string     `json:"desktopFileID,omitempty"`
	BundleID       string     `json:"bundleId,omitempty"`
	RegistryKey    string     `json:"registryKey,omitempty"`
}

// CacheInfo 缓存信息
type CacheInfo struct {
	LastUpdated time.Time `json:"lastUpdated"`
	AppCount    int       `json:"appCount"`
	Version     string    `json:"version"`
}

// AppProvider 应用提供者接口
type AppProvider interface {
	ListApps() ([]*AppInfo, error)
	LaunchApp(appInfo *AppInfo) error
	RefreshCache() error
}

// IconExtractor 图标提取器接口
type IconExtractor interface {
	ExtractIcon(iconPath string) (string, error)
}
