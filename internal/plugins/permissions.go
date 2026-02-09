package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// PermissionManager manages plugin permissions
type PermissionManager struct {
	mu           sync.RWMutex
	file         string
	permissions  map[string]map[Permission]bool // pluginID -> permission -> granted
	dirty        bool
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		permissions: make(map[string]map[Permission]bool),
	}
}

// Load loads permissions from a file
func (pm *PermissionManager) Load(dataDir string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	pm.file = filepath.Join(dataDir, "permissions.json")

	data, err := os.ReadFile(pm.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No permissions file yet
		}
		return err
	}

	return json.Unmarshal(data, &pm.permissions)
}

// Save saves permissions to a file
func (pm *PermissionManager) Save() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.file == "" {
		return nil // No file set, nothing to save
	}

	if !pm.dirty {
		return nil
	}

	data, err := json.MarshalIndent(pm.permissions, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename for atomicity
	tmpFile := pm.file + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, pm.file); err != nil {
		return err
	}

	pm.dirty = false
	return nil
}

// Grant grants a permission to a plugin
func (pm *PermissionManager) Grant(pluginID string, permission Permission) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.permissions[pluginID]; !exists {
		pm.permissions[pluginID] = make(map[Permission]bool)
	}

	pm.permissions[pluginID][permission] = true
	pm.dirty = true
}

// Revoke revokes a permission from a plugin
func (pm *PermissionManager) Revoke(pluginID string, permission Permission) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.permissions[pluginID]; !exists {
		return
	}

	delete(pm.permissions[pluginID], permission)
	pm.dirty = true
}

// IsGranted checks if a permission has been granted to a plugin
func (pm *PermissionManager) IsGranted(pluginID string, permission Permission) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if perms, exists := pm.permissions[pluginID]; exists {
		return perms[permission]
	}

	return false
}

// GetGranted returns all granted permissions for a plugin
func (pm *PermissionManager) GetGranted(pluginID string) []Permission {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if perms, exists := pm.permissions[pluginID]; exists {
		result := make([]Permission, 0, len(perms))
		for perm := range perms {
			result = append(result, perm)
		}
		return result
	}

	return nil
}

// GetRevoked returns all requested but revoked permissions for a plugin
func (pm *PermissionManager) GetRevoked(pluginID string, requested []Permission) []Permission {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []Permission

	if perms, exists := pm.permissions[pluginID]; exists {
		for _, perm := range requested {
			if !perms[perm] {
				result = append(result, perm)
			}
		}
	} else {
		// No permissions granted, all requested are revoked
		result = requested
	}

	return result
}

// Reset resets all permissions for a plugin
func (pm *PermissionManager) Reset(pluginID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.permissions, pluginID)
	pm.dirty = true
}

// ResetAll resets all permissions
func (pm *PermissionManager) ResetAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.permissions = make(map[string]map[Permission]bool)
	pm.dirty = true
}
