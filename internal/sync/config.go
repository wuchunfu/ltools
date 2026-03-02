package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigManager manages the synchronization configuration.
type ConfigManager struct {
	mu       sync.RWMutex
	filePath string
	config   *SyncConfig
}

// NewConfigManager creates a new ConfigManager.
func NewConfigManager(dataDir string) (*ConfigManager, error) {
	filePath := filepath.Join(dataDir, "sync.json")

	cm := &ConfigManager{
		filePath: filePath,
		config:   DefaultSyncConfig(),
	}

	// Try to load existing config
	if err := cm.Load(); err != nil {
		// If file doesn't exist, create it with defaults
		if os.IsNotExist(err) {
			if saveErr := cm.Save(); saveErr != nil {
				return nil, saveErr
			}
		} else {
			return nil, err
		}
	}

	return cm, nil
}

// Load reads the configuration from disk.
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		return err
	}

	var config SyncConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	cm.config = &config
	return nil
}

// Save writes the configuration to disk.
func (cm *ConfigManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(cm.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.filePath, data, 0644)
}

// Get returns the current configuration.
func (cm *ConfigManager) Get() *SyncConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy to prevent external modification
	config := *cm.config
	if cm.config.IgnorePatterns != nil {
		config.IgnorePatterns = make([]string, len(cm.config.IgnorePatterns))
		copy(config.IgnorePatterns, cm.config.IgnorePatterns)
	}
	return &config
}

// Set updates the configuration.
func (cm *ConfigManager) Set(config *SyncConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config = config
	return nil
}

// Update modifies specific fields of the configuration.
func (cm *ConfigManager) Update(fn func(*SyncConfig)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	fn(cm.config)
	return nil
}

// UpdateAndSave modifies the configuration and saves it to disk.
func (cm *ConfigManager) UpdateAndSave(fn func(*SyncConfig)) error {
	cm.mu.Lock()
	fn(cm.config)
	cm.mu.Unlock()
	return cm.Save()
}

// SetEnabled enables or disables synchronization.
func (cm *ConfigManager) SetEnabled(enabled bool) error {
	return cm.UpdateAndSave(func(c *SyncConfig) {
		c.Enabled = enabled
	})
}

// SetRepoURL sets the repository URL.
func (cm *ConfigManager) SetRepoURL(url string) error {
	return cm.UpdateAndSave(func(c *SyncConfig) {
		c.RepoURL = url
	})
}

// SetAutoSync enables or disables automatic synchronization.
func (cm *ConfigManager) SetAutoSync(enabled bool) error {
	return cm.UpdateAndSave(func(c *SyncConfig) {
		c.AutoSync = enabled
	})
}

// SetSyncInterval sets the automatic sync interval in minutes.
func (cm *ConfigManager) SetSyncInterval(minutes int) error {
	return cm.UpdateAndSave(func(c *SyncConfig) {
		c.SyncInterval = minutes
	})
}

// UpdateLastSync updates the last sync time and hash.
func (cm *ConfigManager) UpdateLastSync(hash string) error {
	return cm.UpdateAndSave(func(c *SyncConfig) {
		c.LastSyncTime = Now()
		c.LastSyncHash = hash
	})
}

// Now returns the current time (extracted for testing).
var Now = func() time.Time {
	return time.Now()
}
