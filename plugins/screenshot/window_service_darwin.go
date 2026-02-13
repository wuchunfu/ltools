// +build darwin

package screenshot

import (
	"log"
)

// forceWindowToFrontNonMac 是 macOS 平台的实现，调用 AppleScript 强制置前
func (s *ScreenshotWindowService) forceWindowToFrontNonMac() error {
	log.Printf("[ScreenshotWindowService] Calling AppleScript ForceWindowToFront...")
	if err := ForceWindowToFront(); err != nil {
		log.Printf("[ScreenshotWindowService] AppleScript ForceWindowToFront warning: %v", err)
	}
	// 设置最后一个窗口（截图窗口）到 CGShieldingWindowLevel
	// log.Printf("[ScreenshotWindowService] Setting last window to CGShieldingWindowLevel...")
	// if err := SetLastWindowToShieldingLevel(); err != nil {
	// 	log.Printf("[ScreenshotWindowService] SetLastWindowToShieldingLevel warning: %v", err)
	// }
	return nil
}
