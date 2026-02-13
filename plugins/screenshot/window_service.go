package screenshot

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ScreenshotWindowService manages the screenshot capture and editor window
type ScreenshotWindowService struct {
	app             *application.App
	plugin          *ScreenshotPlugin
	editorWindow    *application.WebviewWindow
	mainWindow      *application.WebviewWindow
	isCapturing     bool
	mu              sync.RWMutex
	currentImageData []byte
	// Track whether main window was hidden for proper recovery
	mainWindowWasHidden bool
}

// NewScreenshotWindowService creates a new screenshot window service
func NewScreenshotWindowService(plugin *ScreenshotPlugin, app *application.App) *ScreenshotWindowService {
	return &ScreenshotWindowService{
		app:         app,
		plugin:      plugin,
		isCapturing: false,
	}
}

// SetMainWindow sets the main window reference
func (s *ScreenshotWindowService) SetMainWindow(window *application.WebviewWindow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mainWindow = window
}

// ServiceStartup is called when the application starts
func (s *ScreenshotWindowService) ServiceStartup(app *application.App) error {
	log.Printf("[ScreenshotWindowService] Starting up...")
	return nil
}

// ServiceShutdown is called when the application shuts down
func (s *ScreenshotWindowService) ServiceShutdown(app *application.App) error {
	log.Printf("[ScreenshotWindowService] Shutting down...")

	// Clean up editor window if it exists
	if s.editorWindow != nil {
		s.editorWindow.Close()
	}

	return nil
}

// ensureMainWindowShown ensures the main window is shown after an error
// This is a helper for error recovery to prevent the app from being "stuck" with
// no visible windows
func (s *ScreenshotWindowService) ensureMainWindowShown() {
	if s.mainWindow != nil && s.mainWindowWasHidden {
		log.Printf("[ScreenshotWindowService] Error recovery: ensuring main window is shown")
		s.mainWindow.Show()
		s.mainWindowWasHidden = false
	}
}

// hideMainWindowForCapture safely hides the main window for screenshot capture
// Returns true if the window was successfully hidden, false otherwise
func (s *ScreenshotWindowService) hideMainWindowForCapture() bool {
	if s.mainWindow != nil {
		log.Printf("[ScreenshotWindowService] Hiding main window for capture...")
		s.mainWindow.Hide()
		s.mainWindowWasHidden = true
		return true
	}
	return false
}

// cleanupAfterError cleans up state after any error during capture
func (s *ScreenshotWindowService) cleanupAfterError() {
	log.Printf("[ScreenshotWindowService] Cleaning up after error...")

	// Ensure main window is shown
	s.ensureMainWindowShown()

	// Reset capturing state
	s.isCapturing = false

	// Clear current image data to prevent stale data
	s.currentImageData = nil

	// Close editor window if it was partially created
	if s.editorWindow != nil {
		log.Printf("[ScreenshotWindowService] Closing editor window due to error...")
		s.editorWindow.Close()
		s.editorWindow = nil
	}
}

// StartCapture starts the screenshot capture process
func (s *ScreenshotWindowService) StartCapture() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[ScreenshotWindowService] Starting capture...")

	// Hide main window before starting capture
	s.hideMainWindowForCapture()

	// Ensure cleanup is called if this function returns an error
	// We use a named return value to allow deferred error handling
	var err error
	defer func() {
		if err != nil {
			log.Printf("[ScreenshotWindowService] StartCapture failed, cleaning up: %v", err)
			s.cleanupAfterError()
			s.emitError(fmt.Sprintf("Capture failed: %v", err))
		}
	}()

	// Capture the primary display
	img, captureErr := s.plugin.CaptureDisplay(0)
	if captureErr != nil {
		err = fmt.Errorf("failed to capture display: %w", captureErr)
		return "", err
	}

	// Store current image data
	var buf bytes.Buffer

	// 使用优化的 PNG 编码器配置 - BestSpeed 模式保持无损质量
	encoder := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	if encodeErr := encoder.Encode(&buf, img); encodeErr != nil {
		err = fmt.Errorf("failed to encode image: %w", encodeErr)
		return "", err
	}
	s.currentImageData = buf.Bytes()

	// Convert to base64 for event transmission (not URL parameter)
	base64Img := base64.StdEncoding.EncodeToString(s.currentImageData)
	base64Img = "data:image/png;base64," + base64Img

	// Create and show editor window immediately
	// We'll send image data right after to minimize black screen time
	if windowErr := s.createAndShowEditorWindow(base64Img); windowErr != nil {
		err = fmt.Errorf("failed to create editor: %w", windowErr)
		return "", err
	}

	// Success - mark as capturing (don't call cleanup in defer)
	s.emitEvent("captured", "image captured")
	s.isCapturing = true

	return base64Img, nil
}

// createAndShowEditorWindow creates or reuses the editor window, then sends image data
func (s *ScreenshotWindowService) createAndShowEditorWindow(base64Img string) error {
	log.Printf("[ScreenshotWindowService] Preparing editor window...")

	// Get screen information for window positioning using Wails v3 API
	// We need to properly handle:
	// 1. Screen width/height for window size
	// 2. Screen X/Y position for multi-monitor support
	// 3. macOS coordinate system (origin at bottom-left)
	var width, height int
	var xPos, yPos int
	var screenName string
	var scaleFactor float32

	// Note: Wails v3 Screen API may not be available in all contexts
	// We'll use the plugin's GetDisplays() which provides the same information
	displays := s.plugin.GetDisplays()

	if len(displays) > 0 {
		// Use the first display (usually the primary one)
		// For multi-monitor support, we could add logic to detect which display has the mouse
		display := displays[0]
		width = display.Width
		height = display.Height
		// Display bounds include X, Y position for multi-monitor setups
		xPos = display.X
		yPos = display.Y
		screenName = display.Name
		scaleFactor = 1.0 // Default scale factor

		log.Printf("[ScreenshotWindowService] Using plugin display info")
		log.Printf("[ScreenshotWindowService] Display: %s", screenName)
		log.Printf("[ScreenshotWindowService] Position: (%d, %d)", xPos, yPos)
		log.Printf("[ScreenshotWindowService] Resolution: %dx%d", width, height)

		// Log all available displays for debugging
		if len(displays) > 1 {
			log.Printf("[ScreenshotWindowService] All displays: %d", len(displays))
			for i, disp := range displays {
				log.Printf("[ScreenshotWindowService]   Display %d: %s at (%d, %d), size %dx%d",
					i, disp.Name, disp.X, disp.Y, disp.Width, disp.Height)
			}
		}
	} else {
		// Ultimate fallback
		width = 1920
		height = 1080
		xPos = 0
		yPos = 0
		screenName = "Default"
		scaleFactor = 1.0
		log.Printf("[ScreenshotWindowService] Using default resolution")
	}

	log.Printf("[ScreenshotWindowService] Screen: %s", screenName)
	log.Printf("[ScreenshotWindowService] Position: (%d, %d)", xPos, yPos)
	log.Printf("[ScreenshotWindowService] Resolution: %dx%d", width, height)
	log.Printf("[ScreenshotWindowService] Scale factor: %.2f", scaleFactor)

	// 创建全屏覆盖窗口，使用目标屏幕的尺寸和位置
	// 注意：在 macOS 上，Wails 会自动处理坐标转换（从左上角原点到底部原点）
	log.Printf("[ScreenshotWindowService] Creating fullscreen window...")

	// 使用固定窗口名称，通过 hide/show 复用窗口
	windowName := "screenshot-editor"

	// 检查窗口是否已存在
	if s.editorWindow != nil {
		// 窗口已存在，复用窗口但使用会话 ID 触发前端状态重置
		log.Printf("[ScreenshotWindowService] Reusing existing window, generating new session...")

		// 在复用窗口时也要进行错误检查
		if err := s.reuseExistingEditorWindow(width, height, xPos, yPos, base64Img); err != nil {
			log.Printf("[ScreenshotWindowService] Failed to reuse existing window: %v", err)
			// 如果复用失败，尝试关闭旧窗口并创建新窗口
			s.editorWindow.Close()
			s.editorWindow = nil
			// 继续执行下面的创建新窗口逻辑
		} else {
			s.emitEvent("started", "editor opened")
			log.Printf("[ScreenshotWindowService] Editor window reused and shown successfully")
			return nil
		}
	}

	// 创建新窗口
	log.Printf("[ScreenshotWindowService] Creating new fullscreen window...")

	return s.createNewEditorWindow(windowName, width, height, xPos, yPos, base64Img)
}

// reuseExistingEditorWindow reuses the existing editor window with a new session
func (s *ScreenshotWindowService) reuseExistingEditorWindow(width, height, xPos, yPos int, base64Img string) error {
	if s.editorWindow == nil {
		return fmt.Errorf("editor window is nil")
	}

	// 更新窗口属性
	s.editorWindow.SetSize(width, height)
	s.editorWindow.SetPosition(xPos, yPos)
	s.editorWindow.Show()
	s.editorWindow.Focus()

	// 等待窗口完全显示
	time.Sleep(100 * time.Millisecond)
	_ = s.forceWindowToFrontNonMac()

	// 等待前端加载
	time.Sleep(200 * time.Millisecond)

	// 生成新的会话 ID，用于触发前端状态重置
	sessionId := fmt.Sprintf("screenshot-%d", time.Now().UnixMilli())
	log.Printf("[ScreenshotWindowService] New session ID: %s", sessionId)

	// 发送会话开始事件（包含会话 ID）
	s.emitEvent("session-start", sessionId)

	// 发送图片数据（包含会话 ID）
	log.Printf("[ScreenshotWindowService] Sending image data with session...")
	s.emitEvent("image-data", base64Img+"|"+sessionId)
	log.Printf("[ScreenshotWindowService] Image data sent, size: %d bytes", len(base64Img))

	return nil
}

// createNewEditorWindow creates a new editor window
func (s *ScreenshotWindowService) createNewEditorWindow(windowName string, width, height, xPos, yPos int, base64Img string) error {
	newWindow := s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:        windowName,
		Title:       "Screenshot Editor",
		Width:       width,
		Height:      height,
		X:           xPos,
		Y:           yPos,
		Frameless:   true,
		AlwaysOnTop: true,
		BackgroundType: application.BackgroundTypeTransparent,
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				Hide: true,
			},
			Backdrop:           application.MacBackdropTransparent,
			// WindowLevel:        application.MacWindowLevelScreenSaver, // 使用最高窗口级别来覆盖菜单栏
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces, // 移除 FullScreenPrimary，避免与窗口级别冲突
		},
		URL: "/screenshot-editor",
	})

	// 检查窗口是否成功创建
	if newWindow == nil {
		return fmt.Errorf("failed to create editor window: returned nil")
	}

	s.editorWindow = newWindow
	log.Printf("[ScreenshotWindowService] Window created, showing...")

	// 显示窗口
	s.editorWindow.Show()
	log.Printf("[ScreenshotWindowService] Window shown")

	// 先设置焦点
	s.editorWindow.Focus()
	log.Printf("[ScreenshotWindowService] Window focused")

	// 等待窗口完全显示后，设置最后一个窗口（截图窗口）到 CGShieldingWindowLevel
	time.Sleep(300 * time.Millisecond)
	_ = s.forceWindowToFrontNonMac()

	// 等待前端加载
	time.Sleep(200 * time.Millisecond)

	// 生成新的会话 ID
	sessionId := fmt.Sprintf("screenshot-%d", time.Now().UnixMilli())
	log.Printf("[ScreenshotWindowService] New session ID: %s", sessionId)

	// 发送会话开始事件
	s.emitEvent("session-start", sessionId)

	// 等待前端处理会话开始事件
	time.Sleep(100 * time.Millisecond)

	// 发送图片数据（包含会话 ID）
	log.Printf("[ScreenshotWindowService] Sending image data with session...")
	s.emitEvent("image-data", base64Img+"|"+sessionId)
	log.Printf("[ScreenshotWindowService] Image data sent, size: %d bytes", len(base64Img))

	s.emitEvent("started", "editor opened")
	log.Printf("[ScreenshotWindowService] Editor window created and shown successfully")

	return nil
}

// CloseEditor hides the screenshot editor window (for reuse later)
func (s *ScreenshotWindowService) CloseEditor() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[ScreenshotWindowService] Hiding editor window...")

	// 发送会话结束事件，让前端清理状态
	s.emitEvent("session-end", "session ended")

	if s.editorWindow != nil {
		s.editorWindow.Hide()
		// 不设置为 nil，保留窗口引用以便复用
	}

	// Show main window again after hiding editor
	s.ensureMainWindowShown()

	s.isCapturing = false
	s.emitEvent("cancelled", "editor closed")

	return nil
}

// GetEditorWindow returns the editor window reference
func (s *ScreenshotWindowService) GetEditorWindow() *application.WebviewWindow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.editorWindow
}

// Toggle toggles the screenshot editor window
func (s *ScreenshotWindowService) Toggle() error {
	s.mu.RLock()
	isCapturing := s.isCapturing
	s.mu.RUnlock()

	if isCapturing {
		return s.CloseEditor()
	}
	_, err := s.StartCapture()
	return err
}

// SaveCurrentImage saves the current captured image to a file
func (s *ScreenshotWindowService) SaveCurrentImage(filename string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.currentImageData) == 0 {
		return "", fmt.Errorf("no image to save")
	}

	// Save using the plugin's storage
	return s.plugin.GetStorage().SaveToFile(s.currentImageData, filename)
}

// CopyCurrentToClipboard copies the current image to clipboard
func (s *ScreenshotWindowService) CopyCurrentToClipboard() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.currentImageData) == 0 {
		return fmt.Errorf("no image to copy")
	}

	return s.plugin.GetClipboard().SetImage(s.currentImageData)
}

// Helper methods to emit events
func (s *ScreenshotWindowService) emitEvent(eventName, data string) {
	if s.app != nil {
		s.app.Event.Emit("screenshot:"+eventName, data)
	}
}

func (s *ScreenshotWindowService) emitError(message string) {
	s.emitEvent("error", message)
}

