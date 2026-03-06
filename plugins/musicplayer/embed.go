package musicplayer

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// 只嵌入打包后的 dist 目录（优化后的 bundle 和插件源文件）
// dist/ 目录包含：
//   - server.bundle.js（打包后的服务代码，~1.7MB）
//   - sources/（音源插件文件，通过 build.js 复制）
//
//go:embed all:lx-music-service/dist
var lxMusicServiceFS embed.FS

// extractEmbeddedService 将嵌入的 lx-music-service 提取到用户目录
// 返回提取后的服务目录路径
func extractEmbeddedService() (string, error) {
	// 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	// 目标目录：~/.config/ltools/lx-music-service (macOS: ~/Library/Application Support/ltools/lx-music-service)
	targetDir := filepath.Join(configDir, "ltools", "lx-music-service")

	// 检查是否已经提取过（server.bundle.js 存在）
	bundlePath := filepath.Join(targetDir, "server.bundle.js")
	if _, err := os.Stat(bundlePath); err == nil {
		log.Printf("[ProcessManager] Using existing extracted service (bundle mode): %s", targetDir)
		return targetDir, nil
	}

	log.Printf("[ProcessManager] Extracting embedded lx-music-service to: %s", targetDir)

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target dir: %w", err)
	}

	// 提取 dist 目录下的所有内容到目标目录
	// dist/ 包含 server.bundle.js 和 sources/ 子目录
	if err := extractDir("lx-music-service/dist", targetDir); err != nil {
		return "", fmt.Errorf("failed to extract dist: %w", err)
	}

	log.Printf("[ProcessManager] Successfully extracted service to: %s", targetDir)
	return targetDir, nil
}

// extractDir 提取嵌入的目录及其所有内容到目标路径
func extractDir(embedPath, targetDir string) error {
	return fs.WalkDir(lxMusicServiceFS, embedPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk error at %s: %w", path, err)
		}

		// 计算相对路径（去掉 embedPath 前缀）
		relPath, err := filepath.Rel(embedPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// 如果是根目录本身，跳过
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			// 创建目录
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			log.Printf("[ProcessManager] Created directory: %s", relPath)
			return nil
		}

		// 读取嵌入的文件内容
		srcFile, err := lxMusicServiceFS.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open embedded file %s: %w", path, err)
		}
		defer srcFile.Close()

		// 创建目标文件
		dstFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}
		defer dstFile.Close()

		// 复制内容
		written, err := io.Copy(dstFile, srcFile)
		if err != nil {
			return fmt.Errorf("failed to copy file %s: %w", targetPath, err)
		}

		log.Printf("[ProcessManager] Extracted: %s (%d bytes)", relPath, written)
		return nil
	})
}
