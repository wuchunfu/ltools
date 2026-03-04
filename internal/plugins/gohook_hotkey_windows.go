//go:build windows

package plugins

import (
	"fmt"
	"sync"
)

// GlobalHotkeyManager stub for Windows (gohook not supported)
type GlobalHotkeyManager struct {
	registeredHotkeys map[string]string
	hotkeyMap         map[string]string
	onHotkeyTriggered func(pluginID string)
	started           bool
	mu                sync.RWMutex
}

// NewGlobalHotkeyManager creates a new global hotkey manager (Windows stub)
func NewGlobalHotkeyManager() *GlobalHotkeyManager {
	return &GlobalHotkeyManager{
		registeredHotkeys: make(map[string]string),
		hotkeyMap:         make(map[string]string),
	}
}

// Register registers a global hotkey (Windows stub)
func (m *GlobalHotkeyManager) Register(normalizedKeyCombo string, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registeredHotkeys[normalizedKeyCombo] = pluginID
	return nil
}

// Unregister unregisters a global hotkey (Windows stub)
func (m *GlobalHotkeyManager) Unregister(normalizedKeyCombo string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.registeredHotkeys, normalizedKeyCombo)
	return nil
}

// SetCallback sets the callback function (Windows stub)
func (m *GlobalHotkeyManager) SetCallback(callback func(pluginID string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onHotkeyTriggered = callback
}

// Start begins listening (Windows stub - not supported)
func (m *GlobalHotkeyManager) Start() error {
	return fmt.Errorf("global hotkeys are not supported on Windows (gohook library unavailable)")
}

// Stop stops listening (Windows stub)
func (m *GlobalHotkeyManager) Stop() {
	// No-op
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
	for k, v := range m.registeredHotkeys {
		result[k] = v
	}
	return result
}
