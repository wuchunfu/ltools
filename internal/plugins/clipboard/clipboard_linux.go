//go:build linux

package clipboard

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Linux clipboard implementation using xclip/wl-paste utilities
// Falls back to X11 via xcb if available

var (
	clipboardMu sync.Mutex
	// Cache for detected clipboard command
	clipboardCmd string
	cmdChecked    bool
)

// detectClipboardCommand detects the available clipboard utility
func detectClipboardCommand() string {
	if cmdChecked {
		return clipboardCmd
	}

	clipboardMu.Lock()
	defer clipboardMu.Unlock()

	// Check if we're in Wayland or X11 session
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	display := os.Getenv("DISPLAY")

	isWayland := sessionType == "wayland" || waylandDisplay != ""
	isX11 := sessionType == "x11" || display != ""

	log.Printf("[Screenshot] Session type: %s, Wayland: %v, X11: %v", sessionType, isWayland, isX11)

	// Priority order for clipboard commands
	commands := []struct {
		cmd    string
		args   []string
		check  string // env var to check for preference
		reason string
	}{
		// Wayland native
		{"wl-copy", []string{"--type", "image/png"}, "WAYLAND_DISPLAY", "Wayland native"},
		{"wl-paste", []string{"--type", "image/png"}, "WAYLAND_DISPLAY", "Wayland native paste"},

		// X11 tools
		{"xclip", []string{"-selection", "clipboard", "-t", "image/png"}, "DISPLAY", "X11 xclip"},
		{"xsel", []string{"--clipboard", "--input"}, "DISPLAY", "X11 xsel"},
	}

	for _, c := range commands {
		// Skip if the required environment is not set
		if c.check != "" && os.Getenv(c.check) == "" {
			continue
		}

		// Check if command exists
		if _, err := exec.LookPath(c.cmd); err == nil {
			log.Printf("[Screenshot] Using clipboard command: %s (%s)", c.cmd, c.reason)
			clipboardCmd = c.cmd
			cmdChecked = true
			return c.cmd
		}
	}

	log.Printf("[Screenshot] Warning: No clipboard utility found")
	cmdChecked = true
	return ""
}

// SetImage sets an image to the clipboard on Linux
func (c *ImageClipboard) SetImage(imgData []byte) error {
	log.Printf("[Screenshot] SetImage called with %d bytes", len(imgData))

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

	cmd := detectClipboardCommand()
	if cmd == "" {
		return fmt.Errorf("no clipboard utility available (install xclip or wl-copy)")
	}

	var cmdExec *exec.Cmd
	switch cmd {
	case "wl-copy":
		// Wayland native clipboard
		cmdExec = exec.Command("wl-copy", "--type", "image/png")
	case "wl-paste":
		// wl-paste doesn't support copying, skip
		return fmt.Errorf("wl-paste does not support copying, use wl-copy instead")
	case "xclip":
		// X11 clipboard with xclip
		cmdExec = exec.Command("xclip", "-selection", "clipboard", "-t", "image/png")
	case "xsel":
		// X11 clipboard with xsel
		cmdExec = exec.Command("xsel", "--clipboard", "--input")
	default:
		return fmt.Errorf("unsupported clipboard command: %s", cmd)
	}

	// Set up stdin with image data
	cmdExec.Stdin = bytes.NewReader(imgData)

	// Run the command
	output, err := cmdExec.CombinedOutput()
	if err != nil {
		log.Printf("[Screenshot] Clipboard command failed: %v, output: %s", err, string(output))
		return fmt.Errorf("clipboard command failed: %w", err)
	}

	log.Printf("[Screenshot] ✓ Successfully set image to clipboard via %s", cmd)
	return nil
}

// GetImage gets an image from the clipboard on Linux
func (c *ImageClipboard) GetImage() ([]byte, error) {
	log.Printf("[Screenshot] Getting image from Linux clipboard...")

	cmd := detectClipboardCommand()
	if cmd == "" {
		return nil, fmt.Errorf("no clipboard utility available")
	}

	var cmdExec *exec.Cmd
	switch cmd {
	case "wl-copy", "wl-paste":
		// Wayland: use wl-paste to get image
		if _, err := exec.LookPath("wl-paste"); err == nil {
			cmdExec = exec.Command("wl-paste", "--type", "image/png")
		} else {
			return nil, fmt.Errorf("wl-paste not available")
		}
	case "xclip":
		// X11: use xclip to get image
		cmdExec = exec.Command("xclip", "-selection", "clipboard", "-t", "image/png", "-o")
	case "xsel":
		// xsel doesn't reliably support image data
		return nil, fmt.Errorf("xsel does not support reading image data from clipboard")
	default:
		return nil, fmt.Errorf("unsupported clipboard command for reading: %s", cmd)
	}

	// Run the command and capture output
	output, err := cmdExec.Output()
	if err != nil {
		log.Printf("[Screenshot] Failed to read clipboard: %v", err)
		return nil, fmt.Errorf("failed to read clipboard: %w", err)
	}

	if len(output) == 0 {
		return nil, &NoImageInClipboardError{}
	}

	log.Printf("[Screenshot] ✓ Retrieved image from clipboard (%d bytes)", len(output))
	return output, nil
}

// SetImageWithFallback tries to set image with multiple fallback methods
func (c *ImageClipboard) SetImageWithFallback(imgData []byte) error {
	// Method 1: Try to use a temporary file (more reliable for some clipboard managers)
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("ltools_clipboard_%d.png", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	// Write to temp file
	if err := os.WriteFile(tmpFile, imgData, 0644); err != nil {
		log.Printf("[Screenshot] Failed to write temp file: %v", err)
		// Fall back to stdin method
		return c.SetImage(imgData)
	}

	// Try with file input
	cmds := [][]string{
		{"xclip", "-selection", "clipboard", "-t", "image/png", "-i", tmpFile},
		{"wl-copy", "--type", "image/png", tmpFile},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if output, err := cmd.CombinedOutput(); err == nil {
			log.Printf("[Screenshot] ✓ Successfully set image via %s (file method)", args[0])
			return nil
		} else {
			log.Printf("[Screenshot] Command %s failed: %v, output: %s", args[0], err, string(output))
		}
	}

	// If all file methods fail, try stdin method
	return c.SetImage(imgData)
}

// GetLinuxDisplayInfo returns display information for debugging
func getLinuxDisplayInfo() map[string]string {
	info := make(map[string]string)

	info["runtime"] = runtime.GOOS + "/" + runtime.GOARCH
	info["session_type"] = os.Getenv("XDG_SESSION_TYPE")
	info["wayland_display"] = os.Getenv("WAYLAND_DISPLAY")
	info["display"] = os.Getenv("DISPLAY")
	info["xdg_current_desktop"] = os.Getenv("XDG_CURRENT_DESKTOP")

	// Check for clipboard utilities
	for _, cmd := range []string{"xclip", "xsel", "wl-copy", "wl-paste"} {
		if path, err := exec.LookPath(cmd); err == nil {
			info[cmd+"_path"] = path
		}
	}

	return info
}
