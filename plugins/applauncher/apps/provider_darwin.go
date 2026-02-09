//go:build darwin

package apps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// darwinProvider macOS 应用提供者
type darwinProvider struct{}

// NewProvider 创建 macOS 应用提供者
func NewProvider() (AppProvider, error) {
	return &darwinProvider{}, nil
}

// ListApps 列出所有已安装的应用程序
func (p *darwinProvider) ListApps() ([]*AppInfo, error) {
	// macOS 应用程序通常位于 /Applications 和 ~/Applications
	appPaths := []string{
		"/Applications",
		filepath.Join(os.Getenv("HOME"), "Applications"),
	}

	var apps []*AppInfo

	for _, appPath := range appPaths {
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			continue
		}

		// 遍历目录查找 .app 应用
		err := filepath.Walk(appPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 跳过无法访问的文件
			}

			// 跳过隐藏文件和目录
			if strings.HasPrefix(info.Name(), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 只处理 .app 目录
			if strings.HasSuffix(info.Name(), ".app") && info.IsDir() {
				appInfo, err := p.parseApp(path)
				if err == nil {
					apps = append(apps, appInfo)
				}
				// 跳过 .app 内部的内容
				return filepath.SkipDir
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to scan apps directory: %w", err)
		}
	}

	return apps, nil
}

// parseApp 解析 .app 应用信息
func (p *darwinProvider) parseApp(appPath string) (*AppInfo, error) {
	// .app 的基本结构：AppName.app/Contents/Info.plist

	infoPlistPath := filepath.Join(appPath, "Contents", "Info.plist")

	// 读取 Info.plist 获取应用信息
	bundleID, name, version := p.readInfoPlist(infoPlistPath)

	// 构建应用信息
	appName := filepath.Base(appPath)
	appName = strings.TrimSuffix(appName, ".app")

	if name == "" {
		name = appName
	}

	// 描述信息
	description := name
	if version != "" {
		description = fmt.Sprintf("%s (%s)", name, version)
	}

	// 可执行文件路径
	execPath := filepath.Join(appPath, "Contents", "MacOS", appName)

	// 存储 .app 路径用于图标提取（不是图标文件路径）
	iconPath := appPath

	return &AppInfo{
		ID:             bundleID,
		Name:           name,
		Description:    description,
		IconPath:       iconPath, // .app 路径，用于图标提取
		IconData:       "",       // 将在刷新时提取
		ExecutablePath: execPath,
		Type:           ResultTypeApp,
		BundleID:       bundleID,
	}, nil
}

// readInfoPlist 读取 Info.plist 文件
func (p *darwinProvider) readInfoPlist(plistPath string) (bundleID, name, version string) {
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return "", "", ""
	}

	// 使用 defaults 命令读取 plist
	// CFBundleIdentifier: 应用标识符
	// CFBundleName: 应用名称
	// CFBundleShortVersionString: 版本号

	bundleID = p.readPlistValue(plistPath, "CFBundleIdentifier")
	name = p.readPlistValue(plistPath, "CFBundleName")
	version = p.readPlistValue(plistPath, "CFBundleShortVersionString")

	if name == "" {
		name = p.readPlistValue(plistPath, "CFBundleDisplayName")
	}

	return bundleID, name, version
}

// readPlistValue 读取 plist 中的值
func (p *darwinProvider) readPlistValue(plistPath, key string) string {
	cmd := exec.Command("defaults", "read", plistPath, key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// LaunchApp 启动应用程序
func (p *darwinProvider) LaunchApp(appInfo *AppInfo) error {
	// 使用 open 命令启动应用
	// 优先使用 Bundle ID 启动
	if appInfo.BundleID != "" {
		cmd := exec.Command("open", "-b", appInfo.BundleID)
		return cmd.Start()
	}

	// 回退到使用 .app 路径
	if appInfo.ExecutablePath != "" {
		// 从可执行文件路径获取 .app 路径
		// ExecutablePath 是 like: /Applications/AppName.app/Contents/MacOS/appname
		// 我们需要获取 /Applications/AppName.app
		appPath := filepath.Dir(filepath.Dir(filepath.Dir(appInfo.ExecutablePath)))
		if !strings.HasSuffix(appPath, ".app") {
			// 如果路径不以 .app 结尾，尝试另一种方式
			appPath = appPath + ".app"
		}
		cmd := exec.Command("open", appPath)
		return cmd.Start()
	}

	return fmt.Errorf("no valid way to launch app: %s", appInfo.Name)
}

// RefreshCache 刷新缓存
func (p *darwinProvider) RefreshCache() error {
	// 在 macOS 上，我们不需要特殊处理
	// 缓存由 Cache 结构管理
	return nil
}
