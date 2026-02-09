package screenshot

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kbinani/screenshot"
)

// DisplayInfo contains information about a display
type DisplayInfo struct {
	Index   int    `json:"index"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Primary bool   `json:"primary"`
	Name    string `json:"name"`
}

// CaptureDisplay captures a specific display by index
func (p *ScreenshotPlugin) CaptureDisplay(displayIndex int) (*image.RGBA, error) {
	// Validate display index
	numDisplays := screenshot.NumActiveDisplays()
	if displayIndex < 0 || displayIndex >= numDisplays {
		return nil, &DisplayNotFoundError{Index: displayIndex, Total: numDisplays}
	}

	// 获取显示边界
	bounds := screenshot.GetDisplayBounds(displayIndex)

	log.Printf("[Screenshot] 捕获显示器 %d (操作系统: %s)", displayIndex, runtime.GOOS)
	log.Printf("[Screenshot] 边界: %v", bounds)
	log.Printf("[Screenshot] 尺寸: %dx%d", bounds.Dx(), bounds.Dy())

	var img *image.RGBA
	var err error

	// 根据操作系统选择合适的截图方法
	switch runtime.GOOS {
	case "darwin":
		// macOS: 使用 screencapture 命令获得物理像素分辨率
		img, err = p.captureWithScreencapture(displayIndex, bounds)
		if err != nil {
			log.Printf("[Screenshot] screencapture 失败，回退到 screenshot 库: %v", err)
			img, err = screenshot.CaptureRect(bounds)
		}
	case "windows":
		// Windows: 使用 screenshot 库，它应该正确处理 DPI
		img, err = p.captureWindows(displayIndex, bounds)
		if err != nil {
			log.Printf("[Screenshot] Windows capture 失败，回退到 screenshot 库: %v", err)
			img, err = screenshot.CaptureRect(bounds)
		}
	default:
		// Linux 和其他系统: 使用 screenshot 库
		img, err = screenshot.CaptureRect(bounds)
	}

	if err != nil {
		log.Printf("[Screenshot] Failed to capture display %d: %v", displayIndex, err)
		return nil, err
	}

	log.Printf("[Screenshot] 捕获成功: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())

	return img, nil
}

// captureWindows 使用 Windows 特定的方法捕获屏幕
// 在 Windows 上，screenshot 库应该正确处理 DPI 缩放
func (p *ScreenshotPlugin) captureWindows(displayIndex int, bounds image.Rectangle) (*image.RGBA, error) {
	log.Printf("[Screenshot] 使用 Windows 截图方法")

	// 在 Windows 上，直接使用 screenshot 库
	// 这个库在 Windows 上使用 GDI+，应该能够正确处理 DPI
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("Windows screenshot failed: %w", err)
	}

	return img, nil
}

// captureWithScreencapture 使用 macOS 原生 screencapture 命令捕获屏幕
// 这个方法可以捕获物理像素分辨率的截图（Retina 屏幕上的 3840x2160）
func (p *ScreenshotPlugin) captureWithScreencapture(displayIndex int, bounds image.Rectangle) (*image.RGBA, error) {
	// 创建临时文件
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("screenshot_%d_%d.png", displayIndex, time.Now().UnixNano()))
	defer os.Remove(tmpFile) // 清理临时文件

	// 捕获整个显示器（不使用 -R 参数，以获得物理像素分辨率）
	// screencapture -D <display> -x -t png <file>
	// -D 指定显示器索引（0=主显示器）
	// -x 表示不播放声音
	// -t png 指定格式
	cmd := exec.Command("screencapture", "-D", fmt.Sprintf("%d", displayIndex+1), "-x", "-t", "png", tmpFile)

	log.Printf("[Screenshot] 执行 screencapture 命令: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[Screenshot] screencapture 命令失败: %v, 输出: %s", err, string(output))
		return nil, fmt.Errorf("screencapture failed: %w", err)
	}

	// 读取捕获的 PNG 文件
	file, err := os.Open(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open screenshot file: %w", err)
	}
	defer file.Close()

	// 解码 PNG
	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	log.Printf("[Screenshot] screencapture 捕获的图片尺寸: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())

	// 转换为 image.RGBA
	var rgba *image.RGBA
	if rgbaImg, ok := img.(*image.RGBA); ok {
		rgba = rgbaImg
	} else {
		// 如果不是 RGBA，转换它
		rgba = image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	return rgba, nil
}

// CaptureAllDisplays captures all active displays and combines them
func (p *ScreenshotPlugin) CaptureAllDisplays() (*image.RGBA, error) {
	numDisplays := screenshot.NumActiveDisplays()
	if numDisplays == 0 {
		return nil, &NoDisplayError{}
	}

	// For simplicity, capture the primary display (index 0)
	// TODO: Implement multi-display capture with proper layout
	return p.CaptureDisplay(0)
}

// GetDisplays returns information about all active displays
func (p *ScreenshotPlugin) GetDisplays() []DisplayInfo {
	numDisplays := screenshot.NumActiveDisplays()
	displays := make([]DisplayInfo, numDisplays)

	for i := 0; i < numDisplays; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displays[i] = DisplayInfo{
			Index:   i,
			Width:   bounds.Dx(),
			Height:  bounds.Dy(),
			X:       bounds.Min.X,
			Y:       bounds.Min.Y,
			Primary: i == 0, // Assume first display is primary
			Name:    getDisplayName(i),
		}
	}

	return displays
}

// ImageToBase64 converts an image to base64 encoded PNG string
func ImageToBase64(img *image.RGBA) (string, error) {
	if img == nil {
		return "", &InvalidImageError{}
	}

	// Create a buffer to store PNG data
	var buf bytes.Buffer

	// 使用优化的 PNG 编码器配置
	// 使用 BestSpeed 模式以加快编码速度，同时保持无损质量
	encoder := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	if err := encoder.Encode(&buf, img); err != nil {
		return "", err
	}

	// Encode to base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + base64Str, nil
}

// ImageToPNG converts an image to PNG byte data
func ImageToPNG(img *image.RGBA) ([]byte, error) {
	if img == nil {
		return nil, &InvalidImageError{}
	}

	var buf bytes.Buffer

	// 使用优化的 PNG 编码器配置
	encoder := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	if err := encoder.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// getDisplayName returns a human-readable name for a display
func getDisplayName(index int) string {
	switch index {
	case 0:
		return "主显示器"
	case 1:
		return "副显示器 1"
	case 2:
		return "副显示器 2"
	default:
		return "显示器"
	}
}

// Custom error types

// DisplayNotFoundError is returned when the requested display is not found
type DisplayNotFoundError struct {
	Index int
	Total int
}

func (e *DisplayNotFoundError) Error() string {
	return fmt.Sprintf("display not found: index %d out of range (total: %d)", e.Index, e.Total)
}

// NoDisplayError is returned when no displays are available
type NoDisplayError struct{}

func (e *NoDisplayError) Error() string {
	return "no active displays found"
}

// InvalidImageError is returned when the image is invalid
type InvalidImageError struct{}

func (e *InvalidImageError) Error() string {
	return "invalid image data"
}
