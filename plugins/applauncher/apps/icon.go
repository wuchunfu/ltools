//go:build !darwin && !windows && !linux

package apps

import "fmt"

// NewIconExtractor 创建平台特定的图标提取器
func NewIconExtractor() IconExtractor {
	return &defaultIconExtractor{}
}

// defaultIconExtractor 默认图标提取器（不支持的平台）
type defaultIconExtractor struct{}

// ExtractIcon 提取图标
func (e *defaultIconExtractor) ExtractIcon(iconPath string) (string, error) {
	return "", fmt.Errorf("icon extraction not supported on this platform")
}

// ExtractIcon 全局函数提取图标
func ExtractIcon(iconPath string) (string, error) {
	extractor := NewIconExtractor()
	return extractor.ExtractIcon(iconPath)
}

// GetAppDefaultIcon 获取应用的默认图标（emoji）
// 其他平台的实现
func GetAppDefaultIcon(appName string) string {
	return "🚀" // 其他平台使用默认图标
}
