package screenshot2

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// WindowManager 管理多显示器截图窗口
// 微信风格：每个显示器一个全屏窗口
type WindowManager struct {
	app                  *application.App
	plugin               *Screenshot2Plugin
	windows              map[int]*application.WebviewWindow // displayIndex -> window
	mainWindow           *application.WebviewWindow         // 主窗口引用
	isCapturing          bool
	mu                   sync.RWMutex
	sessionId            string
	wailsScreens         []*application.Screen // Wails 屏幕信息（正确的坐标系）
	mouseMonitorStopChan chan struct{}         // 停止鼠标监控的信号
	currentFocusDisplay  int                   // 当前聚焦的显示器索引
	frontendReadyChan    chan int              // 等待前端就绪的 channel（传递 displayIndex）
	expectedDisplays     int                   // 期望就绪的显示器数量
	readyDisplays        int                   // 已就绪的显示器数量
}

// NewWindowManager creates a new window manager
func NewWindowManager(plugin *Screenshot2Plugin, app *application.App) *WindowManager {
	return &WindowManager{
		app:               app,
		plugin:            plugin,
		windows:           make(map[int]*application.WebviewWindow),
		frontendReadyChan: make(chan int, 10), // 缓冲区足够大
	}
}

// SetMainWindow sets the main window reference
func (m *WindowManager) SetMainWindow(window *application.WebviewWindow) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mainWindow = window
}

// ServiceStartup is called when the application starts
func (m *WindowManager) ServiceStartup(app *application.App) error {
	log.Printf("[WindowManager] Service startup")
	return nil
}

// ServiceShutdown is called when the application shuts down
func (m *WindowManager) ServiceShutdown(app *application.App) error {
	m.CloseAllWindows()
	return nil
}

// OnFrontendReady 前端调用此方法通知已加载完成
func (m *WindowManager) OnFrontendReady(displayIndex int) {
	log.Printf("[WindowManager] Frontend ready for display %d", displayIndex)
	select {
	case m.frontendReadyChan <- displayIndex:
	default:
		log.Printf("[WindowManager] Frontend ready channel full, dropping notification for display %d", displayIndex)
	}
}

// waitForFrontendReady 等待所有前端就绪，带超时
func (m *WindowManager) waitForFrontendReady(expectedCount int, timeout time.Duration) bool {
	m.readyDisplays = 0
	m.expectedDisplays = expectedCount

	deadline := time.Now().Add(timeout)
	for m.readyDisplays < expectedCount {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			log.Printf("[WindowManager] Timeout waiting for frontend ready, got %d/%d", m.readyDisplays, expectedCount)
			return false
		}

		select {
		case displayIndex := <-m.frontendReadyChan:
			log.Printf("[WindowManager] Received ready from display %d (%d/%d)", displayIndex, m.readyDisplays+1, expectedCount)
			m.readyDisplays++
		case <-time.After(remaining):
			log.Printf("[WindowManager] Timeout waiting for frontend ready, got %d/%d", m.readyDisplays, expectedCount)
			return false
		}
	}

	log.Printf("[WindowManager] All frontends ready (%d/%d)", m.readyDisplays, expectedCount)
	return true
}

// StartCapture 开始截图流程
func (m *WindowManager) StartCapture() (string, error) {
	startTime := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[WindowManager] [TIMING] Starting capture...")

	// 先清理旧窗口（确保没有残留）
	m.closeAllWindowsLocked()
	log.Printf("[WindowManager] [TIMING] After closeAllWindows: %v", time.Since(startTime))

	// 清空 ready channel 中的旧消息
	drained := 0
	for {
		select {
		case <-m.frontendReadyChan:
			drained++
		default:
			goto startCapture
		}
	}
startCapture:
	if drained > 0 {
		log.Printf("[WindowManager] Drained %d old messages from ready channel before capture", drained)
	}

	// 隐藏主窗口（在截图之前）
	if m.mainWindow != nil {
		log.Printf("[WindowManager] Hiding main window before capture...")
		m.mainWindow.Hide()
	}
	log.Printf("[WindowManager] [TIMING] After hide main window: %v", time.Since(startTime))

	// 生成会话 ID
	m.sessionId = generateSessionId()
	log.Printf("[WindowManager] [TIMING] Before getting screen info: %v", time.Since(startTime))

	// 获取 Wails 屏幕信息（使用全局 App 实例）
	app := application.Get()
	log.Printf("[WindowManager] [TIMING] After application.Get(): %v", time.Since(startTime))
	if app == nil {
		log.Printf("[WindowManager] ERROR: application.Get() returned nil!")
	} else if app.Screen == nil {
		log.Printf("[WindowManager] ERROR: app.Screen (from Get()) is nil!")
		log.Printf("[WindowManager] [TIMING] app.Screen is nil: %v", time.Since(startTime))
	} else {
		screens := app.Screen.GetAll()
		log.Printf("[WindowManager] [TIMING] After Screen.GetAll(): %v", time.Since(startTime))
		log.Printf("[WindowManager] application.Get().Screen.GetAll() returned %d screens", len(screens))

		// 如果返回 0 个屏幕，尝试手动初始化 ScreenManager
		// 这是 Wails v3 macOS 的一个已知问题：ScreenManager 没有自动填充
		if len(screens) == 0 {
			log.Printf("[WindowManager] [TIMING] Before initScreensManually: %v", time.Since(startTime))
			log.Printf("[WindowManager] Attempting to manually initialize ScreenManager...")
			screens = m.initScreensManually(app)
			log.Printf("[WindowManager] [TIMING] After initScreensManually: %v", time.Since(startTime))
			log.Printf("[WindowManager] Manual initialization returned %d screens", len(screens))
		}

		if len(screens) > 0 {
			m.wailsScreens = screens
		}
	}
	log.Printf("[WindowManager] [TIMING] After first screen check: %v, wailsScreens=%d", time.Since(startTime), len(m.wailsScreens))

	// 也尝试使用 m.app
	if len(m.wailsScreens) == 0 && m.app != nil {
		log.Printf("[WindowManager] [TIMING] Trying m.app for screens: %v", time.Since(startTime))
		if m.app.Screen == nil {
			log.Printf("[WindowManager] ERROR: m.app.Screen is nil!")
		} else {
			screens := m.app.Screen.GetAll()
			log.Printf("[WindowManager] m.app.Screen.GetAll() returned %d screens", len(screens))

			// 如果返回 0 个屏幕，尝试手动初始化
			if len(screens) == 0 {
				log.Printf("[WindowManager] Attempting to manually initialize ScreenManager via m.app...")
				screens = m.initScreensManually(m.app)
				log.Printf("[WindowManager] Manual initialization returned %d screens", len(screens))
			}

			if len(screens) > 0 {
				m.wailsScreens = screens
			}
		}
	}

	// 如果无法获取 Wails 屏幕信息，使用回退方案
	log.Printf("[WindowManager] [TIMING] Before fallback check: %v, wailsScreens=%d", time.Since(startTime), len(m.wailsScreens))

	// 如果 Wails API 返回 0 个屏幕，回退到 screenshot 库
	if len(m.wailsScreens) == 0 {
		log.Printf("[WindowManager] Warning: Wails returned 0 screens, falling back to screenshot library")
		log.Printf("[WindowManager] [TIMING] Before GetDisplays: %v", time.Since(startTime))
		log.Printf("[WindowManager] [TIMING] Before GetDisplays: %v", time.Since(startTime))

		// 使用 screenshot 库获取显示器信息
		displays := m.plugin.GetDisplays()
		log.Printf("[WindowManager] [TIMING] After GetDisplays: %v", time.Since(startTime))

		// 获取虚拟桌面总高度（用于坐标转换）
		_, _, _, totalHeight := m.plugin.GetVirtualDesktopBounds()
		log.Printf("[WindowManager] Virtual desktop height for coordinate conversion: %d", totalHeight)

		// 捕获所有显示器
		captureResults, err := m.plugin.CaptureAllDisplaysSeparately()
		if err != nil {
			m.showMainWindow()
			return "", err
		}
		log.Printf("[WindowManager] [TIMING] After CaptureAllDisplays: %v", time.Since(startTime))

		// 进入 Kiosk 模式
		EnterKioskMode()
		log.Printf("[WindowManager] [TIMING] After EnterKioskMode: %v", time.Since(startTime))

		// 为每个显示器创建窗口
		// 需要将截图库坐标（左上角原点）转换为 Wails 坐标（macOS 左下角原点）
		for _, display := range displays {
			// 坐标转换：macOS 左下角原点
			// screenshot 库: Y=0 是屏幕顶部
			// macOS/Wails: Y=0 是屏幕底部
			// 转换公式: wailsY = totalHeight - displayY - displayHeight
			// 但由于虚拟桌面可能跨多个显示器，需要更精确的计算

			// 先使用原始坐标测试，看 Wails 的实际行为
			macosX := display.X
			macosY := display.Y

			log.Printf("[WindowManager] Display %d: screenshot coords (%d,%d) -> macOS coords (%d,%d)",
				display.Index, display.X, display.Y, macosX, macosY)

			// 创建转换后的 DisplayInfo
			convertedDisplay := DisplayInfo{
				Index:       display.Index,
				Width:       display.Width,
				Height:      display.Height,
				X:           macosX,
				Y:           macosY,
				Primary:     display.Primary,
				Name:        display.Name,
				ScaleFactor: display.ScaleFactor,
			}

			captureResult := captureResults[display.Index]
			if captureResult == nil {
				log.Printf("[WindowManager] Warning: no capture result for display %d", display.Index)
				continue
			}

			if err := m.createWindowForDisplay(convertedDisplay, captureResult); err != nil {
				log.Printf("[WindowManager] Failed to create window for display %d: %v", display.Index, err)
				continue
			}
		}
		log.Printf("[WindowManager] [TIMING] After create all windows: %v", time.Since(startTime))

		// 等待前端加载完成（使用事件机制）
		log.Printf("[WindowManager] Waiting for frontend to load...")
		waitStart := time.Now()
		m.waitForFrontendReady(len(m.windows), 2*time.Second)
		log.Printf("[WindowManager] [TIMING] After waitForFrontendReady (wait took %v): %v", time.Since(waitStart), time.Since(startTime))

		// 广播显示器信息给所有窗口
		displaysJSON, _ := json.Marshal(displays)
		m.broadcastEvent("displays-info", string(displaysJSON))

		// 广播会话开始
		m.broadcastEvent("session-start", m.sessionId)

		// 发送每个显示器的图片数据
		for displayIndex, result := range captureResults {
			m.sendImageData(displayIndex, result.Base64Data)
		}

		m.isCapturing = true
		m.emitEvent("started", "capture started")

		log.Printf("[WindowManager] [TIMING] Capture started with %d windows, TOTAL: %v", len(m.windows), time.Since(startTime))
		return m.sessionId, nil
	}

	// 使用 Wails 屏幕信息
	log.Printf("[WindowManager] [TIMING] Using Wails screens path: %v", time.Since(startTime))
	for i, screen := range m.wailsScreens {
		log.Printf("[WindowManager] Wails Screen %d: Name=%s, Position=(%d,%d), Size=%dx%d, Scale=%.2f, Primary=%v",
			i, screen.Name, screen.X, screen.Y, screen.Size.Width, screen.Size.Height, screen.ScaleFactor, screen.IsPrimary)
	}

	// 捕获所有显示器
	log.Printf("[WindowManager] [TIMING] Before CaptureAllDisplaysSeparately: %v", time.Since(startTime))
	captureResults, err := m.plugin.CaptureAllDisplaysSeparately()
	if err != nil {
		m.showMainWindow()
		return "", err
	}
	log.Printf("[WindowManager] [TIMING] After CaptureAllDisplaysSeparately: %v", time.Since(startTime))

	// 获取显示器信息（用于前端显示）
	displays := m.plugin.GetDisplays()

	// 进入 Kiosk 模式
	EnterKioskMode()
	log.Printf("[WindowManager] [TIMING] After EnterKioskMode: %v", time.Since(startTime))

	// 为每个显示器创建窗口（每个窗口显示自己的截图）
	for i, screen := range m.wailsScreens {
		// macOS Retina 显示器需要使用物理像素坐标定位窗口
		// 物理坐标 = 逻辑坐标 * ScaleFactor
		physicalX := int(float64(screen.X) * float64(screen.ScaleFactor))
		physicalY := int(float64(screen.Y) * float64(screen.ScaleFactor))

		log.Printf("[WindowManager] Screen %d: logical coords (%d,%d) -> physical coords (%d,%d), scale=%.1f",
			i, screen.X, screen.Y, physicalX, physicalY, screen.ScaleFactor)

		// 创建与 screenshot 库索引对应的 DisplayInfo
		// 窗口尺寸使用逻辑尺寸，位置使用物理坐标
		display := DisplayInfo{
			Index:       i,
			Width:       screen.Size.Width,
			Height:      screen.Size.Height,
			X:           physicalX,
			Y:           physicalY,
			Primary:     screen.IsPrimary,
			Name:        screen.Name,
			ScaleFactor: float64(screen.ScaleFactor),
		}

		captureResult := captureResults[i]
		if captureResult == nil {
			log.Printf("[WindowManager] Warning: no capture result for display %d", i)
			continue
		}

		if err := m.createWindowForDisplay(display, captureResult); err != nil {
			log.Printf("[WindowManager] Failed to create window for display %d: %v", i, err)
			continue
		}
	}
	log.Printf("[WindowManager] [TIMING] After create all windows: %v", time.Since(startTime))

	// 等待前端加载完成（使用事件机制）
	log.Printf("[WindowManager] Waiting for frontend to load...")
	waitStart := time.Now()
	m.waitForFrontendReady(len(m.windows), 2*time.Second)
	log.Printf("[WindowManager] [TIMING] After waitForFrontendReady (wait took %v): %v", time.Since(waitStart), time.Since(startTime))

	// 广播显示器信息给所有窗口
	displaysJSON, _ := json.Marshal(displays)
	m.broadcastEvent("displays-info", string(displaysJSON))

	// 广播会话开始
	m.broadcastEvent("session-start", m.sessionId)

	// 发送每个显示器的图片数据
	for displayIndex, result := range captureResults {
		m.sendImageData(displayIndex, result.Base64Data)
	}

	m.isCapturing = true
	m.emitEvent("started", "capture started")

	// 启动全局鼠标监控
	m.startGlobalMouseMonitor()

	log.Printf("[WindowManager] Capture started with %d windows", len(m.windows))
	return m.sessionId, nil
}

// createWindowForDisplay 为指定显示器创建窗口
func (m *WindowManager) createWindowForDisplay(display DisplayInfo, captureResult *CaptureResult) error {
	log.Printf("[WindowManager] Creating window for display %d at (%d, %d), size: %dx%d, scale: %.1f",
		display.Index, display.X, display.Y, display.Width, display.Height, display.ScaleFactor)

	// 每个窗口使用唯一的名称
	windowName := fmt.Sprintf("screenshot2-overlay-%d", display.Index)

	window := m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:           windowName,
		Title:          fmt.Sprintf("Screenshot - Display %d", display.Index),
		Width:          display.Width,
		Height:         display.Height,
		X:              display.X,
		Y:              display.Y,
		Frameless:      true,
		AlwaysOnTop:    true,
		DisableResize:  true, // 禁止缩放
		BackgroundType: application.BackgroundTypeTransparent,
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				Hide: true,
			},
			Backdrop:           application.MacBackdropTransparent,
			WindowLevel:        application.MacWindowLevelScreenSaver,
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces,
		},
		URL: fmt.Sprintf("/screenshot2-overlay?display=%d&scale=%.1f", display.Index, display.ScaleFactor),
	})

	if window == nil {
		return ErrCreateWindowFailed
	}

	m.windows[display.Index] = window

	// 显示窗口
	window.Show()

	// 显示后再次设置位置（确保在正确的显示器上）
	window.SetPosition(display.X, display.Y)

	// 验证最终位置
	actualX, actualY := window.Position()
	log.Printf("[WindowManager] Window for display %d - requested: (%d, %d), actual after SetPosition: (%d, %d)",
		display.Index, display.X, display.Y, actualX, actualY)

	// 不在这里调用 Focus()，让所有窗口都能接收鼠标事件
	// Focus() 会导致只有一个窗口获得焦点

	log.Printf("[WindowManager] Window created and shown for display %d", display.Index)
	return nil
}

// GetDisplayWindow 获取指定显示器的窗口
func (m *WindowManager) GetDisplayWindow(displayIndex int) *application.WebviewWindow {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.windows[displayIndex]
}

// initScreensManually 手动初始化 ScreenManager
// 这是 Wails v3 macOS 的一个已知问题的变通方案：
// ScreenManager 在 macOS 上没有自动填充屏幕数据
func (m *WindowManager) initScreensManually(app *application.App) []*application.Screen {
	// 使用截图库获取显示器信息
	displays := m.plugin.GetDisplays()
	if len(displays) == 0 {
		log.Printf("[WindowManager] initScreensManually: no displays from screenshot library")
		return nil
	}

	// 将 DisplayInfo 转换为 Wails Screen
	// 重要：PhysicalBounds 必须设置物理像素尺寸（逻辑尺寸 * ScaleFactor）
	// LayoutScreens 会根据 PhysicalBounds 和 ScaleFactor 计算 Bounds
	screens := make([]*application.Screen, len(displays))
	for i, display := range displays {
		// 物理像素尺寸 = 逻辑尺寸 * ScaleFactor
		physicalWidth := int(float64(display.Width) * display.ScaleFactor)
		physicalHeight := int(float64(display.Height) * display.ScaleFactor)

		screens[i] = &application.Screen{
			ID:               fmt.Sprintf("display-%d", display.Index),
			Name:             display.Name,
			X:                display.X,
			Y:                display.Y,
			Size:             application.Size{Width: display.Width, Height: display.Height},
			Bounds:           application.Rect{X: display.X, Y: display.Y, Width: display.Width, Height: display.Height},
			PhysicalBounds:   application.Rect{X: display.X, Y: display.Y, Width: physicalWidth, Height: physicalHeight},
			WorkArea:         application.Rect{X: display.X, Y: display.Y, Width: display.Width, Height: display.Height},
			PhysicalWorkArea: application.Rect{X: display.X, Y: display.Y, Width: physicalWidth, Height: physicalHeight},
			IsPrimary:        display.Primary,
			ScaleFactor:      float32(display.ScaleFactor),
		}
		log.Printf("[WindowManager] initScreensManually: Screen %d - Name=%s, Position=(%d,%d), LogicalSize=%dx%d, PhysicalSize=%dx%d, Scale=%.2f, Primary=%v",
			i, screens[i].Name, screens[i].X, screens[i].Y,
			screens[i].Bounds.Width, screens[i].Bounds.Height,
			screens[i].PhysicalBounds.Width, screens[i].PhysicalBounds.Height,
			screens[i].ScaleFactor, screens[i].IsPrimary)
	}

	// 调用 LayoutScreens 填充 ScreenManager
	err := app.Screen.LayoutScreens(screens)
	if err != nil {
		log.Printf("[WindowManager] initScreensManually: LayoutScreens failed: %v", err)
		return nil
	}

	log.Printf("[WindowManager] initScreensManually: Successfully initialized %d screens", len(screens))
	return app.Screen.GetAll()
}

// sendImageData 发送图片数据到指定显示器的窗口
func (m *WindowManager) sendImageData(displayIndex int, base64Data string) {
	if _, ok := m.windows[displayIndex]; ok {
		// 通过应用事件系统发送
		data := map[string]interface{}{
			"displayIndex": displayIndex,
			"imageData":    base64Data,
			"sessionId":    m.sessionId,
		}
		jsonData, _ := json.Marshal(data)
		log.Printf("[WindowManager] Sending image data for display %d, data length: %d", displayIndex, len(base64Data))
		m.app.Event.Emit("screenshot2:image-data", string(jsonData))
	} else {
		log.Printf("[WindowManager] Warning: no window for display %d", displayIndex)
	}
}

// broadcastEvent 向所有窗口广播事件
func (m *WindowManager) broadcastEvent(eventName, data string) {
	m.app.Event.Emit("screenshot2:"+eventName, data)
}

// CloseAllWindows 关闭所有截图窗口
func (m *WindowManager) CloseAllWindows() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeAllWindowsLocked()
}

// closeAllWindowsLocked 内部方法，关闭所有窗口（需要持有锁）
func (m *WindowManager) closeAllWindowsLocked() {
	// 停止鼠标监控
	m.stopGlobalMouseMonitor()

	// 广播会话结束
	m.broadcastEventLocked("session-end", "session ended")

	// 关闭所有窗口
	for i, window := range m.windows {
		if window != nil {
			window.Close()
			log.Printf("[WindowManager] Closed window %d", i)
		}
	}
	m.windows = make(map[int]*application.WebviewWindow)

	// 退出 Kiosk 模式
	ExitKioskMode()

	// 显示主窗口
	m.showMainWindowLocked()

	m.isCapturing = false
	m.plugin.ClearDisplayImages()
}

// broadcastEventLocked 向所有窗口广播事件（需要持有锁）
func (m *WindowManager) broadcastEventLocked(eventName, data string) {
	m.app.Event.Emit("screenshot2:"+eventName, data)
}

// showMainWindow 显示主窗口
func (m *WindowManager) showMainWindow() {
	if m.mainWindow != nil {
		m.mainWindow.Show()
	}
}

// showMainWindowLocked 显示主窗口（需要持有锁）
func (m *WindowManager) showMainWindowLocked() {
	if m.mainWindow != nil {
		m.mainWindow.Show()
	}
}

// GetMainWindow 获取主窗口
func (m *WindowManager) GetMainWindow() *application.WebviewWindow {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mainWindow
}

// IsCapturing 返回是否正在截图
func (m *WindowManager) IsCapturing() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isCapturing
}

// Helper methods
func (m *WindowManager) emitEvent(eventName, data string) {
	if m.app != nil {
		m.app.Event.Emit("screenshot2:"+eventName, data)
	}
}

// generateSessionId 生成会话 ID
func generateSessionId() string {
	return "screenshot2-" + time.Now().Format("20060102150405")
}

// startGlobalMouseMonitor 启动全局鼠标监控
// 当鼠标移动到不同显示器时，自动聚焦对应的窗口
func (m *WindowManager) startGlobalMouseMonitor() {
	m.mouseMonitorStopChan = make(chan struct{})
	m.currentFocusDisplay = -1

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond) // 每 50ms 检查一次鼠标位置
		defer ticker.Stop()

		for {
			select {
			case <-m.mouseMonitorStopChan:
				log.Printf("[WindowManager] Mouse monitor stopped")
				return
			case <-ticker.C:
				// 获取鼠标位置
				mouseX, mouseY := GetGlobalMousePosition()

				// 确定鼠标在哪个显示器上
				targetDisplay := -1
				for i, screen := range m.wailsScreens {
					// NSEvent mouseLocation 返回的是逻辑坐标，screen.X/Y 也是逻辑坐标
					// 直接比较逻辑坐标，不需要乘以 ScaleFactor
					right := float64(screen.X + screen.Size.Width)
					top := float64(screen.Y + screen.Size.Height)
					if mouseX >= float64(screen.X) && mouseX < right &&
						mouseY >= float64(screen.Y) && mouseY < top {
						targetDisplay = i
						break
					}
				}

				// 如果鼠标移动到了不同的显示器，聚焦对应的窗口
				if targetDisplay != -1 && targetDisplay != m.currentFocusDisplay {
					m.mu.RLock()
					window, exists := m.windows[targetDisplay]
					m.mu.RUnlock()

					if exists && window != nil {
						window.Focus()
						m.currentFocusDisplay = targetDisplay
						log.Printf("[WindowManager] Mouse moved to display %d, focusing window", targetDisplay)
					}
				}
			}
		}
	}()

	log.Printf("[WindowManager] Global mouse monitor started")
}

// stopGlobalMouseMonitor 停止全局鼠标监控
func (m *WindowManager) stopGlobalMouseMonitor() {
	if m.mouseMonitorStopChan != nil {
		close(m.mouseMonitorStopChan)
		m.mouseMonitorStopChan = nil
	}
}

// Error types
var ErrCreateWindowFailed = &WindowError{Message: "failed to create window"}

type WindowError struct {
	Message string
}

func (e *WindowError) Error() string {
	return e.Message
}
