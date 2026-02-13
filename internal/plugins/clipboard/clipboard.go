// Package clipboard provides cross-platform image clipboard operations
// It uses platform-specific native APIs to copy images to the system clipboard
package clipboard

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ImageClipboard handles clipboard operations for images
type ImageClipboard struct {
	app *application.App
}

// NewImageClipboard creates a new clipboard instance
func NewImageClipboard(app *application.App) *ImageClipboard {
	return &ImageClipboard{app: app}
}

// SetImageFromRGBA sets an image to the clipboard from RGBA format
func (c *ImageClipboard) SetImageFromRGBA(img *image.RGBA) error {
	if img == nil {
		return &InvalidImageError{}
	}

	// Convert to PNG bytes
	var buf bytes.Buffer

	// Use optimized PNG encoder configuration
	encoder := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	if err := encoder.Encode(&buf, img); err != nil {
		return err
	}

	// Call platform-specific SetImage implementation
	return c.SetImage(buf.Bytes())
}

// Platform-specific implementations are provided in:
// - clipboard_darwin.go   (macOS) - implements SetImage, GetImage
// - clipboard_windows.go  (Windows) - implements SetImage, GetImage
// - clipboard_linux.go    (Linux) - implements SetImage, GetImage

// Platform-specific error types

// InvalidImageError is returned when image data is invalid
type InvalidImageError struct{}

func (e *InvalidImageError) Error() string {
	return "invalid image data"
}

// ClipboardSetError is returned when setting clipboard fails
type ClipboardSetError struct{}

func (e *ClipboardSetError) Error() string {
	return "failed to set image to clipboard"
}

// UnsupportedPlatformError is returned when clipboard operation is not supported on platform
type UnsupportedPlatformError struct {
	Platform string
}

func (e *UnsupportedPlatformError) Error() string {
	return fmt.Sprintf("clipboard image operation not supported on platform: %s", e.Platform)
}

// NoImageInClipboardError is returned when there's no image in the clipboard
type NoImageInClipboardError struct{}

func (e *NoImageInClipboardError) Error() string {
	return "no image found in clipboard"
}
