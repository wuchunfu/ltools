package plugins

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ShortcutService exposes keyboard shortcut management to the frontend
type ShortcutService struct {
	app              *application.App
	manager          *ShortcutManager
	platform         string
	registeredKeys   map[string]string // keyCombo -> pluginID
	registrationLock sync.Mutex
	// Global hotkey manager using gohook for system-wide shortcuts
	globalHotkeyManager *GlobalHotkeyManager
	useGlobalHotkeys    bool // Whether to use gohook (true) or Wails KeyBinding (false)
	// Main window reference for focusing
	mainWindow      *application.WebviewWindow
}

// NewShortcutService creates a new shortcut service
func NewShortcutService(app *application.App, dataDir string) (*ShortcutService, error) {
	manager, err := NewShortcutManager(dataDir)
	if err != nil {
		return nil, err
	}

	// Detect platform
	platform := runtime.GOOS

	// Create global hotkey manager
	globalHotkeyManager := NewGlobalHotkeyManager()

	return &ShortcutService{
		app:                 app,
		manager:             manager,
		platform:            platform,
		registeredKeys:      make(map[string]string),
		globalHotkeyManager: globalHotkeyManager,
		useGlobalHotkeys:    true, // Enable global hotkeys (testing cmd+6)
	}, nil
}

// ServiceStartup is called when the application starts
func (s *ShortcutService) ServiceStartup(app *application.App) error {
	app.Logger.Info("[ShortcutService] Starting up, registering all keyboard shortcuts...")

	if s.useGlobalHotkeys {
		// Check for accessibility permissions on macOS
		if s.platform == "darwin" {
			app.Logger.Info("[ShortcutService] Checking accessibility permissions on macOS...")
			hasPermission, err := CheckAccessibilityPermissions()
			app.Logger.Info(fmt.Sprintf("[ShortcutService] Accessibility check result: hasPermission=%v, err=%v", hasPermission, err))
			if err != nil {
				app.Logger.Error(fmt.Sprintf("[ShortcutService] Failed to check accessibility permissions: %v", err))
			}

			if !hasPermission {
				app.Logger.Warn("[ShortcutService] Accessibility permissions not granted, global hotkeys will not work in background")
				// Emit event to frontend to show permission request
				// Convert to JSON string for the event
				permissionData := fmt.Sprintf(`{"title":"需要辅助功能权限","message":"要使用全局快捷键（应用在后台时也能响应），需要授予辅助功能权限。","platform":"%s","hasPermission":false}`, s.platform)
				app.Event.Emit("shortcut:permission-required", permissionData)
				// Fall back to Wails KeyBinding
				s.useGlobalHotkeys = false
			} else {
				app.Logger.Info("[ShortcutService] Accessibility permissions GRANTED!")
			}
		}

		app.Logger.Info(fmt.Sprintf("[ShortcutService] useGlobalHotkeys = %v (after permission check)", s.useGlobalHotkeys))

		if s.useGlobalHotkeys {
			// Set up callback for global hotkey triggers
			s.globalHotkeyManager.SetCallback(func(pluginID string) {
				log.Printf("[ShortcutService] *** GLOBAL HOTKEY TRIGGERED: %s ***", pluginID)
				app.Logger.Info(fmt.Sprintf("[ShortcutService] *** GLOBAL HOTKEY TRIGGERED: %s ***", pluginID))

				// Bring window to front, except for search window shortcut
				// Search window should be shown independently without affecting main window visibility
				if s.mainWindow != nil && pluginID != "search.window.builtin" {
					app.Logger.Info("[ShortcutService] Bringing window to front")
					s.mainWindow.Show()
				}

				// Emit event to frontend
				app.Event.Emit("shortcut:triggered", pluginID)
			})

			// Register all enabled shortcuts as global hotkeys
			shortcuts := s.manager.GetAll()
			for keyCombo, shortcut := range shortcuts {
				if shortcut.Enabled {
					if err := s.globalHotkeyManager.Register(keyCombo, shortcut.PluginID); err != nil {
						app.Logger.Error(fmt.Sprintf("[ShortcutService] Failed to register global hotkey %s: %v", keyCombo, err))
					}
				}
			}

			// Start the global hotkey manager
			app.Logger.Info("[ShortcutService] About to call globalHotkeyManager.Start()...")
			if err := s.globalHotkeyManager.Start(); err != nil {
				app.Logger.Error(fmt.Sprintf("[ShortcutService] Failed to start global hotkey manager: %v", err))
				// Fall back to Wails KeyBinding
				s.useGlobalHotkeys = false
			} else {
				app.Logger.Info(fmt.Sprintf("[ShortcutService] Started global hotkey manager with %d hotkeys", len(s.globalHotkeyManager.GetRegisteredHotkeys())))
			}
		}
	}

	// Fall back to Wails KeyBinding if global hotkeys are not enabled or failed
	if !s.useGlobalHotkeys {
		shortcuts := s.manager.GetAll()
		for keyCombo, shortcut := range shortcuts {
			if shortcut.Enabled {
				if err := s.registerShortcut(keyCombo, shortcut.PluginID); err != nil {
					app.Logger.Error(fmt.Sprintf("[ShortcutService] Failed to register shortcut %s: %v", keyCombo, err))
				}
			}
		}
		app.Logger.Info(fmt.Sprintf("[ShortcutService] Registered %d shortcuts (Wails KeyBinding)", len(s.registeredKeys)))
	}

	return nil
}

// ServiceShutdown is called when the application shuts down
func (s *ShortcutService) ServiceShutdown(app *application.App) error {
	app.Logger.Info("[ShortcutService] Shutting down...")

	s.registrationLock.Lock()
	s.registeredKeys = make(map[string]string)
	s.registrationLock.Unlock()

	// Stop global hotkey manager if it's running
	if s.globalHotkeyManager.IsStarted() {
		s.globalHotkeyManager.Stop()
	}

	return nil
}

// ShortcutInfo represents shortcut information for the frontend
type ShortcutInfo struct {
	PluginID  string `json:"pluginId"`
	KeyCombo  string `json:"keyCombo"`
	DisplayTxt string `json:"displayText"` // Formatted for display
	Enabled   bool   `json:"enabled"`
}

// SetMainWindow sets the main window reference for focusing
func (s *ShortcutService) SetMainWindow(window *application.WebviewWindow) {
	s.registrationLock.Lock()
	defer s.registrationLock.Unlock()
	s.mainWindow = window
	log.Printf("[ShortcutService] Main window reference set")
}
func (s *ShortcutService) SetShortcut(keyCombo, pluginID string) error {
	// Check for conflicts first
	if conflict, conflictingPluginID := s.manager.CheckConflict(keyCombo, pluginID); conflict {
		return fmt.Errorf("shortcut %s is already bound to plugin %s", keyCombo, conflictingPluginID)
	}

	// Set the shortcut
	if err := s.manager.Set(keyCombo, pluginID, true); err != nil {
		return err
	}

	// Register the shortcut
	if s.useGlobalHotkeys && s.globalHotkeyManager.IsStarted() {
		if err := s.globalHotkeyManager.Register(keyCombo, pluginID); err != nil {
			// Rollback the save if registration failed
			s.manager.Remove(keyCombo)
			return err
		}
	} else {
		if err := s.registerShortcut(keyCombo, pluginID); err != nil {
			// Rollback the save if registration failed
			s.manager.Remove(keyCombo)
			return err
		}
	}

	return nil
}

// RemoveShortcut removes a shortcut
func (s *ShortcutService) RemoveShortcut(keyCombo string) error {
	// Unregister the shortcut
	if s.useGlobalHotkeys && s.globalHotkeyManager.IsStarted() {
		s.globalHotkeyManager.Unregister(keyCombo)
	} else {
		s.unregisterShortcut(keyCombo)
	}

	// Remove from manager
	return s.manager.Remove(keyCombo)
}

// GetAllShortcuts returns all shortcuts
func (s *ShortcutService) GetAllShortcuts() []ShortcutInfo {
	shortcuts := s.manager.GetAll()
	result := make([]ShortcutInfo, 0, len(shortcuts))

	for _, shortcut := range shortcuts {
		result = append(result, ShortcutInfo{
			PluginID:  shortcut.PluginID,
			KeyCombo:  shortcut.KeyCombo,
			DisplayTxt: FormatKeyCombo(shortcut.KeyCombo, s.platform),
			Enabled:   shortcut.Enabled,
		})
	}

	return result
}

// GetPluginShortcut returns the shortcut for a specific plugin
func (s *ShortcutService) GetPluginShortcut(pluginID string) *ShortcutInfo {
	shortcuts := s.manager.GetByPluginID(pluginID)
	if len(shortcuts) == 0 {
		return nil
	}

	shortcut := shortcuts[0]
	return &ShortcutInfo{
		PluginID:  shortcut.PluginID,
		KeyCombo:  shortcut.KeyCombo,
		DisplayTxt: FormatKeyCombo(shortcut.KeyCombo, s.platform),
		Enabled:   shortcut.Enabled,
	}
}

// EnableShortcut enables a shortcut
func (s *ShortcutService) EnableShortcut(keyCombo string) error {
	shortcut, ok := s.manager.Get(keyCombo)
	if !ok {
		return ErrShortcutNotFound
	}

	// Update manager
	if err := s.manager.Set(keyCombo, shortcut.PluginID, true); err != nil {
		return err
	}

	// Register the key binding
	if s.useGlobalHotkeys && s.globalHotkeyManager.IsStarted() {
		return s.globalHotkeyManager.Register(keyCombo, shortcut.PluginID)
	}
	return s.registerShortcut(keyCombo, shortcut.PluginID)
}

// DisableShortcut disables a shortcut
func (s *ShortcutService) DisableShortcut(keyCombo string) error {
	shortcut, ok := s.manager.Get(keyCombo)
	if !ok {
		return ErrShortcutNotFound
	}

	// Unregister the key binding
	if s.useGlobalHotkeys && s.globalHotkeyManager.IsStarted() {
		s.globalHotkeyManager.Unregister(keyCombo)
	} else {
		s.unregisterShortcut(keyCombo)
	}

	// Update manager
	return s.manager.Set(keyCombo, shortcut.PluginID, false)
}

// CheckConflict checks if a key combo conflicts with existing shortcuts
func (s *ShortcutService) CheckConflict(keyCombo, pluginID string) (bool, string) {
	return s.manager.CheckConflict(keyCombo, pluginID)
}

// FormatShortcut formats a key combo for display
func (s *ShortcutService) FormatShortcut(keyCombo string) string {
	return FormatKeyCombo(keyCombo, s.platform)
}

// CheckPermissions checks if the app has the required permissions for global hotkeys
func (s *ShortcutService) CheckPermissions() (bool, string) {
	if s.platform != "darwin" {
		return true, "not_required"
	}

	hasPermission, err := CheckAccessibilityPermissions()
	if err != nil {
		return false, "error"
	}
	return hasPermission, "accessibility"
}

// RequestPermissions opens the system settings to grant permissions
func (s *ShortcutService) RequestPermissions() error {
	if s.platform != "darwin" {
		return fmt.Errorf("permission request only available on macOS")
	}
	return OpenAccessibilitySettings()
}

// IsUsingGlobalHotkeys returns whether global hotkeys are enabled
func (s *ShortcutService) IsUsingGlobalHotkeys() bool {
	return s.useGlobalHotkeys
}

// registerShortcut registers a keyboard shortcut with Wails
func (s *ShortcutService) registerShortcut(keyCombo, pluginID string) error {
	s.registrationLock.Lock()
	defer s.registrationLock.Unlock()

	// Normalize key combo for Wails
	normalizedKeyCombo := normalizeKeyCombo(keyCombo)

	log.Printf("[ShortcutService] Registering shortcut: %s -> %s", normalizedKeyCombo, pluginID)

	// Wails v3 KeyBinding expects format like "Ctrl+S" or "Cmd+Shift+Z"
	wailsKeyCombo := convertToWailsFormat(normalizedKeyCombo)

	// Create the key binding handler
	handler := func(window application.Window) {
		log.Printf("*** Shortcut TRIGGERED: %s -> %s ***", wailsKeyCombo, pluginID)
		s.app.Logger.Info(fmt.Sprintf("*** Shortcut TRIGGERED: %s -> %s ***", wailsKeyCombo, pluginID))

		// Emit event to frontend
		s.app.Event.Emit("shortcut:triggered", pluginID)
	}

	// Register the key binding with Wails v3
	s.app.KeyBinding.Add(wailsKeyCombo, handler)

	s.registeredKeys[normalizedKeyCombo] = pluginID
	s.registeredKeys[wailsKeyCombo] = pluginID // Store both formats
	return nil
}

// convertToWailsFormat converts a normalized key combo to Wails v3 format
func convertToWailsFormat(keyCombo string) string {
	parts := splitKeyCombo(keyCombo)
	if len(parts) <= 1 {
		return capitalizeKey(keyCombo)
	}

	// Extract modifiers and main key
	var modifiers []string
	var mainKey string

	for _, part := range parts {
		switch toLower(part) {
		case "ctrl", "control":
			modifiers = append(modifiers, "Ctrl")
		case "cmd", "command", "meta":
			modifiers = append(modifiers, "Cmd")
		case "shift":
			modifiers = append(modifiers, "Shift")
		case "alt", "option":
			modifiers = append(modifiers, "Alt")
		default:
			// This is the main key - capitalize it
			mainKey = capitalizeKey(part)
		}
	}

	// Build the final key combo with all modifiers
	if len(modifiers) > 0 && mainKey != "" {
		result := ""
		for i, mod := range modifiers {
			if i > 0 {
				result += "+"
			}
			result += mod
		}
		result += "+" + mainKey
		return result
	}

	// Fallback to original format
	return capitalizeKey(keyCombo)
}

// unregisterShortcut unregisters a keyboard shortcut
func (s *ShortcutService) unregisterShortcut(keyCombo string) {
	s.registrationLock.Lock()
	defer s.registrationLock.Unlock()

	// Normalize key combo
	normalizedKeyCombo := normalizeKeyCombo(keyCombo)
	wailsKeyCombo := convertToWailsFormat(normalizedKeyCombo)

	log.Printf("[ShortcutService] Unregistering shortcut: %s", keyCombo)

	// Try to remove both formats
	s.app.KeyBinding.Remove(wailsKeyCombo)
	s.app.KeyBinding.Remove(normalizedKeyCombo)

	delete(s.registeredKeys, normalizedKeyCombo)
	delete(s.registeredKeys, wailsKeyCombo)
}

// Helper functions for key combo parsing

func splitKeyCombo(keyCombo string) []string {
	var result []string
	var current string
	for _, ch := range keyCombo {
		if ch == '+' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func toLower(s string) string {
	if len(s) == 0 {
		return s
	}
	result := ""
	for _, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			result += string(ch + 32)
		} else {
			result += string(ch)
		}
	}
	return result
}

func capitalizeKey(s string) string {
	if len(s) == 0 {
		return s
	}
	// Capitalize first letter
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
