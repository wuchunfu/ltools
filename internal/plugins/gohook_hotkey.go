package plugins

import (
	"fmt"
	 maps "maps"
	"os"
	"sync"

	hook "github.com/robotn/gohook"
)

// debugLog logs to file for debugging
var debugLog *os.File

func init() {
	var err error
	debugLog, err = os.OpenFile("/tmp/gohook_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open debug log: %v\n", err)
	}
}

func debugLogPrintf(format string, v ...interface{}) {
	if debugLog != nil {
		debugLog.WriteString(fmt.Sprintf(format+"\n", v...))
		debugLog.Sync()
	}
	// Also output to stderr
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}

// GlobalHotkeyManager using robotn/gohook library
// This provides true system-wide global hotkeys that work when app is not focused
type GlobalHotkeyManager struct {
	// evChan is the channel from hook.Start()
	evChan chan hook.Event

	// Registered hotkeys: normalizedKeyCombo -> pluginID
	registeredHotkeys map[string]string

	// Mapping from hook key codes to our key combos
	// Format: "ctrl+shift+d" -> pluginID
	hotkeyMap map[string]string

	// Key state tracking for modifier combinations
	keyState map[string]bool

	// Event emitter callback
	onHotkeyTriggered func(pluginID string)

	// Control
	started  bool
	stopChan chan struct{}
	mu       sync.RWMutex
}

// NewGlobalHotkeyManager creates a new global hotkey manager
func NewGlobalHotkeyManager() *GlobalHotkeyManager {
	return &GlobalHotkeyManager{
		registeredHotkeys: make(map[string]string),
		hotkeyMap:         make(map[string]string),
		keyState:          make(map[string]bool),
		stopChan:          make(chan struct{}),
	}
}

// Register registers a global hotkey
func (m *GlobalHotkeyManager) Register(keyCombo, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	normalizedKeyCombo := normalizeKeyCombo(keyCombo)

	// Check for conflicts
	if existing, exists := m.registeredHotkeys[normalizedKeyCombo]; exists && existing != pluginID {
		return fmt.Errorf("hotkey %s already registered to plugin %s", normalizedKeyCombo, existing)
	}

	m.registeredHotkeys[normalizedKeyCombo] = pluginID
	m.hotkeyMap[normalizedKeyCombo] = pluginID

	debugLogPrintf("[GlobalHotkeyManager] Registered hotkey: %s -> %s", normalizedKeyCombo, pluginID)
	return nil
}

// Unregister removes a global hotkey
func (m *GlobalHotkeyManager) Unregister(keyCombo string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	normalizedKeyCombo := normalizeKeyCombo(keyCombo)

	if _, exists := m.registeredHotkeys[normalizedKeyCombo]; !exists {
		return fmt.Errorf("hotkey %s not registered", normalizedKeyCombo)
	}

	delete(m.registeredHotkeys, normalizedKeyCombo)
	delete(m.hotkeyMap, normalizedKeyCombo)

	debugLogPrintf("[GlobalHotkeyManager] Unregistered hotkey: %s", normalizedKeyCombo)
	return nil
}

// SetCallback sets the callback function for hotkey triggers
func (m *GlobalHotkeyManager) SetCallback(callback func(pluginID string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onHotkeyTriggered = callback
}

// Start begins listening for global hotkeys
// This should be called after all hotkeys are registered
func (m *GlobalHotkeyManager) Start() error {
	debugLogPrintf("[GlobalHotkeyManager] Start() method called")

	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		debugLogPrintf("[GlobalHotkeyManager] Already started, returning")
		return fmt.Errorf("hotkey manager already started")
	}
	m.started = true
	m.mu.Unlock()

	debugLogPrintf("[GlobalHotkeyManager] About to call hook.Start()...")

	// Start the event loop - hook.Start() returns a channel of events
	m.evChan = hook.Start()

	if m.evChan == nil {
		debugLogPrintf("[GlobalHotkeyManager] ERROR: hook.Start() returned nil channel!")
		return fmt.Errorf("hook.Start() returned nil channel - likely no accessibility permission")
	}

	debugLogPrintf("[GlobalHotkeyManager] hook.Start() returned, starting processEvents goroutine")

	// Process events in a separate goroutine
	go m.processEvents()

	debugLogPrintf("[GlobalHotkeyManager] Global hotkey listener started")
	return nil
}

// Stop stops the hotkey listener
func (m *GlobalHotkeyManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return
	}

	debugLogPrintf("[GlobalHotkeyManager] Stopping global hotkey listener...")

	close(m.stopChan)

	// End the hook
	if m.evChan != nil {
		hook.End()
	}

	m.started = false
	debugLogPrintf("[GlobalHotkeyManager] Global hotkey listener stopped")
}

// processEvents processes events from the hook channel
func (m *GlobalHotkeyManager) processEvents() {
	for {
		select {
		case <-m.stopChan:
			return
		case event := <-m.evChan:
			switch event.Kind {
			case hook.KeyDown:
				m.handleKeyDown(event)
			case hook.KeyUp:
				m.handleKeyUp(event)
			}
		}
	}
}

// handleKeyDown handles key down events from channel
func (m *GlobalHotkeyManager) handleKeyDown(event hook.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Debug: log all key events with all available info
	keyName := m.getKeyName(event)
	debugLogPrintf("[GlobalHotkeyManager] KeyDown: keycode=%d rawcode=%d keychar='%c'(%d) keyName='%s' [V2]",
		event.Keycode, event.Rawcode, event.Keychar, event.Keychar, keyName)

	// Track key state - only use getKeyName for consistency
	if keyName != "" {
		m.keyState[keyName] = true
	}

	// Log current key state for debugging
	if keyName != "" {
		debugLogPrintf("[GlobalHotkeyManager] Current key state: %+v", m.keyState)
	}

	// Check if any hotkey combo matches
	for keyCombo, pluginID := range m.registeredHotkeys {
		if m.matchesHotkey(keyCombo) {
			debugLogPrintf("[GlobalHotkeyManager] *** HOTKEY TRIGGERED: %s -> %s ***", keyCombo, pluginID)

			// Call the callback in a goroutine to avoid blocking
			if m.onHotkeyTriggered != nil {
				go m.onHotkeyTriggered(pluginID)
			}
			return
		}
	}
}

// handleKeyUp handles key up events from channel
func (m *GlobalHotkeyManager) handleKeyUp(event hook.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update key state - only use getKeyName for consistency
	keyName := m.getKeyName(event)

	if keyName != "" {
		m.keyState[keyName] = false
	}

	debugLogPrintf("[GlobalHotkeyManager] KeyUp: keyName='%s'", keyName)
}

// matchesHotkey checks if the current key state matches a hotkey combo
func (m *GlobalHotkeyManager) matchesHotkey(keyCombo string) bool {
	parts := splitKeyCombo(keyCombo)

	// Debug: log the matching attempt
	debugLogPrintf("[GlobalHotkeyManager] matchesHotkey: checking '%s', parts=%+v, keyState=%+v", keyCombo, parts, m.keyState)

	// Check if all keys in the combo are currently pressed
	for _, part := range parts {
		normalized := toLower(part)
		if !m.keyState[normalized] {
			debugLogPrintf("[GlobalHotkeyManager] matchesHotkey: '%s' failed - key '%s' not pressed (state=%v)", keyCombo, normalized, m.keyState[normalized])
			return false
		}
	}

	debugLogPrintf("[GlobalHotkeyManager] matchesHotkey: '%s' MATCHED!", keyCombo)
	return true
}

// getKeyName converts a hook event to a normalized key name
//
// Priority order:
// 1. Rawcode mapping (most reliable for gohook)
// 2. Keycode fallback (for edge cases)
// 3. Keychar fallback (last resort, often returns 0xFFFF)
func (m *GlobalHotkeyManager) getKeyName(event hook.Event) string {
	// Primary: use rawcode mapping (gohook's native keycode system)
	rawKeyName := m.mapRawcodeToKey(event.Rawcode)
	if rawKeyName != "" {
		return rawKeyName
	}

	// Secondary: fallback to keycode for edge cases
	// Some key combinations may report different keycode values
	switch event.Keycode {
	case 0x26, 0x37, 0x3A, 0x3D: // Cmd key edge cases
		return "cmd"
	}

	// Tertiary: fallback to keychar (unreliable, often 0xFFFF)
	if event.Keychar != 0 && event.Keychar != 0xFFFF {
		return toLower(string(event.Keychar))
	}

	return ""
}

// mapRawcodeToKey maps gohook library rawcodes to key names
//
// NOTE: gohook uses its own rawcode system, NOT macOS CGKeyCode values.
// The rawcodes are derived from libuiohook and represent physical key positions.
// For number keys, gohook uses values that correspond to numpad positions:
//   - 1-5 use rawcodes 0x53-0x57 (83-87 decimal)
//   - 6-9 use rawcodes 0x31-0x34 (49-52 decimal)
//   - 0 uses rawcode 0x52 (82 decimal)
//
// This mapping was verified through debug logging at /tmp/gohook_debug.log
func (m *GlobalHotkeyManager) mapRawcodeToKey(rawcode uint16) string {
	switch rawcode {
	// ========== Modifier Keys ==========
	// These rawcodes are platform-specific values from libuiohook
	case 0x37, 0x3B: // Left/Right Command (macOS)
		return "cmd"
	case 0x38, 0x3C: // Left/Right Shift
		return "shift"
	case 0x3A, 0x3D: // Left/Right Alt (Option)
		return "alt"
	case 0x36, 0x3E: // Left/Right Control
		return "ctrl"

	// ========== Number Keys ==========
	// gohook uses numpad-position rawcodes for number keys
	// Verified values from actual event logs:
	case 0x52: // 82 decimal - numpad 0 position
		return "0"
	case 0x53: // 83 decimal - numpad 1 position
		return "1"
	case 0x54: // 84 decimal - numpad 2 position
		return "2"
	case 0x55: // 85 decimal - numpad 3 position
		return "3"
	case 0x56: // 86 decimal - numpad 4 position
		return "4"
	case 0x57: // 87 decimal - numpad 5 position
		return "5"
	case 0x31: // 49 decimal - actual key '6'
		return "6"
	case 0x32: // 50 decimal - actual key '7'
		return "7"
	case 0x33: // 51 decimal - actual key '8'
		return "8"
	case 0x34: // 52 decimal - actual key '9'
		return "9"

	// ========== Letter Keys ==========
	// Standard QWERTY layout (verified partial mapping)
	case 0x00: return "a"
	case 0x01: return "s"
	case 0x02: return "d"
	case 0x03: return "f"
	case 0x04: return "h"
	case 0x05: return "g"
	case 0x06: return "z"
	case 0x07: return "x"
	case 0x08: return "c"
	case 0x09: return "v"
	case 0x0B: return "q"
	case 0x0C: return "w"
	case 0x0D: return "e"
	case 0x0E: return "r"
	case 0x0F: return "y"
	case 0x10: return "t"
	// TODO: Add remaining letter keys as needed
	}
	return ""
}

// IsStarted returns whether the manager is started
func (m *GlobalHotkeyManager) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

// GetRegisteredHotkeys returns all registered hotkeys
func (m *GlobalHotkeyManager) GetRegisteredHotkeys() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string, len(m.registeredHotkeys))
	maps.Copy(result, m.registeredHotkeys)
	return result
}
