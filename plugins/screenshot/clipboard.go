package screenshot

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Clipboard handles clipboard operations for images
type Clipboard struct {
	app *application.App
}

// NewClipboard creates a new clipboard instance
func NewClipboard(app *application.App) *Clipboard {
	return &Clipboard{app: app}
}

// Platform-specific error types

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
	return "clipboard image operation not supported on platform: " + e.Platform
}

// SetImage sets an image to the clipboard
func (c *Clipboard) SetImage(imgData []byte) error {
	if len(imgData) == 0 {
		log.Printf("[Screenshot] Error: empty image data")
		return &InvalidImageError{}
	}

	// 验证 PNG 格式（前 8 字节应该是 PNG 签名）
	if len(imgData) < 8 {
		log.Printf("[Screenshot] Error: image data too small (%d bytes)", len(imgData))
		return &InvalidImageError{}
	}

	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.Equal(imgData[:8], pngSignature) {
		log.Printf("[Screenshot] Warning: data may not be PNG format (first 8 bytes: %x)", imgData[:8])
		return &InvalidImageError{}
	}

	log.Printf("[Screenshot] Setting image to clipboard (%d bytes)", len(imgData))

	// 使用平台特定的实现
	switch runtime.GOOS {
	case "darwin":
		return c.setImageMac(imgData)
	case "windows":
		return c.setImageWindows(imgData)
	case "linux":
		return c.setImageLinux(imgData)
	default:
		log.Printf("[Screenshot] Unsupported platform for clipboard image: %s", runtime.GOOS)
		return &UnsupportedPlatformError{Platform: runtime.GOOS}
	}
}

// SetImageFromRGBA sets an image to the clipboard from RGBA format
func (c *Clipboard) SetImageFromRGBA(img *image.RGBA) error {
	if img == nil {
		return &InvalidImageError{}
	}

	// Convert to PNG bytes
	var buf bytes.Buffer

	// 使用优化的 PNG 编码器配置
	encoder := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	if err := encoder.Encode(&buf, img); err != nil {
		return err
	}

	return c.SetImage(buf.Bytes())
}

// GetImage gets an image from the clipboard
func (c *Clipboard) GetImage() ([]byte, error) {
	log.Printf("[Screenshot] Getting image from clipboard")

	// 使用平台特定的实现
	switch runtime.GOOS {
	case "darwin":
		return c.getImageMac()
	case "windows":
		return c.getImageWindows()
	case "linux":
		return c.getImageLinux()
	default:
		return nil, &UnsupportedPlatformError{Platform: runtime.GOOS}
	}
}

// Platform-specific placeholder implementations

// setImageWindows sets image on Windows clipboard
func (c *Clipboard) setImageWindows(imgData []byte) error {
	log.Printf("[Screenshot] Setting image on Windows clipboard - not implemented")
	return &UnsupportedPlatformError{Platform: "windows"}
}

// setImageLinux sets image on Linux clipboard
func (c *Clipboard) setImageLinux(imgData []byte) error {
	log.Printf("[Screenshot] Setting image on Linux clipboard - not implemented")
	return &UnsupportedPlatformError{Platform: "linux"}
}

// getImageMac gets image from macOS clipboard
func (c *Clipboard) getImageMac() ([]byte, error) {
	log.Printf("[Screenshot] Getting image from macOS clipboard - not implemented")
	return nil, &UnsupportedPlatformError{Platform: "darwin"}
}

// getImageWindows gets image from Windows clipboard
func (c *Clipboard) getImageWindows() ([]byte, error) {
	log.Printf("[Screenshot] Getting image from Windows clipboard - not implemented")
	return nil, &UnsupportedPlatformError{Platform: "windows"}
}

// getImageLinux gets image from Linux clipboard
func (c *Clipboard) getImageLinux() ([]byte, error) {
	log.Printf("[Screenshot] Getting image from Linux clipboard - not implemented")
	return nil, &UnsupportedPlatformError{Platform: "linux"}
}

// NoImageInClipboardError is returned when there's no image in the clipboard
type NoImageInClipboardError struct{}

func (e *NoImageInClipboardError) Error() string {
	return "no image found in clipboard"
}
