package plugins

import (
	"os"
	"path/filepath"
	"strings"
)

// PathInfo represents information about a detected file/directory path
type PathInfo struct {
	OriginalInput string // 用户原始输入
	ResolvedPath  string // 解析后的完整路径
	Exists        bool   // 路径是否存在
	IsDirectory   bool   // 是否为目录
	IsValid       bool   // 是否为有效路径
}

// DetectPath 检测输入字符串是否为文件/目录路径
// 支持的格式：
//   - 绝对路径: /Users/xxx/Documents, C:\Users\xxx\Documents
//   - 主目录简写: ~/Documents, ~\Documents
//   - 相对路径: ./file, ../dir
func DetectPath(input string) *PathInfo {
	if input == "" {
		return &PathInfo{IsValid: false}
	}

	input = strings.TrimSpace(input)
	if input == "" || input == "." || input == ".." {
		return &PathInfo{OriginalInput: input, IsValid: false}
	}

	// 路径特征检测：包含路径分隔符或看起来像路径
	hasPathSeparator := strings.ContainsAny(input, "/\\")
	startsWithTilde := strings.HasPrefix(input, "~")
	startsWithDot := strings.HasPrefix(input, ".")
	looksLikePath := hasPathSeparator || startsWithTilde || startsWithDot

	// 不像路径的输入直接返回
	if !looksLikePath {
		return &PathInfo{OriginalInput: input, IsValid: false}
	}

	resolvedPath := input

	// 处理主目录简写 (~)
	if startsWithTilde {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return &PathInfo{OriginalInput: input, IsValid: false}
		}

		if len(input) == 1 {
			resolvedPath = homeDir
		} else if strings.HasPrefix(input[1:], "/") || strings.HasPrefix(input[1:], "\\") {
			resolvedPath = filepath.Join(homeDir, input[2:])
		} else {
			resolvedPath = filepath.Join(homeDir, input[1:])
		}
	}

	// 处理相对路径
	if startsWithDot && (len(input) >= 2 && (input[1] == '/' || input[1] == '\\')) {
		// 获取当前工作目录
		cwd, err := os.Getwd()
		if err != nil {
			return &PathInfo{OriginalInput: input, IsValid: false}
		}
		resolvedPath = filepath.Join(cwd, resolvedPath)
	}

	// 清理路径
	resolvedPath = filepath.Clean(resolvedPath)

	// 检查路径是否存在
	info, err := os.Stat(resolvedPath)
	exists := err == nil
	isDir := exists && info.IsDir()

	return &PathInfo{
		OriginalInput: input,
		ResolvedPath:  resolvedPath,
		Exists:        exists,
		IsDirectory:   isDir,
		IsValid:       true,
	}
}

// GetFileExtension 获取文件扩展名（不含点）
func GetFileExtension(path string) string {
	base := filepath.Base(path)
	if idx := strings.LastIndex(base, "."); idx > 0 {
		return strings.ToLower(base[idx+1:])
	}
	return ""
}
