package screenshot

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// debugLog writes a message to a debug log file with timestamp
func debugLog(filepath, message string) {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
}

// ScreenshotService exposes Screenshot functionality to the frontend
type ScreenshotService struct {
	app             *application.App
	plugin          *ScreenshotPlugin
	windowService   *ScreenshotWindowService
}

// NewScreenshotService creates a new screenshot service
func NewScreenshotService(plugin *ScreenshotPlugin, app *application.App) *ScreenshotService {
	return &ScreenshotService{
		app:    app,
		plugin: plugin,
	}
}

// SetWindowService sets the window service reference
func (s *ScreenshotService) SetWindowService(ws *ScreenshotWindowService) {
	s.windowService = ws
}

// ServiceStartup is called when the application starts
func (s *ScreenshotService) ServiceStartup(app *application.App) error {
	log.Printf("[ScreenshotService] Service startup")
	return nil
}

// ServiceShutdown is called when the application shuts down
func (s *ScreenshotService) ServiceShutdown(app *application.App) error {
	log.Printf("[ScreenshotService] Service shutdown")
	return nil
}

// StartCapture starts a screen capture and opens the editor
func (s *ScreenshotService) StartCapture() (string, error) {
	log.Printf("[ScreenshotService] Starting capture")

	// Capture the primary display
	img, err := s.plugin.CaptureDisplay(0)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Capture failed: %v", err))
		return "", fmt.Errorf("failed to capture display: %w", err)
	}

	// Convert to base64
	base64Img, err := ImageToBase64(img)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Encoding failed: %v", err))
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Store current image
	pngData, err := ImageToPNG(img)
	if err != nil {
		return "", fmt.Errorf("failed to convert to PNG: %w", err)
	}
	s.plugin.SetCurrentImage(pngData)

	s.plugin.emitEvent("captured", "image captured")

	// Open editor window
	_, err = s.plugin.OpenEditorWindow(base64Img)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Editor open failed: %v", err))
		return "", fmt.Errorf("failed to open editor: %w", err)
	}

	return base64Img, nil
}

// GetCapturedImage returns the current captured image as base64
func (s *ScreenshotService) GetCapturedImage() string {
	imgData := s.plugin.GetCurrentImage()
	if len(imgData) == 0 {
		return ""
	}

	base64Str := base64.StdEncoding.EncodeToString(imgData)
	return "data:image/png;base64," + base64Str
}

// SaveImage saves the current screenshot to a file
// If filename is empty, a timestamped filename will be generated
func (s *ScreenshotService) SaveImage(filename string) (string, error) {
	imgData := s.plugin.GetCurrentImage()
	if len(imgData) == 0 {
		s.plugin.emitEvent("error", "no image to save")
		return "", fmt.Errorf("no image to save")
	}

	// Save to file
	savedPath, err := s.plugin.GetStorage().SaveToFile(imgData, filename)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Save failed: %v", err))
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	s.plugin.emitEvent("saved", savedPath)
	log.Printf("[ScreenshotService] Image saved to: %s", savedPath)

	return savedPath, nil
}

// CopyToClipboard copies the current screenshot to clipboard
func (s *ScreenshotService) CopyToClipboard() error {
	imgData := s.plugin.GetCurrentImage()
	if len(imgData) == 0 {
		s.plugin.emitEvent("error", "no image to copy")
		return fmt.Errorf("no image to copy")
	}

	// 使用 clipboard 实现将图片复制到剪贴板
	err := s.plugin.GetClipboard().SetImage(imgData)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("copy to clipboard failed: %v", err))
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	s.plugin.emitEvent("copied", "image copied to clipboard")
	log.Printf("[ScreenshotService] Image copied to clipboard")

	return nil
}

// CancelCapture cancels the current screenshot capture
func (s *ScreenshotService) CancelCapture() {
	// Use window service to close editor and show main window
	if s.windowService != nil {
		s.windowService.CloseEditor()
	} else {
		// Fallback to plugin method if window service not available
		s.plugin.CloseEditorWindow()
	}
	s.plugin.SetCurrentImage(nil)
	s.plugin.emitEvent("cancelled", "capture cancelled")
	log.Printf("[ScreenshotService] Capture cancelled")
}

// GetDisplays returns information about all available displays
func (s *ScreenshotService) GetDisplays() []DisplayInfo {
	return s.plugin.GetDisplays()
}

// CaptureDisplay captures a specific display by index
func (s *ScreenshotService) CaptureDisplay(displayIndex int) (string, error) {
	img, err := s.plugin.CaptureDisplay(displayIndex)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Capture failed: %v", err))
		return "", fmt.Errorf("failed to capture display %d: %w", displayIndex, err)
	}

	base64Img, err := ImageToBase64(img)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Encoding failed: %v", err))
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Store current image
	pngData, err := ImageToPNG(img)
	if err != nil {
		return "", fmt.Errorf("failed to convert to PNG: %w", err)
	}
	s.plugin.SetCurrentImage(pngData)

	s.plugin.emitEvent("captured", fmt.Sprintf("display %d captured", displayIndex))

	return base64Img, nil
}

// CaptureAllDisplays captures all displays
func (s *ScreenshotService) CaptureAllDisplays() (string, error) {
	img, err := s.plugin.CaptureAllDisplays()
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Capture failed: %v", err))
		return "", fmt.Errorf("failed to capture displays: %w", err)
	}

	base64Img, err := ImageToBase64(img)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Encoding failed: %v", err))
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Store current image
	pngData, err := ImageToPNG(img)
	if err != nil {
		return "", fmt.Errorf("failed to convert to PNG: %w", err)
	}
	s.plugin.SetCurrentImage(pngData)

	s.plugin.emitEvent("captured", "all displays captured")

	return base64Img, nil
}

// GetSelectedAnnotation processes annotated image data from the editor
// This receives the base64 data from the canvas after annotations
func (s *ScreenshotService) GetSelectedAnnotation(data string) (string, error) {
	if data == "" {
		return "", fmt.Errorf("no annotation data provided")
	}

	// Parse base64 data URL
	if !strings.HasPrefix(data, "data:image/png;base64,") {
		return "", fmt.Errorf("invalid image data format")
	}

	// Extract base64 string
	base64Str := strings.TrimPrefix(data, "data:image/png;base64,")

	// Decode base64
	imgData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", fmt.Errorf("failed to decode image data: %w", err)
	}

	// Update current image with annotated version
	s.plugin.SetCurrentImage(imgData)

	return data, nil
}

// SaveImageWithData saves image with specific data
func (s *ScreenshotService) SaveImageWithData(imgData, filename string) (string, error) {
	if imgData == "" {
		s.plugin.emitEvent("error", "no image data to save")
		return "", fmt.Errorf("no image data to save")
	}

	// Parse base64 data URL
	var pngData []byte
	var err error

	if strings.HasPrefix(imgData, "data:image/png;base64,") {
		base64Str := strings.TrimPrefix(imgData, "data:image/png;base64,")
		pngData, err = base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			return "", fmt.Errorf("failed to decode image data: %w", err)
		}
	} else {
		// Assume it's already raw data
		pngData = []byte(imgData)
	}

	// Save to file
	savedPath, err := s.plugin.GetStorage().SaveToFile(pngData, filename)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("Save failed: %v", err))
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	s.plugin.emitEvent("saved", savedPath)
	log.Printf("[ScreenshotService] Image saved to: %s", savedPath)

	return savedPath, nil
}

// CopyImageDataToClipboard copies specific image data to clipboard
func (s *ScreenshotService) CopyImageDataToClipboard(imgData string) error {
	log.Printf("[ScreenshotService] CopyImageDataToClipboard called, data length: %d", len(imgData))

	if imgData == "" {
		s.plugin.emitEvent("error", "no image data to copy")
		log.Printf("[ScreenshotService] Error: empty image data")
		return fmt.Errorf("no image data to copy")
	}

	// Parse base64 data URL
	var pngData []byte
	var err error

	if strings.HasPrefix(imgData, "data:image/png;base64,") {
		base64Str := strings.TrimPrefix(imgData, "data:image/png;base64,")
		pngData, err = base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			log.Printf("[ScreenshotService] Error decoding base64: %v", err)
			return fmt.Errorf("failed to decode image data: %w", err)
		}
		log.Printf("[ScreenshotService] Decoded PNG data: %d bytes", len(pngData))
	} else {
		// Assume it's already raw data
		pngData = []byte(imgData)
		log.Printf("[ScreenshotService] Using raw data: %d bytes", len(pngData))
	}

	// Update current image
	s.plugin.SetCurrentImage(pngData)

	// 使用 clipboard 实现将图片复制到剪贴板
	log.Printf("[ScreenshotService] Calling clipboard.SetImage...")
	err = s.plugin.GetClipboard().SetImage(pngData)
	if err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("copy to clipboard failed: %v", err))
		log.Printf("[ScreenshotService] Error copying to clipboard: %v", err)
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	s.plugin.emitEvent("copied", "image copied to clipboard")
	log.Printf("[ScreenshotService] ✓ Image copied to clipboard successfully")

	return nil
}

// SetSaveDirectory sets a custom save directory
func (s *ScreenshotService) SetSaveDirectory(dir string) error {
	return s.plugin.GetStorage().SetSaveDir(dir)
}

// GetSaveDirectory returns the current save directory
func (s *ScreenshotService) GetSaveDirectory() string {
	return s.plugin.GetStorage().GetSaveDir()
}

// GenerateFilename generates a timestamped filename
func (s *ScreenshotService) GenerateFilename() string {
	return s.plugin.GetStorage().GenerateFilename()
}

// GenerateFilenameWithPrefix generates a filename with a custom prefix
func (s *ScreenshotService) GenerateFilenameWithPrefix(prefix string) string {
	return s.plugin.GetStorage().GenerateFilenameWithPrefix(prefix)
}

// GetStorageInfo returns information about stored screenshots
func (s *ScreenshotService) GetStorageInfo() map[string]interface{} {
	return map[string]interface{}{
		"saveDir":    s.plugin.GetStorage().GetSaveDir(),
		"fileCount":  s.plugin.GetStorage().GetFileCount(),
		"totalSize":  s.plugin.GetStorage().GetTotalSize(),
	}
}

// CleanupOldScreenshots removes screenshots older than specified days
func (s *ScreenshotService) CleanupOldScreenshots(days int) error {
	if days <= 0 {
		return fmt.Errorf("days must be positive")
	}

	err := s.plugin.GetStorage().CleanupOldFiles(days)
	if err != nil {
		return fmt.Errorf("failed to cleanup old screenshots: %w", err)
	}

	log.Printf("[ScreenshotService] Cleaned up screenshots older than %d days", days)
	return nil
}

// Trigger manually triggers the screenshot capture from the frontend
// This is called when user clicks the "Start Screenshot" button in the UI
func (s *ScreenshotService) Trigger() error {
	// Debug: Write to file for visibility
	debugLog("/tmp/screenshot-trigger.log", "[ScreenshotService] Trigger() called")
	debugLog("/tmp/screenshot-trigger.log", fmt.Sprintf("[ScreenshotService] windowService is nil: %v", s.windowService == nil))

	log.Printf("[ScreenshotService] Trigger() called")
	log.Printf("[ScreenshotService] windowService is nil: %v", s.windowService == nil)

	if s.windowService == nil {
		log.Printf("[ScreenshotService] ERROR: window service not initialized!")
		debugLog("/tmp/screenshot-trigger.log", "[ScreenshotService] ERROR: window service not initialized!")
		return fmt.Errorf("window service not initialized")
	}

	log.Printf("[ScreenshotService] Triggering screenshot from frontend...")
	debugLog("/tmp/screenshot-trigger.log", "[ScreenshotService] Triggering screenshot from frontend...")
	result, err := s.windowService.StartCapture()
	log.Printf("[ScreenshotService] StartCapture returned: result=%v, err=%v", result, err)
	debugLog("/tmp/screenshot-trigger.log", fmt.Sprintf("[ScreenshotService] StartCapture returned: result=%v, err=%v", result, err))
	return err
}
