//go:build !darwin

package screenshot2

import "log"

// StartGlobalMouseMonitor 开始全局鼠标监控（非 macOS 平台的空实现）
func StartGlobalMouseMonitor(callback func(x, y float64)) {
	log.Printf("[GlobalMouse] Global mouse monitor not implemented on this platform")
}

// StopGlobalMouseMonitor 停止全局鼠标监控
func StopGlobalMouseMonitor() {
	log.Printf("[GlobalMouse] Global mouse monitor stopped")
}

// GetGlobalMousePosition 获取全局鼠标位置
func GetGlobalMousePosition() (x, y float64) {
	return 0, 0
}
