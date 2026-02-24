//go:build !darwin

package screenshot2

import (
	"fmt"
	"os"
	"path/filepath"

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
	}

	// Build dialog using Wails v3 native API
	dialogBuilder := app.Dialog.SaveFile().
		SetMessage(options.Title).
		SetFilename(options.DefaultFilename)

	if len(options.AllowedTypes) > 0 {
		for _, ext := range options.AllowedTypes {
			dialogBuilder.AddFilter("Image", "*."+ext)
		}
	}

	if defaultDir != "" {
		dialogBuilder = dialogBuilder.SetDirectory(defaultDir)
	}

	savePath, err := dialogBuilder.PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("failed to show save dialog: %w", err)
	}

	if savePath == "" {
		return "", fmt.Errorf("user cancelled save")
	}

	saveDir := filepath.Dir(savePath)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(savePath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return savePath, nil
}
