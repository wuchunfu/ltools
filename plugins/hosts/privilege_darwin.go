//go:build darwin

package hosts

import (
	"fmt"
	"os/exec"
	"strings"
)

// WriteHostsWithPrivilege writes content to a file using AppleScript with admin privileges
func WriteHostsWithPrivilege(filePath, content string) error {
	// Escape single quotes in content
	escapedContent := strings.ReplaceAll(content, "'", "'\\''")

	// Build shell command
	script := fmt.Sprintf("echo '%s' > %s", escapedContent, filePath)

	// Use osascript to run with administrator privileges
	cmd := exec.Command("osascript", "-e",
		fmt.Sprintf("do shell script %q with administrator privileges", script))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write with privilege: %w, output: %s", err, string(output))
	}
	return nil
}

// CheckPrivileges checks if we have write access to hosts file
func CheckPrivileges() bool {
	_, err := ReadHostsFile()
	return err == nil
}
