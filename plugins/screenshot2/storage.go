package screenshot2

import (
	"encoding/base64"
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

	if err := os.MkdirAll(s.saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create save directory: %w", err)
	}

	if filename == "" {
		filename = s.GenerateFilename()
	}

	if filepath.Ext(filename) != ".png" {
		filename += ".png"
	}

	fullPath := filepath.Join(s.saveDir, filename)

	if err := os.WriteFile(fullPath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// GetDefaultSavePath returns the default save path for screenshots
func GetDefaultSavePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch os.Getenv("GOOS") {
	case "darwin":
		return filepath.Join(homeDir, "Desktop", "Screenshots"), nil
	case "windows":
		return filepath.Join(homeDir, "Pictures", "Screenshots"), nil
	default:
		return filepath.Join(homeDir, "Pictures", "Screenshots"), nil
	}
}

// GenerateFilename generates a timestamped filename for screenshots
func (s *Storage) GenerateFilename() string {
	now := time.Now()
	return fmt.Sprintf("Screenshot_%s.png", now.Format("2006-01-02_15-04-05"))
}

// GenerateFilenameWithPrefix generates a filename with a custom prefix
func (s *Storage) GenerateFilenameWithPrefix(prefix string) string {
	now := time.Now()
	return fmt.Sprintf("%s_%s.png", prefix, now.Format("2006-01-02_15-04-05"))
}

// GetSaveDir returns the current save directory
func (s *Storage) GetSaveDir() string {
	return s.saveDir
}

// SetSaveDir sets a custom save directory
func (s *Storage) SetSaveDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("save directory cannot be empty")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	s.saveDir = dir
	return nil
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
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			fullPath := filepath.Join(s.saveDir, entry.Name())
			os.Remove(fullPath)
		}
	}

	return nil
}

// decodeBase64 解码 base64 字符串
func decodeBase64(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}
