package screenshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Storage handles file storage operations for screenshots
type Storage struct {
	saveDir string
}

// NewStorage creates a new storage instance
func NewStorage(saveDir string) *Storage {
	return &Storage{saveDir: saveDir}
}

// SaveToFile saves image data to a file
func (s *Storage) SaveToFile(imgData []byte, filename string) (string, error) {
	if len(imgData) == 0 {
		return "", &InvalidImageError{}
	}

	// Ensure save directory exists
	if err := os.MkdirAll(s.saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create save directory: %w", err)
	}

	// If filename is empty, generate one
	if filename == "" {
		filename = s.GenerateFilename()
	}

	// Ensure filename has .png extension
	if filepath.Ext(filename) != ".png" {
		filename += ".png"
	}

	// Full path to save file
	fullPath := filepath.Join(s.saveDir, filename)

	// Write file
	if err := os.WriteFile(fullPath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// GetDefaultSavePath returns the default save path for screenshots
func GetDefaultSavePath() (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create screenshots directory in user's home
	// Using platform-specific conventions
	switch os.Getenv("GOOS") {
	case "darwin":
		// macOS: ~/Desktop/Screenshots or ~/Pictures/Screenshots
		return filepath.Join(homeDir, "Desktop", "Screenshots"), nil
	case "windows":
		// Windows: ~/Pictures/Screenshots
		return filepath.Join(homeDir, "Pictures", "Screenshots"), nil
	default:
		// Linux: ~/Pictures/Screenshots
		return filepath.Join(homeDir, "Pictures", "Screenshots"), nil
	}
}

// GenerateFilename generates a timestamped filename for screenshots
func (s *Storage) GenerateFilename() string {
	now := time.Now()
	// Format: Screenshot_2024-01-15_14-30-25.png
	return fmt.Sprintf("Screenshot_%s.png",
		now.Format("2006-01-02_15-04-05"))
}

// GenerateFilenameWithPrefix generates a filename with a custom prefix
func (s *Storage) GenerateFilenameWithPrefix(prefix string) string {
	now := time.Now()
	return fmt.Sprintf("%s_%s.png", prefix,
		now.Format("2006-01-02_15-04-05"))
}

// EnsureDirectory ensures the save directory exists
func (s *Storage) EnsureDirectory() error {
	return os.MkdirAll(s.saveDir, 0755)
}

// SetSaveDir sets a custom save directory
func (s *Storage) SetSaveDir(dir string) error {
	// Validate directory path
	if dir == "" {
		return fmt.Errorf("save directory cannot be empty")
	}

	// Try to create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	s.saveDir = dir
	return nil
}

// GetSaveDir returns the current save directory
func (s *Storage) GetSaveDir() string {
	return s.saveDir
}

// CleanupOldFiles removes screenshots older than specified days
func (s *Storage) CleanupOldFiles(days int) error {
	if days <= 0 {
		return fmt.Errorf("days must be positive")
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)

	entries, err := os.ReadDir(s.saveDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if file is a screenshot
		if filepath.Ext(entry.Name()) != ".png" {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if file is older than cutoff time
		if info.ModTime().Before(cutoffTime) {
			fullPath := filepath.Join(s.saveDir, entry.Name())
			if err := os.Remove(fullPath); err != nil {
				// Log error but continue with other files
				fmt.Printf("[Screenshot] Failed to remove old file %s: %v\n", fullPath, err)
			} else {
				fmt.Printf("[Screenshot] Removed old screenshot: %s\n", entry.Name())
			}
		}
	}

	return nil
}

// GetFileCount returns the number of screenshot files in the save directory
func (s *Storage) GetFileCount() int {
	entries, err := os.ReadDir(s.saveDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".png" {
			count++
		}
	}

	return count
}

// GetTotalSize returns the total size of all screenshots in bytes
func (s *Storage) GetTotalSize() int64 {
	entries, err := os.ReadDir(s.saveDir)
	if err != nil {
		return 0
	}

	var totalSize int64
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".png" {
			info, err := entry.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
	}

	return totalSize
}
