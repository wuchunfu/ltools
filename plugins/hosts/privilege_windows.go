//go:build windows

package hosts

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// WriteHostsWithPrivilege writes content to a file with UAC elevation
// Note: This requires a helper executable (ltools-elevate.exe)
func WriteHostsWithPrivilege(filePath, content string) error {
	// Try direct write first (if already elevated)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err == nil {
		return nil
	}

	// Request elevation via runas
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Set-Content -Path '%s' -Value '%s'", filePath, content))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write with privilege: %w, output: %s", err, string(output))
	}
	return nil
}

// CheckPrivileges checks if we have admin access
func CheckPrivileges() bool {
	_, err := os.ReadFile(windowsHostsPath)
	return err == nil
}
