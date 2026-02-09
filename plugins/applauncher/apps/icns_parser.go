//go:build darwin

package apps

import (
	"bytes"
	"encoding/binary"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// ICNS 文件格式常量
const (
	icnsMagic = "icns"
)

// ICNS 图标类型常量
const (
	icnsIconPNG  = "icp4" // PNG 数据
	icnsIconJPEG = "icp5" // JPEG 数据
)

// IcnsFile 表示解析后的 ICNS 文件
type IcnsFile struct {
	Magic  string
	Length uint32
	Icons  []IcnsIcon
}

// IcnsIcon 表示 ICNS 文件中的一个图标
type IcnsIcon struct {
	OSType   string // 4 字节类型标识
	Length   uint32
	Data     []byte
	Width    int
	Height   int
	Format   string // "png" 或 "jpeg"
}

// ParseIcnsFile 解析 ICNS 文件
func ParseIcnsFile(data []byte) (*IcnsFile, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("file too small")
	}

	// 读取文件头
	magic := string(data[0:4])
	if magic != icnsMagic {
		return nil, fmt.Errorf("not an icns file (magic: %s)", magic)
	}

	length := binary.BigEndian.Uint32(data[4:8])

	icns := &IcnsFile{
		Magic: magic,
		Length: length,
		Icons:  make([]IcnsIcon, 0),
	}

	// 解析图标条目
	offset := 8
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		// 读取 OSType 和 Length
		ostype := string(data[offset : offset+4])
		iconLength := binary.BigEndian.Uint32(data[offset+4 : offset+8])

		// 验证数据长度
		dataEnd := offset + 8 + int(iconLength)
		if dataEnd > len(data) {
			break
		}

		iconData := data[offset+8 : dataEnd]

		// 创建图标条目
		icon := IcnsIcon{
			OSType: ostype,
			Length: iconLength,
			Data:   iconData,
		}

		// 尝试解析图标信息
		if err := parseIconInfo(&icon); err == nil {
			icns.Icons = append(icns.Icons, icon)
		}

		offset = dataEnd
	}

	return icns, nil
}

// parseIconInfo 解析图标信息（格式、尺寸）
func parseIconInfo(icon *IcnsIcon) error {
	data := icon.Data

	// 检查 PNG 文件头
	if len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		icon.Format = "png"
		// 读取 PNG 尺寸
		if width, height, err := getPNGSize(data); err == nil {
			icon.Width = width
			icon.Height = height
		}
		return nil
	}

	// 检查 JPEG 文件头
	if len(data) >= 4 && bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		icon.Format = "jpeg"
		// JPEG 2000 或其他格式
		return nil
	}

	// 默认格式（可能是压缩数据）
	return fmt.Errorf("unknown icon format")
}

// getPNGSize 从 PNG 数据中读取尺寸
func getPNGSize(data []byte) (int, int, error) {
	if len(data) < 24 {
		return 0, 0, fmt.Errorf("PNG data too small")
	}

	// PNG 文件头：89 50 4E 47 0D 0A 1A 0A
	// IHDR chunk：49 48 44 52

	reader := bytes.NewReader(data)

	// 跳过 PNG 签名（8 字节）
	reader.Seek(8, io.SeekStart)

	// 读取 IHDR chunk
	var chunkLength uint32
	var chunkType [4]byte

	binary.Read(reader, binary.BigEndian, &chunkLength)
	binary.Read(reader, binary.BigEndian, &chunkType)

	if string(chunkType[:]) != "IHDR" {
		return 0, 0, fmt.Errorf("first chunk is not IHDR")
	}

	// IHDR 数据包含：Width(4) + Height(4) + 其他
	var width, height uint32
	binary.Read(reader, binary.BigEndian, &width)
	binary.Read(reader, binary.BigEndian, &height)

	return int(width), int(height), nil
}

// ExtractBestIcon 从 ICNS 文件中提取最好的图标
// 返回最大的 PNG 图标的 Base64 数据
func ExtractBestIcon(icnsData []byte) (string, int, int, error) {
	icns, err := ParseIcnsFile(icnsData)
	if err != nil {
		return "", 0, 0, err
	}

	// 查找最大的 PNG 图标
	var bestIcon *IcnsIcon
	var maxSize int

	for i := range icns.Icons {
		icon := &icns.Icons[i]
		if icon.Format == "png" && icon.Width*icon.Height > maxSize {
			bestIcon = icon
			maxSize = icon.Width * icon.Height
		}
	}

	if bestIcon == nil {
		// 如果没有 PNG，尝试使用最大的图标
		for i := range icns.Icons {
			icon := &icns.Icons[i]
			if len(icon.Data) > maxSize {
				bestIcon = icon
				maxSize = len(icon.Data)
			}
		}
	}

	if bestIcon == nil {
		return "", 0, 0, fmt.Errorf("no icons found in icns file")
	}

	// 编码为 Base64
	base64Str := base64.StdEncoding.EncodeToString(bestIcon.Data)
	dataURL := fmt.Sprintf("data:image/%s;base64,%s", bestIcon.Format, base64Str)

	return dataURL, bestIcon.Width, bestIcon.Height, nil
}

// ExtractIconFromFile 从 .icns 文件路径提取图标
func ExtractIconFromFile(icnsPath string) (string, error) {
	data, err := os.ReadFile(icnsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read icns file: %w", err)
	}

	iconData, _, _, err := ExtractBestIcon(data)
	if err != nil {
		return "", fmt.Errorf("failed to extract icon: %w", err)
	}

	return iconData, nil
}

// DecodeICNS 解析 ICNS 文件
func DecodeICNS(data []byte) (*IcnsFile, error) {
	return ParseIcnsFile(data)
}

// GetIcnsInfo 获取 ICNS 文件信息（用于调试）
func GetIcnsInfo(icnsPath string) (string, error) {
	data, err := os.ReadFile(icnsPath)
	if err != nil {
		return "", err
	}

	icns, err := ParseIcnsFile(data)
	if err != nil {
		return "", err
	}

	info := fmt.Sprintf("ICNS File: %s\n", icnsPath)
	info += fmt.Sprintf("Icons: %d\n", len(icns.Icons))

	for i, icon := range icns.Icons {
		info += fmt.Sprintf("  [%d] Type: %s, Size: %dx%d, Format: %s, Length: %d bytes\n",
			i+1, icon.OSType, icon.Width, icon.Height, icon.Format, icon.Length)
	}

	return info, nil
}
