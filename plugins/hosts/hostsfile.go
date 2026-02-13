//go:build !darwin && !windows

package hosts

import (
	"errors"
	"os"
)

// GetHostsPath returns the hosts file path for current platform
func GetHostsPath() string {
	return "/etc/hosts" // Default for Linux
}

// ReadHostsFile reads the system hosts file
func ReadHostsFile() (string, error) {
	data, err := os.ReadFile(GetHostsPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteHostsFile writes content to hosts file
func WriteHostsFile(content string) error {
	return WriteHostsWithPrivilege(GetHostsPath(), content)
}

// WriteHostsWithPrivilege writes content with privilege (Linux stub)
func WriteHostsWithPrivilege(filePath, content string) error {
	// Try direct write first
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err == nil {
		return nil
	}
	return errors.New("write failed: insufficient privileges (try running with sudo)")
}

// CheckPrivileges checks if we have write access
func CheckPrivileges() bool {
	_, err := ReadHostsFile()
	return err == nil
}
