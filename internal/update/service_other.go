//go:build !windows

package update

import (
	"fmt"
)

// restartWindows Windows 重启逻辑（非 Windows 平台的存根）
func (s *Service) restartWindows(execPath string) error {
	return fmt.Errorf("restart not supported on this platform")
}
