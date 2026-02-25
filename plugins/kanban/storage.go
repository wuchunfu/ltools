package kanban

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configFileName = "boards.json"
	configVersion  = 1
)

// LoadConfig loads the kanban configuration from the data directory
func (p *KanbanPlugin) LoadConfig(dataDir string) error {
	// Ensure kanban directory exists
	kanbanDir := filepath.Join(dataDir, "kanban")
	if err := os.MkdirAll(kanbanDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(kanbanDir, configFileName)

	// Read existing config or create default
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, p.config); err != nil {
			return err
		}
		// If no boards exist, create default board
		if len(p.config.Boards) == 0 {
			p.config.Boards = []Board{*CreateDefaultBoard()}
			return p.SaveConfig(dataDir)
		}
		return nil
	}

	// Create default config with default board
	p.config = &KanbanConfig{
		Version: configVersion,
		Boards:  []Board{*CreateDefaultBoard()},
	}
	return p.SaveConfig(dataDir)
}

// SaveConfig saves the kanban configuration to the data directory
func (p *KanbanPlugin) SaveConfig(dataDir string) error {
	kanbanDir := filepath.Join(dataDir, "kanban")
	if err := os.MkdirAll(kanbanDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(kanbanDir, configFileName)
	data, err := json.MarshalIndent(p.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
