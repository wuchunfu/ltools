//go:build !darwin

package plugins

import (
	"fmt"
	"runtime"
)

// CheckAccessibilityPermissions checks if the app has accessibility permissions
// For non-macOS platforms, this always returns true
func CheckAccessibilityPermissions() (bool, error) {
	return true, nil
}

// OpenAccessibilitySettings opens accessibility settings
// For non-macOS platforms, this returns an error
func OpenAccessibilitySettings() error {
	return fmt.Errorf("accessibility settings are only available on macOS, current platform: %s", runtime.GOOS)
}
