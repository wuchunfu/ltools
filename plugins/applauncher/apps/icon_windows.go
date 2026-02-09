//go:build windows

package apps

import (
	"fmt"
	"os/exec"
	"strings"
)

// windowsIconExtractor Windows 图标提取器
type windowsIconExtractor struct{}

// NewIconExtractor 创建 Windows 图标提取器
func NewIconExtractor() IconExtractor {
	return &windowsIconExtractor{}
}

// ExtractIcon 从 .ico 或 .exe 文件提取图标
func (e *windowsIconExtractor) ExtractIcon(iconPath string) (string, error) {
	if iconPath == "" {
		return "", fmt.Errorf("empty icon path")
	}

	// 检查文件是否存在
	cmd := exec.Command("if", "exist", iconPath, "echo", "found")
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) != "found" {
		return "", fmt.Errorf("icon file not found: %s", iconPath)
	}

	// 对于 Windows，我们暂时返回空
	// 真正的图标提取需要使用 Windows API 或第三方库
	// 这里可以留作后续扩展
	return "", fmt.Errorf("icon extraction not implemented for Windows yet")
}

// ExtractIcon 全局函数提取图标
func ExtractIcon(iconPath string) (string, error) {
	extractor := NewIconExtractor()
	return extractor.ExtractIcon(iconPath)
}

// GetAppDefaultIcon 获取应用的默认图标（emoji）
// Windows 平台的实现
func GetAppDefaultIcon(appName string) string {
	return "🚀" // Windows 平台暂时使用默认图标
}
