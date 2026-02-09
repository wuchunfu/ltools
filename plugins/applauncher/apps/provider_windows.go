//go:build windows

package apps

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// windowsProvider Windows 应用提供者
type windowsProvider struct{}

// NewProvider 创建 Windows 应用提供者
func NewProvider() (AppProvider, error) {
	return &windowsProvider{}, nil
}

// ListApps 列出所有已安装的应用程序
// Windows 应用信息存储在注册表中
func (p *windowsProvider) ListApps() ([]*AppInfo, error) {
	// 注册表路径列表
	registryPaths := []string{
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	apps := make(map[string]*AppInfo)

	for _, regPath := range registryPaths {
		p.readRegistryApps(regPath, apps)
	}

	// 转换为切片
	result := make([]*AppInfo, 0, len(apps))
	for _, app := range apps {
		result = append(result, app)
	}

	return result, nil
}

// readRegistryApps 从注册表路径读取应用信息
func (p *windowsProvider) readRegistryApps(regPath string, apps map[string]*AppInfo) {
	// 使用 reg query 命令读取注册表
	// 格式: reg query "HKLM\path" /s
	cmd := exec.Command("reg", "query", regPath, "/s")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	// 解析输出
	lines := strings.Split(string(output), "\r\n")
	var currentApp *AppInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// 检查是否是注册表键（应用）
		if strings.Contains(line, "HKEY") && !strings.Contains(line, "REG_") {
			// 保存前一个应用
			if currentApp != nil && currentApp.Name != "" {
				apps[currentApp.ID] = currentApp
			}

			// 创建新应用
			currentApp = &AppInfo{
				ID:          line,
				Type:        ResultTypeApp,
				RegistryKey: line,
			}
		}

		// 解析值
		if currentApp != nil && strings.Contains(line, "REG_") {
			parts := strings.SplitN(line, "    ", 3)
			if len(parts) >= 3 {
				name := strings.TrimSpace(parts[0])
				typeStr := strings.TrimSpace(parts[1])
				value := strings.TrimSpace(parts[2])

				if typeStr == "REG_SZ" || typeStr == "REG_EXPAND_SZ" {
					// 提取有用的信息
					switch {
					case strings.HasSuffix(strings.ToUpper(name), "DISPLAYNAME"):
						currentApp.Name = value
					case strings.HasSuffix(strings.ToUpper(name), "DISPLAYVERSION"):
						if currentApp.Name != "" {
							currentApp.Description = fmt.Sprintf("%s %s", currentApp.Name, value)
						}
					case strings.HasSuffix(strings.ToUpper(name), "INSTALLLOCATION"):
						if value != "" {
							iconPath := p.findIconInDir(value)
							if iconPath != "" {
								currentApp.IconPath = iconPath
							}
						}
					case strings.HasSuffix(strings.ToUpper(name), "DISPLAYICON"):
						currentApp.IconPath = value
						currentApp.ExecutablePath = value
					}
				}
			}
		}
	}

	// 保存最后一个应用
	if currentApp != nil && currentApp.Name != "" {
		if currentApp.Description == "" {
			currentApp.Description = currentApp.Name
		}
		apps[currentApp.ID] = currentApp
	}
}

// findIconInDir 在目录中查找图标
func (p *windowsProvider) findIconInDir(dir string) string {
	if dir == "" {
		return ""
	}

	// 常见的图标文件名
	iconNames := []string{
		"app.ico",
		"icon.ico",
		"main.ico",
		"program.ico",
		"app.exe",
		"main.exe",
	}

	for _, iconName := range iconNames {
		iconPath := filepath.Join(dir, iconName)
		cmd := exec.Command("if", "exist", iconPath, "echo", "found")
		if output, _ := cmd.Output(); strings.TrimSpace(string(output)) == "found" {
			return iconPath
		}
	}

	return ""
}

// LaunchApp 启动应用程序
func (p *windowsProvider) LaunchApp(appInfo *AppInfo) error {
	var exePath string

	if appInfo.ExecutablePath != "" {
		exePath = appInfo.ExecutablePath
	} else if appInfo.IconPath != "" {
		exePath = appInfo.IconPath
	} else {
		return fmt.Errorf("no executable path for app: %s", appInfo.Name)
	}

	// 使用 start 命令启动应用
	cmd := exec.Command("cmd", "/c", "start", "", exePath)
	return cmd.Start()
}

// RefreshCache 刷新缓存
func (p *windowsProvider) RefreshCache() error {
	return nil
}
