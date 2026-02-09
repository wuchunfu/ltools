// +build darwin

package screenshot

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// SetWindowOverlayOnMac 使用 macOS 特定的方法设置窗口覆盖
func SetWindowOverlayOnMac() error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	log.Printf("[Screenshot] Setting up macOS window overlay via AppleScript...")

	// 使用 System Events 操作进程，确保应用激活
	script := `
tell application "System Events"
    try
        set ltoolsProcess to first process whose name contains "ltools"
        tell ltoolsProcess
            set frontmost to true
            -- 尝试提升窗口
            try
                perform action "AXRaise" of window 1
            end try
        end tell
    end try
end tell
`

	log.Printf("[Screenshot] Executing SetWindowOverlayOnMac AppleScript...")

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	log.Printf("[Screenshot] SetWindowOverlayOnMac output: %s", string(output))
	if err != nil {
		log.Printf("[Screenshot] SetWindowOverlayOnMac error: %v", err)
	}

	log.Printf("[Screenshot] Window overlay attempt completed")
	return nil
}

// ForceWindowToFront 强制窗口到最前面并覆盖菜单栏
func ForceWindowToFront() error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	log.Printf("[Screenshot] Forcing window to front and covering menu bar...")

	// 只强制窗口到最前面，不调整窗口位置和大小
	// 让窗口大小保持与屏幕一致，通过 CGShieldingWindowLevel 来覆盖菜单栏
	script := `
tell application "System Events"
    try
        set ltoolsProcess to first process whose name contains "ltools"
        set frontmost of ltoolsProcess to true
        tell ltoolsProcess
            -- 强制提升窗口到最前面
            try
                perform action "AXRaise" of window 1
            end try
        end tell
    end try
end tell

delay 0.1

-- 确保应用是最前面的应用
tell application "ltools"
    activate
end try

delay 0.1

-- 再次确保窗口在最前面
tell application "System Events"
    try
        tell process "ltools"
            set frontmost to true
            perform action "AXRaise" of window 1
        end tell
    end try
end tell
`

	log.Printf("[Screenshot] Executing ForceWindowToFront AppleScript...")

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	log.Printf("[Screenshot] ForceWindowToFront output: %s", string(output))
	if err != nil {
		log.Printf("[Screenshot] ForceWindowToFront error: %v", err)
		// 继续执行
	}

	log.Printf("[Screenshot] Window force attempt completed")
	return nil
}

// SetWindowBounds 通过 AppleScript 设置窗口边界以覆盖菜单栏
func SetWindowToBounds(x, y, width, height int) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	log.Printf("[Screenshot] Setting window bounds via AppleScript: %d,%d %dx%d", x, y, width, height)

	// 使用 AppleScript 设置窗口位置和大小
	script := fmt.Sprintf(`
tell application "System Events"
    tell process "ltools"
        try
            set frontmost to true
            tell window 1
                set position to {%d, %d}
                set size to {%d, %d}
            end tell
        end try
    end tell
end tell
`, x, y, width, height)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[Screenshot] AppleScript bounds warning: %v, output: %s", err, string(output))
		return nil
	}

	log.Printf("[Screenshot] Window bounds set successfully")
	return nil
}
