//go:build !windows

package clipboard

import (
	"ltools/internal/plugins/clipboard"
)
// NewClipboardPlugin creates a new clipboard plugin instance
func NewClipboardPlugin() *ClipboardPlugin {
	return &ClipboardPlugin{clipboard: clipboardpkg.ImageClipboard}
}
