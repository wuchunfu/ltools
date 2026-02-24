//go:build !darwin

package screenshot2

import "log"

// EnterKioskMode 进入 Kiosk 模式（非 macOS 平台的空实现）
func EnterKioskMode() {
	log.Printf("[Screenshot2] Kiosk mode not supported on this platform")
}

// ExitKioskMode 退出 Kiosk 模式（非 macOS 平台的空实现）
func ExitKioskMode() {
	log.Printf("[Screenshot2] Kiosk mode not supported on this platform")
}
