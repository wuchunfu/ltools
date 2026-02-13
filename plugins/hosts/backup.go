package hosts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	backupDirName          = "backups/hosts"
	maxBackupsPerScenario = 20
)

// CreateBackup creates a backup of the current hosts file
// Returns backup ID and error
func (p *HostsPlugin) CreateBackup(dataDir string, scenarioID string) (*Backup, error) {
	// Read current hosts content
	content, err := ReadHostsFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read hosts file: %w", err)
	}

	// Create backup directory
	backupDir := filepath.Join(dataDir, backupDirName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup ID using content hash
	hash := sha256.Sum256([]byte(content))
	backupID := hex.EncodeToString(hash[:])[:16]

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("hosts_%s_%s.backup", scenarioID, timestamp)
	backupPath := filepath.Join(backupDir, filename)

	// Write backup file (atomic write)
	tmpPath := backupPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write backup: %w", err)
	}
	if err := os.Rename(tmpPath, backupPath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to finalize backup: %w", err)
	}

	// Get file size
	info, _ := os.Stat(backupPath)

	// Create backup record
	backup := &Backup{
		ID:         backupID,
		ScenarioID: scenarioID,
		CreatedAt:  time.Now(),
		Size:       info.Size(),
	}

	// Add to config
	p.config.Backups = append(p.config.Backups, *backup)

	// Emit event
	p.emitEvent("backup:created", backupID)

	return backup, nil
}

// ListBackups returns all backups, sorted by creation time (newest first)
func (p *HostsPlugin) ListBackups() []Backup {
	backups := make([]Backup, len(p.config.Backups))
	copy(backups, p.config.Backups)

	// Sort by creation time, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups
}

// RestoreBackup restores a backup from file
func (p *HostsPlugin) RestoreBackup(dataDir string, backupID string) error {
	// Find backup record
	var backup *Backup
	for i := range p.config.Backups {
		if p.config.Backups[i].ID == backupID {
			backup = &p.config.Backups[i]
			break
		}
	}
	if backup == nil {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Find backup file
	backupDir := filepath.Join(dataDir, backupDirName)
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backupPath string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "hosts_"+backup.ScenarioID+"_") {
			backupPath = filepath.Join(backupDir, file.Name())
			break
		}
	}

	if backupPath == "" {
		return fmt.Errorf("backup file not found for: %s", backupID)
	}

	// Read backup content
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Write to hosts file
	if err := WriteHostsFile(string(content)); err != nil {
		return fmt.Errorf("failed to restore hosts file: %w", err)
	}

	// Emit event
	p.emitEvent("backup:restored", backupID)

	return nil
}

// DeleteBackup deletes a backup record and its file
func (p *HostsPlugin) DeleteBackup(dataDir string, backupID string) error {
	// Find and remove from config
	for i, backup := range p.config.Backups {
		if backup.ID == backupID {
			// Remove backup file
			backupDir := filepath.Join(dataDir, backupDirName)
			files, _ := os.ReadDir(backupDir)
			for _, file := range files {
				if strings.HasPrefix(file.Name(), "hosts_"+backup.ScenarioID+"_") {
					backupPath := filepath.Join(backupDir, file.Name())
					os.Remove(backupPath)
					break
				}
			}

			// Remove from config
			p.config.Backups = append(p.config.Backups[:i], p.config.Backups[i+1:]...)

			// Emit event
			p.emitEvent("backup:deleted", backupID)

			return nil
		}
	}

	return fmt.Errorf("backup not found: %s", backupID)
}

// CleanOldBackups removes old backups, keeping only the most recent ones
func (p *HostsPlugin) CleanOldBackups(dataDir string, keepCount int) int {
	if keepCount <= 0 {
		keepCount = maxBackupsPerScenario
	}

	// Group backups by scenario
	byScenario := make(map[string][]Backup)
	for _, backup := range p.config.Backups {
		byScenario[backup.ScenarioID] = append(byScenario[backup.ScenarioID], backup)
	}

	// Sort each group by time and remove old ones
	removed := 0
	for _, backups := range byScenario {
		sort.Slice(backups, func(i, j int) bool {
			return backups[i].CreatedAt.After(backups[j].CreatedAt)
		})

		if len(backups) > keepCount {
			// Remove old backups
			for i := keepCount; i < len(backups); i++ {
				p.DeleteBackup(dataDir, backups[i].ID)
				removed++
			}
		}
	}

	return removed
}
