// +build !darwin

package screenshot

// createAndShowEditorWindowNonMac 是非 macOS 平台的实现
func (s *ScreenshotWindowService) createAndShowEditorWindowNonMac(base64Img string) error {
	// 非 macOS 平台不需要 AppleScript 调用
	return nil
}

// forceWindowToFrontNonMac 是非 macOS 平台的实现
func (s *ScreenshotWindowService) forceWindowToFrontNonMac() error {
	// 非 macOS 平台不需要 AppleScript 调用
	return nil
}

// SetLastWindowToShieldingLevel 是非 macOS 平台的实现（空操作）
func (s *ScreenshotWindowService) SetLastWindowToShieldingLevel() error {
	// 非 macOS 平台不需要设置窗口级别
	return nil
}
