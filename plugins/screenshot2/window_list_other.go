//go:build !darwin

package screenshot2

import (
	"fmt"
	"log"
)

// WindowInfo 窗口信息
type WindowInfo struct {
	OwnerPID   int    `json:"ownerPid"`
	OwnerName  string `json:"ownerName"`
	WindowID   int    `json:"windowId"`
	WindowName string `json:"windowName"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

// GetWindowAtPoint 获取指定坐标下的窗口信息（非 macOS 平台）
func GetWindowAtPoint(x, y int, excludeSelf bool) (*WindowInfo, error) {
	log.Printf("[Screenshot2] GetWindowAtPoint(%d, %d) - not supported", x, y)
	return nil, fmt.Errorf("window detection not supported on this platform")
}

// GetWindowList 获取所有可见窗口列表（非 macOS 平台）
func GetWindowList(excludeSelf bool) ([]WindowInfo, error) {
	log.Printf("[Screenshot2] GetWindowList() - not supported")
	return nil, fmt.Errorf("window detection not supported on this platform")
}
