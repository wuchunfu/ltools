//go:build linux

package apps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// linuxProvider Linux 应用提供者
type linuxProvider struct{}

// NewProvider 创建 Linux 应用提供者
func NewProvider() (AppProvider, error) {
	return &linuxProvider{}, nil
}

// ListApps 列出所有已安装的应用程序
// Linux 应用通过 .desktop 文件定义，位于：
// - /usr/share/applications (系统应用)
// - ~/.local/share/applications (用户应用)
// - /var/lib/snapd/desktop/applications (Snap 应用)
// - /var/lib/flatpak/exports/share/applications (Flatpak 应用)
func (p *linuxProvider) ListApps() ([]*AppInfo, error) {
	// .desktop 文件搜索路径
	desktopPaths := []string{
		"/usr/share/applications",
		filepath.Join(os.Getenv("HOME"), ".local/share/applications"),
		"/var/lib/snapd/desktop/applications",
		"/var/lib/flatpak/exports/share/applications",
	}

	var apps []*AppInfo

	for _, desktopPath := range desktopPaths {
		if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
			continue
		}

		// 遍历目录查找 .desktop 文件
		err := filepath.Walk(desktopPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 跳过无法访问的文件
			}

			if !info.IsDir() && strings.HasSuffix(info.Name(), ".desktop") {
				appInfo, err := p.parseDesktopFile(path)
				if err == nil {
					apps = append(apps, appInfo)
				}
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to scan desktop directory: %w", err)
		}
	}

	return apps, nil
}

// parseDesktopFile 解析 .desktop 文件
func (p *linuxProvider) parseDesktopFile(desktopPath string) (*AppInfo, error) {
	// 使用 grep 命令提取关键字段
	// 或者直接读取文件解析

	content, err := os.ReadFile(desktopPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	var name, description, execCmd, icon string
	isHidden := false
	inDesktopEntry := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查是否在 [Desktop Entry] 部分
		if strings.HasPrefix(line, "[Desktop Entry]") {
			inDesktopEntry = true
			continue
		}

		// 跳过其他部分
		if strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[Desktop Entry]") {
			break
		}

		// 解析字段
		if inDesktopEntry && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "Name":
					name = value
				case "Comment":
					description = value
				case "Exec":
					execCmd = value
				case "Icon":
					icon = value
				case "NoDisplay":
					if value == "true" {
						isHidden = true
					}
				}
			}
		}
	}

	// 跳过隐藏的应用
	if isHidden || name == "" {
		return nil, fmt.Errorf("skipping hidden or invalid desktop entry")
	}

	if description == "" {
		description = name
	}

	// 获取图标路径
	iconPath := p.resolveIconPath(icon)

	return &AppInfo{
		ID:             filepath.Base(desktopPath),
		Name:           name,
		Description:    description,
		IconPath:       iconPath,
		IconData:       "",
		ExecutablePath: execCmd,
		Type:           ResultTypeApp,
		DesktopFileID:  filepath.Base(desktopPath),
	}, nil
}

// resolveIconPath 解析图标路径
// Linux 图标可以是绝对路径或图标名称
func (p *linuxProvider) resolveIconPath(icon string) string {
	if icon == "" {
		return ""
	}

	// 如果是绝对路径
	if strings.HasPrefix(icon, "/") {
		if _, err := os.Stat(icon); err == nil {
			return icon
		}
	}

	// 搜索图标主题目录
	iconSizes := []string{"512", "256", "128", "64", "48", "32", "16"}
	iconFormats := []string{".png", ".svg", ".xpm"}

	// 图标搜索路径
	iconPaths := []string{
		"/usr/share/icons",
		"/usr/share/pixmaps",
		filepath.Join(os.Getenv("HOME"), ".local/share/icons"),
		"/usr/share/app-install/icons",
	}

	for _, iconBasePath := range iconPaths {
		// 首先尝试精确匹配
		for _, format := range iconFormats {
			iconPath := filepath.Join(iconBasePath, icon+format)
			if _, err := os.Stat(iconPath); err == nil {
				return iconPath
			}
		}

		// 然后尝试不同尺寸
		for _, size := range iconSizes {
			for _, format := range iconFormats {
				iconPath := filepath.Join(iconBasePath, size+"x"+size, icon+format)
				if _, err := os.Stat(iconPath); err == nil {
					return iconPath
				}

				// 也尝试 hicolor 主题
				iconPath = filepath.Join(iconBasePath, "hicolor", size+"x"+size, "apps", icon+format)
				if _, err := os.Stat(iconPath); err == nil {
					return iconPath
				}
			}
		}
	}

	return ""
}

// LaunchApp 启动应用程序
func (p *linuxProvider) LaunchApp(appInfo *AppInfo) error {
	// 使用 gtk-launch 或直接执行 Exec 命令
	if appInfo.DesktopFileID != "" {
		// 首先尝试 gtk-launch
		cmd := exec.Command("gtk-launch", appInfo.DesktopFileID)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// 回退到直接执行
	if appInfo.ExecutablePath != "" {
		// 清理 Exec 命令（移除 %f, %u 等占位符）
		execCmd := strings.Fields(appInfo.ExecutablePath)
		if len(execCmd) > 0 {
			cmd := exec.Command(execCmd[0], execCmd[1:]...)
			return cmd.Start()
		}
	}

	return fmt.Errorf("no valid way to launch app: %s", appInfo.Name)
}

// RefreshCache 刷新缓存
func (p *linuxProvider) RefreshCache() error {
	return nil
}
