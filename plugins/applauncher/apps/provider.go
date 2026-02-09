//go:build !darwin && !windows && !linux

package apps

import (
	"fmt"
)

// NewProvider 创建平台特定的应用提供者
// 默认实现返回错误（不支持的平台）
func NewProvider() (AppProvider, error) {
	return nil, fmt.Errorf("app launcher not supported on this platform")
}

// NewIconExtractor 创建平台特定的图标提取器
func NewIconExtractor() IconExtractor {
	return &defaultIconExtractor{}
}

// ExtractIcon 提取图标（默认实现）
func ExtractIcon(iconPath string) (string, error) {
	extractor := NewIconExtractor()
	return extractor.ExtractIcon(iconPath)
}
