//go:build windows

package clipboard
import (
	"fmt"
)

// SetImage sets an image to the clipboard from RGBA format
func (c *ImageClipboard) SetImage(img *image.RGBA) error {
	if img == nil {
		return &InvalidImageError{}
    }
    // Windows clipboard is not supported for this plugin
    return fmt.Errorf("clipboard image operations not supported on Windows")
}

// GetImage gets an image from the clipboard
func (c *ImageClipboard) GetImage() ([]byte, error) {
	return nil, fmt.Errorf("clipboard image operations not supported on Windows")
}
