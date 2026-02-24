//go:build darwin

package screenshot2

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

	if options.DefaultFilename == "" {
		options.DefaultFilename = s.GenerateFilename()
	}

	if options.Title == "" {
		options.Title = "保存截图"
	}

	defaultDir := options.DefaultDir
	if defaultDir == "" {
		defaultDir = s.saveDir
		if defaultDir == "" {
			homeDir, _ := os.UserHomeDir()
			defaultDir = filepath.Join(homeDir, "Desktop")
		}
	}

	log.Printf("[Screenshot2] Showing save dialog: filename=%s, dir=%s", options.DefaultFilename, defaultDir)

	// Hide parent window before showing dialog
	var parentWindow *application.WebviewWindow
	if options.ParentWindow != nil {
		if win, ok := options.ParentWindow.(*application.WebviewWindow); ok {
			parentWindow = win
			parentWindow.Hide()
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Build dialog using Wails v3 native API
	dialogBuilder := app.Dialog.SaveFile().
		SetMessage(options.Title).
		SetFilename(options.DefaultFilename)

	if len(options.AllowedTypes) > 0 {
		for _, ext := range options.AllowedTypes {
			dialogBuilder.AddFilter("图片文件", "*."+ext)
		}
		dialogBuilder.AddFilter("所有文件", "*.*")
	}

	if defaultDir != "" {
		dialogBuilder = dialogBuilder.SetDirectory(defaultDir)
	}

	savePath, err := dialogBuilder.PromptForSingleSelection()

	// Restore parent window after dialog closes
	if parentWindow != nil {
		parentWindow.Show()
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		if strings.Contains(err.Error(), "cancelled") || savePath == "" {
			return "", fmt.Errorf("user cancelled save")
		}
		return "", fmt.Errorf("failed to show save dialog: %w", err)
	}

	if savePath == "" {
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

	saveDir := filepath.Dir(savePath)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(savePath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("[Screenshot2] File saved to: %s", savePath)
	return savePath, nil
}
