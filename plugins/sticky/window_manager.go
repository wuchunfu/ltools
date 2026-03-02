package sticky

import (
	"fmt"
	"log"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// WindowManager manages multiple sticky note windows
type WindowManager struct {
	app     *application.App
	windows map[string]*application.WebviewWindow // keyed by noteID
	mutex   sync.RWMutex
}

// NewWindowManager creates a new sticky note window manager
func NewWindowManager(app *application.App) *WindowManager {
	return &WindowManager{
		app:     app,
		windows: make(map[string]*application.WebviewWindow),
	}
}

// CreateWindow creates a new sticky note window
func (wm *WindowManager) CreateWindow(noteID string, x, y int, width, height int, color string) (*application.WebviewWindow, error) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// Check if window already exists
	if existingWindow, exists := wm.windows[noteID]; exists {
		log.Printf("[StickyWindow] Window for note %s already exists, focusing it", noteID)
		existingWindow.Show()
		existingWindow.Focus()
		return existingWindow, nil
	}

	// Set default size if not provided
	if width <= 0 {
		width = 280
	}
	if height <= 0 {
		height = 320
	}

	// Wails v3 uses logical pixels for window positioning
	// Wails handles DPI scaling automatically
	log.Printf("[StickyWindow] Creating window at logical position (%d, %d)", x, y)

	// Create the window
	// Note: InitialPosition must be set to WindowXY for X/Y coordinates to work
	// Otherwise, window defaults to centered position
	window := wm.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            fmt.Sprintf("sticky-window-%s", noteID),
		Title:           "Sticky Note",
		InitialPosition: application.WindowXY, // Required to use X/Y coordinates
		Width:           width,
		Height:          height,
		X:               x,
		Y:               y,
		Frameless:       true,
		AlwaysOnTop:     true,
		DisableResize:   false,
		MinWidth:        200,
		MinHeight:       200,
		BackgroundType:  application.BackgroundTypeTransparent,
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				Hide: true,
			},
			Backdrop:           application.MacBackdropTransparent,
			WindowLevel:        application.MacWindowLevelFloating,
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces,
		},
		URL: fmt.Sprintf("/sticky-window?id=%s&color=%s", noteID, color),
	})

	if window == nil {
		return nil, fmt.Errorf("failed to create sticky window for note %s", noteID)
	}

	// Store the window
	wm.windows[noteID] = window

	// Show the window
	window.Show()

	log.Printf("[StickyWindow] Created sticky window for note %s at (%d, %d), size: %dx%d", noteID, x, y, width, height)

	return window, nil
}

// CloseWindow closes a specific sticky note window
func (wm *WindowManager) CloseWindow(noteID string) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	window, exists := wm.windows[noteID]
	if !exists {
		return fmt.Errorf("window for note %s not found", noteID)
	}

	// Close the window
	if window != nil {
		window.Close()
	}

	// Remove from map
	delete(wm.windows, noteID)

	log.Printf("[StickyWindow] Closed sticky window for note %s", noteID)

	return nil
}

// CloseAll closes all sticky note windows
func (wm *WindowManager) CloseAll() {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	for noteID, window := range wm.windows {
		if window != nil {
			window.Close()
			log.Printf("[StickyWindow] Closed sticky window for note %s", noteID)
		}
	}

	// Clear the map
	wm.windows = make(map[string]*application.WebviewWindow)

	log.Printf("[StickyWindow] All sticky windows closed")
}

// GetWindowCount returns the number of active sticky windows
func (wm *WindowManager) GetWindowCount() int {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()

	return len(wm.windows)
}

// GetWindow gets a window by noteID
func (wm *WindowManager) GetWindow(noteID string) (*application.WebviewWindow, bool) {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()

	window, exists := wm.windows[noteID]
	return window, exists
}

// HasWindow checks if a window exists for the given noteID
func (wm *WindowManager) HasWindow(noteID string) bool {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()

	_, exists := wm.windows[noteID]
	return exists
}

// UpdateWindowPosition updates the position of a sticky window
func (wm *WindowManager) UpdateWindowPosition(noteID string, x, y int) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	window, exists := wm.windows[noteID]
	if !exists {
		return fmt.Errorf("window for note %s not found", noteID)
	}

	if window != nil {
		// Wails v3 uses logical pixels for window positioning
		window.SetPosition(x, y)
		log.Printf("[StickyWindow] Updated position for note %s: (%d, %d)",
			noteID, x, y)
	}

	return nil
}

// UpdateWindowSize updates the size of a sticky window
func (wm *WindowManager) UpdateWindowSize(noteID string, width, height int) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	window, exists := wm.windows[noteID]
	if !exists {
		return fmt.Errorf("window for note %s not found", noteID)
	}

	if window != nil {
		window.SetSize(width, height)
	}

	return nil
}

// FocusWindow brings a sticky window to the front
func (wm *WindowManager) FocusWindow(noteID string) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	window, exists := wm.windows[noteID]
	if !exists {
		return fmt.Errorf("window for note %s not found", noteID)
	}

	if window != nil {
		// 先显示窗口
		window.Show()
		// 然后聚焦
		window.Focus()
		log.Printf("[StickyWindow] Focused sticky window for note %s", noteID)
	}

	return nil
}

// GetAllWindowIDs returns all active note IDs
func (wm *WindowManager) GetAllWindowIDs() []string {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()

	ids := make([]string, 0, len(wm.windows))
	for id := range wm.windows {
		ids = append(ids, id)
	}

	return ids
}
