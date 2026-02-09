//go:build linux

package apps

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
)

// linuxIconExtractor Linux 图标提取器
type linuxIconExtractor struct{}

// NewIconExtractor 创建 Linux 图标提取器
func NewIconExtractor() IconExtractor {
	return &linuxIconExtractor{}
}

// ExtractIcon 从图标文件提取 Base64 编码的数据
func (e *linuxIconExtractor) ExtractIcon(iconPath string) (string, error) {
	if iconPath == "" {
		return "", fmt.Errorf("empty icon path")
	}

	// 检查文件是否存在
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return "", fmt.Errorf("icon file not found: %s", iconPath)
	}

	// 读取图标文件
	data, err := os.ReadFile(iconPath)
	if err != nil {
		return "", fmt.Errorf("failed to read icon file: %w", err)
	}

	// 根据文件类型处理
	var imageData []byte
	var mimeType string

	if strings.HasSuffix(strings.ToLower(iconPath), ".png") {
		imageData = data
		mimeType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(iconPath), ".svg") {
		// SVG 直接作为文本编码
		mimeType = "image/svg+xml"
		imageData = data
	} else if strings.HasSuffix(strings.ToLower(iconPath), ".xpm") {
		// XPM 需要转换（这里简化处理）
		mimeType = "image/x-xpm"
		imageData = data
	} else {
		// 尝试解码为 PNG
		_, err := png.Decode(bytes.NewReader(data))
		if err == nil {
			imageData = data
			mimeType = "image/png"
		} else {
			return "", fmt.Errorf("unsupported icon format: %s", iconPath)
		}
	}

	// 编码为 Base64
	base64Str := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)

	return dataURL, nil
}

// ExtractIconSize 提取指定尺寸的图标
// 对于 SVG，可以无损缩放
// 对于位图，需要缩放
func (e *linuxIconExtractor) ExtractIconSize(iconPath string, width, height int) (string, error) {
	if iconPath == "" {
		return "", fmt.Errorf("empty icon path")
	}

	// 检查文件是否存在
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return "", fmt.Errorf("icon file not found: %s", iconPath)
	}

	// 读取图标文件
	data, err := os.ReadFile(iconPath)
	if err != nil {
		return "", fmt.Errorf("failed to read icon file: %w", err)
	}

	// 如果是 SVG，直接返回（可以任意缩放）
	if strings.HasSuffix(strings.ToLower(iconPath), ".svg") {
		base64Str := base64.StdEncoding.EncodeToString(data)
		return fmt.Sprintf("data:image/svg+xml;base64,%s", base64Str), nil
	}

	// 对于 PNG 等位图，这里简化处理，直接返回原图
	// 实际应用中可以使用 imaging 等库进行缩放
	base64Str := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// ExtractIcon 全局函数提取图标
func ExtractIcon(iconPath string) (string, error) {
	extractor := NewIconExtractor()
	return extractor.ExtractIcon(iconPath)
}

// GetAppDefaultIcon 获取应用的默认图标（emoji）
// Linux 平台的实现
func GetAppDefaultIcon(appName string) string {
	return "🚀" // Linux 平台暂时使用默认图标
}

// DecodeImage 解码图片为 image.Image
func DecodeImage(data []byte, ext string) (image.Image, error) {
	switch strings.ToLower(ext) {
	case ".png":
		return png.Decode(bytes.NewReader(data))
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
}
