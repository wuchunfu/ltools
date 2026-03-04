//go:build windows

package update

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// restartWindows Windows 重启逻辑
func (s *Service) restartWindows(execPath string) error {
	log.Println("[UpdateService] Restarting on Windows...")

	// 创建一个延迟启动的命令
	// 使用 cmd /c 来执行：等待 1 秒后启动新进程
	cmd := exec.Command("cmd", "/c",
		fmt.Sprintf("timeout /t 1 /nobreak >nul & start \"\" \"%s\"", execPath))

	// Windows 特定：隐藏窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new instance: %w", err)
	}

	log.Println("[UpdateService] New instance started, exiting current...")

	// 退出当前应用
	os.Exit(0)

	return nil
}
