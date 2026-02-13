//go:build linux

package plugins

import (
	"os/exec"
)

// OpenPathWithDefaultApp 使用系统默认应用打开文件或目录 (Linux)
func OpenPathWithDefaultApp(path string) error {
	// Linux 使用 xdg-open 命令
	cmd := exec.Command("xdg-open", path)
	return cmd.Start()
}

// OpenPathWithApp 使用指定应用打开文件或目录 (Linux)
func OpenPathWithApp(path, appName string) error {
	cmd := exec.Command(appName, path)
	return cmd.Start()
}
