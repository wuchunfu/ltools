//go:build darwin

package screenshot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// SaveFileWithDialog shows a system native save file dialog using Wails v3 API
func (s *Storage) SaveFileWithDialog(imgData []byte, app *application.App, options SaveDialogOptions) (string, error) {
	if len(imgData) == 0 {
		return "", &InvalidImageError{}
	}

	// Set default options if not provided
	if options.DefaultFilename == "" {
		options.DefaultFilename = s.GenerateFilename()
		log.Printf("[Screenshot] Generated default filename: %s", options.DefaultFilename)
	}

	if options.Title == "" {
		options.Title = "保存截图"
	}

	// Set default directory
	defaultDir := options.DefaultDir
	if defaultDir == "" {
		defaultDir = s.saveDir
		if defaultDir == "" {
			homeDir, _ := os.UserHomeDir()
			defaultDir = filepath.Join(homeDir, "Desktop")
		}
	}

	log.Printf("[Screenshot] Showing Wails v3 save dialog: filename=%s, dir=%s", options.DefaultFilename, defaultDir)

	// Hide editor window before showing dialog
	if options.ParentWindow != nil {
		log.Printf("[Screenshot] Hiding editor window before dialog")
		options.ParentWindow.Hide()
		// Wait for window to hide completely
		time.Sleep(200 * time.Millisecond)
	}

	// Build dialog using Wails v3 native API
	dialogBuilder := app.Dialog.SaveFile().
		SetMessage(options.Title).
		SetFilename(options.DefaultFilename)

	// Add filters for allowed types
	if len(options.AllowedTypes) > 0 {
		for _, ext := range options.AllowedTypes {
			dialogBuilder.AddFilter("图片文件", "*."+ext)
		}
		dialogBuilder.AddFilter("所有文件", "*.*")
	}

	// Set default directory
	if defaultDir != "" {
		dialogBuilder = dialogBuilder.SetDirectory(defaultDir)
	}

	// Show the dialog
	log.Printf("[Screenshot] Showing save dialog")
	savePath, err := dialogBuilder.PromptForSingleSelection()
	log.Printf("[Screenshot] Dialog closed: path=%s, err=%v", savePath, err)

	// Restore editor window after dialog closes
	if options.ParentWindow != nil {
		log.Printf("[Screenshot] Restoring editor window after dialog")
		options.ParentWindow.Show()
		// Wait for window to show
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		log.Printf("[Screenshot] Save dialog error: %v", err)
		// Check if user cancelled
		if strings.Contains(err.Error(), "cancelled") || savePath == "" {
			return "", fmt.Errorf("user cancelled save")
		}
		return "", fmt.Errorf("failed to show save dialog: %w", err)
	}

	if savePath == "" {
		log.Printf("[Screenshot] User cancelled save dialog")
		return "", fmt.Errorf("user cancelled save")
	}

	// Ensure the file has the correct extension
	ext := strings.ToLower(filepath.Ext(savePath))
	hasAllowedExt := false
	for _, allowedExt := range options.AllowedTypes {
		if strings.TrimPrefix(ext, ".") == allowedExt {
			hasAllowedExt = true
			break
		}
	}

	if !hasAllowedExt && len(options.AllowedTypes) > 0 {
		savePath = savePath + "." + options.AllowedTypes[0]
	}

	// Ensure directory exists
	saveDir := filepath.Dir(savePath)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(savePath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("[Screenshot] File saved to: %s", savePath)
	return savePath, nil
}
