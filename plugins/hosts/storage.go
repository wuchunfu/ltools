package hosts

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configFileName = "hosts.json"
	configVersion  = 1
)

// LoadConfig loads the hosts configuration from the data directory
func (p *HostsPlugin) LoadConfig(dataDir string) error {
	configPath := filepath.Join(dataDir, configFileName)

	// Read existing config or create default
	if data, err := os.ReadFile(configPath); err == nil {
		return json.Unmarshal(data, p.config)
	}

	// Create default config
	p.config = &HostsConfig{
		Version:         configVersion,
		Scenarios:       []Scenario{},
		Backups:         []Backup{},
		CurrentScenario: "",
	}
	return nil
}

// SaveConfig saves the hosts configuration to the data directory
func (p *HostsPlugin) SaveConfig(dataDir string) error {
	configPath := filepath.Join(dataDir, configFileName)
	data, err := json.MarshalIndent(p.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
