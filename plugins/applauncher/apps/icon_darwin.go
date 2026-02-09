//go:build darwin

package apps

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// darwinIconExtractor macOS 图标提取器
type darwinIconExtractor struct{}

// NewIconExtractor 创建 macOS 图标提取器
func NewIconExtractor() IconExtractor {
	return &darwinIconExtractor{}
}

// ExtractIcon 从 .app bundle 提取图标
// 使用多种策略查找图标文件
func (e *darwinIconExtractor) ExtractIcon(appPath string) (string, error) {
	if appPath == "" {
		return "", fmt.Errorf("empty app path")
	}

	// 检查是否是 .app 路径
	if !strings.HasSuffix(appPath, ".app") {
		return "", fmt.Errorf("not an .app bundle: %s", appPath)
	}

	resourcesDir := filepath.Join(appPath, "Contents", "Resources")

	// 策略 1: 查找常见的图标文件名
	commonIconNames := []string{
		"AppIcon.icns",
		"app.icns",
		"application.icns",
		"AppIcon-highres.icns",
		"AppIcon-512.icns",
		"icon.icns",
	}

	for _, iconName := range commonIconNames {
		icnsPath := filepath.Join(resourcesDir, iconName)
		if info, err := os.Stat(icnsPath); err == nil && !info.IsDir() {
			if iconData, err := ExtractIconFromFile(icnsPath); err == nil && iconData != "" {
				return iconData, nil
			}
		}
	}

	// 策略 2: 扫描 Resources 目录中的所有 .icns 文件
	if entries, err := os.ReadDir(resourcesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".icns") {
				icnsPath := filepath.Join(resourcesDir, entry.Name())
				if iconData, err := ExtractIconFromFile(icnsPath); err == nil && iconData != "" {
					return iconData, nil
				}
			}
		}
	}

	// 策略 3: 使用 macOS 的 sips 命令将图标转换为 PNG
	// 从 Info.plist 读取 CFBundleIconFile
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	if iconFile, err := getIconFileFromPlist(plistPath); err == nil && iconFile != "" {
		// 尝试直接读取（可能没有扩展名）
		icnsPath := filepath.Join(resourcesDir, iconFile)
		if !strings.HasSuffix(icnsPath, ".icns") {
			icnsPath += ".icns"
		}
		if iconData, err := ExtractIconFromFile(icnsPath); err == nil && iconData != "" {
			return iconData, nil
		}

		// 尝试使用 sips 转换
		if pngData, err := convertIconToPNG(icnsPath); err == nil && pngData != "" {
			return pngData, nil
		}
	}

	// 策略 4: 使用 macOS 的 iconutil 命令
	if iconData, err := extractIconUsingIconutil(appPath); err == nil && iconData != "" {
		return iconData, nil
	}

	// 没有找到图标文件，返回空使用默认图标
	return "", nil
}

// getIconFileFromPlist 从 Info.plist 读取 CFBundleIconFile
func getIconFileFromPlist(plistPath string) (string, error) {
	cmd := exec.Command("/usr/libexec/PlistBuddy", "-c", "Print :CFBundleIconFile", plistPath)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// convertIconToPNG 使用 sips 命令将图标转换为 PNG
func convertIconToPNG(icnsPath string) (string, error) {
	// 创建临时文件
	tmpFile := "/tmp/tmp_icon.png"
	defer os.Remove(tmpFile)

	// 使用 sips 转换
	cmd := exec.Command("sips", "-s", "format", "png", icnsPath, "--out", tmpFile)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// 读取转换后的 PNG
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", err
	}

	// 编码为 Base64
	base64Str := encodeBase64(data)
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// extractIconUsingIconutil 使用 iconutil 提取图标
func extractIconUsingIconutil(appPath string) (string, error) {
	// iconutil 需要 .iconset 目录，这里我们使用简化的方式
	// 直接尝试从 bundle 获取图标
	return "", fmt.Errorf("iconutil not implemented")
}

// ExtractIconFromAppPath 直接从 .app 路径提取图标
func ExtractIconFromAppPath(appPath string) (string, error) {
	extractor := NewIconExtractor()
	return extractor.ExtractIcon(appPath)
}

// ExtractIcon 全局函数提取图标
// 这个函数接收图标文件路径或 .app 路径
func ExtractIcon(iconPath string) (string, error) {
	if iconPath == "" {
		return "", fmt.Errorf("empty icon path")
	}

	// 如果是 .app 路径，查找并提取图标
	if strings.HasSuffix(iconPath, ".app") {
		return ExtractIconFromAppPath(iconPath)
	}

	// 如果是 .icns 文件，直接提取
	if strings.HasSuffix(iconPath, ".icns") {
		return ExtractIconFromFile(iconPath)
	}

	// 其他情况返回空
	return "", nil
}

// GetAppDefaultIcon 获取应用的默认图标（emoji）
// 当图标提取失败时使用
func GetAppDefaultIcon(appName string) string {
	appName = strings.ToLower(appName)

	// 浏览器
	if strings.Contains(appName, "safari") {
		return "🧭"
	}
	if strings.Contains(appName, "chrome") {
		return "🌐"
	}
	if strings.Contains(appName, "firefox") {
		return "🦊"
	}
	if strings.Contains(appName, "edge") {
		return "📱"
	}

	// 开发工具
	if strings.Contains(appName, "xcode") {
		return "🛠️"
	}
	if strings.Contains(appName, "visual") && strings.Contains(appName, "studio") {
		return "💻"
	}
	if strings.Contains(appName, "intellij") {
		return "💡"
	}
	if strings.Contains(appName, "atom") || strings.Contains(appName, "vscode") {
		return "💻"
	}

	// 通讯工具
	if strings.Contains(appName, "wechat") || strings.Contains(appName, "微信") {
		return "💬"
	}
	if strings.Contains(appName, "qq") {
		return "🐧"
	}
	if strings.Contains(appName, "slack") {
		return "💼"
	}
	if strings.Contains(appName, "telegram") {
		return "✈️"
	}
	if strings.Contains(appName, "discord") {
		return "🎮"
	}

	// 云服务
	if strings.Contains(appName, "drive") || strings.Contains(appName, "云") || strings.Contains(appName, "cloud") {
		return "☁️"
	}
	if strings.Contains(appName, "dropbox") {
		return "📦"
	}
	if strings.Contains(appName, "onedrive") {
		return "📥"
	}

	// 媒体
	if strings.Contains(appName, "music") || strings.Contains(appName, "音乐") {
		return "🎵"
	}
	if strings.Contains(appName, "photo") || strings.Contains(appName, "照片") {
		return "📷"
	}
	if strings.Contains(appName, "video") || strings.Contains(appName, "视频") {
		return "🎬"
	}

	// 设计工具
	if strings.Contains(appName, "figma") {
		return "🎨"
	}
	if strings.Contains(appName, "sketch") {
		return "🎨"
	}

	// 办公
	if strings.Contains(appName, "word") {
		return "📝"
	}
	if strings.Contains(appName, "excel") {
		return "📊"
	}
	if strings.Contains(appName, "powerpoint") || strings.Contains(appName, "keynote") {
		return "📽"
	}
	if strings.Contains(appName, "pages") {
		return "📄"
	}

	// 默认
	return "🚀"
}

// encodeBase64 编码数据为 Base64
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
