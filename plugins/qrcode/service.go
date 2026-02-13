package qrcode

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	clipboardpkg "ltools/internal/plugins/clipboard"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// QrcodeService exposes Qrcode functionality to frontend
// File saving and clipboard operations are handled by backend for Wails WebView compatibility
type QrcodeService struct {
	app       *application.App
	plugin    *QrcodePlugin
	saveDir   string                       // For default directory in dialogs and listing
	clipboard *clipboardpkg.ImageClipboard // Shared clipboard for image operations
}

// NewQrcodeService creates a new qrcode service
func NewQrcodeService(plugin *QrcodePlugin, app *application.App) *QrcodeService {
	// Get default save directory for listing and dialog default
	saveDir, _ := getDefaultSaveDir()

	return &QrcodeService{
		app:       app,
		plugin:    plugin,
		saveDir:   saveDir,
		clipboard: clipboardpkg.NewImageClipboard(app),
	}
}

// ServiceStartup is called when application starts
func (s *QrcodeService) ServiceStartup(app *application.App) error {
	return nil
}

// ServiceShutdown is called when application shuts down
func (s *QrcodeService) ServiceShutdown(app *application.App) error {
	return nil
}

// SaveToFile saves QR code image data to a file with system save dialog
// Input is base64 encoded PNG data (from canvas.toDataURL())
// Shows a system-native save dialog for user to choose location
func (s *QrcodeService) SaveToFile(base64Data string, defaultFilename string) (string, error) {
	if s.app == nil {
		return "", fmt.Errorf("app instance not available")
	}

	// Generate default filename if not provided
	if defaultFilename == "" {
		defaultFilename = generateFilename()
	}

	// Ensure filename has .png extension
	if !hasSuffix(defaultFilename, ".png") {
		defaultFilename += ".png"
	}

	// Decode base64 data (strip data URL prefix if present)
	imgData, err := decodeBase64Data(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Show system save dialog
	path, err := s.app.Dialog.SaveFile().
		SetMessage("保存二维码").
		SetFilename(defaultFilename).
		AddFilter("PNG Files", "*.png").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()

	if err != nil || path == "" {
		return "", err
	}

	// Write file
	if err := os.WriteFile(path, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Emit saved event
	s.plugin.Saved(filepath.Base(path))

	return path, nil
}

// CopyToClipboard copies QR code image data to clipboard
// Input is base64 encoded PNG data (from canvas.toDataURL())
// Uses platform-specific native clipboard API (required for WebView compatibility)
func (s *QrcodeService) CopyToClipboard(base64Data string) error {
	// Decode base64 data (strip data URL prefix if present)
	imgData, err := decodeBase64Data(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Use shared clipboard to copy image
	if err := s.clipboard.SetImage(imgData); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	// Emit copied event for tracking
	s.plugin.Copied("qrcode")

	return nil
}

// TrackGenerated tracks when a QR code is generated
func (s *QrcodeService) TrackGenerated(content string) {
	s.plugin.Generated(content)
}

// TrackCopied tracks when a QR code is copied to clipboard
func (s *QrcodeService) TrackCopied(content string) {
	s.plugin.Copied(content)
}

// TrackSaved tracks when a QR code is saved to file
func (s *QrcodeService) TrackSaved(filename string) {
	s.plugin.Saved(filename)
}

// GetSaveDir returns current save directory
func (s *QrcodeService) GetSaveDir() string {
	return s.saveDir
}

// SetSaveDir sets a custom save directory
func (s *QrcodeService) SetSaveDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("save directory cannot be empty")
	}

	s.saveDir = dir
	return nil
}

// ListSavedFiles returns list of saved QR code files
func (s *QrcodeService) ListSavedFiles() ([]FileInfo, error) {
	entries, err := os.ReadDir(s.saveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:     entry.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
		})
	}

	return files, nil
}

// FileInfo represents information about a saved file
type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"` // Unix timestamp
}

// emitNotification emits a notification event for the frontend
func (s *QrcodeService) emitNotification(title, message string) {
	if s.app != nil {
		s.app.Event.Emit("notification:show", fmt.Sprintf("%s|%s", title, message))
	}
}

// getDefaultSaveDir determines the default save directory
// Returns absolute path to ~/Pictures/QRCodes (or platform equivalent)
func getDefaultSaveDir() (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Build save directory path
	// macOS: ~/Pictures/QRCodes
	// Windows: %USERPROFILE%\Pictures\QRCodes (or use FOLDERID_OneDrive\Desktop etc.)
	// Linux: ~/Pictures/QRCodes
	// Use filepath.Join for cross-platform compatibility
	saveDir := filepath.Join(homeDir, "Pictures", "QRCodes")

	return saveDir, nil
}

// decodeBase64Data decodes base64 image data
// Strips the "data:image/png;base64," prefix if present
func decodeBase64Data(base64Data string) ([]byte, error) {
	// Check for data URL prefix
	if len(base64Data) > 0 {
		// Check for "data:image/png;base64," prefix (with comma)
		prefix := "data:image/png;base64,"
		if len(base64Data) > len(prefix) && base64Data[:len(prefix)] == prefix {
			// Strip the prefix
			base64Data = base64Data[len(prefix):]
		}
	}

	// Decode the base64 data
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, err
	}

	return decodedData, nil
}

// Helper function to check if string has a suffix
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// generateFilename generates a timestamped filename for QR codes
func generateFilename() string {
	now := time.Now()
	return fmt.Sprintf("qrcode_%s.png", now.Format("20060102_150405"))
}
