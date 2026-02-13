//go:build darwin

package hosts

import (
	"os"
)

const darwinHostsPath = "/etc/hosts"

// GetHostsPath returns the hosts file path on Darwin
func GetHostsPath() string {
	return darwinHostsPath
}

// ReadHostsFile reads the system hosts file
func ReadHostsFile() (string, error) {
	data, err := os.ReadFile(darwinHostsPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteHostsFile writes content to hosts file (requires privileges)
func WriteHostsFile(content string) error {
	return WriteHostsWithPrivilege(darwinHostsPath, content)
}
