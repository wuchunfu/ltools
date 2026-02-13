//go:build darwin

package plugins

import (
	"os/exec"
)

// OpenPathWithDefaultApp 使用系统默认应用打开文件或目录 (macOS)
func OpenPathWithDefaultApp(path string) error {
	// macOS 使用 open 命令
	cmd := exec.Command("open", path)
	return cmd.Start()
}

// OpenPathWithApp 使用指定应用打开文件或目录 (macOS)
func OpenPathWithApp(path, appName string) error {
	// macOS 使用 open -a 指定应用
	cmd := exec.Command("open", "-a", appName, path)
	return cmd.Start()
}
