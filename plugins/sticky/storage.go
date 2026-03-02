package sticky

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configFileName = "sticky.json"
	configVersion  = 1
)

// LoadConfig loads the sticky configuration from the data directory
func (p *StickyPlugin) LoadConfig(dataDir string) error {
	configPath := filepath.Join(dataDir, configFileName)

	// Read existing config or create default
	if data, err := os.ReadFile(configPath); err == nil {
		return json.Unmarshal(data, p.config)
	}

	// Create default config
	p.config = &StickyConfig{
		Version: configVersion,
		Notes:   []StickyNote{},
	}
	return nil
}

// SaveConfig saves the sticky configuration to the data directory
func (p *StickyPlugin) SaveConfig(dataDir string) error {
	configPath := filepath.Join(dataDir, configFileName)
	data, err := json.MarshalIndent(p.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
