//go:build windows

package plugins

import (
	"os/exec"
	"syscall"
)

// OpenPathWithDefaultApp 使用系统默认应用打开文件或目录 (Windows)
func OpenPathWithDefaultApp(path string) error {
	// Windows 使用 start 命令或直接 explorer
	cmd := exec.Command("explorer", path)
	return cmd.Start()
}

// OpenPathWithApp 使用指定应用打开文件或目录 (Windows)
func OpenPathWithApp(path, appName string) error {
	// Windows 使用 start 命令指定应用
	cmd := exec.Command("cmd", "/c", "start", "", appName, path)
	return cmd.Start()
}

// 特别处理：在 Windows 上，打开目录可能需要特殊处理
// Windows 上打开目录用 explorer，打开文件用 start
func openWindowsPath(path string, isDir bool) error {
	if isDir {
		cmd := exec.Command("explorer", path)
		return cmd.Start()
	}

	// 使用 ShellExecute 打开文件（关联应用）
	cmd := exec.Command("cmd", "/c", "start", "", "\"\"", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
