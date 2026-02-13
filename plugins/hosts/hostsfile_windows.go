//go:build windows

package hosts

import (
	"os"
)

const windowsHostsPath = `C:\Windows\System32\drivers\etc\hosts`

// GetHostsPath returns the hosts file path on Windows
func GetHostsPath() string {
	return windowsHostsPath
}

// ReadHostsFile reads the system hosts file
func ReadHostsFile() (string, error) {
	data, err := os.ReadFile(windowsHostsPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteHostsFile writes content to hosts file (requires admin privileges)
func WriteHostsFile(content string) error {
	return WriteHostsWithPrivilege(windowsHostsPath, content)
}
