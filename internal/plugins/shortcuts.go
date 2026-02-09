package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Shortcut represents a keyboard shortcut binding
type Shortcut struct {
	PluginID string `json:"pluginId"`
	KeyCombo string `json:"keyCombo"`   // e.g., "ctrl+shift+d", "cmd+1", "alt+space"
	Enabled  bool   `json:"enabled"`
}

// ShortcutManager manages keyboard shortcuts
type ShortcutManager struct {
	mu        sync.RWMutex
	file      string
	shortcuts map[string]*Shortcut // key: keyCombo
	dirty     bool
}

// NewShortcutManager creates a new shortcut manager
func NewShortcutManager(dataDir string) (*ShortcutManager, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	file := filepath.Join(dataDir, "shortcuts.json")
	sm := &ShortcutManager{
		file:      file,
		shortcuts: make(map[string]*Shortcut),
	}

	if err := sm.load(); err != nil {
		// If file doesn't exist, that's okay - start fresh
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	return sm, nil
}

// load loads shortcuts from disk
func (sm *ShortcutManager) load() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &sm.shortcuts)
}

// save saves shortcuts to disk
func (sm *ShortcutManager) save() error {
	if !sm.dirty {
		return nil
	}

	data, err := json.MarshalIndent(sm.shortcuts, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename for atomicity
	tmpFile := sm.file + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, sm.file); err != nil {
		return err
	}

	sm.dirty = false
	return nil
}

// Set sets a shortcut for a plugin
func (sm *ShortcutManager) Set(keyCombo, pluginID string, enabled bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Normalize key combo
	normalizedKeyCombo := normalizeKeyCombo(keyCombo)

	// Check if this key combo is already bound to another plugin
	if existing, ok := sm.shortcuts[normalizedKeyCombo]; ok && existing.PluginID != pluginID {
		return fmt.Errorf("shortcut %s is already bound to plugin %s", normalizedKeyCombo, existing.PluginID)
	}

	sm.shortcuts[normalizedKeyCombo] = &Shortcut{
		PluginID: pluginID,
		KeyCombo: normalizedKeyCombo,
		Enabled:  enabled,
	}
	sm.dirty = true

	return sm.save()
}

// Remove removes a shortcut
func (sm *ShortcutManager) Remove(keyCombo string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	normalizedKeyCombo := normalizeKeyCombo(keyCombo)

	if _, ok := sm.shortcuts[normalizedKeyCombo]; !ok {
		return ErrShortcutNotFound
	}

	delete(sm.shortcuts, normalizedKeyCombo)
	sm.dirty = true

	return sm.save()
}

// Get retrieves a shortcut by key combo
func (sm *ShortcutManager) Get(keyCombo string) (*Shortcut, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	normalizedKeyCombo := normalizeKeyCombo(keyCombo)
	shortcut, ok := sm.shortcuts[normalizedKeyCombo]
	return shortcut, ok
}

// GetByPluginID retrieves all shortcuts for a plugin
func (sm *ShortcutManager) GetByPluginID(pluginID string) []*Shortcut {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*Shortcut
	for _, shortcut := range sm.shortcuts {
		if shortcut.PluginID == pluginID {
			result = append(result, shortcut)
		}
	}

	return result
}

// GetAll returns all shortcuts
func (sm *ShortcutManager) GetAll() map[string]*Shortcut {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*Shortcut, len(sm.shortcuts))
	for k, v := range sm.shortcuts {
		result[k] = v
	}

	return result
}

// CheckConflict checks if a key combo conflicts with existing shortcuts
func (sm *ShortcutManager) CheckConflict(keyCombo, pluginID string) (bool, string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	normalizedKeyCombo := normalizeKeyCombo(keyCombo)

	if existing, ok := sm.shortcuts[normalizedKeyCombo]; ok {
		if existing.PluginID != pluginID {
			return true, existing.PluginID
		}
	}

	return false, ""
}

// normalizeKeyCombo normalizes a key combo string
// - Converts to lowercase
// - Removes spaces
// - Ensures consistent separator (+)
// - Maps meta/cmd appropriately
func normalizeKeyCombo(keyCombo string) string {
	// Convert to lowercase
	normalized := strings.ToLower(keyCombo)

	// Remove spaces
	normalized = strings.ReplaceAll(normalized, " ", "")

	// Ensure consistent separator
	normalized = strings.ReplaceAll(normalized, "-", "+")
	normalized = strings.ReplaceAll(normalized, "_", "+")

	return normalized
}

// FormatKeyCombo formats a key combo for display
// Handles cross-platform display (e.g., macOS shows Cmd, Windows/Linux shows Ctrl)
func FormatKeyCombo(keyCombo string, platform string) string {
	normalized := normalizeKeyCombo(keyCombo)

	parts := strings.Split(normalized, "+")
	var modifiers []string
	var key string

	for _, part := range parts {
		switch part {
		case "ctrl", "control":
			if platform == "darwin" {
				modifiers = append(modifiers, "⌘") // Cmd on macOS
			} else {
				modifiers = append(modifiers, "Ctrl")
			}
		case "cmd", "command", "meta":
			if platform == "darwin" {
				modifiers = append(modifiers, "⌘")
			} else {
				modifiers = append(modifiers, "Win")
			}
		case "shift":
			if platform == "darwin" {
				modifiers = append(modifiers, "⇧")
			} else {
				modifiers = append(modifiers, "Shift")
			}
		case "alt", "option":
			if platform == "darwin" {
				modifiers = append(modifiers, "⌥")
			} else {
				modifiers = append(modifiers, "Alt")
			}
		default:
			// This is the main key
			key = strings.ToUpper(part)
		}
	}

	result := strings.Join(modifiers, "+")
	if key != "" {
		if result != "" {
			result += "+"
		}
		result += key
	}

	return result
}

// ParseKeyCombo parses a key combo into Wails v3 key binding format
// Returns modifier flags and key
func ParseKeyCombo(keyCombo string) (modifiers string, key string) {
	normalized := normalizeKeyCombo(keyCombo)

	parts := strings.Split(normalized, "+")
	var modList []string
	var mainKey string

	for _, part := range parts {
		switch part {
		case "ctrl", "control":
			modList = append(modList, "ctrl")
		case "cmd", "command", "meta":
			modList = append(modList, "cmd")
		case "shift":
			modList = append(modList, "shift")
		case "alt", "option":
			modList = append(modList, "alt")
		default:
			// This is the main key
			mainKey = strings.ToLower(part)
		}
	}

	modifiers = strings.Join(modList, "+")
	key = mainKey

	return
}

// Errors
var (
	ErrShortcutNotFound = errors.New("shortcut not found")
	ErrShortcutConflict = errors.New("shortcut conflict")
)
