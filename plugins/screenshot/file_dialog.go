package screenshot

import "github.com/wailsapp/wails/v3/pkg/application"

// SaveDialogOptions contains options for the save file dialog
type SaveDialogOptions struct {
	DefaultFilename string   // Suggested filename
	AllowedTypes   []string // File extensions to allow (e.g., ["png", "jpg"])
	Title          string   // Dialog title (used as message)
	DefaultDir     string   // Default directory
	ParentWindow   *application.WebviewWindow // Parent window for the dialog
}

// Platform-specific implementations are provided in:
// - file_dialog_darwin.go   (macOS) - implements SaveFileWithDialog with native NSSavePanel
// - file_dialog_windows.go  (Windows) - implements SaveFileWithDialog
// - file_dialog_linux.go    (Linux) - implements SaveFileWithDialog
